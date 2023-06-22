package eventing

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/morimar32/helpers/outbox"
	"github.com/morimar32/helpers/retry"

	"github.com/google/uuid"
)

/*
--DROP TABLE Outbox
CREATE TABLE Outbox (

	Id UNIQUEIDENTIFIER NOT NULL PRIMARY KEY DEFAULT(NEWID()),
	Topic VARCHAR(255) NOT NULL,
	EventName VARCHAR(255) NOT NULL,
	Payload TEXT NOT NULL,
	[Status] VARCHAR(100) NOT NULL DEFAULT('Unpublished'),
	CreatedDateTime DATETIME NOT NULL DEFAULT(CURRENT_TIMESTAMP),
	PublishedDateTime DATETIME NULL,
	ErrorCount INT NOT NULL DEFAULT(0),
	ErrorMessage VARCHAR(255) NULL

);
GO

*/

const (
	AddEventSql           = "INSERT INTO Outbox ( Id, Topic, EventName, Payload, [Status], CreatedDateTime, ErrorCount ) VALUES (@Id, @Topic, @EventName, @Payload, @Status, @CreatedDateTime, 0 )"
	GetPendingMessagesSql = "SELECT TOP 50 Id, Topic, EventName, CreatedDateTime, Payload FROM Outbox WITH (NOLOCK) WHERE PublishedDateTime IS NULL AND ErrorCount < 10 ORDER BY CreatedDateTime"
	ClearEventSql         = "UPDATE Outbox SET [Status] = 'Published', PublishedDateTime = @PublishedDateTime, ErrorCount = 0 WHERE Id = @Id;"
	ErroredEventSql       = "UPDATE O SET O.[Status] = 'Error', O.ErrorMessage = @ErrorMessage, O.ErrorCount = P.ErrorCount + 1 FROM Outbox O INNER JOIN Outbox P ON O.Id = P.Id AND  O.Id = @Id;"
)

type OutboxStorage struct {
	conn                   *sql.DB
	dbOnce                 sync.Once
	policy                 *retry.DbRetry
	AddEventStmt           *sql.Stmt
	GetPendingMessagesStmt *sql.Stmt
	ClearEventStmt         *sql.Stmt
	ErroredEventStmt       *sql.Stmt
}

func NewOutboxStorage(policy *retry.DbRetry) *OutboxStorage {
	o := &OutboxStorage{}
	o.policy = policy
	return o
}

func (o *OutboxStorage) Init(db *sql.DB) error {
	o.dbOnce.Do(func() {
		var err error
		o.conn = db
		o.AddEventStmt, err = o.conn.Prepare(AddEventSql)
		if err != nil {
			log.Fatal(fmt.Errorf("outbox: Failed to Prepare AddEventSql: %w", err))
		}
		o.GetPendingMessagesStmt, err = o.conn.Prepare(GetPendingMessagesSql)
		if err != nil {
			log.Fatal(fmt.Errorf("outbox: Failed to Prepare GetPendingMessagesSql: %w", err))
		}
		o.ClearEventStmt, err = o.conn.Prepare(ClearEventSql)
		if err != nil {
			log.Fatal(fmt.Errorf("outbox: Failed to Prepare ClearEventSql: %w", err))
		}
		o.ErroredEventStmt, err = o.conn.Prepare(ErroredEventSql)
		if err != nil {
			log.Fatal(fmt.Errorf("outbox: Failed to Prepare ErroredEventSql: %w", err))
		}
	})

	return nil
}

func (o *OutboxStorage) CreateTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, err := o.conn.BeginTx(ctx, &sql.TxOptions{})
	return tx, err
}
func (o *OutboxStorage) AddEvent(ctx context.Context, tx *sql.Tx, topic string, eventName string, payload string) error {
	tx_started := false
	var err error
	if tx == nil {
		tx, err = o.conn.BeginTx(ctx, &sql.TxOptions{})
		tx_started = true
		defer tx.Rollback()
		if err != nil {
			return err
		}
	}

	_, err = o.policy.ExecContext(ctx, tx, o.AddEventStmt,
		sql.Named("Id", uuid.NewString()),
		sql.Named("Topic", topic),
		sql.Named("EventName", eventName),
		sql.Named("Payload", payload),
		sql.Named("Status", "Unpublished"),
		sql.Named("CreatedDateTime", time.Now().UTC()))
	if err == nil && tx_started {
		tx.Commit()
	}
	return err
}
func (o *OutboxStorage) GetPendingMessages(ctx context.Context, tx *sql.Tx) (*[]outbox.Message, error) {
	tx_started := false
	var err error
	if tx == nil {
		tx, err = o.conn.BeginTx(ctx, &sql.TxOptions{})
		tx_started = true
		defer tx.Rollback()
		if err != nil {
			return nil, err
		}
	}

	rows, err := o.policy.QueryContext(ctx, tx, o.GetPendingMessagesStmt)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()
	messages := make([]outbox.Message, 0)
	for rows.Next() {
		var (
			db_Id              string
			db_Topic           string
			db_EventName       string
			db_CreatedDateTime time.Time
			db_Payload         string
		)
		err = rows.Scan(&db_Id, &db_Topic, &db_EventName, &db_CreatedDateTime, &db_Payload)
		if err != nil {
			return nil, err
		}
		msg := &outbox.Message{
			Id:            db_Id,
			EventName:     db_EventName,
			EventDateTime: db_CreatedDateTime,
			Payload:       db_Payload,
		}

		messages = append(messages, *msg)
	}
	if err == nil && tx_started {
		tx.Commit()
	}
	return &messages, nil
}
func (o *OutboxStorage) ClearEvent(ctx context.Context, tx *sql.Tx, id string) error {
	tx_started := false
	var err error
	if tx == nil {
		tx, err = o.conn.BeginTx(ctx, &sql.TxOptions{})
		tx_started = true
		defer tx.Rollback()
		if err != nil {
			return err
		}
	}

	_, err = o.policy.ExecContext(ctx, tx, o.ClearEventStmt, sql.Named("PublishedDateTime", time.Now().UTC()), sql.Named("Id", []byte(id)))
	if err == nil && tx_started {
		tx.Commit()
	}
	return err
}
func (o *OutboxStorage) ErroredEvent(ctx context.Context, tx *sql.Tx, id string, errorMessage error) error {
	tx_started := false
	var err error
	if tx == nil {
		tx, err = o.conn.BeginTx(ctx, &sql.TxOptions{})
		tx_started = true
		defer tx.Rollback()
		if err != nil {
			return err
		}
	}

	_, err = o.policy.ExecContext(ctx, tx, o.ErroredEventStmt, sql.Named("ErrorMessage", errorMessage.Error()), sql.Named("Id", []byte(id)))
	if err == nil && tx_started {
		tx.Commit()
	}
	return err
}
