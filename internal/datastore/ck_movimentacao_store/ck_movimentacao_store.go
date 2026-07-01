package ckmovimentacaostore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type CKMovimentacao struct {
	ID        string     `json:"id"`
	InsumoID  string     `json:"insumo_id"`
	OrderID   *string    `json:"order_id,omitempty"`
	Tipo      string     `json:"tipo"`
	Quantidade float64   `json:"quantidade"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type RegistrarMovimentacaoInput struct {
	InsumoID  string
	OrderID   *string
	Tipo      string
	Quantidade float64
}

type CKMovimentacaoStore interface {
	t.StoreType

	// Registrar persiste uma movimentação. Para tipo='baixa', é idempotente por (order_id, insumo_id).
	Registrar(ctx context.Context, input *RegistrarMovimentacaoInput) (*CKMovimentacao, error)
	ListByInsumo(ctx context.Context, insumoID string, limit int) ([]CKMovimentacao, error)
	ListByOrder(ctx context.Context, orderID string) ([]CKMovimentacao, error)
}

// Garantia de compilação: CKMovimentacaoStore embeds StoreType
var _ t.StoreType = (*struct{ CKMovimentacaoStore })(nil)
