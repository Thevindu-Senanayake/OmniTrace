package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/handler"
)

// New assembles the chi router with all routes and middleware.
func New(products *handler.ProductHandler, health *handler.HealthHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Health)
	r.Get("/products", products.List)
	r.Get("/products/{id}", products.GetByID)

	return r
}
