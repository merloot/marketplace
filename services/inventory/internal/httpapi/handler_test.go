package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/merloot/marketplace/services/inventory/internal/app"
	"github.com/merloot/marketplace/services/inventory/internal/domain"
)

type mockAddStockUseCase struct {
	id      uuid.UUID
	err     error
	gotIn   app.AddStockInput
	callCnt int
}

func (m *mockAddStockUseCase) Execute(ctx context.Context, input app.AddStockInput) (uuid.UUID, error) {
	m.callCnt++
	m.gotIn = input
	return m.id, m.err
}

func newTestHandler(uc AddStockUseCase) *StockHandler {
	return &StockHandler{
		addStockUC: uc,
		log:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestStockHandler_Add(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name          string
		body          string
		mockID        uuid.UUID
		mockErr       error
		wantStatus    int
		wantCode      string
		wantUCInput   *app.AddStockInput
		wantProductID string
	}{
		{
			name:          "success",
			body:          `{"product_id":"` + validID.String() + `","quantity":5}`,
			mockID:        validID,
			wantStatus:    http.StatusOK,
			wantUCInput:   &app.AddStockInput{ProductID: validID, Quantity: 5},
			wantProductID: validID.String(),
		},
		{
			name:          "success with zero quantity",
			body:          `{"product_id":"` + validID.String() + `","quantity":0}`,
			mockID:        validID,
			wantStatus:    http.StatusOK,
			wantUCInput:   &app.AddStockInput{ProductID: validID, Quantity: 0},
			wantProductID: validID.String(),
		},
		{
			name:       "invalid json",
			body:       `{"product_id":`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_json",
		},
		{
			name:       "unknown field",
			body:       `{"product_id":"` + validID.String() + `","quantity":5,"extra":1}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_json",
		},
		{
			name:       "empty body",
			body:       ``,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_json",
		},
		{
			name:       "invalid uuid",
			body:       `{"product_id":"not-a-uuid","quantity":5}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation_error",
		},
		{
			name:       "empty product_id",
			body:       `{"product_id":"","quantity":5}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation_error",
		},
		{
			name:       "invalid quantity domain error",
			body:       `{"product_id":"` + validID.String() + `","quantity":-1}`,
			mockID:     uuid.Nil,
			mockErr:    domain.ErrInvalidQuantity,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation_error",
		},
		{
			name:       "wrapped domain error",
			body:       `{"product_id":"` + validID.String() + `","quantity":-1}`,
			mockID:     uuid.Nil,
			mockErr:    errors.Join(errors.New("context"), domain.ErrInvalidQuantity),
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation_error",
		},
		{
			name:       "unexpected internal error",
			body:       `{"product_id":"` + validID.String() + `","quantity":5}`,
			mockID:     uuid.Nil,
			mockErr:    errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockAddStockUseCase{id: tt.mockID, err: tt.mockErr}
			h := newTestHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/stock/add", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			h.Add(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}

			var resp map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if tt.wantCode != "" {
				if resp["code"] != tt.wantCode {
					t.Errorf("code = %q, want %q", resp["code"], tt.wantCode)
				}
				if resp["message"] == "" {
					t.Errorf("message is empty, want non-empty")
				}
			}
			if tt.wantProductID != "" {
				if resp["product_id"] != tt.wantProductID {
					t.Errorf("product_id = %q, want %q", resp["product_id"], tt.wantProductID)
				}
			}

			if tt.wantUCInput != nil {
				if uc.callCnt != 1 {
					t.Fatalf("use case calls = %d, want 1", uc.callCnt)
				}
				if uc.gotIn.ProductID != tt.wantUCInput.ProductID {
					t.Errorf("input ProductID = %v, want %v", uc.gotIn.ProductID, tt.wantUCInput.ProductID)
				}
				if uc.gotIn.Quantity != tt.wantUCInput.Quantity {
					t.Errorf("input Quantity = %d, want %d", uc.gotIn.Quantity, tt.wantUCInput.Quantity)
				}
			}
		})
	}
}

// TestStockHandler_Add_DoesNotCallUseCaseOnValidationFailure — при ошибках
// декодирования/валидации use case не должен вызываться вовсе.
func TestStockHandler_Add_DoesNotCallUseCaseOnValidationFailure(t *testing.T) {
	bodies := map[string]string{
		"bad json":       `{`,
		"unknown field":  `{"product_id":"` + uuid.NewString() + `","quantity":5,"x":1}`,
		"invalid uuid":   `{"product_id":"nope","quantity":5}`,
		"empty product":  `{"product_id":"","quantity":5}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			uc := &mockAddStockUseCase{}
			h := newTestHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/stock/add", bytes.NewBufferString(body))
			rec := httptest.NewRecorder()

			h.Add(rec, req)

			if rec.Code < 400 || rec.Code >= 500 {
				t.Fatalf("status = %d, want 4xx", rec.Code)
			}
			if uc.callCnt != 0 {
				t.Errorf("use case calls = %d, want 0", uc.callCnt)
			}
		})
	}
}

// TestStockHandler_Add_PropagatesContext — проверяем, что request context
// прокидывается в use case (важно для таймаутов/отмены).
func TestStockHandler_Add_PropagatesContext(t *testing.T) {
	type ctxKey string
	const key ctxKey = "k"

	uc := &mockAddStockUseCase{id: uuid.New()}
	h := newTestHandler(uc)

	ctx := context.WithValue(context.Background(), key, "v")
	req := httptest.NewRequest(http.MethodPost, "/stock/add",
		bytes.NewBufferString(`{"product_id":"`+uuid.NewString()+`","quantity":1}`)).WithContext(ctx)
	rec := httptest.NewRecorder()

	// заменяем мок на проверяющий контекст
	var gotCtx context.Context
	uc2 := &ctxCheckingUseCase{id: uuid.New(), onExecute: func(c context.Context) { gotCtx = c }}
	h = newTestHandler(uc2)
	h.Add(rec, req)

	if gotCtx == nil {
		t.Fatal("context was not passed to use case")
	}
	if gotCtx.Value(key) != "v" {
		t.Errorf("context value = %v, want %v", gotCtx.Value(key), "v")
	}
}

type ctxCheckingUseCase struct {
	id        uuid.UUID
	onExecute func(context.Context)
}

func (c *ctxCheckingUseCase) Execute(ctx context.Context, _ app.AddStockInput) (uuid.UUID, error) {
	if c.onExecute != nil {
		c.onExecute(ctx)
	}
	return c.id, nil
}

// TestStockHandler_Add_MessageNotEmpty — для всех ошибок message не должен быть пустым.
func TestStockHandler_Add_MessageNotEmpty(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		mockErr  error
		wantCode string
	}{
		{"bad json", `{`, nil, "invalid_json"},
		{"bad uuid", `{"product_id":"x","quantity":1}`, nil, "validation_error"},
		{"domain", `{"product_id":"` + uuid.NewString() + `","quantity":1}`, domain.ErrInvalidQuantity, "validation_error"},
		{"internal", `{"product_id":"` + uuid.NewString() + `","quantity":1}`, errors.New("x"), "internal_error"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			uc := &mockAddStockUseCase{err: c.mockErr}
			h := newTestHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/stock/add", bytes.NewBufferString(c.body))
			rec := httptest.NewRecorder()
			h.Add(rec, req)

			var resp map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp["message"] == "" {
				t.Errorf("empty message for code %q", c.wantCode)
			}
		})
	}
}