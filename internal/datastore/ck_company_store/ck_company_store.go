package ckcompanystore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type CKCompany struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Logo         *string    `json:"logo,omitempty"`
	BrandColor   string     `json:"brand_color"`
	Cuisine      string     `json:"cuisine"`
	ActiveStatus string     `json:"active_status"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	Channels     []string   `json:"channels"`
}

type CreateCKCompanyInput struct {
	Name         string
	Logo         *string
	BrandColor   string
	Cuisine      string
	ActiveStatus string
	Channels     []string
}

type UpdateCKCompanyInput struct {
	ID           string
	Name         *string
	Logo         *string
	BrandColor   *string
	Cuisine      *string
	ActiveStatus *string
	Channels     []string
}

type CKCompanyFilters struct {
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
