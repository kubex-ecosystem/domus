package ckinsumostore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type CKInsumo struct {
	ID            string     `json:"id"`
	CompanyID     string     `json:"company_id"`
	Nome          string     `json:"nome"`
	Unidade       string     `json:"unidade"`
	Estoque       float64    `json:"estoque"`
	EstoqueMinimo float64    `json:"estoque_minimo"`
	CustoUnitario float64    `json:"custo_unitario"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type CreateCKInsumoInput struct {
	CompanyID     string
	Nome          string
	Unidade       string
	Estoque       float64
	EstoqueMinimo float64
	CustoUnitario float64
}

type UpdateCKInsumoInput struct {
	ID            string
	Nome          *string
	Unidade       *string
	Estoque       *float64
	EstoqueMinimo *float64
	CustoUnitario *float64
}

type CKInsumoFilters struct {
	CompanyID    *string
	BaixoEstoque *bool
	Page         int
	Limit        int
}

type CKInsumoStore interface {
	t.StoreType

	Create(ctx context.Context, input *CreateCKInsumoInput) (*CKInsumo, error)
	GetByID(ctx context.Context, id string) (*CKInsumo, error)
	Update(ctx context.Context, input *UpdateCKInsumoInput) (*CKInsumo, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters *CKInsumoFilters) (*t.PaginatedResult[CKInsumo], error)
	AjustarEstoque(ctx context.Context, id string, delta float64) (*CKInsumo, error)
}
