package utils

import (
	"context"
	"crypto/tls"
	"encoding/json"

	"github.com/bengosborn/cue/helpers"
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
	prefix string
}

// Initialize new broker
func NewBrokerKafka(ctx context.Context, username string, password string, endpoint string, topicName string, prefix string) (*BrokerKafka, error) {
	broker := BrokerKafka{
		ctx:    ctx,
		prefix: prefix,
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

		key := helpers.FormatKey(b.prefix, msg.Id)

		go func() {
			if lock != nil {
				lock.Lock(key)
				defer lock.Unlock(key, false)

				if processed, err := lock.IsProcessed(key); processed || err != nil {
					return
				}

				ok := fn(&msg)
				lock.Unlock(key, ok)
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
	prefix  string
}

// Initialize new broker
func NewBrokerRedis(ctx context.Context, redis *redis.Client, channel string, prefix string) *BrokerRedis {
	return &BrokerRedis{client: redis, ctx: ctx, channel: channel, prefix: prefix}
}

// Listen to broker events
func (b *BrokerRedis) Listen(fn func(*BrokerMessage) bool, lock *ResourceLockDistributed) error {
	pubsub := b.client.Subscribe(b.ctx, b.channel)
	ch := pubsub.Channel()
	defer pubsub.Close()

	for rawMsg := range ch {
		var msg BrokerMessage
		if err := json.Unmarshal([]byte(rawMsg.Payload), &msg); err != nil {
			continue
		}

		key := helpers.FormatKey(b.prefix, msg.Id)

		go func() {
			if lock != nil {
				lock.Lock(key)
				defer lock.Unlock(key, false)

				if processed, err := lock.IsProcessed(key); processed || err != nil {
					return
				}

				ok := fn(&msg)
				lock.Unlock(key, ok)
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
