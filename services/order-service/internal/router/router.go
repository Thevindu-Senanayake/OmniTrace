package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/handler"
)

// New assembles the chi router with all routes and middleware, wrapped in
// otelhttp so every request extracts any inbound W3C traceparent and runs under
// a server span. That span is what the api-gateway hop links to and what the
// downstream catalog call and db queries hang their child spans off.
func New(orders *handler.OrderHandler, health *handler.HealthHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(otelRouteName)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Health)
	r.Post("/order", orders.Create)

	return otelhttp.NewHandler(r, "http.server")
}

// otelRouteName upgrades the otelhttp server span from the bare HTTP method
// ("POST") to "POST /order" once chi has matched the route. otelhttp wraps the
// mux from the outside and names the span before routing runs, so the route
// pattern is not known at span creation; chi only fills RoutePattern() during
// next.ServeHTTP, so we read it on the way back out and rename the span while
// it is still open. Registered outermost so it also runs after Recoverer
// absorbs a panic. Without this every request collapses onto its method
// ("GET", "POST") in the RED span-metrics, so /health and real routes share one
// series and health probes can't be filtered out of the dashboards.
func otelRouteName(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			if pattern := rctx.RoutePattern(); pattern != "" {
				span := trace.SpanFromContext(r.Context())
				span.SetName(r.Method + " " + pattern)
				span.SetAttributes(attribute.String("http.route", pattern))
			}
		}
	})
}
