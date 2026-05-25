package gorm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	lctests "github.com/guidomantilla/yarumo/common/lifecycle/tests"
	cdatasource "github.com/guidomantilla/yarumo/datasource"
)

// newSqliteContext returns a Context targeting an in-memory SQLite
// database. Tests rely on the sqlite OpenFn instead of postgres so
// they don't touch the network.
func newSqliteContext(t *testing.T) cdatasource.Context {
	t.Helper()

	return cdatasource.NewContext(":memory:", "u", "p", "host", "svc")
}

// newSqliteOpen returns an OpenFn that opens a pure-Go sqlite
// dialector ignoring the DSN-rendered URL (which sqlite would treat as
// a file path).
func newSqliteOpen() OpenFn {
	return func(_ string) gorm.Dialector {
		return sqlite.Open(":memory:")
	}
}

// newTestConn returns a connection backed by in-memory SQLite. Each
// call yields a fresh, isolated database; the cleanup callback closes
// it.
func newTestConn(t *testing.T) cdatasource.Connection {
	t.Helper()

	conn := NewConnection(newSqliteContext(t), newSqliteOpen(), WithMaxOpenConns(1))

	t.Cleanup(func() { _ = conn.Close(context.Background()) })

	return conn
}

type item struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func mustMigrate(t *testing.T, conn cdatasource.Connection) *gorm.DB {
	t.Helper()

	raw, err := conn.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	gdb, ok := raw.(*gorm.DB)
	if !ok {
		t.Fatal("expected *gorm.DB")
	}

	err = gdb.AutoMigrate(&item{})
	if err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	return gdb
}

func TestNewConnection(t *testing.T) {
	t.Parallel()

	t.Run("does not open the underlying handle", func(t *testing.T) {
		t.Parallel()

		conn := NewConnection(newSqliteContext(t), newSqliteOpen())

		c, ok := conn.(*connection)
		if !ok {
			t.Fatal("expected *connection")
		}

		if c.db != nil {
			t.Fatal("expected lazy open")
		}
	})

	t.Run("Context returns the supplied datasource.Context", func(t *testing.T) {
		t.Parallel()

		dsCtx := newSqliteContext(t)
		conn := NewConnection(dsCtx, newSqliteOpen())

		if conn.Context() != dsCtx {
			t.Fatal("Context did not return the supplied datasource.Context")
		}
	})
}

func TestConnection_Connect(t *testing.T) {
	t.Parallel()

	t.Run("opens the *gorm.DB on first call and reuses on second", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)

		raw1, err := conn.Connect(context.Background())
		if err != nil {
			t.Fatalf("Connect 1: %v", err)
		}

		raw2, err := conn.Connect(context.Background())
		if err != nil {
			t.Fatalf("Connect 2: %v", err)
		}

		if raw1 != raw2 {
			t.Fatal("expected the same *gorm.DB on second call")
		}
	})

	t.Run("ErrOpen when dialector fails", func(t *testing.T) {
		t.Parallel()

		badOpen := OpenFn(func(_ string) gorm.Dialector {
			return sqlite.Open("file:/nonexistent/dir/db.sqlite?mode=rw")
		})

		conn := NewConnection(newSqliteContext(t), badOpen)
		t.Cleanup(func() { _ = conn.Close(context.Background()) })

		_, err := conn.Connect(context.Background())
		if err == nil {
			t.Fatal("expected open error")
		}

		if !errors.Is(err, ErrOpenFailed) {
			t.Fatalf("expected ErrOpenFailed, got %v", err)
		}

		if !errors.Is(err, cdatasource.ErrConnectFailed) {
			t.Fatalf("expected datasource.ErrConnectFailed, got %v", err)
		}
	})
}

func TestConnection_DB(t *testing.T) {
	t.Parallel()

	t.Run("returns the typed *gorm.DB handle", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)

		c, ok := conn.(*connection)
		if !ok {
			t.Fatal("expected *connection")
		}

		gdb, err := c.DB(context.Background())
		if err != nil {
			t.Fatalf("DB: %v", err)
		}

		if gdb == nil {
			t.Fatal("expected non-nil *gorm.DB")
		}
	})
}

func TestConnection_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns service@server", func(t *testing.T) {
		t.Parallel()

		conn := NewConnection(cdatasource.NewContext("u", "user", "pwd", "myhost:5432", "mydb"), newSqliteOpen())
		t.Cleanup(func() { _ = conn.Close(context.Background()) })

		c, ok := conn.(*connection)
		if !ok {
			t.Fatal("expected *connection")
		}

		got := c.Name()
		want := "mydb@myhost:5432"

		if got != want {
			t.Fatalf("Name = %q, want %q", got, want)
		}
	})
}

func TestConnection_Close(t *testing.T) {
	t.Parallel()

	t.Run("closes once and Done is closed", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)

		// Make sure it's open.
		_, err := conn.Connect(context.Background())
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}

		err = conn.Close(context.Background())
		if err != nil {
			t.Fatalf("first Close: %v", err)
		}

		c, ok := conn.(*connection)
		if !ok {
			t.Fatal("expected *connection")
		}

		select {
		case <-c.Done():
		case <-time.After(time.Second):
			t.Fatal("Done not closed")
		}

		err = conn.Close(context.Background())
		if err != nil {
			t.Fatalf("second Close should be a no-op, got %v", err)
		}
	})

	t.Run("Close before Connect closes Done without error", func(t *testing.T) {
		t.Parallel()

		conn := NewConnection(newSqliteContext(t), newSqliteOpen())

		err := conn.Close(context.Background())
		if err != nil {
			t.Fatalf("Close: %v", err)
		}

		c, ok := conn.(*connection)
		if !ok {
			t.Fatal("expected *connection")
		}

		select {
		case <-c.Done():
		case <-time.After(time.Second):
			t.Fatal("Done not closed")
		}
	})
}

func TestConnection_StartAndStop(t *testing.T) {
	t.Parallel()

	t.Run("Start opens and pings successfully", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)

		c, ok := conn.(*connection)
		if !ok {
			t.Fatal("expected *connection")
		}

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
	})

	t.Run("Start returns lifecycle.ErrStart when open fails", func(t *testing.T) {
		t.Parallel()

		badOpen := OpenFn(func(_ string) gorm.Dialector {
			return sqlite.Open("file:/nonexistent/dir/db.sqlite?mode=rw")
		})

		conn := NewConnection(newSqliteContext(t), badOpen)
		t.Cleanup(func() { _ = conn.Close(context.Background()) })

		c, ok := conn.(*connection)
		if !ok {
			t.Fatal("expected *connection")
		}

		err := c.Start(context.Background())
		if !errors.Is(err, lifecycle.ErrStartFailed) {
			t.Fatalf("expected lifecycle.ErrStartFailed, got %v", err)
		}

		if !errors.Is(err, ErrOpenFailed) {
			t.Fatalf("expected wrap of ErrOpenFailed, got %v", err)
		}
	})

	t.Run("StopIsIdempotent satisfies lifecycle contract", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)

		c, ok := conn.(*connection)
		if !ok {
			t.Fatal("expected *connection")
		}

		lctests.AssertIdempotentStop(t, c)
	})
}

func TestBuildDB(t *testing.T) {
	t.Parallel()

	t.Run("wires lifecycle.Build and returns a CloseFn", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		errChan := make(chan error, 1)

		conn, closeFn, err := BuildDB(ctx, newSqliteContext(t), newSqliteOpen(), errChan)
		if err != nil {
			t.Fatalf("BuildDB: %v", err)
		}

		if conn == nil || closeFn == nil {
			t.Fatal("expected non-nil conn and closeFn")
		}

		closeFn(ctx, time.Second)

		c, ok := conn.(*connection)
		if !ok {
			t.Fatal("expected *connection")
		}

		select {
		case <-c.Done():
		case <-time.After(time.Second):
			t.Fatal("Done not closed after CloseFn")
		}
	})
}
