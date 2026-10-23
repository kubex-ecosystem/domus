package execution

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HTTPExecutor define a interface "de verdade" para trabalhar com Postgres
// via pgx/pgxpool, no modo nativo.
//
// Essa interface é o que os Stores vão usar para escrever SQL explícito
// sem conhecer pool/driver interno do DS.
type HTTPExecutor interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row

	BeginTx(ctx context.Context) (pgx.Tx, error)

	// Pool expõe o *pgxpool.Pool em último caso (ex: uso avançado).
	Pool() *pgxpool.Pool
}

// httpExecutor é a implementação padrão baseada em *pgxpool.Pool.
type httpExecutor struct {
	pool *pgxpool.Pool
}

// NewHTTPExecutor cria um HTTPExecutor a partir de um *pgxpool.Pool.
func NewHTTPExecutor(pool *pgxpool.Pool) HTTPExecutor {
	return &httpExecutor{pool: pool}
}

func (e *httpExecutor) Exec(ctx context.Context, q string, args ...any) (pgconn.CommandTag, error) {
	return e.pool.Exec(ctx, q, args...)
}

func (e *httpExecutor) Query(ctx context.Context, q string, args ...any) (pgx.Rows, error) {
	return e.pool.Query(ctx, q, args...)
}

func (e *httpExecutor) QueryRow(ctx context.Context, q string, args ...any) pgx.Row {
	return e.pool.QueryRow(ctx, q, args...)
}

func (e *httpExecutor) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return e.pool.Begin(ctx)
}

func (e *httpExecutor) Pool() *pgxpool.Pool {
	return e.pool
}
