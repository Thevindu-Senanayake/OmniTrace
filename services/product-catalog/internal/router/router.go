package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/handler"
)

// New assembles the chi router with all routes and middleware, wrapped in
// otelhttp so every request extracts any inbound W3C traceparent and runs under
// a server span — this is what links the order-service catalog call into the
// same trace and roots the db query spans.
func New(products *handler.ProductHandler, health *handler.HealthHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Health)
	r.Get("/products", products.List)
	r.Get("/products/{id}", products.GetByID)

	return otelhttp.NewHandler(r, "http.server")
}
