package gorm

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cdatasource "github.com/guidomantilla/yarumo/datasource"
)

// DatasourceGormType is the domain tag attached to every Error produced
// by the GORM datasource driver.
const DatasourceGormType = "datasource-gorm"

// Error is the domain error for the GORM datasource driver.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for GORM-specific failure modes. Each public factory
// joins the matching cross-driver sentinel from
// modules/datasource/errors.go so consumers can branch on either layer
// via errors.Is.
var (
	// ErrOpenFailed is joined when gorm.Open or the underlying *sql.DB
	// handshake fails.
	ErrOpenFailed = errors.New("gorm open failed")
	// ErrSQLDBUnavailable is joined when *gorm.DB.DB() cannot return
	// the underlying *sql.DB (driver was not configured correctly).
	ErrSQLDBUnavailable = errors.New("gorm: underlying sql.DB unavailable")
)

// ErrOpen creates a driver domain error joining the given causes with
// ErrOpenFailed and the cross-driver datasource.ErrConnectFailed +
// datasource.ErrDatasourceFailed sentinels.
func ErrOpen(causes ...error) error {
	return newError(append(causes, ErrOpenFailed, cdatasource.ErrConnectFailed, cdatasource.ErrDatasourceFailed)...)
}

// ErrClose creates a driver domain error joining the given causes with
// the cross-driver datasource.ErrCloseFailed +
// datasource.ErrDatasourceFailed sentinels.
func ErrClose(causes ...error) error {
	return newError(append(causes, cdatasource.ErrCloseFailed, cdatasource.ErrDatasourceFailed)...)
}

// ErrTransaction creates a driver domain error joining the given causes
// with the cross-driver datasource.ErrTransactionFailed +
// datasource.ErrDatasourceFailed sentinels. Panics from the
// user-supplied TxFn are surfaced through this factory with
// datasource.ErrTxPanic also joined.
func ErrTransaction(causes ...error) error {
	return newError(append(causes, cdatasource.ErrTransactionFailed, cdatasource.ErrDatasourceFailed)...)
}

// ErrSQLDB creates a driver domain error joining the given causes with
// ErrSQLDBUnavailable and the cross-driver
// datasource.ErrDatasourceFailed sentinel.
func ErrSQLDB(causes ...error) error {
	return newError(append(causes, ErrSQLDBUnavailable, cdatasource.ErrDatasourceFailed)...)
}

// newError is the shared internal constructor used by every public
// factory.
func newError(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: DatasourceGormType,
			Err:  errors.Join(causes...),
		},
	}
}
