// Package companystore fornece o store de domínio para empresas.
package companystore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// Company representa a entidade de empresa conforme schema do banco.
type Company struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Slug          string     `json:"slug"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
	PlanExpiresAt *time.Time `json:"plan_expires_at,omitempty"`
	IsTrial       *bool      `json:"is_trial,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
	Domain        *string    `json:"domain,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	Address       *string    `json:"address,omitempty"`
}

// CreateCompanyInput define dados para criação de empresa.
type CreateCompanyInput struct {
	Name          string
	Slug          string
	IsTrial       *bool
	IsActive      *bool
	Domain        *string
	Phone         *string
	Address       *string
	PlanExpiresAt *time.Time
}

// UpdateCompanyInput define campos atualizáveis.
type UpdateCompanyInput struct {
	ID            string
	Name          *string
	Slug          *string
	IsTrial       *bool
	IsActive      *bool
	Domain        *string
	Phone         *string
	Address       *string
	PlanExpiresAt *time.Time
}

// CompanyFilters define filtros para listagem de empresas.
type CompanyFilters struct {
	Name     *string
	Slug     *string
	IsActive *bool
	Page     int
	Limit    int
}

// CompanyStore define operações específicas de persistência de empresas.
type CompanyStore interface {
	t.StoreType

	// Create cria uma nova empresa e retorna a entidade criada.
	Create(ctx context.Context, input *CreateCompanyInput) (*Company, error)

	// GetByID busca empresa por ID.
	// Retorna (nil, nil) se não encontrado.
	GetByID(ctx context.Context, id string) (*Company, error)

	// GetBySlug busca empresa por slug (case-insensitive).
	// Retorna (nil, nil) se não encontrado.
	GetBySlug(ctx context.Context, slug string) (*Company, error)

	// Update atualiza dados da empresa.
	// Apenas campos não-nil em UpdateCompanyInput são atualizados.
	Update(ctx context.Context, input *UpdateCompanyInput) (*Company, error)

	// Delete remove a empresa (hard delete).
	Delete(ctx context.Context, id string) error

	// List retorna empresas paginadas com filtros opcionais.
	List(ctx context.Context, filters *CompanyFilters) (*t.PaginatedResult[Company], error)

	// Count retorna o total de empresas (com filtros opcionais).
	Count(ctx context.Context, filters *CompanyFilters) (int64, error)
}
