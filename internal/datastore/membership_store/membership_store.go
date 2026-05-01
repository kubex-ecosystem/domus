// Package membershipstore fornece o store de domínio para tenant_membership (RBAC multi-tenant).
package membershipstore

import (
	"context"
	"time"

	t "github.com/faelmori/domus-pulse/internal/types"
)

// MemberWithPermissions representa um membro de um tenant com role e permissões expandidas.
// Corresponde ao resultado da query de JOIN usada na API de gerenciamento de acesso.
type MemberWithPermissions struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	LastName    string    `json:"last_name"`
	Status      string    `json:"status"`
	TenantID    string    `json:"tenant_id"`
	RoleID      string    `json:"role_id"`
	RoleCode    string    `json:"role_code"`
	RoleName    string    `json:"role_name"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	Permissions []string  `json:"permissions"`
}

// UpsertMembershipInput define os dados para criar ou atualizar uma membership.
type UpsertMembershipInput struct {
	UserID   string
	TenantID string
	RoleID   string // UUID do role já resolvido
	IsActive bool
}

// UpdateRoleInput define os dados para atualizar o role de um membro existente.
type UpdateRoleInput struct {
	UserID   string
	TenantID string
	RoleID   string // UUID do role já resolvido
}

// MembershipStore define operações de persistência de tenant_membership.
type MembershipStore interface {
	t.StoreType

	// ListByTenantWithPermissions retorna todos os membros de um tenant com roles e permissões.
	// Executa o JOIN completo: tenant_membership → user → role → role_permission → permission.
	ListByTenantWithPermissions(ctx context.Context, tenantID string) ([]MemberWithPermissions, error)

	// Upsert cria ou atualiza uma membership (ON CONFLICT user_id + tenant_id).
	Upsert(ctx context.Context, input *UpsertMembershipInput) error

	// UpdateRole atualiza o role de um membro existente.
	UpdateRole(ctx context.Context, input *UpdateRoleInput) error
}
