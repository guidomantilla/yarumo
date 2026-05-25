package gorm

import (
	"context"
	"sync"

	"gorm.io/gorm"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/lifecycle"
	cdatasource "github.com/guidomantilla/yarumo/datasource"
)

// connection is the GORM-backed datasource.Connection. It owns the
// *gorm.DB handle and the underlying *sql.DB pool tuning. The
// connection additionally satisfies the workspace lifecycle.Component
// contract:
//   - Start opens (or re-uses) the live *gorm.DB and performs an
//     initial Ping. Worker-style: Start returns once the connection is
//     ready.
//   - Stop closes the underlying *sql.DB pool exactly once and closes
//     Done. Idempotent.
type connection struct {
	context_ cdatasource.Context
	openFn   OpenFn
	opts     *Options

	mutex sync.Mutex
	db    *gorm.DB

	done chan struct{}
	once sync.Once
}

// NewConnection constructs a Connection without opening the underlying
// *gorm.DB. The handle is opened lazily by Connect or by the lifecycle
// Start path; that mirrors the reference go-feather-lib datasource and
// allows test doubles to construct a Connection without touching the
// network.
func NewConnection(ctx cdatasource.Context, openFn OpenFn, opts ...Option) cdatasource.Connection {
	cassert.NotNil(ctx, "context is nil")
	cassert.NotNil(openFn, "openFn is nil")

	return &connection{
		context_: ctx,
		openFn:   openFn,
		opts:     NewOptions(opts...),
		done:     make(chan struct{}),
	}
}

// BuildDB wires a connection through the workspace's lifecycle.Build
// pattern: NewConnection + lifecycle.Build. It dispatches the
// worker-style Start goroutine that opens and pings the backend,
// returns the Connection so repositories can mount on the typed *gorm.DB
// via DB(ctx), and surfaces a CloseFn the caller invokes on shutdown.
func BuildDB(ctx context.Context, c cdatasource.Context, openFn OpenFn, errChan lifecycle.ErrChan, opts ...Option) (cdatasource.Connection, lifecycle.CloseFn, error) {
	cassert.NotNil(c, "context is nil")
	cassert.NotNil(openFn, "openFn is nil")

	conn := NewConnection(c, openFn, opts...)

	component, ok := conn.(lifecycle.Component)
	if !ok {
		return nil, nil, ErrOpen()
	}

	closeFn, err := lifecycle.Build(ctx, component, errChan)
	if err != nil {
		return nil, nil, err
	}

	return conn, closeFn, nil
}

// Name returns the connection identity, derived from the Context's
// service + server pair. Lifecycle helpers log this string at the
// startup / shutdown boundary.
func (c *connection) Name() string {
	cassert.NotNil(c, "connection is nil")

	return c.context_.Service() + "@" + c.context_.Server()
}

// Connect returns the live *gorm.DB, opening it on first use. The
// returned value is typed as any so it satisfies the cross-driver
// Connection interface; callers in the gorm package should use DB(ctx)
// to obtain the typed handle without an explicit assertion.
func (c *connection) Connect(ctx context.Context) (any, error) {
	cassert.NotNil(c, "connection is nil")

	return c.openOrReuse(ctx)
}

// DB returns the live *gorm.DB, opening it on first use. Repositories
// in the gorm package use this method instead of Connect so they avoid
// type-asserting the any returned by the cross-driver interface.
func (c *connection) DB(ctx context.Context) (*gorm.DB, error) {
	cassert.NotNil(c, "connection is nil")

	return c.openOrReuse(ctx)
}

// Close releases the underlying *sql.DB pool. Idempotent: only the
// first call closes the pool; subsequent calls return nil. Close also
// closes the Done channel so callers waiting on Done() are released.
func (c *connection) Close(_ context.Context) error {
	cassert.NotNil(c, "connection is nil")

	var closeErr error

	c.once.Do(func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		if c.db == nil {
			close(c.done)

			return
		}

		sqlDB, err := c.db.DB()
		if err != nil {
			closeErr = ErrSQLDB(err)

			close(c.done)

			return
		}

		err = sqlDB.Close()
		if err != nil {
			closeErr = ErrClose(err)
		}

		close(c.done)
	})

	return closeErr
}

// Context returns the Context that produced this connection.
func (c *connection) Context() cdatasource.Context {
	cassert.NotNil(c, "connection is nil")

	return c.context_
}

// Start opens the underlying *gorm.DB and verifies it with a Ping. It
// satisfies the lifecycle.Component worker-style contract: returns
// immediately on success, or returns a lifecycle.ErrStart wrapping the
// driver-domain ErrOpen / ErrSQLDB when the backend is unreachable.
func (c *connection) Start(ctx context.Context) error {
	cassert.NotNil(c, "connection is nil")

	gdb, err := c.openOrReuse(ctx)
	if err != nil {
		return lifecycle.ErrStart(err)
	}

	sqlDB, sqlErr := gdb.DB()
	if sqlErr != nil {
		return lifecycle.ErrStart(ErrSQLDB(sqlErr))
	}

	pingErr := sqlDB.PingContext(ctx)
	if pingErr != nil {
		return lifecycle.ErrStart(ErrOpen(pingErr))
	}

	return nil
}

// Stop closes the underlying connection pool. Idempotent (see Close).
func (c *connection) Stop(ctx context.Context) error {
	cassert.NotNil(c, "connection is nil")

	return c.Close(ctx)
}

// Done returns the channel that is closed when Close / Stop has been
// invoked.
func (c *connection) Done() <-chan struct{} {
	cassert.NotNil(c, "connection is nil")

	return c.done
}

// openOrReuse returns the cached *gorm.DB or opens a fresh one using
// the configured dialector and pool tuning. Safe for concurrent use.
func (c *connection) openOrReuse(_ context.Context) (*gorm.DB, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.db != nil {
		return c.db, nil
	}

	dialector := c.openFn(c.context_.Url())

	gdb, err := gorm.Open(dialector, c.opts.gormConfig)
	if err != nil {
		return nil, ErrOpen(err)
	}

	sqlDB, sqlErr := gdb.DB()
	if sqlErr != nil {
		return nil, ErrSQLDB(sqlErr)
	}

	sqlDB.SetMaxIdleConns(c.opts.maxIdleConns)
	sqlDB.SetMaxOpenConns(c.opts.maxOpenConns)
	sqlDB.SetConnMaxLifetime(c.opts.connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(c.opts.connMaxIdleTime)

	c.db = gdb

	return gdb, nil
}
