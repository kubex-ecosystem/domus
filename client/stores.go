package client

import (
	"context"

	"github.com/faelmori/domus-pulse/internal/adapter"
	store "github.com/faelmori/domus-pulse/internal/datastore"
	commissionstore "github.com/faelmori/domus-pulse/internal/datastore/commission_store"
	company "github.com/faelmori/domus-pulse/internal/datastore/company_store"
	externalmetadata "github.com/faelmori/domus-pulse/internal/datastore/external_metadata_store"
	invite "github.com/faelmori/domus-pulse/internal/datastore/invite_store"
	leadconvstore "github.com/faelmori/domus-pulse/internal/datastore/lead_conversation_store"
	leadeventstore "github.com/faelmori/domus-pulse/internal/datastore/lead_event_store"
	leadstore "github.com/faelmori/domus-pulse/internal/datastore/lead_store"
	membershipstore "github.com/faelmori/domus-pulse/internal/datastore/membership_store"
	notificationstore "github.com/faelmori/domus-pulse/internal/datastore/notification_store"
	pendingaccess "github.com/faelmori/domus-pulse/internal/datastore/pending_access_store"
	pipelinecardstore "github.com/faelmori/domus-pulse/internal/datastore/pipeline_card_store"
	propsharestore "github.com/faelmori/domus-pulse/internal/datastore/property_share_store"
	propertystore "github.com/faelmori/domus-pulse/internal/datastore/property_store"
	rolestore "github.com/faelmori/domus-pulse/internal/datastore/role_store"
	transactionstore "github.com/faelmori/domus-pulse/internal/datastore/transaction_store"
	user "github.com/faelmori/domus-pulse/internal/datastore/user_store"
	visiteventstore "github.com/faelmori/domus-pulse/internal/datastore/visit_event_store"
	wappmsgstore "github.com/faelmori/domus-pulse/internal/datastore/wapp_message_store"
	wappwebhookstore "github.com/faelmori/domus-pulse/internal/datastore/wapp_webhook_store"
	"github.com/faelmori/domus-pulse/internal/execution"
	t "github.com/faelmori/domus-pulse/internal/types"
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
	PendingAccessStatus             = pendingaccess.PendingAccessStatus
	PendingAccessRequest            = pendingaccess.PendingAccessRequest
	CreatePendingAccessRequestInput = pendingaccess.CreatePendingAccessRequestInput
	UpdatePendingAccessRequestInput = pendingaccess.UpdatePendingAccessRequestInput
	PendingAccessFilters            = pendingaccess.PendingAccessFilters

	// External metadata types
	ExternalMetadataRecord      = externalmetadata.ExternalMetadataRecord
	UpsertExternalMetadataInput = externalmetadata.UpsertExternalMetadataInput
	ExternalMetadataFilters     = externalmetadata.ExternalMetadataFilters

	// Property types (Colonial NetImóveis)
	Property            = propertystore.Property
	PropertyImage       = propertystore.PropertyImage
	CreatePropertyInput = propertystore.CreatePropertyInput
	UpdatePropertyInput = propertystore.UpdatePropertyInput
	PropertyFilters     = propertystore.PropertyFilters

	// PropertyShare types (Colonial NetImóveis)
	ShareRecord   = propsharestore.ShareRecord
	ShareSummary  = propsharestore.ShareSummary
	LogShareInput = propsharestore.LogShareInput

	// WappMessage types (Colonial NetImóveis)
	WappMessage      = wappmsgstore.WappMessage
	ThreadSummary    = wappmsgstore.ThreadSummary
	SaveMessageInput = wappmsgstore.SaveMessageInput

	// LeadConversation types (Colonial NetImóveis)
	LeadConversation        = leadconvstore.LeadConversation
	CreateConversationInput = leadconvstore.CreateConversationInput
	ConversationFilters     = leadconvstore.ConversationFilters

	// WappWebhook types (raw Meta webhook events)
	WebhookMessage          = wappwebhookstore.WebhookMessage
	WebhookStatus           = wappwebhookstore.WebhookStatus
	SaveWebhookMessageInput = wappwebhookstore.SaveWebhookMessageInput
	SaveWebhookStatusInput  = wappwebhookstore.SaveWebhookStatusInput

	// Notification types (Calm UX — Colonial NetImóveis)
	Notification                = notificationstore.Notification
	CreateNotificationInput     = notificationstore.CreateNotificationInput
	NotificationFilters         = notificationstore.NotificationFilters
	NotificationPref            = notificationstore.NotificationPref
	UpsertNotificationPrefInput = notificationstore.UpsertNotificationPrefInput

	// Commission types (Financial — Colonial NetImóveis)
	Commission            = commissionstore.Commission
	CreateCommissionInput = commissionstore.CreateCommissionInput
	CommissionFilters     = commissionstore.CommissionFilters

	// LeadEvent types (timeline — Colonial NetImóveis)
	LeadEvent            = leadeventstore.LeadEvent
	CreateLeadEventInput = leadeventstore.CreateLeadEventInput

	// Lead types (CRM — Colonial NetImóveis)
	Lead            = leadstore.Lead
	CreateLeadInput = leadstore.CreateLeadInput
	UpdateLeadInput = leadstore.UpdateLeadInput
	LeadFilters     = leadstore.LeadFilters

	// PipelineCard types (kanban CRM — Colonial NetImóveis)
	PipelineCard            = pipelinecardstore.PipelineCard
	CreatePipelineCardInput = pipelinecardstore.CreatePipelineCardInput
	PipelineCardFilters     = pipelinecardstore.PipelineCardFilters

	// Role/Membership types (RBAC)
	Role                  = rolestore.Role
	MemberWithPermissions = membershipstore.MemberWithPermissions
	UpsertMembershipInput = membershipstore.UpsertMembershipInput
	UpdateRoleInput       = membershipstore.UpdateRoleInput

	// Transaction types (Colonial NetImóveis — fechamento imobiliário)
	Transaction            = transactionstore.Transaction
	CreateTransactionInput = transactionstore.CreateTransactionInput
	TransactionFilters     = transactionstore.TransactionFilters
	TransactionStep        = transactionstore.TransactionStep
	UpdateStepInput        = transactionstore.UpdateStepInput
	TransactionDoc         = transactionstore.TransactionDoc
	AddDocInput            = transactionstore.AddDocInput

	// Visit agenda types (Colonial NetImóveis — roteiro de visitas)
	VisitEvent            = visiteventstore.VisitEvent
	CreateVisitEventInput = visiteventstore.CreateVisitEventInput
	VisitEventFilters     = visiteventstore.VisitEventFilters

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

// GetPropertyStore é um helper para obter PropertyStore diretamente.
func (c *DSClientImpl) GetPropertyStore(ctx context.Context, dbName string) (PropertyStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}

	factory := store.NewStoreFactory(conn.Driver)
	return factory.PropertyStore(ctx)
}

// GetPropertyShareStore é um helper para obter PropertyShareStore diretamente.
func (c *DSClientImpl) GetPropertyShareStore(ctx context.Context, dbName string) (PropertyShareStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.PropertyShareStore(ctx)
}

// GetWappMessageStore é um helper para obter WappMessageStore diretamente.
func (c *DSClientImpl) GetWappMessageStore(ctx context.Context, dbName string) (WappMessageStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.WappMessageStore(ctx)
}

// GetLeadConversationStore é um helper para obter LeadConversationStore diretamente.
func (c *DSClientImpl) GetLeadConversationStore(ctx context.Context, dbName string) (LeadConversationStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.LeadConversationStore(ctx)
}

// GetWappWebhookStore é um helper para obter WappWebhookStore diretamente.
func (c *DSClientImpl) GetWappWebhookStore(ctx context.Context, dbName string) (WappWebhookStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.WappWebhookStore(ctx)
}

// GetRoleStore é um helper para obter RoleStore diretamente.
func (c *DSClientImpl) GetRoleStore(ctx context.Context, dbName string) (RoleStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.RoleStore(ctx)
}

// GetMembershipStore é um helper para obter MembershipStore diretamente.
func (c *DSClientImpl) GetMembershipStore(ctx context.Context, dbName string) (MembershipStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.MembershipStore(ctx)
}

// NewWappWebhookStore cria um WappWebhookStore a partir de uma conexão.
func NewWappWebhookStore(ctx context.Context, conn *BackendConnection) (WappWebhookStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.WappWebhookStore(ctx)
}

// NewRoleStore cria um RoleStore a partir de uma conexão.
func NewRoleStore(ctx context.Context, conn *BackendConnection) (RoleStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.RoleStore(ctx)
}

// NewMembershipStore cria um MembershipStore a partir de uma conexão.
func NewMembershipStore(ctx context.Context, conn *BackendConnection) (MembershipStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.MembershipStore(ctx)
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

// NewPropertyStore cria um PropertyStore a partir de uma conexão.
func NewPropertyStore(ctx context.Context, conn *BackendConnection) (PropertyStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.PropertyStore(ctx)
}

// NewPropertyShareStore cria um PropertyShareStore a partir de uma conexão.
func NewPropertyShareStore(ctx context.Context, conn *BackendConnection) (PropertyShareStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.PropertyShareStore(ctx)
}

// NewWappMessageStore cria um WappMessageStore a partir de uma conexão.
func NewWappMessageStore(ctx context.Context, conn *BackendConnection) (WappMessageStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.WappMessageStore(ctx)
}

// NewLeadConversationStore cria um LeadConversationStore a partir de uma conexão.
func NewLeadConversationStore(ctx context.Context, conn *BackendConnection) (LeadConversationStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.LeadConversationStore(ctx)
}

func NewIntegrationStore(ctx context.Context, conn *BackendConnection, mKey []byte) (IntegrationStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.IntegrationStore(ctx, mKey)
}

// GetLeadEventStore é um helper para obter LeadEventStore diretamente.
func (c *DSClientImpl) GetLeadEventStore(ctx context.Context, dbName string) (LeadEventStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.LeadEventStore(ctx)
}

// NewLeadEventStore cria um LeadEventStore a partir de uma conexão.
func NewLeadEventStore(ctx context.Context, conn *BackendConnection) (LeadEventStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.LeadEventStore(ctx)
}

// GetLeadStore é um helper para obter LeadStore diretamente.
func (c *DSClientImpl) GetLeadStore(ctx context.Context, dbName string) (LeadStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.LeadStore(ctx)
}

// GetPipelineCardStore é um helper para obter PipelineCardStore diretamente.
func (c *DSClientImpl) GetPipelineCardStore(ctx context.Context, dbName string) (PipelineCardStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.PipelineCardStore(ctx)
}

// NewLeadStore cria um LeadStore a partir de uma conexão.
func NewLeadStore(ctx context.Context, conn *BackendConnection) (LeadStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.LeadStore(ctx)
}

// NewPipelineCardStore cria um PipelineCardStore a partir de uma conexão.
func NewPipelineCardStore(ctx context.Context, conn *BackendConnection) (PipelineCardStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.PipelineCardStore(ctx)
}

// GetNotificationStore é um helper para obter NotificationStore diretamente.
func (c *DSClientImpl) GetNotificationStore(ctx context.Context, dbName string) (NotificationStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.NotificationStore(ctx)
}

// GetNotificationPrefStore é um helper para obter NotificationPrefStore diretamente.
func (c *DSClientImpl) GetNotificationPrefStore(ctx context.Context, dbName string) (NotificationPrefStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.NotificationPrefStore(ctx)
}

// NewNotificationStore cria um NotificationStore a partir de uma conexão.
func NewNotificationStore(ctx context.Context, conn *BackendConnection) (NotificationStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.NotificationStore(ctx)
}

// NewNotificationPrefStore cria um NotificationPrefStore a partir de uma conexão.
func NewNotificationPrefStore(ctx context.Context, conn *BackendConnection) (NotificationPrefStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.NotificationPrefStore(ctx)
}

// GetCommissionStore é um helper para obter CommissionStore diretamente.
func (c *DSClientImpl) GetCommissionStore(ctx context.Context, dbName string) (CommissionStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.CommissionStore(ctx)
}

// NewCommissionStore cria um CommissionStore a partir de uma conexão.
func NewCommissionStore(ctx context.Context, conn *BackendConnection) (CommissionStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.CommissionStore(ctx)
}

// GetTransactionStore é um helper para obter TransactionStore diretamente.
func (c *DSClientImpl) GetTransactionStore(ctx context.Context, dbName string) (TransactionStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.TransactionStore(ctx)
}

// NewTransactionStore cria um TransactionStore a partir de uma conexão.
func NewTransactionStore(ctx context.Context, conn *BackendConnection) (TransactionStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.TransactionStore(ctx)
}

// GetTransactionStepStore é um helper para obter TransactionStepStore diretamente.
func (c *DSClientImpl) GetTransactionStepStore(ctx context.Context, dbName string) (TransactionStepStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.TransactionStepStore(ctx)
}

// NewTransactionStepStore cria um TransactionStepStore a partir de uma conexão.
func NewTransactionStepStore(ctx context.Context, conn *BackendConnection) (TransactionStepStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.TransactionStepStore(ctx)
}

// GetTransactionDocStore é um helper para obter TransactionDocStore diretamente.
func (c *DSClientImpl) GetTransactionDocStore(ctx context.Context, dbName string) (TransactionDocStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.TransactionDocStore(ctx)
}

// NewTransactionDocStore cria um TransactionDocStore a partir de uma conexão.
func NewTransactionDocStore(ctx context.Context, conn *BackendConnection) (TransactionDocStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.TransactionDocStore(ctx)
}

// GetVisitEventStore é um helper para obter VisitEventStore diretamente.
func (c *DSClientImpl) GetVisitEventStore(ctx context.Context, dbName string) (VisitEventStore, error) {
	conn, err := c.GetConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	factory := store.NewStoreFactory(conn.Driver)
	return factory.VisitEventStore(ctx)
}

// NewVisitEventStore cria um VisitEventStore a partir de uma conexão.
func NewVisitEventStore(ctx context.Context, conn *BackendConnection) (VisitEventStore, error) {
	factory := store.NewStoreFactory(conn.Driver)
	return factory.VisitEventStore(ctx)
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
