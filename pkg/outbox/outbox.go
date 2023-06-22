package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type Outboxer interface {
	Init(db *sql.DB, shutdown <-chan bool, errors chan<- error) error
	Publish(ctx context.Context, tx *sql.Tx, topic string, eventName string, payload any) error
}

type Outbox struct {
	db         *sql.DB
	databaser  Databaser
	publisher  Publisher
	pollDelay  time.Duration
	msgBatchTx bool
}

func New(opts ...OutboxOption) (Outboxer, error) {
	o := Outbox{}
	o.msgBatchTx = false
	o.pollDelay = time.Second * 1

	for _, opt := range opts {
		opt(&o)
	}
	if o.databaser == nil {
		return nil, errors.New("a Db implementation must be provided to the outbox")
	}
	if o.publisher == nil {
		return nil, errors.New("a Publisher implementation must be provided to the outbox")
	}
	return &o, nil
}

func (o *Outbox) Init(db *sql.DB, shutdown <-chan bool, errors chan<- error) error {
	err := o.databaser.Init(db)
	if err != nil {
		return err
	}

	go o.pollMessages(shutdown, errors)
	return nil
}

func (o *Outbox) Publish(ctx context.Context, tx *sql.Tx, topic string, eventName string, payload any) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	err = o.databaser.AddEvent(ctx, tx, topic, eventName, string(payloadJSON))
	return err
}

func (o *Outbox) pollMessages(shutdown <-chan bool, errors chan<- error) {
	ticker := time.NewTicker(o.pollDelay)

	for {
		select {
		case <-shutdown:
			o.publisher.Shutdown()
			return
		case <-ticker.C:
			o.timerElapsed(errors)
		}
	}
}

func (o *Outbox) timerElapsed(errors chan<- error) {
	var tx *sql.Tx
	var err error
	tx_started := false
	if o.msgBatchTx {
		tx, err = o.databaser.CreateTransaction(context.Background())
		if err != nil {
			errors <- err
			return
		}
		tx_started = true
		defer tx.Rollback()
	}

	messages, err := o.databaser.GetPendingMessages(context.Background(), tx)
	if err != nil {
		errors <- err
		return
	}
	if messages == nil || len(*messages) <= 0 {
		return
	}

	for _, msg := range *messages {
		err = o.publisher.PublishToQueue(context.Background(), msg)
		if err != nil {
			errors <- err
			err = o.databaser.ErroredEvent(context.Background(), tx, msg.Id, err)
			if err != nil {
				errors <- err
			}
		} else {
			err = o.databaser.ClearEvent(context.Background(), tx, msg.Id)
			if err != nil {
				errors <- err
			}
		}
	}
	if tx_started {
		tx.Commit()
	}
}
