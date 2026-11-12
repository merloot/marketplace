package postgres

import (
	"context"

	pgbase "github.com/merloot/marketplace/pkg/postgres"
)

type Provider struct {
	repo *pgbase.Repo
	uow  *pgbase.UnitOfWork
}

type ExecutorProvider interface {
	Executor(ctx context.Context) pgbase.DBExecutor
}

func NewProvider(ctx context.Context, cfg pgbase.Config, opts ...pgbase.Options) (*Provider, error) {
	repo, err := pgbase.New(ctx, cfg, opts...)
	if err != nil {
		return nil, err
	}

	return &Provider{
		repo: repo,
		uow:  repo.UnitOfWork(),
	}, nil
}

func (p *Provider) UnitOfWork() *pgbase.UnitOfWork                 { return p.uow }
func (p *Provider) Executor(ctx context.Context) pgbase.DBExecutor { return p.repo.Executor(ctx) }
func (p *Provider) Close()                                         { p.repo.Close() }
func (p *Provider) WithTransaction(ctx context.Context, fn func(ctx context.Context, exec pgbase.DBExecutor) error) error {
	return p.repo.WithTransaction(ctx, fn)
}
