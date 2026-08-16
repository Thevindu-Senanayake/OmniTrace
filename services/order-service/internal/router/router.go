package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/handler"
)

// New assembles the chi router with all routes and middleware, wrapped in
// otelhttp so every request extracts any inbound W3C traceparent and runs under
// a server span. That span is what the api-gateway hop links to and what the
// downstream catalog call and db queries hang their child spans off.
func New(orders *handler.OrderHandler, health *handler.HealthHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Health)
	r.Post("/order", orders.Create)

	return otelhttp.NewHandler(r, "http.server")
}
