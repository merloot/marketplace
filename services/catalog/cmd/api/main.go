package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	pgbase "github.com/merloot/marketplace/pkg/postgres"
	"github.com/merloot/marketplace/services/catalog/internal/app"
	"github.com/merloot/marketplace/services/catalog/internal/config"
	"github.com/merloot/marketplace/services/catalog/internal/httpapi"
	"github.com/merloot/marketplace/services/catalog/internal/storage/postgres"
	"golang.org/x/sync/errgroup"
)

func main() {
	_ = godotenv.Load()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbConfig := pgbase.Config{URL: cfg.Database.URL, SearchPath: cfg.Database.SearchPath, MaxConns: cfg.Database.MaxConns, MinConns: cfg.Database.MinConns, MaxConnLifetime: cfg.Database.MaxConnLifetime, MaxConnIdleTime: cfg.Database.MaxConnIdleTime, ConnectionTimeout: cfg.Database.ConnectionTimeout}
	provider, err := postgres.NewProvider(ctx, dbConfig, pgbase.WithLogger(log), pgbase.WithPingRetries(5))
	if err != nil {
		return err
	}
	defer provider.Close()

	productRepo := postgres.NewProductRepo(provider)
	outboxRepo := postgres.NewOutboxRepo(provider)
	uow := provider.UnitOfWork()

	useCase := app.NewCreateProductUseCase(uow, productRepo, outboxRepo)
	handler := httpapi.NewProductHandler(useCase)

	router := httpapi.NewRouter(handler)
	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})

	return g.Wait()
}
