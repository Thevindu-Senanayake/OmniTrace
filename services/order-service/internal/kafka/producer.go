package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	kafkago "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/domain"
)

const TopicOrderCreated = "order.created"

// Producer publishes order events. Message key = orderId so all events
// for one order land on the same partition (ordering guarantee).
type Producer struct {
	writer *kafkago.Writer
	tracer trace.Tracer
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
		tracer: otel.Tracer("order-service/kafka"),
		logger: logger,
	}
}

// PublishOrderCreated emits the order.created event with the request ID carried
// in a header for correlation. It opens a producer span and injects the W3C
// traceparent into the message headers so the inventory consumer continues this
// same trace across the async hop.
func (p *Producer) PublishOrderCreated(ctx context.Context, event domain.OrderCreatedEvent, requestID string) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, span := p.tracer.Start(ctx, TopicOrderCreated+" publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination.name", TopicOrderCreated),
			attribute.String("order_id", event.OrderID),
		),
	)
	defer span.End()

	msg := kafkago.Message{
		Key:   []byte(event.OrderID),
		Value: payload,
	}
	if requestID != "" {
		msg.Headers = append(msg.Headers, kafkago.Header{Key: "x-request-id", Value: []byte(requestID)})
	}
	otel.GetTextMapPropagator().Inject(ctx, headerCarrier{&msg.Headers})

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		span.RecordError(err)
		return err
	}
	p.logger.InfoContext(ctx, "published order.created", "order_id", event.OrderID, "request_id", requestID)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
