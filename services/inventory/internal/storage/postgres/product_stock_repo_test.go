package postgres_test

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	pgbase "github.com/merloot/marketplace/pkg/postgres"
	"github.com/merloot/marketplace/services/inventory/internal/domain"
	repo "github.com/merloot/marketplace/services/inventory/internal/storage/postgres"
)

// setupTestDB поднимает Postgres в Docker и применяет миграции
func setupTestDB(t *testing.T) (*pgbase.Repo, func()) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("inventory_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second)),
	)
	require.NoError(t, err, "failed to start postgres container")

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	dbConfig := pgbase.Config{
		URL:      connStr,
		MaxConns: 5,
		MinConns: 1,
		ConnectionTimeout: 10 * time.Second,
	}
	db, err := pgbase.New(ctx, dbConfig, pgbase.WithLogger(log))
	require.NoError(t, err)

	_, err = db.Executor(ctx).Exec(ctx, `
		CREATE TABLE product_stocks (
			product_id UUID PRIMARY KEY,
			sku VARCHAR(255) UNIQUE NOT NULL,
			quantity INT NOT NULL DEFAULT 0,
			reserved INT NOT NULL DEFAULT 0
		);
	`)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

func TestProductStockRepo_Save(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	r := repo.NewProductStockRepo(db)

	productID := uuid.New()
	sku, _ := domain.NewSKU("SKU-TEST-1")
	stock, _ := domain.NewProductStock(productID, sku)

	t.Run("successful save", func(t *testing.T) {
		err := r.Save(ctx, *stock)
		assert.NoError(t, err)
	})

	t.Run("duplicate product ID should return ErrDuplicateSKU", func(t *testing.T) {
		err := r.Save(ctx, *stock)
		assert.ErrorIs(t, err, domain.ErrDuplicateSKU)
	})
}

func TestProductStockRepo_TryReserve(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	r := repo.NewProductStockRepo(db)

	productID := uuid.New()
	sku, _ := domain.NewSKU("SKU-RESERVE-1")

	_, err := db.Executor(ctx).Exec(ctx,
		"INSERT INTO product_stocks (product_id, sku, quantity, reserved) VALUES ($1, $2, $3, $4)",
		productID, sku.String(), 10, 0)
	require.NoError(t, err)

	t.Run("successful reserve", func(t *testing.T) {
		ok, err := r.TryReserve(ctx, productID, 5)
		assert.NoError(t, err)
		assert.True(t, ok)

		var reserved int
		err = db.Executor(ctx).QueryRow(ctx, "SELECT reserved FROM product_stocks WHERE product_id = $1", productID).Scan(&reserved)
		assert.NoError(t, err)
		assert.Equal(t, 5, reserved)
	})

	t.Run("insufficient stock", func(t *testing.T) {
		ok, err := r.TryReserve(ctx, productID, 6)
		assert.NoError(t, err)
		assert.False(t, ok)

		var reserved int
		err = db.Executor(ctx).QueryRow(ctx, "SELECT reserved FROM product_stocks WHERE product_id = $1", productID).Scan(&reserved)
		assert.NoError(t, err)
		assert.Equal(t, 5, reserved)
	})

	t.Run("concurrent reserve on last item", func(t *testing.T) {
		_, err := db.Executor(ctx).Exec(ctx,
			"UPDATE product_stocks SET quantity = 1, reserved = 0 WHERE product_id = $1", productID)
		require.NoError(t, err)

		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ok, err := r.TryReserve(ctx, productID, 1)
				if err == nil && ok {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		assert.Equal(t, 1, successCount, "only one concurrent reserve should succeed")

		var reserved int
		err = db.Executor(ctx).QueryRow(ctx, "SELECT reserved FROM product_stocks WHERE product_id = $1", productID).Scan(&reserved)
		assert.NoError(t, err)
		assert.Equal(t, 1, reserved)
	})
}

func TestProductStockRepo_AddStock(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	r := repo.NewProductStockRepo(db)

	productID := uuid.New()
	sku, _ := domain.NewSKU("SKU-ADD-1")
	_, err := db.Executor(ctx).Exec(ctx,
		"INSERT INTO product_stocks (product_id, sku, quantity, reserved) VALUES ($1, $2, $3, $4)",
		productID, sku.String(), 5, 0)
	require.NoError(t, err)

	t.Run("successful add", func(t *testing.T) {
		err := r.AddStock(ctx, productID, 10)
		assert.NoError(t, err)

		var quantity int
		err = db.Executor(ctx).QueryRow(ctx, "SELECT quantity FROM product_stocks WHERE product_id = $1", productID).Scan(&quantity)
		assert.NoError(t, err)
		assert.Equal(t, 15, quantity)
	})

	t.Run("invalid quantity", func(t *testing.T) {
		err := r.AddStock(ctx, productID, -5)
		assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
	})

	t.Run("product not found", func(t *testing.T) {
		err := r.AddStock(ctx, uuid.New(), 5)
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})
}

func TestProductStockRepo_GetById(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	r := repo.NewProductStockRepo(db)

	productID := uuid.New()
	sku, _ := domain.NewSKU("SKU-GET-1")
	_, err := db.Executor(ctx).Exec(ctx,
		"INSERT INTO product_stocks (product_id, sku, quantity, reserved) VALUES ($1, $2, $3, $4)",
		productID, sku.String(), 7, 2)
	require.NoError(t, err)

	t.Run("successful get", func(t *testing.T) {
		stock, err := r.GetById(ctx, db.Executor(ctx), productID)
		assert.NoError(t, err)
		assert.Equal(t, productID, stock.ProductID())
		assert.Equal(t, "SKU-GET-1", stock.SKU().String())
		assert.Equal(t, 7, stock.Quantity())
		assert.Equal(t, 2, stock.Reserved())
	})

	t.Run("not found", func(t *testing.T) {
		_, err := r.GetById(ctx, db.Executor(ctx), uuid.New())
		assert.ErrorIs(t, err, domain.ErrProductNotFound) // pgx.ErrNoRows
	})
}
