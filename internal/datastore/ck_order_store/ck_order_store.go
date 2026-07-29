package ckorderstore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type CKOrderItem struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"order_id"`
	Name      string  `json:"name"`
	ProdutoID *string `json:"produto_id,omitempty"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Notes     *string `json:"notes,omitempty"`
}

type CKOrder struct {
	ID          string        `json:"id"`
	CompanyID   string        `json:"company_id"`
	Ticket      string        `json:"ticket"`
	ClientName  string        `json:"client_name"`
	Channel     string        `json:"channel"`
	Status      string        `json:"status"`
	Destination *string       `json:"destination,omitempty"`
	Note        *string       `json:"note,omitempty"`
	TotalValue  float64       `json:"total_value"`
	// ExternalID é o ID do pedido na plataforma de origem (ex. order.id do iFood).
	// Chave de idempotência estável — o Ticket é o número curto visível na comanda,
	// que pode repetir entre lojas/dias. Nulo em pedidos manuais/sem origem externa.
	ExternalID  *string       `json:"external_id,omitempty"`
	CreatedAt   *time.Time    `json:"created_at,omitempty"`
	UpdatedAt   *time.Time    `json:"updated_at,omitempty"`
	Items       []CKOrderItem `json:"items,omitempty"`
}

type CreateCKOrderItemInput struct {
	Name      string
	ProdutoID *string
	Quantity  float64
	UnitPrice float64
	Notes     *string
}

type CreateCKOrderInput struct {
	CompanyID   string
	Ticket      string
	ClientName  string
	Channel     string
	Destination *string
	Note        *string
	TotalValue  float64
	ExternalID  *string // ID na plataforma de origem (idempotência); nulo se manual
	Items       []CreateCKOrderItemInput
}

type UpdateCKOrderStatusInput struct {
	ID     string
	Status string
}

type CKOrderFilters struct {
	CompanyID *string
	Status    *string
	Channel   *string
	Page      int
	Limit     int
}

type CKOrderStore interface {
	t.StoreType

	Create(ctx context.Context, input *CreateCKOrderInput) (*CKOrder, error)
	GetByID(ctx context.Context, id string) (*CKOrder, error)
	GetByTicket(ctx context.Context, companyID, ticket string) (*CKOrder, error)
	// GetByExternalID busca por (empresa, ID da plataforma de origem) — a via de
	// idempotência de pedidos de marketplace. (nil,nil) se não existir.
	GetByExternalID(ctx context.Context, companyID, externalID string) (*CKOrder, error)
	UpdateStatus(ctx context.Context, input *UpdateCKOrderStatusInput) (*CKOrder, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters *CKOrderFilters) (*t.PaginatedResult[CKOrder], error)
	ListItems(ctx context.Context, orderID string) ([]CKOrderItem, error)
}
