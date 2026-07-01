package datastore

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kubex-ecosystem/domus/internal/execution"
	"github.com/kubex-ecosystem/domus/internal/provider/flavors"

	bmoristore "github.com/kubex-ecosystem/domus/internal/datastore/bmori_store"
	ckcompanystore "github.com/kubex-ecosystem/domus/internal/datastore/ck_company_store"
	ckinsumostore "github.com/kubex-ecosystem/domus/internal/datastore/ck_insumo_store"
	ckmessagestore "github.com/kubex-ecosystem/domus/internal/datastore/ck_message_store"
	ckmovimentacaostore "github.com/kubex-ecosystem/domus/internal/datastore/ck_movimentacao_store"
	ckorderstore "github.com/kubex-ecosystem/domus/internal/datastore/ck_order_store"
	ckprodutostore "github.com/kubex-ecosystem/domus/internal/datastore/ck_produto_store"
	commissionstore "github.com/kubex-ecosystem/domus/internal/datastore/commission_store"
	c "github.com/kubex-ecosystem/domus/internal/datastore/company_store"
	externalmetadata "github.com/kubex-ecosystem/domus/internal/datastore/external_metadata_store"
	integration "github.com/kubex-ecosystem/domus/internal/datastore/integration_store"
	i "github.com/kubex-ecosystem/domus/internal/datastore/invite_store"
	leadconv "github.com/kubex-ecosystem/domus/internal/datastore/lead_conversation_store"
	leadeventstore "github.com/kubex-ecosystem/domus/internal/datastore/lead_event_store"
	leadstore "github.com/kubex-ecosystem/domus/internal/datastore/lead_store"
	membershipstore "github.com/kubex-ecosystem/domus/internal/datastore/membership_store"
	notificationstore "github.com/kubex-ecosystem/domus/internal/datastore/notification_store"
	p "github.com/kubex-ecosystem/domus/internal/datastore/pending_access_store"
	pipelinecardstore "github.com/kubex-ecosystem/domus/internal/datastore/pipeline_card_store"
	propshare "github.com/kubex-ecosystem/domus/internal/datastore/property_share_store"
	prop "github.com/kubex-ecosystem/domus/internal/datastore/property_store"
	rolestore "github.com/kubex-ecosystem/domus/internal/datastore/role_store"
	s "github.com/kubex-ecosystem/domus/internal/datastore/schemas_store"
	session "github.com/kubex-ecosystem/domus/internal/datastore/session_store"
	transactionstore "github.com/kubex-ecosystem/domus/internal/datastore/transaction_store"
	u "github.com/kubex-ecosystem/domus/internal/datastore/user_store"
	visiteventstore "github.com/kubex-ecosystem/domus/internal/datastore/visit_event_store"
	wappmsg "github.com/kubex-ecosystem/domus/internal/datastore/wapp_message_store"
	wappwebhook "github.com/kubex-ecosystem/domus/internal/datastore/wapp_webhook_store"
	t "github.com/kubex-ecosystem/domus/internal/types"
	gl "github.com/kubex-ecosystem/logz"
)

func init() {
	if registry == nil {
		var err error
		if registry, err = s.NewSchemaRegistry(); err != nil {
			gl.Fatalf("falha ao inicializar SchemaRegistry: %v", err)
		}
	}
}

// StoreFactory fornece métodos para construir stores a partir de drivers.
type StoreFactory struct {
	driver     t.Driver
	driverName string
}

// NewStoreFactory cria uma factory baseada em um Driver.
func NewStoreFactory(driver t.Driver) *StoreFactory {
	return &StoreFactory{
		driver:     driver,
		driverName: GetDriverName(driver),
	}
}

// Create cria um store por nome usando o registry.
func (f *StoreFactory) Create(ctx context.Context, storeName string) (t.StoreType, error) {
	factory, ok := GetStoreFactory(f.driverName, storeName)
	if !ok {
		return nil, fmt.Errorf("store %s not found for driver %s", storeName, f.driverName)
	}
	return factory(ctx, f.driver)
}

// ListAvailable retorna stores disponíveis para este driver.
func (f *StoreFactory) ListAvailable() []string {
	return ListStoresByDriver(f.driverName)
}

// GetMetadata retorna metadados de um store.
func (f *StoreFactory) GetMetadata(storeName string) (*t.StoreMetadata, error) {
	return GetStoreMetadata(f.driverName, storeName)
}

// BMoriStore cria um BMoriStore (helper específico).
func (f *StoreFactory) BMoriStore(ctx context.Context) (bmoristore.BMoriStore, error) {
	store, err := f.Create(ctx, "bmori")
	if err != nil {
		return nil, err
	}
	bMoriStore, ok := store.(bmoristore.BMoriStore)
	if !ok {
		return nil, fmt.Errorf("store is not a BMoriStore")
	}
	return bMoriStore, nil
}

// UserStore cria um UserStore (helper específico).
func (f *StoreFactory) UserStore(ctx context.Context) (u.UserStore, error) {
	store, err := f.Create(ctx, "user")
	if err != nil {
		return nil, err
	}
	userStore, ok := store.(u.UserStore)
	if !ok {
		return nil, fmt.Errorf("store is not a UserStore")
	}
	return userStore, nil
}

// SessionStore cria um SessionStore (helper específico).
func (f *StoreFactory) SessionStore(ctx context.Context) (session.SessionStore, error) {
	store, err := f.Create(ctx, "session")
	if err != nil {
		return nil, err
	}
	sessionStore, ok := store.(session.SessionStore)
	if !ok {
		return nil, fmt.Errorf("store is not a SessionStore")
	}
	return sessionStore, nil
}

// InviteStore cria um InviteStore (helper específico).
func (f *StoreFactory) InviteStore(ctx context.Context) (i.InviteStore, error) {
	available := f.ListAvailable()
	found := false
	for _, name := range available {
		if name == "invite" {
			found = true
			break
		}
	}
	if !found {
		gl.Warnf("InviteStore not available for driver %s", f.driverName)
	}
	executor, err := f.driver.Executor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get executor from driver: %v", err)
	}
	store, err := NewInviteStoreFromExecutor(executor)
	if err != nil {
		return nil, err
	}
	return store, nil
}

// CompanyStore cria um CompanyStore (helper específico).
func (f *StoreFactory) CompanyStore(ctx context.Context) (c.CompanyStore, error) {
	store, err := f.Create(ctx, "company")
	if err != nil {
		return nil, err
	}
	companyStore, ok := store.(c.CompanyStore)
	if !ok {
		return nil, fmt.Errorf("store is not a CompanyStore")
	}
	return companyStore, nil
}

// PendingAccessStore cria um PendingAccessStore (helper específico).
func (f *StoreFactory) PendingAccessStore(ctx context.Context) (p.PendingAccessStore, error) {
	store, err := f.Create(ctx, "pending_access_request")
	if err != nil {
		return nil, err
	}
	pendingStore, ok := store.(p.PendingAccessStore)
	if !ok {
		return nil, fmt.Errorf("store is not a PendingAccessStore")
	}
	return pendingStore, nil
}

// ExternalMetadataStore cria um ExternalMetadataStore (helper específico).
func (f *StoreFactory) ExternalMetadataStore(ctx context.Context) (externalmetadata.ExternalMetadataStore, error) {
	store, err := f.Create(ctx, "external_metadata")
	if err != nil {
		return nil, err
	}
	externalMetadataStore, ok := store.(externalmetadata.ExternalMetadataStore)
	if !ok {
		return nil, fmt.Errorf("store is not an ExternalMetadataStore")
	}
	return externalMetadataStore, nil
}

// PropertyStore cria um PropertyStore (helper específico).
func (f *StoreFactory) PropertyStore(ctx context.Context) (prop.PropertyStore, error) {
	store, err := f.Create(ctx, "property")
	if err != nil {
		return nil, err
	}
	propertyStore, ok := store.(prop.PropertyStore)
	if !ok {
		return nil, fmt.Errorf("store is not a PropertyStore")
	}
	return propertyStore, nil
}

func (f *StoreFactory) IntegrationStore(ctx context.Context, mKey []byte) (integration.IntegrationStore, error) {
	store, err := f.Create(ctx, "integration")
	if err != nil {
		return nil, err
	}
	integrationStore, ok := store.(integration.IntegrationStore)
	if !ok {
		return nil, fmt.Errorf("store is not an IntegrationStore")
	}
	return integrationStore, nil
}

// GetDriverName retorna o nome do driver usando reflection.
func GetDriverName(drv t.Driver) string {
	// Tenta usar Driver.Name() se disponível
	if drv != nil {
		if name := drv.Name(); name != "" {
			return name
		}
	}

	// Fallback: usa registry de flavors
	registry := flavors.AllMap()
	drvType := reflect.TypeOf(drv)

	for name, provider := range registry {
		if reflect.TypeOf(provider) == drvType {
			return name
		}
	}

	return "unknown"
}

// Funções helper para criar stores diretamente de um Executor

// NewUserStoreFromExecutor cria UserStore diretamente de um Executor.
func NewUserStoreFromExecutor(exec execution.Executor) (u.UserStore, error) {
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("user store requires Postgres backend, got %s", exec.Kind())
	}

	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}

	return u.NewPGUserStore(pgExec), nil
}

// NewSessionStoreFromExecutor cria SessionStore diretamente de um Executor.
func NewSessionStoreFromExecutor(exec execution.Executor) (session.SessionStore, error) {
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("session store requires Postgres backend, got %s", exec.Kind())
	}

	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}

	return session.NewPGSessionStore(pgExec), nil
}

// NewInviteStoreFromExecutor cria InviteStore diretamente de um Executor.
func NewInviteStoreFromExecutor(exec execution.Executor) (i.InviteStore, error) {
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("invite store requires Postgres backend, got %s", exec.Kind())
	}

	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}

	return i.NewPGInviteStore(pgExec), nil
}

// NewCompanyStoreFromExecutor cria CompanyStore diretamente de um Executor.
func NewCompanyStoreFromExecutor(exec execution.Executor) (c.CompanyStore, error) {
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("company store requires Postgres backend, got %s", exec.Kind())
	}

	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}

	return c.NewPGCompanyStore(pgExec), nil
}

// NewPendingAccessStoreFromExecutor cria PendingAccessStore diretamente de um Executor.
func NewPendingAccessStoreFromExecutor(exec execution.Executor) (p.PendingAccessStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("pending access store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return p.NewPGPendingAccessStore(pgExec), nil
}

// NewExternalMetadataStoreFromExecutor cria ExternalMetadataStore diretamente de um Executor.
func NewExternalMetadataStoreFromExecutor(exec execution.Executor) (externalmetadata.ExternalMetadataStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("external metadata store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return externalmetadata.NewPGExternalMetadataStore(pgExec), nil
}

// NewPropertyStoreFromExecutor cria PropertyStore diretamente de um Executor.
func NewPropertyStoreFromExecutor(exec execution.Executor) (prop.PropertyStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("property store requires Postgres backend, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return prop.NewPGPropertyStore(pgExec), nil
}

// PropertyShareStore cria um PropertyShareStore (helper específico).
func (f *StoreFactory) PropertyShareStore(ctx context.Context) (propshare.PropertyShareStore, error) {
	store, err := f.Create(ctx, "property_share")
	if err != nil {
		return nil, err
	}
	shareStore, ok := store.(propshare.PropertyShareStore)
	if !ok {
		return nil, fmt.Errorf("store is not a PropertyShareStore")
	}
	return shareStore, nil
}

// WappMessageStore cria um WappMessageStore (helper específico).
func (f *StoreFactory) WappMessageStore(ctx context.Context) (wappmsg.WappMessageStore, error) {
	store, err := f.Create(ctx, "wapp_message")
	if err != nil {
		return nil, err
	}
	wappStore, ok := store.(wappmsg.WappMessageStore)
	if !ok {
		return nil, fmt.Errorf("store is not a WappMessageStore")
	}
	return wappStore, nil
}

// LeadConversationStore cria um LeadConversationStore (helper específico).
func (f *StoreFactory) LeadConversationStore(ctx context.Context) (leadconv.LeadConversationStore, error) {
	store, err := f.Create(ctx, "lead_conversation")
	if err != nil {
		return nil, err
	}
	convStore, ok := store.(leadconv.LeadConversationStore)
	if !ok {
		return nil, fmt.Errorf("store is not a LeadConversationStore")
	}
	return convStore, nil
}

// NewPropertyShareStoreFromExecutor cria PropertyShareStore diretamente de um Executor.
func NewPropertyShareStoreFromExecutor(exec execution.Executor) (propshare.PropertyShareStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("property share store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return propshare.NewPGPropertyShareStore(pgExec), nil
}

// NewWappMessageStoreFromExecutor cria WappMessageStore diretamente de um Executor.
func NewWappMessageStoreFromExecutor(exec execution.Executor) (wappmsg.WappMessageStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("wapp message store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return wappmsg.NewPGWappMessageStore(pgExec), nil
}

// NewLeadConversationStoreFromExecutor cria LeadConversationStore diretamente de um Executor.
func NewLeadConversationStoreFromExecutor(exec execution.Executor) (leadconv.LeadConversationStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("lead conversation store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return leadconv.NewPGLeadConversationStore(pgExec), nil
}

func NewIntegrationStoreFromExecutor(exec execution.Executor, mKey []byte) (integration.IntegrationStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("integration store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return integration.NewPGIntegrationStore(pgExec, mKey), nil
}

// WappWebhookStore cria um WappWebhookStore (eventos brutos de webhook Meta).
func (f *StoreFactory) WappWebhookStore(ctx context.Context) (wappwebhook.WappWebhookStore, error) {
	store, err := f.Create(ctx, "wapp_webhook")
	if err != nil {
		return nil, err
	}
	webhookStore, ok := store.(wappwebhook.WappWebhookStore)
	if !ok {
		return nil, fmt.Errorf("store is not a WappWebhookStore")
	}
	return webhookStore, nil
}

// NewWappWebhookStoreFromExecutor cria WappWebhookStore diretamente de um Executor.
func NewWappWebhookStoreFromExecutor(exec execution.Executor) (wappwebhook.WappWebhookStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("wapp webhook store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return wappwebhook.NewPGWappWebhookStore(pgExec), nil
}

// RoleStore cria um RoleStore (leitura de roles RBAC).
func (f *StoreFactory) RoleStore(ctx context.Context) (rolestore.RoleStore, error) {
	store, err := f.Create(ctx, "role")
	if err != nil {
		return nil, err
	}
	rs, ok := store.(rolestore.RoleStore)
	if !ok {
		return nil, fmt.Errorf("store is not a RoleStore")
	}
	return rs, nil
}

// NewRoleStoreFromExecutor cria RoleStore diretamente de um Executor.
func NewRoleStoreFromExecutor(exec execution.Executor) (rolestore.RoleStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("role store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return rolestore.NewPGRoleStore(pgExec), nil
}

// MembershipStore cria um MembershipStore (tenant_membership com joins).
func (f *StoreFactory) MembershipStore(ctx context.Context) (membershipstore.MembershipStore, error) {
	store, err := f.Create(ctx, "membership")
	if err != nil {
		return nil, err
	}
	ms, ok := store.(membershipstore.MembershipStore)
	if !ok {
		return nil, fmt.Errorf("store is not a MembershipStore")
	}
	return ms, nil
}

// NewMembershipStoreFromExecutor cria MembershipStore diretamente de um Executor.
func NewMembershipStoreFromExecutor(exec execution.Executor) (membershipstore.MembershipStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("membership store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return membershipstore.NewPGMembershipStore(pgExec), nil
}

// LeadStore cria um LeadStore (CRM de leads — Colonial NetImóveis).
func (f *StoreFactory) LeadStore(ctx context.Context) (leadstore.LeadStore, error) {
	store, err := f.Create(ctx, "lead")
	if err != nil {
		return nil, err
	}
	ls, ok := store.(leadstore.LeadStore)
	if !ok {
		return nil, fmt.Errorf("store is not a LeadStore")
	}
	return ls, nil
}

// NewLeadStoreFromExecutor cria LeadStore diretamente de um Executor.
func NewLeadStoreFromExecutor(exec execution.Executor) (leadstore.LeadStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("lead store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return leadstore.NewPGLeadStore(pgExec), nil
}

// PipelineCardStore cria um PipelineCardStore (kanban CRM — Colonial NetImóveis).
func (f *StoreFactory) PipelineCardStore(ctx context.Context) (pipelinecardstore.PipelineCardStore, error) {
	store, err := f.Create(ctx, "pipeline_card")
	if err != nil {
		return nil, err
	}
	ps, ok := store.(pipelinecardstore.PipelineCardStore)
	if !ok {
		return nil, fmt.Errorf("store is not a PipelineCardStore")
	}
	return ps, nil
}

// NewPipelineCardStoreFromExecutor cria PipelineCardStore diretamente de um Executor.
func NewPipelineCardStoreFromExecutor(exec execution.Executor) (pipelinecardstore.PipelineCardStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("pipeline card store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return pipelinecardstore.NewPGPipelineCardStore(pgExec), nil
}

// CommissionStore cria um CommissionStore (módulo financeiro — Colonial NetImóveis).
func (f *StoreFactory) CommissionStore(ctx context.Context) (commissionstore.CommissionStore, error) {
	store, err := f.Create(ctx, "commission")
	if err != nil {
		return nil, err
	}
	cs, ok := store.(commissionstore.CommissionStore)
	if !ok {
		return nil, fmt.Errorf("store is not a CommissionStore")
	}
	return cs, nil
}

// NewCommissionStoreFromExecutor cria CommissionStore diretamente de um Executor.
func NewCommissionStoreFromExecutor(exec execution.Executor) (commissionstore.CommissionStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("commission store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return commissionstore.NewPGCommissionStore(pgExec), nil
}

// LeadEventStore cria um LeadEventStore (timeline de atividades — Colonial NetImóveis).
func (f *StoreFactory) LeadEventStore(ctx context.Context) (leadeventstore.LeadEventStore, error) {
	store, err := f.Create(ctx, "lead_event")
	if err != nil {
		return nil, err
	}
	les, ok := store.(leadeventstore.LeadEventStore)
	if !ok {
		return nil, fmt.Errorf("store is not a LeadEventStore")
	}
	return les, nil
}

// NewLeadEventStoreFromExecutor cria LeadEventStore diretamente de um Executor.
func NewLeadEventStoreFromExecutor(exec execution.Executor) (leadeventstore.LeadEventStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("lead event store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return leadeventstore.NewPGLeadEventStore(pgExec), nil
}

// NotificationStore cria um NotificationStore (centro de notificações).
func (f *StoreFactory) NotificationStore(ctx context.Context) (notificationstore.NotificationStore, error) {
	store, err := f.Create(ctx, "notification")
	if err != nil {
		return nil, err
	}
	ns, ok := store.(notificationstore.NotificationStore)
	if !ok {
		return nil, fmt.Errorf("store is not a NotificationStore")
	}
	return ns, nil
}

// NewNotificationStoreFromExecutor cria NotificationStore diretamente de um Executor.
func NewNotificationStoreFromExecutor(exec execution.Executor) (notificationstore.NotificationStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("notification store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return notificationstore.NewPGNotificationStore(pgExec), nil
}

// NotificationPrefStore cria um NotificationPrefStore (preferências Calm UX).
func (f *StoreFactory) NotificationPrefStore(ctx context.Context) (notificationstore.NotificationPrefStore, error) {
	store, err := f.Create(ctx, "notification_pref")
	if err != nil {
		return nil, err
	}
	nps, ok := store.(notificationstore.NotificationPrefStore)
	if !ok {
		return nil, fmt.Errorf("store is not a NotificationPrefStore")
	}
	return nps, nil
}

// NewNotificationPrefStoreFromExecutor cria NotificationPrefStore diretamente de um Executor.
func NewNotificationPrefStoreFromExecutor(exec execution.Executor) (notificationstore.NotificationPrefStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("notification pref store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return notificationstore.NewPGNotificationPrefStore(pgExec), nil
}

// TransactionStore cria um TransactionStore (negociações imobiliárias — Colonial NetImóveis).
func (f *StoreFactory) TransactionStore(ctx context.Context) (transactionstore.TransactionStore, error) {
	store, err := f.Create(ctx, "transaction")
	if err != nil {
		return nil, err
	}
	ts, ok := store.(transactionstore.TransactionStore)
	if !ok {
		return nil, fmt.Errorf("store is not a TransactionStore")
	}
	return ts, nil
}

// NewTransactionStoreFromExecutor cria TransactionStore diretamente de um Executor.
func NewTransactionStoreFromExecutor(exec execution.Executor) (transactionstore.TransactionStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("transaction store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return transactionstore.NewPGTransactionStore(pgExec), nil
}

// TransactionStepStore cria um TransactionStepStore (checklist — Colonial NetImóveis).
func (f *StoreFactory) TransactionStepStore(ctx context.Context) (transactionstore.TransactionStepStore, error) {
	store, err := f.Create(ctx, "transaction_step")
	if err != nil {
		return nil, err
	}
	ss, ok := store.(transactionstore.TransactionStepStore)
	if !ok {
		return nil, fmt.Errorf("store is not a TransactionStepStore")
	}
	return ss, nil
}

// NewTransactionStepStoreFromExecutor cria TransactionStepStore diretamente de um Executor.
func NewTransactionStepStoreFromExecutor(exec execution.Executor) (transactionstore.TransactionStepStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("transaction step store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return transactionstore.NewPGTransactionStepStore(pgExec), nil
}

// TransactionDocStore cria um TransactionDocStore (documentos — Colonial NetImóveis).
func (f *StoreFactory) TransactionDocStore(ctx context.Context) (transactionstore.TransactionDocStore, error) {
	store, err := f.Create(ctx, "transaction_doc")
	if err != nil {
		return nil, err
	}
	ds, ok := store.(transactionstore.TransactionDocStore)
	if !ok {
		return nil, fmt.Errorf("store is not a TransactionDocStore")
	}
	return ds, nil
}

// NewTransactionDocStoreFromExecutor cria TransactionDocStore diretamente de um Executor.
func NewTransactionDocStoreFromExecutor(exec execution.Executor) (transactionstore.TransactionDocStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("transaction doc store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return transactionstore.NewPGTransactionDocStore(pgExec), nil
}

// VisitEventStore cria um VisitEventStore (agenda de visitas — Colonial NetImóveis).
func (f *StoreFactory) VisitEventStore(ctx context.Context) (visiteventstore.VisitEventStore, error) {
	store, err := f.Create(ctx, "visit_event")
	if err != nil {
		return nil, err
	}
	vs, ok := store.(visiteventstore.VisitEventStore)
	if !ok {
		return nil, fmt.Errorf("store is not a VisitEventStore")
	}
	return vs, nil
}

// NewVisitEventStoreFromExecutor cria VisitEventStore diretamente de um Executor.
func NewVisitEventStoreFromExecutor(exec execution.Executor) (visiteventstore.VisitEventStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("visit event store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return visiteventstore.NewPGVisitEventStore(pgExec), nil
}

// ── Cozinha Pass (KDS) ───────────────────────────────────────────────────────

func (f *StoreFactory) CKCompanyStore(ctx context.Context) (ckcompanystore.CKCompanyStore, error) {
	store, err := f.Create(ctx, "ck_company")
	if err != nil {
		return nil, err
	}
	s, ok := store.(ckcompanystore.CKCompanyStore)
	if !ok {
		return nil, fmt.Errorf("store is not a CKCompanyStore")
	}
	return s, nil
}

func NewCKCompanyStoreFromExecutor(exec execution.Executor) (ckcompanystore.CKCompanyStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("ck_company store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return ckcompanystore.NewPGCKCompanyStore(pgExec), nil
}

func (f *StoreFactory) CKOrderStore(ctx context.Context) (ckorderstore.CKOrderStore, error) {
	store, err := f.Create(ctx, "ck_order")
	if err != nil {
		return nil, err
	}
	s, ok := store.(ckorderstore.CKOrderStore)
	if !ok {
		return nil, fmt.Errorf("store is not a CKOrderStore")
	}
	return s, nil
}

func NewCKOrderStoreFromExecutor(exec execution.Executor) (ckorderstore.CKOrderStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("ck_order store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return ckorderstore.NewPGCKOrderStore(pgExec), nil
}

func (f *StoreFactory) CKInsumoStore(ctx context.Context) (ckinsumostore.CKInsumoStore, error) {
	store, err := f.Create(ctx, "ck_insumo")
	if err != nil {
		return nil, err
	}
	s, ok := store.(ckinsumostore.CKInsumoStore)
	if !ok {
		return nil, fmt.Errorf("store is not a CKInsumoStore")
	}
	return s, nil
}

func NewCKInsumoStoreFromExecutor(exec execution.Executor) (ckinsumostore.CKInsumoStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("ck_insumo store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return ckinsumostore.NewPGCKInsumoStore(pgExec), nil
}

func (f *StoreFactory) CKProdutoStore(ctx context.Context) (ckprodutostore.CKProdutoStore, error) {
	store, err := f.Create(ctx, "ck_produto")
	if err != nil {
		return nil, err
	}
	s, ok := store.(ckprodutostore.CKProdutoStore)
	if !ok {
		return nil, fmt.Errorf("store is not a CKProdutoStore")
	}
	return s, nil
}

func NewCKProdutoStoreFromExecutor(exec execution.Executor) (ckprodutostore.CKProdutoStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("ck_produto store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return ckprodutostore.NewPGCKProdutoStore(pgExec), nil
}

func (f *StoreFactory) CKMessageStore(ctx context.Context) (ckmessagestore.CKMessageStore, error) {
	store, err := f.Create(ctx, "ck_message")
	if err != nil {
		return nil, err
	}
	s, ok := store.(ckmessagestore.CKMessageStore)
	if !ok {
		return nil, fmt.Errorf("store is not a CKMessageStore")
	}
	return s, nil
}

func NewCKMessageStoreFromExecutor(exec execution.Executor) (ckmessagestore.CKMessageStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("ck_message store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return ckmessagestore.NewPGCKMessageStore(pgExec), nil
}

func (f *StoreFactory) CKMovimentacaoStore(ctx context.Context) (ckmovimentacaostore.CKMovimentacaoStore, error) {
	store, err := f.Create(ctx, "ck_movimentacao")
	if err != nil {
		return nil, err
	}
	s, ok := store.(ckmovimentacaostore.CKMovimentacaoStore)
	if !ok {
		return nil, fmt.Errorf("store is not a CKMovimentacaoStore")
	}
	return s, nil
}

func NewCKMovimentacaoStoreFromExecutor(exec execution.Executor) (ckmovimentacaostore.CKMovimentacaoStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("ck_movimentacao store requires Postgres, got %s", exec.Kind())
	}
	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}
	return ckmovimentacaostore.NewPGCKMovimentacaoStore(pgExec), nil
}

// suppress unused import warnings for reflect (used in init.go via same package)
var _ = reflect.TypeOf
