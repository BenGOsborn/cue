package gateway

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
	closed bool
}

// Initialize new queue topic
func NewQueue(username string, password string, endpoint string, topicName string, logger *log.Logger) (*Queue, error) {
	queue := Queue{}

	mechanism, err := scram.Mechanism(scram.SHA512, username, password)

	queue.closed = false

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
func (queue *Queue) Close() {
	queue.closed = true

	queue.reader.Close()
	queue.writer.Close()
}

// Listen to queue events
func (queue *Queue) Listen(fn func(*QueueMessage) error) error {
	for {
		msg, err := queue.reader.ReadMessage(context.TODO())

		if err != nil {
			return err
		}

		var queueMessage QueueMessage
		err = json.Unmarshal([]byte(msg.Value), &queueMessage)

		if err != nil {
			return err
		}

		if err := fn(&queueMessage); err != nil {
			return err
		}
	}
}

// Send message
func (queue *Queue) Send(msg *QueueMessage) error {
	data, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	return queue.writer.WriteMessages(context.TODO(), kafka.Message{Value: []byte(data)})
}
