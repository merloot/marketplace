package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/merloot/marketplace/pkg/kafka"
	pgbase "github.com/merloot/marketplace/pkg/postgres"
	"github.com/merloot/marketplace/services/inventory/internal/app"
	"github.com/merloot/marketplace/services/inventory/internal/config"
	"github.com/merloot/marketplace/services/inventory/internal/consumer"
	"github.com/merloot/marketplace/services/inventory/internal/storage/postgres"
	"golang.org/x/sync/errgroup"
)

func main() {
	_ = godotenv.Load()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, log); err != nil {
		log.Error("consumer application failed", "err", err)
		os.Exit(1)
	}
	log.Info("consumer application stopped gracefully")
}

func run(ctx context.Context, log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dbConfig := pgbase.Config{URL: cfg.Database.URL, SearchPath: cfg.Database.SearchPath, MaxConns: cfg.Database.MaxConns, MinConns: cfg.Database.MinConns, MaxConnLifetime: cfg.Database.MaxConnLifetime, MaxConnIdleTime: cfg.Database.MaxConnIdleTime, ConnectionTimeout: cfg.Database.ConnectionTimeout}

	provider, err := postgres.NewProvider(ctx, dbConfig, pgbase.WithLogger(log), pgbase.WithPingRetries(5))
	if err != nil {
		return err
	}
	defer provider.Close()

	productStockRepo := postgres.NewProductStockRepo(provider)

	syncUC := app.NewSyncProductUseCase(productStockRepo, log)

	readerConfig := kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		Topic:          cfg.Kafka.Topic,
		GroupID:        cfg.Kafka.GroupID,
		CommitInterval: cfg.Kafka.CommitInterval,
	}
	reader, err := kafka.NewReader(readerConfig, log)
	if err != nil {
		log.Error("failed to create kafka reader", "error", err.Error())
		return err
	}
	defer reader.Close()

	consumer := consumer.NewProductConsumer(reader, syncUC, log)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return consumer.Run(gctx)
	})

	return g.Wait()
}
