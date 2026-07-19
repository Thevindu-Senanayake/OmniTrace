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

// Create persists the order header and all its line items in one transaction,
// so an order is never half-written.
func (r *PostgresOrderRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful commit

	err = tx.QueryRow(ctx,
		`INSERT INTO orders (customer_id, total_amount, status)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		order.CustomerID, order.TotalAmount, string(domain.StatusPending),
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return domain.Order{}, err
	}

	for _, item := range order.Items {
		if _, err := tx.Exec(ctx,
			`INSERT INTO order_items (order_id, item_id, quantity, unit_price)
			 VALUES ($1, $2, $3, $4)`,
			order.ID, item.ItemID, item.Quantity, item.UnitPrice,
		); err != nil {
			return domain.Order{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
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
