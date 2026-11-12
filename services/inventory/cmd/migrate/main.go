package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	pgbase "github.com/merloot/marketplace/pkg/postgres"

	"github.com/merloot/marketplace/services/inventory/internal/config"
	"github.com/merloot/marketplace/services/inventory/internal/storage/postgres"
)

func main() {
	_ = godotenv.Load()

	if err := run(); err != nil {
		slog.Error("Migrate failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) == 0 {
		return fmt.Errorf("Usage: migrate <up|down|status|version up-to> [args]")
	}
	command := args[0]

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("Config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dbConfig := pgbase.Config{URL: cfg.Database.URL, SearchPath: cfg.Database.SearchPath, MaxConns: cfg.Database.MaxConns, MinConns: cfg.Database.MinConns, MaxConnLifetime: cfg.Database.MaxConnLifetime, MaxConnIdleTime: cfg.Database.MaxConnIdleTime, ConnectionTimeout: cfg.Database.ConnectionTimeout}

	provider, err := postgres.NewProvider(ctx, dbConfig, pgbase.WithLogger(log), pgbase.WithPingRetries(5))
	if err != nil {
		return fmt.Errorf("Postgres: %w", err)
	}
	defer provider.Close()

	switch command {
	case "up":
		return provider.MigrateUp(ctx)
	case "down":
		return provider.MigrateDown(ctx)
	case "status":
		return provider.MigrateStatus(ctx)
	case "version":
		return provider.MigrateVersion(ctx)
	default:
		return fmt.Errorf("Unknown command: %q", command)
	}
}
