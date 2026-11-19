package ckorderstore

import (
	"context"
	"errors"
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
	ID          string  `json:"id"`
	CompanyID   string  `json:"company_id"`
	Ticket      string  `json:"ticket"`
	ClientName  string  `json:"client_name"`
	Channel     string  `json:"channel"`
	Status      string  `json:"status"`
	Destination *string `json:"destination,omitempty"`
	Note        *string `json:"note,omitempty"`
	TotalValue  float64 `json:"total_value"`
	// ExternalID é o ID do pedido na plataforma de origem (ex. order.id do iFood).
	// Chave de idempotência estável — o Ticket é o número curto visível na comanda,
	// que pode repetir entre lojas/dias. Nulo em pedidos manuais/sem origem externa.
	ExternalID *string       `json:"external_id,omitempty"`
	CreatedAt  *time.Time    `json:"created_at,omitempty"`
	UpdatedAt  *time.Time    `json:"updated_at,omitempty"`
	Items      []CKOrderItem `json:"items,omitempty"`
}

type CreateCKOrderItemInput struct {
	Name      string  `json:"name"`
	ProdutoID *string `json:"produto_id"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Notes     *string `json:"notes"`
}

type CreateCKOrderInput struct {
	CompanyID   string                   `json:"company_id"`
	Ticket      string                   `json:"ticket"`
	ClientName  string                   `json:"client_name"`
	Channel     string                   `json:"channel"`
	Destination *string                  `json:"destination"`
	Note        *string                  `json:"note"`
	TotalValue  float64                  `json:"total_value"`
	ExternalID  *string                  `json:"external_id"` // ID na plataforma de origem (idempotência); nulo se manual
	Items       []CreateCKOrderItemInput `json:"items"`
}

// ErrTransicaoInvalida: o pedido não pode ir do status atual para o pedido.
// O gateway responde 409 a esse erro (errors.Is).
var ErrTransicaoInvalida = errors.New("transição de status inválida")

// statusAtivos são os status em que o pedido ainda está na cozinha. Entre eles
// o pedido anda para frente ou volta (operador arrastou para a coluna errada).
var statusAtivos = map[string]bool{"new": true, "preparing": true, "ready": true}

// TransicaoValida diz se um pedido pode ir de `de` para `para`.
//
//   - entre new, preparing e ready: livre, nos dois sentidos;
//   - de qualquer status ativo para delivered ou cancelled: permitido;
//   - delivered e cancelled são finais: nada sai deles. É isso que garante que
//     a baixa de estoque da entrega acontece uma vez só;
//   - mesmo status: permitido (no-op idempotente).
func TransicaoValida(de, para string) bool {
	if de == para {
		return statusValido(para)
	}
	if !statusAtivos[de] {
		return false
	}
	return statusAtivos[para] || para == "delivered" || para == "cancelled"
}

func statusValido(s string) bool {
	return statusAtivos[s] || s == "delivered" || s == "cancelled"
}

type UpdateCKOrderStatusInput struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type CKOrderFilters struct {
	CompanyID *string
	// CompanyIDs restringe às empresas listadas. nil = sem filtro;
	// slice vazio = nenhum registro (fail-closed).
	CompanyIDs []string
	Status     *string
	Channel    *string
	Page       int
	Limit      int
}

type CKOrderStore interface {
	t.StoreType

	Create(ctx context.Context, input *CreateCKOrderInput) (*CKOrder, error)
	GetByID(ctx context.Context, id string) (*CKOrder, error)
	GetByTicket(ctx context.Context, companyID, ticket string) (*CKOrder, error)
	// GetByExternalID busca por (empresa, ID da plataforma de origem) — a via de
	// idempotência de pedidos de marketplace. (nil,nil) se não existir.
	GetByExternalID(ctx context.Context, companyID, externalID string) (*CKOrder, error)
	// UpdateStatus muda o status numa transação: trava o pedido, valida a
	// transição (ErrTransicaoInvalida se não vale) e, ao entrar em delivered,
	// dá baixa no estoque explodindo a ficha técnica de cada item até os
	// insumos. A baixa é idempotente por (pedido, insumo).
	UpdateStatus(ctx context.Context, input *UpdateCKOrderStatusInput) (*CKOrder, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters *CKOrderFilters) (*t.PaginatedResult[CKOrder], error)
	ListItems(ctx context.Context, orderID string) ([]CKOrderItem, error)
}
