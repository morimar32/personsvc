package outbox

import (
	"database/sql"
	retry "personsvc/internal/retry"
)

type Outboxer interface {
	Publish(tx *sql.Tx, topic string, eventName string, payload any) error
}

type Outbox struct {
	db     *sql.DB
	policy *retry.DbRetry
}

func New(opts ...OutboxOption) (Outboxer, error) {
	o := Outbox{}

	for _, opt := range opts {
		opt(&o)
	}

	return &o, nil
}

func (o *Outbox) Publish(tx *sql.Tx, topic string, eventName string, payload any) error {
	return nil
}
