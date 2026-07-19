package domain

import "time"

type OrderStatus string

// The saga moves an order from PENDING to exactly one terminal state.
const (
	StatusPending       OrderStatus = "PENDING"
	StatusConfirmed     OrderStatus = "CONFIRMED"
	StatusPaymentFailed OrderStatus = "PAYMENT_FAILED"
	StatusOutOfStock    OrderStatus = "OUT_OF_STOCK"
)

// OrderItem is a single line in an order, priced from the catalog at order time.
type OrderItem struct {
	ItemID    string  `json:"itemId"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

// Order is a customer order backed by orders_db (header + order_items lines).
type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customerId"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"totalAmount"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// EventItem is one line as carried on the wire in order.created.
type EventItem struct {
	ItemID    string  `json:"itemId"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

// OrderCreatedEvent is published to the order.created topic.
type OrderCreatedEvent struct {
	EventType   string      `json:"eventType"`
	Version     int         `json:"version"`
	OrderID     string      `json:"orderId"`
	CustomerID  string      `json:"customerId"`
	Items       []EventItem `json:"items"`
	TotalAmount float64     `json:"totalAmount"`
	Timestamp   time.Time   `json:"timestamp"`
}

// SagaEvent is the minimal shape consumed from payment.success,
// payment.failed and inventory.failed — only orderId is needed to
// apply the status transition.
type SagaEvent struct {
	OrderID string `json:"orderId"`
}
