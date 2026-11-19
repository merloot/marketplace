package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/merloot/marketplace/services/inventory/internal/app"
	"github.com/merloot/marketplace/services/inventory/internal/domain"
)

type addStockRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type AddStockUseCase interface {
	Execute(ctx context.Context, input app.AddStockInput) (uuid.UUID, error)
}

type StockHandler struct {
	addStockUC AddStockUseCase
	log        *slog.Logger
}

func NewStockHandler(addStockUC AddStockUseCase, log *slog.Logger) *StockHandler {
	return &StockHandler{addStockUC: addStockUC, log: log}
}

func (h *StockHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req addStockRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", domain.ErrInvalidProduct.Error())
		return
	}
	if req.Quantity < 0 {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", domain.ErrInvalidQuantity.Error())
		return
	}
	id, err := h.addStockUC.Execute(r.Context(), app.AddStockInput{
		ProductID: productID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		h.log.Error("failed to add stock", "error", err)
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"product_id": id.String()})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidProduct):
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
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
