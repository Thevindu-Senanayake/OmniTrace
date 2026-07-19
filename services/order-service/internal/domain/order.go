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

// Order is a customer order backed by orders_db.
type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customerId"`
	ItemID      string      `json:"itemId"`
	Quantity    int         `json:"quantity"`
	TotalAmount float64     `json:"totalAmount"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// OrderCreatedEvent is published to the order.created topic.
type OrderCreatedEvent struct {
	EventType   string    `json:"eventType"`
	Version     int       `json:"version"`
	OrderID     string    `json:"orderId"`
	CustomerID  string    `json:"customerId"`
	ItemID      string    `json:"itemId"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"totalAmount"`
	Timestamp   time.Time `json:"timestamp"`
}

// SagaEvent is the minimal shape consumed from payment.success,
// payment.failed and inventory.failed — only orderId is needed to
// apply the status transition.
type SagaEvent struct {
	OrderID string `json:"orderId"`
}
