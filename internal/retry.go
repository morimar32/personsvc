package retry

import (
	"context"
	"database/sql"
	"errors"
	"time"

	mssql "github.com/denisenkom/go-mssqldb"
)

type EvalShouldRetry func(error) bool

type DbRetryOption func(*DbRetry)

func WithRetry(retry int) DbRetryOption {
	return func(o *DbRetry) {
		o.Retry = retry
	}
}

func WithDelay(delay time.Duration) DbRetryOption {
	return func(o *DbRetry) {
		o.Delay = delay
	}
}

func WithMSSQLSupport() DbRetryOption {
	return func(o *DbRetry) {
		o.evalError = func(err error) bool {
			if mssqlerr, ok := err.(mssql.Error); ok {
				switch mssqlerr.Number {
				case 1205, 1231: // deadlock
					return true
				case -2: // timeout
				default:
					return false
				}
			}
			return false
		}
	}
}

type DbRetry struct {
	Retry     int
	Delay     time.Duration
	evalError EvalShouldRetry
}

func New(opts ...DbRetryOption) (*DbRetry, error) {
	r := DbRetry{
		Retry: 3,
		Delay: time.Millisecond * 5,
	}
	for _, opt := range opts {
		opt(&r)
	}

	if r.evalError == nil {
		return nil, errors.New("a provider specific error evaluation function must be provided")
	}
	return &r, nil
}

func (policy *DbRetry) QueryRowContext(ctx context.Context, tx *sql.Tx, query *sql.Stmt, args ...any) *sql.Row {
	stmt := tx.StmtContext(ctx, query)
	var ret *sql.Row = nil
	shouldRetry := false

	for i := 0; i <= int(policy.Retry); i++ {
		ret = stmt.QueryRowContext(ctx, args...)
		err := ret.Err()
		if err != nil {
			if err == sql.ErrNoRows {
				return ret
			}
			shouldRetry = policy.evalError(err)
		}

		if !shouldRetry {
			return ret
		}
		time.Sleep(policy.Delay)
	}
	return ret
}

func (policy *DbRetry) QueryContext(ctx context.Context, tx *sql.Tx, query *sql.Stmt, args ...any) (*sql.Rows, error) {
	stmt := tx.StmtContext(ctx, query)
	var rows *sql.Rows = nil
	var err error = nil
	shouldRetry := false

	for i := 0; i <= policy.Retry; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		err = nil
		rows, err = stmt.QueryContext(ctx, args...)
		if err != nil {
			shouldRetry = policy.evalError(err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if !shouldRetry {
			return rows, nil
		}
		time.Sleep(policy.Delay)
	}

	return rows, err
}

func (policy *DbRetry) ExecContext(ctx context.Context, tx *sql.Tx, cmd *sql.Stmt, cmdargs ...any) (sql.Result, error) {
	stmt := tx.StmtContext(ctx, cmd)
	var result sql.Result = nil
	var err error = nil
	shouldRetry := false

	for i := 0; i <= policy.Retry; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		err = nil
		result, err = stmt.ExecContext(ctx, cmdargs...)
		if err != nil {
			shouldRetry = policy.evalError(err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if !shouldRetry {
			return result, nil
		}
		time.Sleep(policy.Delay)
	}
	return result, err
}
