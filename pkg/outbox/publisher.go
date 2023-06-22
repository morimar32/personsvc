package outbox

import (
	"context"
	"time"
)

type Message struct {
	Id            string
	EventName     string
	EventDateTime time.Time
	Payload       string
}

type Publisher interface {
	PublishToQueue(ctx context.Context, msg Message) error
	Shutdown()
}
