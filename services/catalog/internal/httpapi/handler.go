package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/merloot/marketplace/services/catalog/internal/app"
	"github.com/merloot/marketplace/services/catalog/internal/domain"
)

type createProductRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	PriceAmount int64  `json:"price_amount"`
	Currency    string `json:"currency"`
	SKU         string `json:"sku"`
}

type CreateProductUseCase interface {
	Execute(ctx context.Context, input app.CreateProductInput) (uuid.UUID, error)
}

type ProductHandler struct {
	createUC CreateProductUseCase
}

type ProductUseCase interface {
	Execute(ctx context.Context, input app.CreateProductInput) (uuid.UUID, error)
}

func NewProductHandler(createUC CreateProductUseCase) *ProductHandler {
	return &ProductHandler{createUC: createUC}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	id, err := h.createUC.Execute(r.Context(), app.CreateProductInput{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		PriceAmount: req.PriceAmount,
		Currency:    req.Currency,
		SKU:         req.SKU,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"product_id": id.String()})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrEmptyTitle),
		errors.Is(err, domain.ErrNegativePrice),
		errors.Is(err, domain.ErrInvalidSKU):
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
	case errors.Is(err, domain.ErrDuplicateSKU):
		writeError(w, http.StatusConflict, "duplicate_sku", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"code": code, "message": msg})
}
