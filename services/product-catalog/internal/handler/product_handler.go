package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/repository"
)

// ProductHandler serves catalog read endpoints.
type ProductHandler struct {
	repo   repository.ProductRepository
	logger *slog.Logger
}

func NewProductHandler(repo repository.ProductRepository, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{repo: repo, logger: logger}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.repo.List(r.Context())
	if err != nil {
		h.logger.Error("query products failed", "request_id", r.Header.Get("X-Request-ID"), "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	h.logger.Info("list products", "request_id", r.Header.Get("X-Request-ID"), "count", len(products))
	writeJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.repo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		h.logger.Warn("product not found", "request_id", r.Header.Get("X-Request-ID"), "product_id", id)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found", "id": id})
		return
	}
	if err != nil {
		h.logger.Error("query product failed", "request_id", r.Header.Get("X-Request-ID"), "product_id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	h.logger.Info("get product", "request_id", r.Header.Get("X-Request-ID"), "product_id", id)
	writeJSON(w, http.StatusOK, p)
}
