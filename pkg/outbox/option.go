package outbox

import (
	"database/sql"
	"time"
)

type OutboxOption func(*Outbox)

func WithConnection(conn *sql.DB) OutboxOption {
	return func(o *Outbox) {
		o.db = conn
	}
}

func WithPublisher(publisher Publisher) OutboxOption {
	return func(o *Outbox) {
		o.publisher = publisher
	}
}

func WithDatabase(dber Databaser) OutboxOption {
	return func(o *Outbox) {
		o.databaser = dber
	}
}

func WithPollingInterval(delay time.Duration) OutboxOption {
	return func(o *Outbox) {
		o.pollDelay = delay
	}
}

func WithTransactionBehavior(batchMessages bool) OutboxOption {
	return func(o *Outbox) {
		o.msgBatchTx = batchMessages
	}
}
