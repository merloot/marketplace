package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/merloot/marketplace/services/catalog/internal/app"
	"github.com/merloot/marketplace/services/catalog/internal/domain"
	"github.com/merloot/marketplace/services/catalog/internal/httpapi"
	"github.com/stretchr/testify/assert"
)

type MockUseCase struct {
	ID  uuid.UUID
	Err error
}

func (m *MockUseCase) Execute(ctx context.Context, input app.CreateProductInput) (uuid.UUID, error) {
	return m.ID, m.Err
}

func TestProductHandler_Create(t *testing.T) {
	validUUID := uuid.New()
	validBody := `{"title":"Test","description":"Desc","category":"Cat","price_amount":1000,"currency":"RUB","sku":"SKU-123"}`

	t.Run("success 201", func(t *testing.T) {
		mockUC := &MockUseCase{ID: validUUID}
		handler := httpapi.NewProductHandler(mockUC)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(validBody))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, validUUID.String(), resp["product_id"])
	})

	t.Run("invalid json 400", func(t *testing.T) {
		mockUC := &MockUseCase{}
		handler := httpapi.NewProductHandler(mockUC)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(`{bad json}`))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("domain error 422", func(t *testing.T) {
		mockUC := &MockUseCase{Err: domain.ErrEmptyTitle}
		handler := httpapi.NewProductHandler(mockUC)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(validBody))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("duplicate sku 409", func(t *testing.T) {
		mockUC := &MockUseCase{Err: domain.ErrDuplicateSKU}
		handler := httpapi.NewProductHandler(mockUC)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(validBody))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}
