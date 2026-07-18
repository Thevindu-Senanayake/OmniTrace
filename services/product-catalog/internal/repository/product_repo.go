package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/domain"
)

// ErrNotFound is returned when a product does not exist.
var ErrNotFound = errors.New("product not found")

// ProductRepository defines read access to the product catalog.
type ProductRepository interface {
	List(ctx context.Context) ([]domain.Product, error)
	GetByID(ctx context.Context, id string) (domain.Product, error)
}

// PostgresProductRepository implements ProductRepository against catalog_db.
type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProductRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{pool: pool}
}

const selectColumns = `id, name, COALESCE(description, ''), price, COALESCE(category, ''), created_at`

func (r *PostgresProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+selectColumns+` FROM products ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []domain.Product{}
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	var p domain.Product
	err := r.pool.QueryRow(ctx, `SELECT `+selectColumns+` FROM products WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Product{}, ErrNotFound
	}
	return p, err
}
