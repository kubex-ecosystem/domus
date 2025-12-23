// Package pendingaccessstore fornece o store de domínio para solicitações de acesso pendentes.
package pendingaccessstore

import (
	"context"
	"encoding/json"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// PendingAccessStatus define o status da solicitação.
type PendingAccessStatus string

const (
	StatusPending  PendingAccessStatus = "pending"
	StatusApproved PendingAccessStatus = "approved"
	StatusRejected PendingAccessStatus = "rejected"
)

// PendingAccessRequest representa uma solicitação de acesso pendente.
type PendingAccessRequest struct {
	ID                 string              `json:"id"`
	Email              string              `json:"email"`
	Provider           string              `json:"provider"`
	ProviderUserID     *string             `json:"provider_user_id,omitempty"`
	Name               *string             `json:"name,omitempty"`
	AvatarURL          *string             `json:"avatar_url,omitempty"`
	Status             PendingAccessStatus `json:"status"`
	RequesterIP        *string             `json:"requester_ip,omitempty"`
	RequesterUserAgent *string             `json:"requester_user_agent,omitempty"`
	TenantID           *string             `json:"tenant_id,omitempty"`
	RoleCode           *string             `json:"role_code,omitempty"`
	Metadata           json.RawMessage     `json:"metadata,omitempty"`
	ReviewedBy         *string             `json:"reviewed_by,omitempty"`
	ReviewedAt         *time.Time          `json:"reviewed_at,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          *time.Time          `json:"updated_at,omitempty"`
}

// CreatePendingAccessRequestInput define dados para criação.
type CreatePendingAccessRequestInput struct {
	Email              string
	Provider           string
	ProviderUserID     *string
	Name               *string
	AvatarURL          *string
	Status             *PendingAccessStatus
	RequesterIP        *string
	RequesterUserAgent *string
	TenantID           *string
	RoleCode           *string
	Metadata           json.RawMessage
	ReviewedBy         *string
	ReviewedAt         *time.Time
}

// UpdatePendingAccessRequestInput define campos atualizáveis.
type UpdatePendingAccessRequestInput struct {
	ID         string
	Status     *PendingAccessStatus
	ReviewedBy *string
	ReviewedAt *time.Time
}

// PendingAccessFilters define filtros para listagem.
type PendingAccessFilters struct {
	Email    *string
	Provider *string
	Status   *PendingAccessStatus
	TenantID *string
	RoleCode *string
	Page     int
	Limit    int
}

// PendingAccessStore define operações de persistência.
type PendingAccessStore interface {
	t.StoreType

	// Create cria uma nova solicitação (ou atualiza se já existir).
	Create(ctx context.Context, input *CreatePendingAccessRequestInput) (*PendingAccessRequest, error)

	// GetByID busca solicitação por ID.
	GetByID(ctx context.Context, id string) (*PendingAccessRequest, error)

	// Update atualiza status e campos de revisão.
	Update(ctx context.Context, input *UpdatePendingAccessRequestInput) (*PendingAccessRequest, error)

	// List retorna solicitações paginadas com filtros.
	List(ctx context.Context, filters *PendingAccessFilters) (*t.PaginatedResult[PendingAccessRequest], error)

	// Count retorna total de solicitações com filtros.
	Count(ctx context.Context, filters *PendingAccessFilters) (int64, error)
}
