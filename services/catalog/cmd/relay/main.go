package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	pgbase "github.com/merloot/marketplace/pkg/postgres"
	"github.com/merloot/marketplace/services/catalog/internal/config"
	"github.com/merloot/marketplace/services/catalog/internal/relay"
	"github.com/merloot/marketplace/services/catalog/internal/storage/postgres"
	"github.com/segmentio/kafka-go"
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
		log.Error("relay application failed", "err", err)
		os.Exit(1)
	}
	log.Info("relay application stopped gracefully")
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

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Kafka.Brokers...),
		Topic:        cfg.Kafka.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	defer writer.Close()

	outboxRepo := postgres.NewOutboxRepo(provider)
	relayWorker := relay.New(provider, outboxRepo, writer, log)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Info("starting outbox relay worker")
		return relayWorker.Run(gctx)
	})

	<-gctx.Done()
	log.Info("shutting down relay, waiting for current batch to finish...")

	if err := g.Wait(); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)
	return nil
}
