package utils

import (
	"context"
	"crypto/tls"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type Broker interface {
	Listen(fn func(*BrokerMessage) bool, lock *ResourceLockDistributed) error
	Send(msg *BrokerMessage) error
}

type BrokerKafka struct {
	dialer *kafka.Dialer
	reader *kafka.Reader
	writer *kafka.Writer
	ctx    context.Context
}

// Initialize new broker
func NewBrokerKafka(ctx context.Context, username string, password string, endpoint string, topicName string) (*BrokerKafka, error) {
	broker := BrokerKafka{
		ctx: ctx,
	}

	mechanism, err := scram.Mechanism(scram.SHA512, username, password)
	if err != nil {
		return nil, err
	}

	broker.dialer = &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}

	broker.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{endpoint},
		Topic:   topicName,
		Dialer:  broker.dialer,
	})

	broker.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{endpoint},
		Topic:   topicName,
		Dialer:  broker.dialer,
	})

	return &broker, nil
}

// Close the broker
func (b *BrokerKafka) Close() error {
	if err := b.reader.Close(); err != nil {
		return err
	}

	return b.writer.Close()
}

// Listen to broker events
func (b *BrokerKafka) Listen(fn func(*BrokerMessage) bool, lock *ResourceLockDistributed) error {
	for {
		rawMsg, err := b.reader.ReadMessage(b.ctx)
		if err != nil {
			return err
		}

		var msg BrokerMessage
		if err := json.Unmarshal([]byte(rawMsg.Value), &msg); err != nil {
			continue
		}

		go func() {
			if lock != nil {
				if err := lock.Lock(msg.Id); err != nil {
					return
				}
				defer lock.Unlock(msg.Id, false)

				if processed, err := lock.IsProcessed(msg.Id); processed || err != nil {
					return
				}

				ok := fn(&msg)
				lock.Unlock(msg.Id, ok)
			} else {
				fn(&msg)
			}
		}()
	}
}

// Send message
func (b *BrokerKafka) Send(msg *BrokerMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return b.writer.WriteMessages(b.ctx, kafka.Message{Value: []byte(data)})
}

type BrokerRedis struct {
	client  *redis.Client
	ctx     context.Context
	channel string
}

// Initialize new broker
func NewBrokerRedis(ctx context.Context, redis *redis.Client, channel string) *BrokerRedis {
	return &BrokerRedis{client: redis, ctx: ctx, channel: channel}
}

// Listen to broker events
func (b *BrokerRedis) Listen(fn func(*BrokerMessage) bool, lock *ResourceLockDistributed) error {
	pubsub := b.client.Subscribe(b.ctx, b.channel)
	ch := pubsub.Channel()

	for rawMsg := range ch {
		var msg BrokerMessage
		if err := json.Unmarshal([]byte(rawMsg.Payload), &msg); err != nil {
			continue
		}

		go func() {
			if lock != nil {
				if err := lock.Lock(msg.Id); err != nil {
					return
				}
				defer lock.Unlock(msg.Id, false)

				if processed, err := lock.IsProcessed(msg.Id); processed || err != nil {
					return
				}

				ok := fn(&msg)
				lock.Unlock(msg.Id, ok)
			} else {
				fn(&msg)
			}
		}()
	}

	return nil
}

// Send message
func (b *BrokerRedis) Send(msg *BrokerMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return b.client.Publish(b.ctx, b.channel, data).Err()
}
