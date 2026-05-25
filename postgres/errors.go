package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/leobishop234/util/srverr"
)

type sqlStater interface {
	SQLState() string
}

func Error(message string, err error, fallback srverr.ErrCode) error {
	return srverr.New(classifyPostgresErrCode(err, fallback), message, err)
}

func classifyPostgresErrCode(err error, fallback srverr.ErrCode) srverr.ErrCode {
	if err == nil {
		return fallback
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return srverr.ErrCodeTimeout
	}

	var stateErr sqlStater
	if !errors.As(err, &stateErr) {
		return fallback
	}

	sqlState := stateErr.SQLState()
	switch {
	// SQLSTATE class 28: invalid authorisation specification (invalid user/password).
	case strings.HasPrefix(sqlState, "28"):
		return srverr.ErrCodeDependencyFailure
	// SQLSTATE 42501: insufficient privilege for the configured database user.
	case sqlState == "42501":
		return srverr.ErrCodeDependencyFailure
	// SQLSTATE 23503: foreign key violation (referenced dependency does not exist).
	case sqlState == "23503":
		return srverr.ErrCodeValidation
	// SQLSTATE 23514: check constraint violation (domain constraint failed).
	case sqlState == "23514":
		return srverr.ErrCodeValidation
	// SQLSTATE 23502: not null violation (required field missing).
	case sqlState == "23502":
		return srverr.ErrCodeValidation
	// SQLSTATE 23505: unique constraint violation (resource already exists).
	case sqlState == "23505":
		return srverr.ErrCodeStateConflict
	// Remaining SQLSTATE class 23 values are treated as integrity failures.
	case strings.HasPrefix(sqlState, "23"):
		return srverr.ErrCodeInternal
	// SQLSTATE 40001/40P01: serialisation failure or deadlock detected.
	case sqlState == "40001" || sqlState == "40P01":
		return srverr.ErrCodeStateConflict
	// SQLSTATE class 08: connection exceptions between service and database.
	case strings.HasPrefix(sqlState, "08"):
		return srverr.ErrCodeDependencyFailure
	default:
		return fallback
	}
}
