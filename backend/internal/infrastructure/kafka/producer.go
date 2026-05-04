package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	brokers []string
}

func NewProducer(brokers []string) *Producer {
	return &Producer{brokers: brokers}
}

func (p *Producer) Publish(ctx context.Context, topic, key string, value []byte) error {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		RequiredAcks: kafka.RequireOne,
		BatchTimeout: 20 * time.Millisecond,
	}
	defer writer.Close()
	return writer.WriteMessages(ctx, kafka.Message{Key: []byte(key), Value: value, Time: time.Now().UTC()})
}

func NewReader(brokers []string, groupID, topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topic:       topic,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
}
