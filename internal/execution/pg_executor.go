package execution

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGExecutor define a interface "de verdade" para trabalhar com Postgres
// via pgx/pgxpool, no modo nativo.
//
// Essa interface é o que os Stores vão usar para escrever SQL explícito
// sem conhecer pool/driver interno do DS.
type PGExecutor interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row

	BeginTx(ctx context.Context) (pgx.Tx, error)

	// Pool expõe o *pgxpool.Pool em último caso (ex: uso avançado).
	Pool() *pgxpool.Pool
}

// pgExecutor é a implementação padrão baseada em *pgxpool.Pool.
type pgExecutor struct {
	pool *pgxpool.Pool
}

// NewPGExecutor cria um PGExecutor a partir de um *pgxpool.Pool.
func NewPGExecutor(pool *pgxpool.Pool) PGExecutor {
	return &pgExecutor{pool: pool}
}

func (e *pgExecutor) Exec(ctx context.Context, q string, args ...any) (pgconn.CommandTag, error) {
	return e.pool.Exec(ctx, q, args...)
}

func (e *pgExecutor) Query(ctx context.Context, q string, args ...any) (pgx.Rows, error) {
	return e.pool.Query(ctx, q, args...)
}

func (e *pgExecutor) QueryRow(ctx context.Context, q string, args ...any) pgx.Row {
	return e.pool.QueryRow(ctx, q, args...)
}

func (e *pgExecutor) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return e.pool.Begin(ctx)
}

func (e *pgExecutor) Pool() *pgxpool.Pool {
	return e.pool
}
