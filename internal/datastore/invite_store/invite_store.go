// Package invitestore fornece o store de domínio para convites (invitations).
package invitestore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// InvitationType define o tipo de convite.
type InvitationType string

const (
	TypePartner  InvitationType = "partner"
	TypeInternal InvitationType = "internal"
	TypeDemo     InvitationType = "demo"
	TypeTrial    InvitationType = "trial_partner"
	TypeEmployee InvitationType = "employee"
	TypeGuest    InvitationType = "guest"
)

// InvitationStatus define o status do convite.
type InvitationStatus string

const (
	StatusPending  InvitationStatus = "pending"
	StatusAccepted InvitationStatus = "accepted"
	StatusRevoked  InvitationStatus = "revoked"
	StatusExpired  InvitationStatus = "expired"
)

// Invitation representa um convite unificado (partner ou internal).
type Invitation struct {
	ID         string           `json:"id"`
	Type       InvitationType   `json:"type"`
	Name       string           `json:"name"`
	Email      string           `json:"email"`
	Role       string           `json:"role"`
	Token      string           `json:"token"`
	TenantID   string           `json:"tenant_id"`
	TeamID     *string          `json:"team_id,omitempty"`
	InvitedBy  string           `json:"invited_by"`
	Status     InvitationStatus `json:"status"`
	ExpiresAt  time.Time        `json:"expires_at"`
	AcceptedAt *time.Time       `json:"accepted_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  *time.Time       `json:"updated_at,omitempty"`
}

// CreateInvitationInput define dados para criação de convite.
type CreateInvitationInput struct {
	Type      InvitationType
	Name      string
	Email     string
	Role      string
	Token     string
	TenantID  string
	TeamID    *string
	InvitedBy string
	ExpiresAt *time.Time // Se nil, usa TTL padrão
}

// UpdateInvitationInput define campos atualizáveis.
type UpdateInvitationInput struct {
	ID         string
	Type       InvitationType
	Status     *InvitationStatus
	AcceptedAt *time.Time
	ExpiresAt  *time.Time
}

// InvitationFilters define filtros para listagem.
type InvitationFilters struct {
	Type      *InvitationType
	Email     *string
	TenantID  *string
	TeamID    *string
	Status    *InvitationStatus
	InvitedBy *string
	Page      int
	Limit     int
}

// InviteStore define operações de persistência para convites.
type InviteStore interface {
	t.StoreType

	// Create cria um novo convite (partner ou internal).
	Create(ctx context.Context, input *CreateInvitationInput) (*Invitation, error)

	// GetByID busca convite por ID e tipo.
	// Retorna (nil, nil) se não encontrado.
	GetByID(ctx context.Context, id string, invType InvitationType) (*Invitation, error)

	// GetByToken busca convite por token em ambas as tabelas.
	// Retorna (nil, nil) se não encontrado.
	GetByToken(ctx context.Context, token string) (*Invitation, error)

	// Update atualiza campos do convite.
	// Apenas campos não-nil em UpdateInvitationInput são atualizados.
	Update(ctx context.Context, input *UpdateInvitationInput) (*Invitation, error)

	// Accept marca o convite como aceito usando token.
	// Usa transação para garantir atomicidade (FOR UPDATE).
	Accept(ctx context.Context, token string) (*Invitation, error)

	// Revoke altera o status para revoked.
	Revoke(ctx context.Context, id string, invType InvitationType) error

	// Delete remove o convite (hard delete).
	Delete(ctx context.Context, id string, invType InvitationType) error

	// List retorna convites paginados com filtros.
	// Type é obrigatório em filters.
	List(ctx context.Context, filters *InvitationFilters) (*t.PaginatedResult[Invitation], error)

	// Count retorna total de convites com filtros.
	Count(ctx context.Context, filters *InvitationFilters) (int64, error)
}
