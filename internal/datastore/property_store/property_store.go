// Package propertystore fornece o store de domínio para imóveis (Colonial NetImóveis).
package propertystore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// PropertyImage representa uma imagem associada a um imóvel.
type PropertyImage struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Property representa a entidade de imóvel conforme schema do banco.
type Property struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	Type         string          `json:"type"`
	Transaction  string          `json:"transaction"`
	Neighborhood string          `json:"neighborhood"`
	Address      string          `json:"address"`
	Price        int64           `json:"price"`
	Area         int             `json:"area"`
	Bedrooms     int             `json:"bedrooms"`
	Suites       int             `json:"suites"`
	Bathrooms    int             `json:"bathrooms"`
	Parking      int             `json:"parking"`
	Description  string          `json:"description"`
	Highlights   string          `json:"highlights"`
	ContactPhone string          `json:"contact_phone"`
	Images       []PropertyImage `json:"images"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CreatePropertyInput define dados para criação de imóvel.
type CreatePropertyInput struct {
	TenantID     string
	Type         string
	Transaction  string
	Neighborhood string
	Address      string
	Price        int64
	Area         int
	Bedrooms     int
	Suites       int
	Bathrooms    int
	Parking      int
	Description  string
	Highlights   string
	ContactPhone string
	Images       []PropertyImage
}

// UpdatePropertyInput define campos atualizáveis.
// Apenas campos não-nil são aplicados no UPDATE.
type UpdatePropertyInput struct {
	ID           string
	TenantID     string
	Type         *string
	Transaction  *string
	Neighborhood *string
	Address      *string
	Price        *int64
	Area         *int
	Bedrooms     *int
	Suites       *int
	Bathrooms    *int
	Parking      *int
	Description  *string
	Highlights   *string
	ContactPhone *string
}

// PropertyFilters define filtros para listagem de imóveis.
type PropertyFilters struct {
	TenantID     string
	Transaction  *string
	Neighborhood *string
	Type         *string
	Page         int
	Limit        int
}

// PropertyStore define operações específicas de persistência de imóveis.
type PropertyStore interface {
	t.StoreType

	// Create cria um novo imóvel e retorna a entidade completa.
	Create(ctx context.Context, input *CreatePropertyInput) (*Property, error)

	// GetByID busca imóvel por ID e tenant.
	// Retorna (nil, nil) se não encontrado.
	GetByID(ctx context.Context, tenantID, id string) (*Property, error)

	// Update atualiza dados do imóvel.
	// Apenas campos não-nil em UpdatePropertyInput são atualizados.
	Update(ctx context.Context, input *UpdatePropertyInput) (*Property, error)

	// Delete remove o imóvel (hard delete), validando tenant.
	Delete(ctx context.Context, tenantID, id string) error

	// List retorna imóveis paginados com filtros opcionais.
	List(ctx context.Context, filters *PropertyFilters) (*t.PaginatedResult[Property], error)

	// AddImages adiciona imagens ao imóvel e retorna a lista atualizada.
	AddImages(ctx context.Context, tenantID, propertyID string, images []PropertyImage) ([]PropertyImage, error)

	// DeleteImage remove uma imagem específica do imóvel.
	DeleteImage(ctx context.Context, tenantID, propertyID, imageID string) error
}
