package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/domain"
)

const TopicOrderCreated = "order.created"

// Producer publishes order events. Message key = orderId so all events
// for one order land on the same partition (ordering guarantee).
type Producer struct {
	writer *kafkago.Writer
	logger *slog.Logger
}

func NewProducer(brokers []string, logger *slog.Logger) *Producer {
	return &Producer{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        TopicOrderCreated,
			Balancer:     &kafkago.Hash{},
			RequiredAcks: kafkago.RequireAll,
		},
		logger: logger,
	}
}

// PublishOrderCreated emits the order.created event with the request ID
// carried in a header for correlation.
func (p *Producer) PublishOrderCreated(ctx context.Context, event domain.OrderCreatedEvent, requestID string) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafkago.Message{
		Key:   []byte(event.OrderID),
		Value: payload,
	}
	if requestID != "" {
		msg.Headers = append(msg.Headers, kafkago.Header{Key: "x-request-id", Value: []byte(requestID)})
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return err
	}
	p.logger.Info("published order.created", "order_id", event.OrderID, "request_id", requestID)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
