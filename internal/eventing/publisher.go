package eventing

import (
	"context"
	"fmt"
	"personsvc/pkg/outbox"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn         *amqp.Connection
	channelRef   *amqp.Channel
	queueAddress string
	queueName    string
	exchangeName string
}

func NewPublisher(queueAddr string, queuename string) (*Publisher, error) {
	var p = &Publisher{
		queueAddress: queueAddr,
		queueName:    queuename,
	}
	var err error
	p.conn, err = amqp.Dial(p.queueAddress)
	if err != nil {
		return nil, err
	}
	p.channelRef, err = p.conn.Channel()
	if err != nil {
		p.conn.Close()
		return nil, err
	}
	p.exchangeName = fmt.Sprintf("%sExchange", p.queueName)
	err = p.channelRef.ExchangeDeclare(
		p.exchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		p.channelRef.Close()
		p.conn.Close()
		return nil, err
	}

	return p, nil
}
func (p *Publisher) PublishToQueue(ctx context.Context, msg outbox.Message) error {
	userId := ""
	v := ctx.Value("UserId")
	if v != nil {
		userId = v.(string)
	}
	err := p.channelRef.PublishWithContext(
		ctx,
		p.exchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/json",
			MessageId:   msg.Id,
			Timestamp:   msg.EventDateTime,
			Type:        msg.EventName,
			UserId:      userId,
			Body:        []byte(msg.Payload),
		},
	)
	return err
}

func (p *Publisher) Shutdown() {
	p.channelRef.Close()
	p.conn.Close()
}
