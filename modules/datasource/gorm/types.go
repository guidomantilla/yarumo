// Package gorm provides a GORM-backed implementation of the
// datasource.Connection and datasource.TransactionHandler contracts.
//
// The driver mirrors the structure of
// github.com/guidomantilla/go-feather-lib/pkg/datasource/gorm: a
// connection that wraps *gorm.DB and a transaction handler that
// delegates to gorm.DB.Transaction. It additionally satisfies the
// workspace's lifecycle.Component contract so it can be wired through
// common/lifecycle.Build.
//
// Backend selection is caller-driven: NewConnection takes an OpenFn
// returning a gorm.Dialector built from the DSN. Helpers
// PostgresOpener and SqliteOpener ship as convenience defaults; any
// other GORM dialector (mysql, sqlserver, clickhouse, ...) plugs in by
// passing its package-level Open function.
package gorm

import (
	cdatasource "github.com/guidomantilla/yarumo/datasource"

	"github.com/guidomantilla/yarumo/common/lifecycle"

	"gorm.io/gorm"
)

// Type compliance: connection implements both Connection and
// lifecycle.Component; transactionHandler implements the cross-driver
// TransactionHandler.
var (
	_ cdatasource.Connection         = (*connection)(nil)
	_ cdatasource.TransactionHandler = (*transactionHandler)(nil)
	_ lifecycle.Component            = (*connection)(nil)
	_ error                          = (*Error)(nil)
)

// OpenFn is the dialector factory consumed by NewConnection. Callers
// pass a function that wraps a gorm.Dialector constructor (for example
// postgres.Open or sqlite.Open) so the driver can defer dialector
// construction until the Context's URL is known.
type OpenFn func(dsn string) gorm.Dialector

// txCtxKey is the context.Context key under which HandleTransaction
// stores the active transactional *gorm.DB. Unexported so callers
// cannot collide.
type txCtxKey struct{}
