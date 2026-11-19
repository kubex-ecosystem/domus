package ckcompanystore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type CKCompany struct {
	ID           string     `json:"id"`
	TenantID     *string    `json:"tenant_id,omitempty"`
	Name         string     `json:"name"`
	Logo         *string    `json:"logo,omitempty"`
	BrandColor   string     `json:"brand_color"`
	Cuisine      string     `json:"cuisine"`
	ActiveStatus string     `json:"active_status"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	Channels     []string   `json:"channels"`
}

type CreateCKCompanyInput struct {
	// TenantID é o tenant dono da empresa. Opcional por compatibilidade, mas
	// empresas sem tenant não aparecem em listagens filtradas por TenantIDs.
	TenantID     string   `json:"tenant_id"`
	Name         string   `json:"name"`
	Logo         *string  `json:"logo"`
	BrandColor   string   `json:"brand_color"`
	Cuisine      string   `json:"cuisine"`
	ActiveStatus string   `json:"active_status"`
	Channels     []string `json:"channels"`
}

type UpdateCKCompanyInput struct {
	ID           string   `json:"id"`
	Name         *string  `json:"name"`
	Logo         *string  `json:"logo"`
	BrandColor   *string  `json:"brand_color"`
	Cuisine      *string  `json:"cuisine"`
	ActiveStatus *string  `json:"active_status"`
	Channels     []string `json:"channels"`
}

type CKCompanyFilters struct {
	// TenantIDs restringe a listagem às empresas desses tenants. nil = sem
	// filtro (uso interno); slice vazio = nenhuma empresa (fail-closed).
	TenantIDs    []string
	Name         *string
	ActiveStatus *string
	Page         int
	Limit        int
}

type CKCompanyStore interface {
	t.StoreType

	Create(ctx context.Context, input *CreateCKCompanyInput) (*CKCompany, error)
	GetByID(ctx context.Context, id string) (*CKCompany, error)
	Update(ctx context.Context, input *UpdateCKCompanyInput) (*CKCompany, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters *CKCompanyFilters) (*t.PaginatedResult[CKCompany], error)
	SetChannels(ctx context.Context, companyID string, channels []string) error
	GetChannels(ctx context.Context, companyID string) ([]string, error)
}
