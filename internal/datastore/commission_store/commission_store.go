// Package commissionstore fornece o store de comissões do módulo financeiro (Colonial NetImóveis).
package commissionstore

import (
	"context"
	"time"

	t "github.com/faelmori/domus-pulse/internal/types"
)

// Commission representa uma comissão registrada para um imóvel negociado.
type Commission struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	UserID     string     `json:"user_id"`
	PropertyID *string    `json:"property_id,omitempty"`
	GrossValue int64      `json:"gross_value"` // centavos
	Pct        float64    `json:"pct"`         // percentual ex: 6.00
	NetValue   int64      `json:"net_value"`   // centavos (coluna gerada no DB)
	Status     string     `json:"status"`      // pending|received|overdue
	Notes      string     `json:"notes"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// CreateCommissionInput define os dados para criar uma comissão.
type CreateCommissionInput struct {
	TenantID   string
	UserID     string
	PropertyID *string
	GrossValue int64
	Pct        float64
	Status     string
	Notes      string
	PaidAt     *time.Time
}

// CommissionFilters define filtros para listagem de comissões.
type CommissionFilters struct {
	TenantID   string
	UserID     *string
	PropertyID *string
	Status     *string
	From       *time.Time
	To         *time.Time
	Page       int
	Limit      int
}

// CommissionStore define operações de persistência do módulo financeiro.
type CommissionStore interface {
	t.StoreType

	// Create registra uma nova comissão.
	Create(ctx context.Context, input *CreateCommissionInput) (*Commission, error)

	// GetByID retorna uma comissão pelo ID, dentro do tenant.
	GetByID(ctx context.Context, tenantID, id string) (*Commission, error)

	// List retorna comissões paginadas com filtros opcionais.
	List(ctx context.Context, filters *CommissionFilters) (*t.PaginatedResult[Commission], error)

	// UpdateStatus atualiza o status de uma comissão e seta paid_at se status = 'received'.
	UpdateStatus(ctx context.Context, tenantID, id, status string) error

	// SumByPeriod retorna a soma de net_value (centavos) para um tenant/usuário em um período.
	// Usa paid_at quando status = 'received', caso contrário created_at.
	SumByPeriod(ctx context.Context, tenantID, userID string, from, to time.Time) (int64, error)
}
