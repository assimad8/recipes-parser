package rabbitmq

import (
	"encoding/json"
	"log"
	"parser/internal/domain/valueobjects"
)

func (q *Queue) Consume(handler func(valueobjects.Job) error) error {
    // Ensure queue exists (durable)
    _, err := q.client.channel.QueueDeclare(
        q.client.queue,
        true,  // durable
        false, // delete when unused
        false,
        false,
        nil,
    )
    if err != nil {
        return err
    }

    msgs, err := q.client.channel.Consume(
        q.client.queue,
        "",    // consumer name
        false, // autoAck = false
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return err
    }

    go func() {
        for msg := range msgs {
            var job valueobjects.Job
            if err := json.Unmarshal(msg.Body, &job); err != nil {
                log.Println("invalid job payload:", err)
                _ = msg.Nack(false, false) // discard bad message
                continue
            }

            if err := handler(job); err != nil {
                log.Println("job failed:", err)
                _ = msg.Nack(false, true) // requeue
                continue
            }

            _ = msg.Ack(false)
        }
    }()

    log.Println("Consumer registered for queue:", q.client.queue)
    return nil
}
