package ckprodutostore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type CKComponente struct {
	ID           string  `json:"id"`
	ProdutoID    string  `json:"produto_id"`
	RefTipo      string  `json:"ref_tipo"`
	InsumoRefID  *string `json:"insumo_ref_id,omitempty"`
	ProdutoRefID *string `json:"produto_ref_id,omitempty"`
	Quantidade   float64 `json:"quantidade"`
}

type CKProduto struct {
	ID          string         `json:"id"`
	CompanyID   string         `json:"company_id"`
	Nome        string         `json:"nome"`
	Tipo        string         `json:"tipo"`
	PrecoVenda  *float64       `json:"preco_venda,omitempty"`
	CreatedAt   *time.Time     `json:"created_at,omitempty"`
	UpdatedAt   *time.Time     `json:"updated_at,omitempty"`
	Componentes []CKComponente `json:"componentes,omitempty"`
}

type CreateCKComponenteInput struct {
	RefTipo      string  `json:"ref_tipo"`
	InsumoRefID  *string `json:"insumo_ref_id"`
	ProdutoRefID *string `json:"produto_ref_id"`
	Quantidade   float64 `json:"quantidade"`
}

type CreateCKProdutoInput struct {
	CompanyID   string                    `json:"company_id"`
	Nome        string                    `json:"nome"`
	Tipo        string                    `json:"tipo"`
	PrecoVenda  *float64                  `json:"preco_venda"`
	Componentes []CreateCKComponenteInput `json:"componentes"`
}

type UpdateCKProdutoInput struct {
	ID          string                    `json:"id"`
	Nome        *string                   `json:"nome"`
	Tipo        *string                   `json:"tipo"`
	PrecoVenda  *float64                  `json:"preco_venda"`
	Componentes []CreateCKComponenteInput `json:"componentes"`
}

type CKProdutoFilters struct {
	CompanyID *string
	// CompanyIDs restringe às empresas listadas. nil = sem filtro;
	// slice vazio = nenhum registro (fail-closed).
	CompanyIDs []string
	Tipo       *string
	Page       int
	Limit      int
}

type CKProdutoStore interface {
	t.StoreType

	Create(ctx context.Context, input *CreateCKProdutoInput) (*CKProduto, error)
	GetByID(ctx context.Context, id string) (*CKProduto, error)
	Update(ctx context.Context, input *UpdateCKProdutoInput) (*CKProduto, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters *CKProdutoFilters) (*t.PaginatedResult[CKProduto], error)
	SetComponentes(ctx context.Context, produtoID string, componentes []CreateCKComponenteInput) error
	ListComponentes(ctx context.Context, produtoID string) ([]CKComponente, error)
}
