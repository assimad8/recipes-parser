package rabbitmq

import (
	"encoding/json"
	"parser/internal/domain/ports"
	"parser/internal/domain/valueobjects"

	"github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	client *Client
}

var _ ports.Queue = (*Queue)(nil)

func NewQueue(client *Client) *Queue {
	return &Queue{client: client}
}


func (q *Queue) Publish(job valueobjects.Job) error {
	body,err := json.Marshal(job)
	if err!=nil {
		return err
	}

	return q.client.channel.Publish(
		"",//default exchange
		q.client.queue,//routing key
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			DeliveryMode: amqp091.Persistent,
			Body: body,
		},
	)
}