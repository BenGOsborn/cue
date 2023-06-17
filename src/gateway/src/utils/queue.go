package gateway

import (
	"context"
	"crypto/tls"
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
// **** Perhaps it should be able to create the new topic by itself
func NewQueue(username string, password string, endpoint string, groupName string, topicName string, logger *log.Logger) *Queue {
	queue := Queue{}

	mechanism, err := scram.Mechanism(scram.SHA512, username, password)

	queue.closed = false

	if err != nil {
		logger.Fatalln(err)
	}

	queue.dialer = &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}

	queue.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{endpoint},
		GroupID: groupName,
		Topic:   topicName,
		Dialer:  queue.dialer,
	})

	queue.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{endpoint},
		Topic:   topicName,
		Dialer:  queue.dialer,
	})

	return &queue
}

// Close the queue
func (queue *Queue) Close() {
	queue.closed = true

	queue.reader.Close()
	queue.writer.Close()
}

// Listen to queue events
func (queue *Queue) Listen(fn func(string) error) error {
	for {
		msg, err := queue.reader.ReadMessage(context.TODO())

		if err != nil {
			return err
		}

		if err := fn(string(msg.Value)); err != nil {
			return err
		}
	}
}

// Send message
func (queue *Queue) Send(msg Message) error {
	return queue.writer.WriteMessages(context.TODO(), kafka.Message{Value: []byte(msg.Message)})
}
