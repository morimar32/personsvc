package outbox

import (
	"database/sql"
	retry "personsvc/internal/retry"
)

type OutboxOption func(*Outbox)

func WithConnection(conn *sql.DB) OutboxOption {
	return func(o *Outbox) {
		o.db = conn
	}
}

func WithPolicy(pol *retry.DbRetry) OutboxOption {
	return func(o *Outbox) {
		o.policy = pol
	}
}
