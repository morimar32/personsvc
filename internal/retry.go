package retry

import (
	"context"
	"database/sql"
	"time"
)

type DbRetryOption func(*DbRetry)

func WithRetry(retry int32) DbRetryOption {
	return func(o *DbRetry) {
		o.Retry = retry
	}
}

func WithDelay(delay time.Duration) DbRetryOption {
	return func(o *DbRetry) {
		o.Delay = delay
	}
}

type DbRetry struct {
	Retry int32
	Delay time.Duration
}

func New(opts ...DbRetryOption) (*DbRetry, error) {
	r := DbRetry{
		Retry: 3,
		Delay: time.Millisecond * 5,
	}
	for _, opt := range opts {
		opt(&r)
	}
	return &r, nil
}

func (r *DbRetry) QueryRowContext(ctx context.Context, tx *sql.Tx, query *sql.Stmt, args ...any) *sql.Row {
	stmt := tx.StmtContext(ctx, query)
	var ret *sql.Row = nil
	shouldRetry := false

	for i := 0; i <= int(r.Retry); i++ {
		ret = stmt.QueryRowContext(ctx, args...)
		err := ret.Err()
		if err != nil {
			if err == sql.ErrNoRows {
				return ret
			}
			//TODO: find deadlock error list
		}

		if !shouldRetry {
			return ret
		}
		time.Sleep(r.Delay)
	}
	return ret
}

func (r *DbRetry) QueryContext(ctx context.Context, tx *sql.Tx, query *sql.Stmt, args ...any) (*sql.Rows, error) {
	stmt := tx.StmtContext(ctx, query)
	var rows *sql.Rows = nil
	var err error = nil
	shouldRetry := false

	for i := 0; i <= int(r.Retry); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		err = nil
		rows, err = stmt.QueryContext(ctx, args...)
		if err != nil {
			//TODO: find deadlock error list

		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if !shouldRetry {
			return rows, nil
		}
		time.Sleep(r.Delay)
	}

	return rows, err
}
