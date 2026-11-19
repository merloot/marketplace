package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/merloot/marketplace/services/inventory/internal/domain"
)

type mockProductStockRepo struct {
	err     error
	gotID   uuid.UUID
	gotQty  int
	callCnt int
}

func (m *mockProductStockRepo) AddStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	m.callCnt++
	m.gotID = productID
	m.gotQty = quantity
	return m.err
}

func (m *mockProductStockRepo) GetById(ctx context.Context, id uuid.UUID) (domain.ProductStock, error) {
	return domain.ProductStock{}, nil
}

func (m *mockProductStockRepo) TryReserve(ctx context.Context, productID uuid.UUID, quantity int) (bool, error) {
	return false, nil
}

func (m *mockProductStockRepo) Save(ctx context.Context, stock domain.ProductStock) error {
	return nil
}

func intPtr(v int) *int { return &v }

func TestAddStockUseCase_Execute(t *testing.T) {
	validID := uuid.New()
	repoSentinel := errors.New("db is down")

	tests := []struct {
		name        string
		input       AddStockInput
		repoErr     error
		wantID      uuid.UUID
		wantErr     error
		wantCall    int
		wantRepoID  *uuid.UUID
		wantRepoQty *int
	}{
		{
			name:        "success",
			input:       AddStockInput{ProductID: validID, Quantity: 5},
			wantID:      validID,
			wantCall:    1,
			wantRepoID:  &validID,
			wantRepoQty: intPtr(5),
		},
		{
			name:        "success with zero quantity",
			input:       AddStockInput{ProductID: validID, Quantity: 0},
			wantID:      validID,
			wantCall:    1,
			wantRepoID:  &validID,
			wantRepoQty: intPtr(0),
		},
		{
			name:     "nil product id — domain.ErrInvalidProductID",
			input:    AddStockInput{ProductID: uuid.Nil, Quantity: 1},
			wantErr:  domain.ErrInvalidProduct,
			wantCall: 0,
		},
		{
			name:     "negative quantity — domain.ErrInvalidQuantity",
			input:    AddStockInput{ProductID: validID, Quantity: -1},
			wantErr:  domain.ErrInvalidQuantity,
			wantCall: 0,
		},
		{
			name:     "both invalid — product id проверяется первым",
			input:    AddStockInput{ProductID: uuid.Nil, Quantity: -1},
			wantErr:  domain.ErrInvalidProduct,
			wantCall: 0,
		},
		{
			name:        "repo error — пробрасывается, id = uuid.Nil",
			input:       AddStockInput{ProductID: validID, Quantity: 3},
			repoErr:     repoSentinel,
			wantErr:     repoSentinel,
			wantID:      uuid.Nil,
			wantCall:    1,
			wantRepoID:  &validID,
			wantRepoQty: intPtr(3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductStockRepo{err: tt.repoErr}
			uc := NewAddStockUseCase(repo)

			gotID, err := uc.Execute(context.Background(), tt.input)

			switch {
			case tt.wantErr == nil && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.wantErr != nil && err == nil:
				t.Fatalf("expected error %v, got nil", tt.wantErr)
			case tt.wantErr != nil && !errors.Is(err, tt.wantErr):
				t.Errorf("error = %v, want errors.Is(_, %v) = true", err, tt.wantErr)
			}

			if gotID != tt.wantID {
				t.Errorf("returned id = %v, want %v", gotID, tt.wantID)
			}
			if repo.callCnt != tt.wantCall {
				t.Fatalf("repo calls = %d, want %d", repo.callCnt, tt.wantCall)
			}
			if tt.wantRepoID != nil && repo.gotID != *tt.wantRepoID {
				t.Errorf("repo got ProductID = %v, want %v", repo.gotID, *tt.wantRepoID)
			}
			if tt.wantRepoQty != nil && repo.gotQty != *tt.wantRepoQty {
				t.Errorf("repo got Quantity = %d, want %d", repo.gotQty, *tt.wantRepoQty)
			}
		})
	}
}

// Явный регрессионный тест: uuid.Nil не должен доходить до репозитория.
func TestAddStockUseCase_Execute_NilProductIDDoesNotCallRepo(t *testing.T) {
	repo := &mockProductStockRepo{}
	uc := NewAddStockUseCase(repo)

	id, err := uc.Execute(context.Background(), AddStockInput{
		ProductID: uuid.Nil,
		Quantity:  1,
	})

	if !errors.Is(err, domain.ErrInvalidProduct) {
		t.Fatalf("error = %v, want domain.ErrInvalidProductID", err)
	}
	if id != uuid.Nil {
		t.Errorf("id = %v, want uuid.Nil", id)
	}
	if repo.callCnt != 0 {
		t.Errorf("repo calls = %d, want 0", repo.callCnt)
	}
}

// Явный регрессионный тест: отрицательное количество не должно доходить до репозитория.
func TestAddStockUseCase_Execute_NegativeQuantityDoesNotCallRepo(t *testing.T) {
	repo := &mockProductStockRepo{}
	uc := NewAddStockUseCase(repo)

	_, err := uc.Execute(context.Background(), AddStockInput{
		ProductID: uuid.New(),
		Quantity:  -1,
	})

	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Fatalf("error = %v, want domain.ErrInvalidQuantity", err)
	}
	if repo.callCnt != 0 {
		t.Errorf("repo calls = %d, want 0", repo.callCnt)
	}
}

// Ошибка репозитория не должна проглатываться.
func TestAddStockUseCase_Execute_RepoErrorNotSwallowed(t *testing.T) {
	sentinel := errors.New("sentinel")
	repo := &mockProductStockRepo{err: sentinel}
	uc := NewAddStockUseCase(repo)

	id, err := uc.Execute(context.Background(), AddStockInput{
		ProductID: uuid.New(),
		Quantity:  1,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("error = %v, want errors.Is(err, sentinel)", err)
	}
	if id != uuid.Nil {
		t.Errorf("id = %v, want uuid.Nil on error", id)
	}
}

// Context должен прокидываться в репозиторий.
func TestAddStockUseCase_Execute_PropagatesContext(t *testing.T) {
	type ctxKey string
	const key ctxKey = "k"

	repo := &ctxCheckingRepo{}
	uc := NewAddStockUseCase(repo)

	ctx := context.WithValue(context.Background(), key, "v")
	if _, err := uc.Execute(ctx, AddStockInput{ProductID: uuid.New(), Quantity: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.gotCtx == nil {
		t.Fatal("context was not passed to repo")
	}
	if repo.gotCtx.Value(key) != "v" {
		t.Errorf("context value = %v, want %v", repo.gotCtx.Value(key), "v")
	}
}

type ctxCheckingRepo struct{ gotCtx context.Context }

func (c *ctxCheckingRepo) AddStock(ctx context.Context, _ uuid.UUID, _ int) error {
	c.gotCtx = ctx
	return nil
}

func (ctxCheckingRepo) TryReserve(context.Context, uuid.UUID, int) (bool, error) {
	return false, nil
}

func (ctxCheckingRepo) Save(context.Context, domain.ProductStock) error {
	return nil
}

func (ctxCheckingRepo) GetById(context.Context, uuid.UUID) (domain.ProductStock, error) {
	return domain.ProductStock{}, nil
}
