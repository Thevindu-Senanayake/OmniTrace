package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/handler"
)

// New assembles the chi router with all routes and middleware.
func New(orders *handler.OrderHandler, health *handler.HealthHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Health)
	r.Post("/order", orders.Create)

	return r
}
