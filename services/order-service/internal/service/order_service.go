package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/catalog"
	"github.com/ieee-yp/ecommerce-observability/order-service/internal/domain"
	kafkapkg "github.com/ieee-yp/ecommerce-observability/order-service/internal/kafka"
	"github.com/ieee-yp/ecommerce-observability/order-service/internal/repository"
)

// Sentinel errors the HTTP layer maps to status codes.
var (
	ErrValidation      = errors.New("validation error")
	ErrProductNotFound = errors.New("product not found")
)

// PlaceOrderRequest is the incoming order payload.
type PlaceOrderRequest struct {
	CustomerID string `json:"customerId"`
	ItemID     string `json:"itemId"`
	Quantity   int    `json:"quantity"`
}

// OrderService orchestrates order placement and saga status transitions.
type OrderService struct {
	repo     repository.OrderRepository
	catalog  *catalog.Client
	producer *kafkapkg.Producer
	logger   *slog.Logger
}

func New(repo repository.OrderRepository, catalogClient *catalog.Client, producer *kafkapkg.Producer, logger *slog.Logger) *OrderService {
	return &OrderService{repo: repo, catalog: catalogClient, producer: producer, logger: logger}
}

// PlaceOrder validates the request, prices it from the catalog, persists a
// PENDING order and publishes order.created. The price always comes from the
// catalog — never from the client.
func (s *OrderService) PlaceOrder(ctx context.Context, req PlaceOrderRequest, requestID string) (domain.Order, error) {
	if req.CustomerID == "" {
		return domain.Order{}, fmt.Errorf("%w: customerId is required", ErrValidation)
	}
	if _, err := uuid.Parse(req.ItemID); err != nil {
		return domain.Order{}, fmt.Errorf("%w: itemId must be a valid UUID", ErrValidation)
	}
	if req.Quantity <= 0 {
		return domain.Order{}, fmt.Errorf("%w: quantity must be positive", ErrValidation)
	}

	product, err := s.catalog.GetProduct(ctx, req.ItemID, requestID)
	if errors.Is(err, catalog.ErrProductNotFound) {
		return domain.Order{}, fmt.Errorf("%w: %s", ErrProductNotFound, req.ItemID)
	}
	if err != nil {
		return domain.Order{}, err
	}

	total := math.Round(product.Price*float64(req.Quantity)*100) / 100
	order, err := s.repo.Create(ctx, domain.Order{
		CustomerID:  req.CustomerID,
		ItemID:      req.ItemID,
		Quantity:    req.Quantity,
		TotalAmount: total,
	})
	if err != nil {
		return domain.Order{}, err
	}
	s.logger.Info("order persisted", "order_id", order.ID, "request_id", requestID, "status", order.Status)

	event := domain.OrderCreatedEvent{
		EventType:   kafkapkg.TopicOrderCreated,
		Version:     1,
		OrderID:     order.ID,
		CustomerID:  order.CustomerID,
		ItemID:      order.ItemID,
		Quantity:    order.Quantity,
		TotalAmount: order.TotalAmount,
		Timestamp:   time.Now().UTC(),
	}
	if err := s.producer.PublishOrderCreated(ctx, event, requestID); err != nil {
		// Order stays PENDING in the DB; surface the failure honestly.
		return domain.Order{}, fmt.Errorf("order %s persisted but event publish failed: %w", order.ID, err)
	}
	return order, nil
}

// HandleSagaEvent applies the status transition for a consumed saga event.
// Idempotent: an already-transitioned order is a no-op.
func (s *OrderService) HandleSagaEvent(ctx context.Context, topic, orderID string) error {
	var newStatus domain.OrderStatus
	switch topic {
	case kafkapkg.TopicPaymentSuccess:
		newStatus = domain.StatusConfirmed
	case kafkapkg.TopicPaymentFailed:
		newStatus = domain.StatusPaymentFailed
	case kafkapkg.TopicInventoryFailed:
		newStatus = domain.StatusOutOfStock
	default:
		s.logger.Error("unknown saga topic", "topic", topic, "order_id", orderID)
		return nil
	}

	applied, err := s.repo.TransitionFromPending(ctx, orderID, newStatus)
	if err != nil {
		return err
	}
	if applied {
		s.logger.Info("order status updated", "order_id", orderID, "status", newStatus, "topic", topic)
	} else {
		s.logger.Info("saga event skipped (already transitioned or unknown order)", "order_id", orderID, "topic", topic)
	}
	return nil
}
