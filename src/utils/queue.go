package utils

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type Queue struct {
	dialer *kafka.Dialer
	reader *kafka.Reader
	writer *kafka.Writer
	ctx    context.Context
}

// Initialize new queue topic
func NewQueue(ctx context.Context, username string, password string, endpoint string, topicName string, logger *log.Logger) (*Queue, error) {
	queue := Queue{
		ctx: ctx,
	}

	mechanism, err := scram.Mechanism(scram.SHA512, username, password)
	if err != nil {
		return nil, err
	}

	queue.dialer = &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}

	queue.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{endpoint},
		Topic:   topicName,
		Dialer:  queue.dialer,
	})

	queue.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{endpoint},
		Topic:   topicName,
		Dialer:  queue.dialer,
	})

	return &queue, nil
}

// Close the queue
func (q *Queue) Close() {
	q.reader.Close()
	q.writer.Close()
}

// Listen to queue events
func (q *Queue) Listen(fn func(*QueueMessage)) error {
	for {
		msg, err := q.reader.ReadMessage(q.ctx)
		if err != nil {
			return err
		}

		var queueMessage QueueMessage

		if err := json.Unmarshal([]byte(msg.Value), &queueMessage); err != nil {
			continue
		}

		go fn(&queueMessage)
	}
}

// Consume queue events (in a group)

// Send message
func (q *Queue) Send(msg *QueueMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return q.writer.WriteMessages(q.ctx, kafka.Message{Value: []byte(data)})
}
