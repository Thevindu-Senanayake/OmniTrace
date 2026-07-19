package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/domain"
)

// OrderRepository defines persistence for orders.
type OrderRepository interface {
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
	// TransitionFromPending updates the status only if the order is still
	// PENDING. Returns false when no row changed (already transitioned or
	// unknown order) — this makes saga event handling idempotent under
	// Kafka's at-least-once delivery.
	TransitionFromPending(ctx context.Context, orderID string, newStatus domain.OrderStatus) (bool, error)
}

// PostgresOrderRepository implements OrderRepository against orders_db.
type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO orders (customer_id, item_id, quantity, total_amount, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		order.CustomerID, order.ItemID, order.Quantity, order.TotalAmount, string(domain.StatusPending),
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return domain.Order{}, err
	}
	order.Status = domain.StatusPending
	return order, nil
}

func (r *PostgresOrderRepository) TransitionFromPending(ctx context.Context, orderID string, newStatus domain.OrderStatus) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = now()
		 WHERE id = $2 AND status = $3`,
		string(newStatus), orderID, string(domain.StatusPending))
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
