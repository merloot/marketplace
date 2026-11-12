package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func (r *Repo) MigrateUp(ctx context.Context, fsys fs.FS, dir string) error {
	return r.runGoose(ctx, fsys, dir, func(db *sql.DB) error {
		return goose.UpContext(ctx, db, dir)
	}, "up")
}

func (r *Repo) MigrateDown(ctx context.Context, fsys fs.FS, dir string) error {
	return r.runGoose(ctx, fsys, dir, func(db *sql.DB) error {
		return goose.DownContext(ctx, db, dir)
	}, "down")
}

func (r *Repo) MigrateStatus(ctx context.Context, fsys fs.FS, dir string) error {
	return r.runGoose(ctx, fsys, dir, func(db *sql.DB) error {
		return goose.StatusContext(ctx, db, dir)
	}, "status")
}

func (r *Repo) MigrateVersion(ctx context.Context, fsys fs.FS, dir string) error {
	return r.runGoose(ctx, fsys, dir, func(db *sql.DB) error {
		v, err := goose.GetDBVersionContext(ctx, db)
		if err != nil {
			return err
		}
		r.log.Info("Current migration version", "version", v)
		return nil
	}, "version")
}

func (r *Repo) MigrateUpTo(ctx context.Context, fsys fs.FS, dir string, version int64) error {
	return r.runGoose(ctx, fsys, dir, func(db *sql.DB) error {
		return goose.UpToContext(ctx, db, dir, version)
	}, "up-to")
}

func (r *Repo) runGoose(ctx context.Context, fsys fs.FS, dir string, fn func(*sql.DB) error, op string) error {
	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("Goose dialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(r.pool)
	defer db.Close()

	r.log.Info("Migrate start", "op", op, "dir", dir)
	if err := fn(db); err != nil {
		return fmt.Errorf("Goose %s: %w", op, err)
	}

	r.log.Info("Migrate done", "op", op)
	return nil
}
