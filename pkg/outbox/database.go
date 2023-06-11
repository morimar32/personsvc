package outbox

import (
	"context"
	"database/sql"
)

type Databaser interface {
	Init(db *sql.DB) error
	CreateTransaction(ctx context.Context) (*sql.Tx, error)
	AddEvent(ctx context.Context, tx *sql.Tx, topic string, eventName string, payload string) error
	GetPendingMessages(ctx context.Context, tx *sql.Tx) (*[]Message, error)
	ClearEvent(ctx context.Context, tx *sql.Tx, id string) error
	ErroredEvent(ctx context.Context, tx *sql.Tx, id string, errorMessage error) error
}
