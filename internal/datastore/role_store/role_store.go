// Package rolestore fornece o store de domínio para roles e permissões do RBAC.
package rolestore

import (
	"context"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// Role representa um papel no sistema RBAC.
type Role struct {
	ID              string `json:"id"`
	Code            string `json:"code"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description,omitempty"`
	IsSystemRole    bool   `json:"is_system_role"`
	PermissionCount int    `json:"permission_count,omitempty"`
}

// RoleStore define operações de leitura sobre roles do sistema.
// Roles são gerenciados pelo seed/migration — não há CRUD de criação via API.
type RoleStore interface {
	t.StoreType

	// GetByCode retorna um role pelo seu código único (ex: "admin", "viewer").
	GetByCode(ctx context.Context, code string) (*Role, error)

	// GetByID retorna um role pelo seu UUID.
	GetByID(ctx context.Context, id string) (*Role, error)

	// List retorna todos os roles disponíveis com contagem de permissões ativas.
	List(ctx context.Context) ([]Role, error)
}
