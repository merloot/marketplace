package postgres

import (
	"context"
	"embed"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const migrationsDir = "migrations"

func (p *Provider) MigrateUp(ctx context.Context) error {
	return p.repo.MigrateUp(ctx, migrationsFS, migrationsDir)
}

func (p *Provider) MigrateDown(ctx context.Context) error {
	return p.repo.MigrateDown(ctx, migrationsFS, migrationsDir)
}

func (p *Provider) MigrateStatus(ctx context.Context) error {
	return p.repo.MigrateStatus(ctx, migrationsFS, migrationsDir)
}

func (p *Provider) MigrateVersion(ctx context.Context) error {
	return p.repo.MigrateVersion(ctx, migrationsFS, migrationsDir)
}
