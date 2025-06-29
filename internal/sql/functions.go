package sql

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type TransactionCtxKey struct{}

type TransactionHandlerFn func(ctx context.Context, tx *sqlx.Tx) error

func HandleTransaction(ctx context.Context, driver string, url string, fn TransactionHandlerFn) error {

	connection, err := sqlx.Open(driver, url)
	if err != nil {
		return err
	}

	tx, err := connection.Beginx()
	if err != nil {
		return err
	}

	err = fn(ctx, tx)
	if err != nil {
		return tx.Rollback()
	}

	return tx.Commit()
}

func Query(ctx context.Context, tx *sqlx.Tx, sql string, columns int, params ...any) ([]map[string]any, error) {
	rows, err := tx.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	ret := make([]map[string]any, 0)
	for rows.Next() {
		colVals := make([]any, columns)
		for i := range colVals {
			colVals[i] = new(any)
		}
		err = rows.Scan(colVals...)
		if err != nil {
			return nil, err
		}
		colNames, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		these := make(map[string]any)
		for idx, name := range colNames {
			these[name] = *colVals[idx].(*any)
		}
		ret = append(ret, these)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}
