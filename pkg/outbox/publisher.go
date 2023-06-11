package outbox

import "time"

type Message struct {
	Id            string
	EventName     string
	EventDateTime time.Time
	Payload       string
}

type Publisher interface {
	PublishToQueue(msg Message) error
}
