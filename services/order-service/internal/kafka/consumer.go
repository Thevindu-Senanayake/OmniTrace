package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/domain"
)

// Saga topics consumed by the Order Service.
const (
	TopicPaymentSuccess  = "payment.success"
	TopicPaymentFailed   = "payment.failed"
	TopicInventoryFailed = "inventory.failed"
)

const consumerGroup = "order-service"

// SagaHandler applies an order status transition for an event from a topic.
type SagaHandler func(ctx context.Context, topic, orderID string) error

// RunSagaConsumers starts one reader goroutine per saga topic and blocks
// until ctx is cancelled. Messages are committed only after the handler
// succeeds, so failures are redelivered (handlers are idempotent).
func RunSagaConsumers(ctx context.Context, brokers []string, handle SagaHandler, logger *slog.Logger) {
	topics := []string{TopicPaymentSuccess, TopicPaymentFailed, TopicInventoryFailed}

	var wg sync.WaitGroup
	for _, topic := range topics {
		wg.Add(1)
		go func(topic string) {
			defer wg.Done()
			consumeTopic(ctx, brokers, topic, handle, logger)
		}(topic)
	}
	wg.Wait()
}

func consumeTopic(ctx context.Context, brokers []string, topic string, handle SagaHandler, logger *slog.Logger) {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: brokers,
		GroupID: consumerGroup,
		Topic:   topic,
	})
	defer reader.Close()

	logger.Info("saga consumer started", "topic", topic)
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, ctx.Err()) {
				logger.Info("saga consumer stopping", "topic", topic)
				return
			}
			logger.Error("fetch message failed", "topic", topic, "error", err)
			continue
		}

		var event domain.SagaEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil || event.OrderID == "" {
			// Malformed message: log and commit — redelivery cannot fix it.
			logger.Error("malformed saga event, skipping", "topic", topic, "error", err)
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		if err := handle(ctx, topic, event.OrderID); err != nil {
			// Do not commit: message will be redelivered and retried.
			logger.Error("saga handler failed, will retry", "topic", topic, "order_id", event.OrderID, "error", err)
			continue
		}
		if err := reader.CommitMessages(ctx, msg); err != nil {
			logger.Error("commit failed", "topic", topic, "order_id", event.OrderID, "error", err)
		}
	}
}
