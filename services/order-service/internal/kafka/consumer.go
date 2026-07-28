package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

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

// Fetch failures (broker down, rebalance) are retried with capped backoff. A
// reader whose broker vanished can stay wedged, so after this many consecutive
// failures it is recreated to force a fresh connection and group rejoin.
const (
	fetchRetryBackoff    = 2 * time.Second
	maxFetchRetryBackoff = 30 * time.Second
	failuresBeforeReopen = 5
)

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

// consumeTopic keeps a reader alive for topic until ctx is cancelled. If the
// reader wedges (broker restarted, group rebalance never completed) it is
// closed and reopened, so the consumer survives Kafka being unavailable at
// startup or restarting mid-life.
func consumeTopic(ctx context.Context, brokers []string, topic string, handle SagaHandler, logger *slog.Logger) {
	for ctx.Err() == nil {
		if reopen := runReader(ctx, brokers, topic, handle, logger); !reopen {
			return
		}
		logger.Warn("recreating saga consumer reader", "topic", topic)
	}
}

// runReader owns one reader's lifetime. It returns true when the caller should
// open a fresh reader, false when ctx was cancelled and it should stop.
func runReader(ctx context.Context, brokers []string, topic string, handle SagaHandler, logger *slog.Logger) bool {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: brokers,
		GroupID: consumerGroup,
		Topic:   topic,
	})
	defer reader.Close()

	logger.Info("saga consumer started", "topic", topic)

	failures := 0
	backoff := fetchRetryBackoff
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Info("saga consumer stopping", "topic", topic)
				return false
			}
			failures++
			logger.Error("fetch message failed", "topic", topic, "failures", failures, "error", err)
			if failures >= failuresBeforeReopen {
				return true
			}
			if !sleepCtx(ctx, backoff) {
				return false
			}
			if backoff < maxFetchRetryBackoff {
				backoff *= 2
			}
			continue
		}
		failures, backoff = 0, fetchRetryBackoff

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

// sleepCtx waits for d, returning false if ctx was cancelled first.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
