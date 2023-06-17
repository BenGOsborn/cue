package gateway

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type KafkaTopic struct {
	dialer *kafka.Dialer
	reader *kafka.Reader
	writer *kafka.Writer
	closed bool
}

// Initialize new Kafka topic
func (kafkaTopic *KafkaTopic) NewKafkaTopic(username string, password string, logger *log.Logger, endpoint string, groupName string, topicName string) {
	mechanism, err := scram.Mechanism(scram.SHA512, username, password)

	kafkaTopic.closed = false

	if err != nil {
		logger.Fatalln(err)
	}

	// **** Perhaps it should be able to create the new topic by itself

	kafkaTopic.dialer = &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}

	kafkaTopic.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{endpoint},
		GroupID: groupName,
		Topic:   topicName,
		Dialer:  kafkaTopic.dialer,
	})

	kafkaTopic.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{endpoint},
		Topic:   topicName,
		Dialer:  kafkaTopic.dialer,
	})
}

// Close the topic
func (kafkaTopic *KafkaTopic) Close() {
	kafkaTopic.closed = true

	kafkaTopic.reader.Close()
	kafkaTopic.writer.Close()
}

// Register event listener
func (kafkaTopic *KafkaTopic) Listen(fn func(string) error) error {
	for {
		msg, err := kafkaTopic.reader.ReadMessage(context.TODO())

		if err != nil {
			return err
		}

		fn(string(msg.Value))
	}
}

// Send message
func (kafkaTopic *KafkaTopic) Send(msg Message) error {
	return kafkaTopic.writer.WriteMessages(context.TODO(), kafka.Message{Value: []byte(msg.Message)})
}
