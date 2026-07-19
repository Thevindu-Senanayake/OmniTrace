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

// OrderLine is a single requested item in an order request.
type OrderLine struct {
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

// PlaceOrderRequest is the incoming order payload (one or more line items).
type PlaceOrderRequest struct {
	CustomerID string      `json:"customerId"`
	Items      []OrderLine `json:"items"`
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

// PlaceOrder validates the request, prices every line from the catalog,
// persists a PENDING order with its items and publishes order.created. Prices
// always come from the catalog — never from the client.
func (s *OrderService) PlaceOrder(ctx context.Context, req PlaceOrderRequest, requestID string) (domain.Order, error) {
	if req.CustomerID == "" {
		return domain.Order{}, fmt.Errorf("%w: customerId is required", ErrValidation)
	}
	if len(req.Items) == 0 {
		return domain.Order{}, fmt.Errorf("%w: at least one item is required", ErrValidation)
	}

	seen := make(map[string]struct{}, len(req.Items))
	items := make([]domain.OrderItem, 0, len(req.Items))
	total := 0.0
	for _, line := range req.Items {
		if _, err := uuid.Parse(line.ItemID); err != nil {
			return domain.Order{}, fmt.Errorf("%w: itemId must be a valid UUID", ErrValidation)
		}
		if line.Quantity <= 0 {
			return domain.Order{}, fmt.Errorf("%w: quantity must be positive for item %s", ErrValidation, line.ItemID)
		}
		if _, dup := seen[line.ItemID]; dup {
			return domain.Order{}, fmt.Errorf("%w: item %s listed more than once", ErrValidation, line.ItemID)
		}
		seen[line.ItemID] = struct{}{}

		product, err := s.catalog.GetProduct(ctx, line.ItemID, requestID)
		if errors.Is(err, catalog.ErrProductNotFound) {
			return domain.Order{}, fmt.Errorf("%w: %s", ErrProductNotFound, line.ItemID)
		}
		if err != nil {
			return domain.Order{}, err
		}

		items = append(items, domain.OrderItem{
			ItemID:    line.ItemID,
			Quantity:  line.Quantity,
			UnitPrice: product.Price,
		})
		total += product.Price * float64(line.Quantity)
	}
	total = math.Round(total*100) / 100

	order, err := s.repo.Create(ctx, domain.Order{
		CustomerID:  req.CustomerID,
		Items:       items,
		TotalAmount: total,
	})
	if err != nil {
		return domain.Order{}, err
	}
	s.logger.Info("order persisted", "order_id", order.ID, "request_id", requestID, "status", order.Status, "items", len(order.Items))

	eventItems := make([]domain.EventItem, len(order.Items))
	for i, it := range order.Items {
		eventItems[i] = domain.EventItem{ItemID: it.ItemID, Quantity: it.Quantity, UnitPrice: it.UnitPrice}
	}
	event := domain.OrderCreatedEvent{
		EventType:   kafkapkg.TopicOrderCreated,
		Version:     1,
		OrderID:     order.ID,
		CustomerID:  order.CustomerID,
		Items:       eventItems,
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
