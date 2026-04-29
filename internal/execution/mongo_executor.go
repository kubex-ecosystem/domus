package execution

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MongoExecutor define a interface "de verdade" para trabalhar com Postgres
// via pgx/pgxpool, no modo nativo.
//
// Essa interface é o que os Stores vão usar para escrever SQL explícito
// sem conhecer pool/driver interno do DS.
type MongoExecutor interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row

	BeginTx(ctx context.Context) (pgx.Tx, error)

	// Pool expõe o *pgxpool.Pool em último caso (ex: uso avançado).
	Pool() *pgxpool.Pool
}

// mongoExecutor é a implementação padrão baseada em *pgxpool.Pool.
type mongoExecutor struct {
	pool *pgxpool.Pool
}

// NewMongoExecutor cria um MongoExecutor a partir de um *pgxpool.Pool.
func NewMongoExecutor(pool *pgxpool.Pool) MongoExecutor {
	return &mongoExecutor{pool: pool}
}

func (e *mongoExecutor) Exec(ctx context.Context, q string, args ...any) (pgconn.CommandTag, error) {
	return e.pool.Exec(ctx, q, args...)
}

func (e *mongoExecutor) Query(ctx context.Context, q string, args ...any) (pgx.Rows, error) {
	return e.pool.Query(ctx, q, args...)
}

func (e *mongoExecutor) QueryRow(ctx context.Context, q string, args ...any) pgx.Row {
	return e.pool.QueryRow(ctx, q, args...)
}

func (e *mongoExecutor) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return e.pool.Begin(ctx)
}

func (e *mongoExecutor) Pool() *pgxpool.Pool {
	return e.pool
}
