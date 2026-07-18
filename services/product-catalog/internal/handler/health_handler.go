package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler reports service liveness including DB reachability.
type HealthHandler struct {
	pool        *pgxpool.Pool
	serviceName string
}

func NewHealthHandler(pool *pgxpool.Pool, serviceName string) *HealthHandler {
	return &HealthHandler{pool: pool, serviceName: serviceName}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.pool.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "degraded", "service": h.serviceName, "error": "db unreachable",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": h.serviceName})
}
