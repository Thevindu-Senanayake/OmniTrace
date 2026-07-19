package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/service"
)

// OrderHandler serves the order placement endpoint.
type OrderHandler struct {
	svc    *service.OrderService
	logger *slog.Logger
}

func NewOrderHandler(svc *service.OrderService, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, logger: logger}
}

// Create handles POST /order and returns 202 Accepted — the saga decides
// the final order status asynchronously.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")

	var req service.PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	order, err := h.svc.PlaceOrder(r.Context(), req, requestID)
	switch {
	case errors.Is(err, service.ErrValidation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrProductNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case err != nil:
		h.logger.Error("place order failed", "request_id", requestID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	default:
		writeJSON(w, http.StatusAccepted, map[string]any{
			"orderId": order.ID,
			"status":  order.Status,
		})
	}
}
