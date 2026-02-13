package rabbitmq

import (
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type Client struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   string
}

func NewClient(amqURL,queueName string) (*Client,error) {
	conn,err := amqp091.DialConfig(amqURL,amqp091.Config{
		Heartbeat: 10*time.Second,
	})

	if err!=nil {
		return nil,err
	}

	ch, err := conn.Channel()
	if err!=nil {
		return  nil,err
	}

	_,err = ch.QueueDeclare(
		queueName,
		true, // durable
		false, // auto-delete
		false, // exclusive
		false,
		nil,
	)
	if err !=nil {
		return nil ,err
	}

	return &Client{
		conn: conn,
		channel: ch,
		queue: queueName,
	},nil
}