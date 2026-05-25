package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/leobishop234/util/srverr"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	NamedQuery(query string, arg any) (*sqlx.Rows, error)
}

func (db *DB) WithTx(ctx context.Context, fn func(tx DBTX) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return Error("failed to begin transaction", err, srverr.ErrCodeDependencyFailure)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, Error("failed to rollback transaction", rbErr, srverr.ErrCodeDependencyFailure))
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return Error("failed to commit transaction", err, srverr.ErrCodeDependencyFailure)
	}

	return nil
}
