package client

import (
	"context"

	"github.com/kubex-ecosystem/domus/internal/adapter"
	store "github.com/kubex-ecosystem/domus/internal/datastore"
	company "github.com/kubex-ecosystem/domus/internal/datastore/company_store"
	externalmetadata "github.com/kubex-ecosystem/domus/internal/datastore/external_metadata_store"
	invite "github.com/kubex-ecosystem/domus/internal/datastore/invite_store"
	pendingaccess "github.com/kubex-ecosystem/domus/internal/datastore/pending_access_store"
	user "github.com/kubex-ecosystem/domus/internal/datastore/user_store"
	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

// Type aliases para exports públicos dos stores (entidades e inputs)
type (
	// User types (UserStore já declarado em client.go)
	User            = user.User
	CreateUserInput = user.CreateUserInput
	UpdateUserInput = user.UpdateUserInput
	UserFilters     = user.UserFilters

	Executor = execution.Executor

	// Invite types (InviteStore já declarado em client.go)

	Invitation            = invite.Invitation
	InvitationType        = invite.InvitationType
	InvitationStatus      = invite.InvitationStatus
	CreateInvitationInput = invite.CreateInvitationInput
	UpdateInvitationInput = invite.UpdateInvitationInput
	InvitationFilters     = invite.InvitationFilters

	// Company types (CompanyStore já declarado em client.go)
	Company            = company.Company
	CreateCompanyInput = company.CreateCompanyInput
	UpdateCompanyInput = company.UpdateCompanyInput
	CompanyFilters     = company.CompanyFilters

	// Pending access types

	PendingAccessRequest            = pendingaccess.PendingAccessRequest
	CreatePendingAccessRequestInput = pendingaccess.CreatePendingAccessRequestInput
	UpdatePendingAccessRequestInput = pendingaccess.UpdatePendingAccessRequestInput
	PendingAccessFilters            = pendingaccess.PendingAccessFilters

	// External metadata types
	ExternalMetadataRecord      = externalmetadata.ExternalMetadataRecord
	UpsertExternalMetadataInput = externalmetadata.UpsertExternalMetadataInput
	ExternalMetadataFilters     = externalmetadata.ExternalMetadataFilters

	// Adapter types - apenas os não-genéricos

	AdapterFactory   = adapter.AdapterFactory
	RepositoryConfig = adapter.RepositoryConfig
)

// Type aliases genéricos (devem ser instanciados com tipo concreto)
type (
	DSRepository[T any]    = adapter.DSRepository[T]
	ORMRepository[T any]   = adapter.ORMRepository[T]
	Repository[T any]      = store.Repository[T]
	PaginatedResult[T any] = t.PaginatedResult[T] // Export para uso externo
)

// Invite type constants
const (
	TypePartner  = invite.TypePartner
	TypeInternal = invite.TypeInternal
)

// Invite status constants
const (
	StatusPending  = invite.StatusPending
	StatusAccepted = invite.StatusAccepted
	StatusRevoked  = invite.StatusRevoked
	StatusExpired  = invite.StatusExpired
)

// GetUserStore é um helper para obter UserStore diretamente.
func (c *DSClientImpl) GetUserStore(ctx context.Context, dbName string) (UserStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}

	factory := store.NewStoreFactory(conn.Driver)
	return factory.UserStore(ctx)
}

// GetInviteStore é um helper para obter InviteStore diretamente.
func (c *DSClientImpl) GetInviteStore(ctx context.Context, dbName string) (InviteStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}

	factory := store.NewStoreFactory(conn.Driver)
	return factory.InviteStore(ctx)
}

// GetCompanyStore é um helper para obter CompanyStore diretamente.
func (c *DSClientImpl) GetCompanyStore(ctx context.Context, dbName string) (CompanyStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}

	factory := store.NewStoreFactory(conn.Driver)
	return factory.CompanyStore(ctx)
}

// GetPendingAccessStore é um helper para obter PendingAccessStore diretamente.
func (c *DSClientImpl) GetPendingAccessStore(ctx context.Context, dbName string) (PendingAccessStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}

	factory := store.NewStoreFactory(conn.Driver)
	return factory.PendingAccessStore(ctx)
}

// GetExternalMetadataStore é um helper para obter ExternalMetadataStore diretamente.
func (c *DSClientImpl) GetExternalMetadataStore(ctx context.Context, dbName string) (ExternalMetadataStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}

	factory := store.NewStoreFactory(conn.Driver)
	return factory.ExternalMetadataStore(ctx)
}

// Funções standalone para criação de stores a partir de BackendConnection

// NewUserStore cria um UserStore a partir de uma conexão.
func NewUserStore(ctx context.Context, conn *BackendConnection) (UserStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.UserStore(ctx)
}

// NewInviteStore cria um InviteStore a partir de uma conexão.
func NewInviteStore(ctx context.Context, conn *BackendConnection) (InviteStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.InviteStore(ctx)
}

// NewCompanyStore cria um CompanyStore a partir de uma conexão.
func NewCompanyStore(ctx context.Context, conn *BackendConnection) (CompanyStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.CompanyStore(ctx)
}

// NewPendingAccessStore cria um PendingAccessStore a partir de uma conexão.
func NewPendingAccessStore(ctx context.Context, conn *BackendConnection) (PendingAccessStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.PendingAccessStore(ctx)
}

// NewExternalMetadataStore cria um ExternalMetadataStore a partir de uma conexão.
func NewExternalMetadataStore(ctx context.Context, conn *BackendConnection) (ExternalMetadataStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.ExternalMetadataStore(ctx)
}

func NewIntegrationStore(ctx context.Context, conn *BackendConnection, mKey []byte) (IntegrationStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.IntegrationStore(ctx, mKey)
}

// DefaultRepositoryConfig retorna a configuração padrão do adapter.
// Prefere Store com fallback automático para ORM.
func DefaultRepositoryConfig() *RepositoryConfig {
	return adapter.DefaultConfig()
}

// StoreOnlyConfig retorna configuração que força uso APENAS de Store.
func StoreOnlyConfig() *RepositoryConfig {
	return &RepositoryConfig{
		PreferStore:   true,
		FallbackToORM: false,
		ForceStore:    true,
		ForceORM:      false,
	}
}

// ORMOnlyConfig retorna configuração que força uso APENAS de ORM.
func ORMOnlyConfig() *RepositoryConfig {
	return &RepositoryConfig{
		PreferStore:   false,
		FallbackToORM: false,
		ForceStore:    false,
		ForceORM:      true,
	}
}

// CreateAdapter cria um DSRepository[T] adapter unificado.
//
// Wrapper público para adapter.CreateAdapter que pode ser chamado de outros módulos.
//
// Parâmetros:
// - factory: AdapterFactory obtida via DSClient.NewAdapterFactory()
// - ctx: contexto para operações
// - storeName: nome do store no registry (ex: "user", "company")
// - ormRepoFactory: função que cria ORMRepository[T] a partir de *gorm.DB
//
// Retorna DSRepository[T] que implementa Repository[T] unificado.
//
// Exemplo:
//
//	factory, _ := dsClient.NewAdapterFactory(ctx, "gnyx", gormDB, nil)
//	userRepo, _ := client.CreateAdapter[Users](factory, ctx, "user", users.NewRepository)
func CreateAdapter[T any](
	factory *AdapterFactory,
	ctx context.Context,
	storeName string,
	ormRepoFactory func(execution.Executor, error) ORMRepository[T],
) (*DSRepository[T], error) {
	return adapter.CreateAdapter(factory, ctx, storeName, ormRepoFactory)
}
