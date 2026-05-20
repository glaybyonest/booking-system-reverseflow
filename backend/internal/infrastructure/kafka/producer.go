package kafka

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	brokers []string
}

func NewProducer(brokers []string) *Producer {
	return &Producer{brokers: brokers}
}

func EnsureTopics(ctx context.Context, brokers []string, topics []string) error {
	if len(brokers) == 0 || len(topics) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerAddr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	controllerConn, err := kafka.DialContext(ctx, "tcp", controllerAddr)
	if err != nil {
		return err
	}
	defer controllerConn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = controllerConn.SetDeadline(deadline)
	}

	configs := make([]kafka.TopicConfig, 0, len(topics))
	seen := make(map[string]struct{}, len(topics))
	for _, topic := range topics {
		if topic == "" {
			continue
		}
		if _, ok := seen[topic]; ok {
			continue
		}
		seen[topic] = struct{}{}
		configs = append(configs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}
	if len(configs) == 0 {
		return nil
	}

	if err := controllerConn.CreateTopics(configs...); err != nil && !errors.Is(err, kafka.TopicAlreadyExists) {
		return err
	}
	return nil
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
