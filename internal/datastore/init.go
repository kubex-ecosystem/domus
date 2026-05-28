package datastore

import (
	"context"
	"reflect"

	commissionstore "github.com/kubex-ecosystem/domus/internal/datastore/commission_store"
	company "github.com/kubex-ecosystem/domus/internal/datastore/company_store"
	externalmetadata "github.com/kubex-ecosystem/domus/internal/datastore/external_metadata_store"
	invite "github.com/kubex-ecosystem/domus/internal/datastore/invite_store"
	leadconversationstore "github.com/kubex-ecosystem/domus/internal/datastore/lead_conversation_store"
	leadeventstore "github.com/kubex-ecosystem/domus/internal/datastore/lead_event_store"
	leadstore "github.com/kubex-ecosystem/domus/internal/datastore/lead_store"
	membershipstore "github.com/kubex-ecosystem/domus/internal/datastore/membership_store"
	notificationstore "github.com/kubex-ecosystem/domus/internal/datastore/notification_store"
	pendingaccess "github.com/kubex-ecosystem/domus/internal/datastore/pending_access_store"
	pipelinecardstore "github.com/kubex-ecosystem/domus/internal/datastore/pipeline_card_store"
	propertysharesstore "github.com/kubex-ecosystem/domus/internal/datastore/property_share_store"
	propertystore "github.com/kubex-ecosystem/domus/internal/datastore/property_store"
	rolestore "github.com/kubex-ecosystem/domus/internal/datastore/role_store"
	transactionstore "github.com/kubex-ecosystem/domus/internal/datastore/transaction_store"
	userstore "github.com/kubex-ecosystem/domus/internal/datastore/user_store"
	visiteventstore "github.com/kubex-ecosystem/domus/internal/datastore/visit_event_store"
	wappmessagestore "github.com/kubex-ecosystem/domus/internal/datastore/wapp_message_store"
	wappwebhookstore "github.com/kubex-ecosystem/domus/internal/datastore/wapp_webhook_store"
	"github.com/kubex-ecosystem/domus/internal/execution"
	"github.com/kubex-ecosystem/domus/internal/provider/flavors"

	t "github.com/kubex-ecosystem/domus/internal/types"
	gl "github.com/kubex-ecosystem/logz"
)

func init() {
	if globalRegistry == nil {
		gl.Notice("Initializing store registry")
		globalRegistry = newStoreRegistry()
	}

	// Registra stores para cada driver disponível
	registerPostgresStores()
	registerMongoStores()
	registerRedisStores()
	registerRabbitStores()

	gl.Notice("Store registry initialized")
}

// registerPostgresStores registra todos os stores para PostgreSQL.
func registerPostgresStores() {
	driverName := "postgres"

	// Valida que o driver existe; se não, registra mesmo assim para permitir uso direto.
	if _, ok := flavors.Get(driverName); !ok {
		gl.Noticef("Driver %s not found in flavors registry", driverName)
	}

	// Registra UserStore
	userMetadata := &t.StoreMetadata{
		Name:        "user",
		Type:        reflect.TypeFor[userstore.User](),
		DriverName:  driverName,
		Description: "User management store",
		Version:     "1.0.0",
		Capabilities: []string{
			"crud",
			"email_lookup",
			"pagination",
			"filtering",
		},
	}

	userFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}

		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("user store requires Postgres, got %s", exec.Kind())
		}

		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}

		return userstore.NewPGUserStore(pgExec), nil
	}

	if err := RegisterStore(driverName, "user", userMetadata, userFactory); err != nil {
		gl.Errorf("Failed to register user store: %v", err)
	}

	// Registra InviteStore
	inviteMetadata := &t.StoreMetadata{
		Name:        "invite",
		Type:        reflect.TypeFor[invite.Invitation](),
		DriverName:  driverName,
		Description: "Invitation management store (partner + internal)",
		Version:     "1.0.0",
		Capabilities: []string{
			"crud",
			"token_lookup",
			"pagination",
			"filtering",
			"transactions",
			"dual_table", // partner_invitation + internal_invitation
		},
	}

	inviteFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}

		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("invite store requires Postgres, got %s", exec.Kind())
		}

		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}

		return invite.NewPGInviteStore(pgExec), nil
	}

	if err := RegisterStore(driverName, "invite", inviteMetadata, inviteFactory); err != nil {
		gl.Errorf("Failed to register invite store: %v", err)
	}

	// Registra CompanyStore
	companyMetadata := &t.StoreMetadata{
		Name:        "company",
		Type:        reflect.TypeFor[company.Company](),
		DriverName:  driverName,
		Description: "Company management store",
		Version:     "1.0.0",
		Capabilities: []string{
			"crud",
			"slug_lookup",
			"pagination",
			"filtering",
		},
	}

	companyFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}

		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("company store requires Postgres, got %s", exec.Kind())
		}

		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}

		return company.NewPGCompanyStore(pgExec), nil
	}

	if err := RegisterStore(driverName, "company", companyMetadata, companyFactory); err != nil {
		gl.Errorf("Failed to register company store: %v", err)
	}

	// Registra PendingAccessStore
	pendingMetadata := &t.StoreMetadata{
		Name:        "pending_access_request",
		Type:        reflect.TypeFor[pendingaccess.PendingAccessRequest](),
		DriverName:  driverName,
		Description: "Pending access requests store (OAuth, external access)",
		Version:     "1.0.0",
		Capabilities: []string{
			"crud",
			"pagination",
			"filtering",
		},
	}

	pendingFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}

		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("pending access store requires Postgres, got %s", exec.Kind())
		}

		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}

		return pendingaccess.NewPGPendingAccessStore(pgExec), nil
	}

	if err := RegisterStore(driverName, "pending_access_request", pendingMetadata, pendingFactory); err != nil {
		gl.Errorf("Failed to register pending access store: %v", err)
	}

	externalMetadata := &t.StoreMetadata{
		Name:        "external_metadata",
		Type:        reflect.TypeFor[externalmetadata.ExternalMetadataRecord](),
		DriverName:  driverName,
		Description: "External metadata registry store",
		Version:     "1.0.0",
		Capabilities: []string{
			"upsert",
			"dataset_lookup",
			"pagination",
			"filtering",
		},
	}

	externalMetadataFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}

		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("external metadata store requires Postgres, got %s", exec.Kind())
		}

		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}

		return externalmetadata.NewPGExternalMetadataStore(pgExec), nil
	}

	if err := RegisterStore(driverName, "external_metadata", externalMetadata, externalMetadataFactory); err != nil {
		gl.Errorf("Failed to register external metadata store: %v", err)
	}

	// Registra PropertyStore (Colonial NetImóveis)
	propertyMetadata := &t.StoreMetadata{
		Name:        "property",
		Type:        reflect.TypeFor[propertystore.Property](),
		DriverName:  driverName,
		Description: "Property management store (Colonial NetImóveis)",
		Version:     "1.0.0",
		Capabilities: []string{
			"crud",
			"pagination",
			"filtering",
			"image_management",
		},
	}

	propertyFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("property store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return propertystore.NewPGPropertyStore(pgExec), nil
	}

	if err := RegisterStore(driverName, "property", propertyMetadata, propertyFactory); err != nil {
		gl.Errorf("Failed to register property store: %v", err)
	}

	// Registra PropertyShareStore (Colonial NetImóveis)
	propertyShareMetadata := &t.StoreMetadata{
		Name:        "property_share",
		Type:        reflect.TypeFor[propertysharesstore.ShareRecord](),
		DriverName:  driverName,
		Description: "Property share events store (Colonial NetImóveis)",
		Version:     "1.0.0",
		Capabilities: []string{
			"append",
			"phone_lookup",
			"aggregation",
		},
	}

	propertyShareFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("property share store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return propertysharesstore.NewPGPropertyShareStore(pgExec), nil
	}

	if err := RegisterStore(driverName, "property_share", propertyShareMetadata, propertyShareFactory); err != nil {
		gl.Errorf("Failed to register property share store: %v", err)
	}

	// Registra WappMessageStore (Colonial NetImóveis)
	wappMessageMetadata := &t.StoreMetadata{
		Name:        "wapp_message",
		Type:        reflect.TypeFor[wappmessagestore.WappMessage](),
		DriverName:  driverName,
		Description: "WhatsApp threaded message store (Colonial NetImóveis)",
		Version:     "1.0.0",
		Capabilities: []string{
			"append",
			"thread_listing",
			"phone_lookup",
		},
	}

	wappMessageFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("wapp message store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return wappmessagestore.NewPGWappMessageStore(pgExec), nil
	}

	if err := RegisterStore(driverName, "wapp_message", wappMessageMetadata, wappMessageFactory); err != nil {
		gl.Errorf("Failed to register wapp message store: %v", err)
	}

	// Registra LeadConversationStore (Colonial NetImóveis)
	leadConvMetadata := &t.StoreMetadata{
		Name:        "lead_conversation",
		Type:        reflect.TypeFor[leadconversationstore.LeadConversation](),
		DriverName:  driverName,
		Description: "Lead qualification conversation store (Colonial NetImóveis)",
		Version:     "1.0.0",
		Capabilities: []string{
			"crud",
			"pagination",
			"filtering",
			"status_update",
		},
	}

	leadConvFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("lead conversation store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return leadconversationstore.NewPGLeadConversationStore(pgExec), nil
	}

	if err := RegisterStore(driverName, "lead_conversation", leadConvMetadata, leadConvFactory); err != nil {
		gl.Errorf("Failed to register lead conversation store: %v", err)
	}

	// Registra WappWebhookStore (eventos brutos de webhook Meta)
	wappWebhookMetadata := &t.StoreMetadata{
		Name:         "wapp_webhook",
		Type:         reflect.TypeFor[wappwebhookstore.WebhookMessage](),
		DriverName:   driverName,
		Description:  "Raw WhatsApp webhook event store (public.whatsapp_messages + whatsapp_statuses)",
		Version:      "1.0.0",
		Capabilities: []string{"append", "idempotent"},
	}
	wappWebhookFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("wapp webhook store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return wappwebhookstore.NewPGWappWebhookStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "wapp_webhook", wappWebhookMetadata, wappWebhookFactory); err != nil {
		gl.Errorf("Failed to register wapp webhook store: %v", err)
	}

	// Registra RoleStore (leitura de roles do RBAC)
	roleMetadata := &t.StoreMetadata{
		Name:         "role",
		Type:         reflect.TypeFor[rolestore.Role](),
		DriverName:   driverName,
		Description:  "Role and permission store (RBAC)",
		Version:      "1.0.0",
		Capabilities: []string{"read", "code_lookup"},
	}
	roleFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("role store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return rolestore.NewPGRoleStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "role", roleMetadata, roleFactory); err != nil {
		gl.Errorf("Failed to register role store: %v", err)
	}

	// Registra MembershipStore (tenant_membership com roles e permissões)
	membershipMetadata := &t.StoreMetadata{
		Name:         "membership",
		Type:         reflect.TypeFor[membershipstore.MemberWithPermissions](),
		DriverName:   driverName,
		Description:  "Tenant membership store with role and permission joins",
		Version:      "1.0.0",
		Capabilities: []string{"upsert", "update_role", "list_with_permissions"},
	}
	membershipFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("membership store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return membershipstore.NewPGMembershipStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "membership", membershipMetadata, membershipFactory); err != nil {
		gl.Errorf("Failed to register membership store: %v", err)
	}

	// Registra LeadStore (CRM de leads — Colonial NetImóveis)
	leadMetadata := &t.StoreMetadata{
		Name:         "lead",
		Type:         reflect.TypeFor[leadstore.Lead](),
		DriverName:   driverName,
		Description:  "Lead CRM store (Colonial NetImóveis)",
		Version:      "1.0.0",
		Capabilities: []string{"crud", "phone_lookup", "pagination", "filtering", "notes"},
	}
	leadFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("lead store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return leadstore.NewPGLeadStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "lead", leadMetadata, leadFactory); err != nil {
		gl.Errorf("Failed to register lead store: %v", err)
	}

	// Registra PipelineCardStore (kanban CRM — Colonial NetImóveis)
	pipelineCardMetadata := &t.StoreMetadata{
		Name:         "pipeline_card",
		Type:         reflect.TypeFor[pipelinecardstore.PipelineCard](),
		DriverName:   driverName,
		Description:  "Pipeline kanban card store (Colonial NetImóveis)",
		Version:      "1.0.0",
		Capabilities: []string{"crud", "column_move", "reorder"},
	}
	pipelineCardFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("pipeline card store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return pipelinecardstore.NewPGPipelineCardStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "pipeline_card", pipelineCardMetadata, pipelineCardFactory); err != nil {
		gl.Errorf("Failed to register pipeline card store: %v", err)
	}

	// Registra CommissionStore (módulo financeiro — Colonial NetImóveis)
	commissionMetadata := &t.StoreMetadata{
		Name:         "commission",
		Type:         reflect.TypeFor[commissionstore.Commission](),
		DriverName:   driverName,
		Description:  "Commission tracking store (Colonial NetImóveis)",
		Version:      "1.0.0",
		Capabilities: []string{"crud", "pagination", "filtering", "sum_by_period"},
	}
	commissionFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("commission store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return commissionstore.NewPGCommissionStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "commission", commissionMetadata, commissionFactory); err != nil {
		gl.Errorf("Failed to register commission store: %v", err)
	}

	// Registra LeadEventStore (timeline de atividades — Colonial NetImóveis)
	leadEventMetadata := &t.StoreMetadata{
		Name:         "lead_event",
		Type:         reflect.TypeFor[leadeventstore.LeadEvent](),
		DriverName:   driverName,
		Description:  "Lead activity timeline store (Colonial NetImóveis)",
		Version:      "1.0.0",
		Capabilities: []string{"append", "list_by_lead"},
	}
	leadEventFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("lead event store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return leadeventstore.NewPGLeadEventStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "lead_event", leadEventMetadata, leadEventFactory); err != nil {
		gl.Errorf("Failed to register lead event store: %v", err)
	}

	// Registra NotificationStore (centro de notificações — Colonial NetImóveis)
	notifMetadata := &t.StoreMetadata{
		Name:         "notification",
		Type:         reflect.TypeFor[notificationstore.Notification](),
		DriverName:   driverName,
		Description:  "Notification store (Colonial NetImóveis) — Calm UX",
		Version:      "1.0.0",
		Capabilities: []string{"append", "mark_read", "unread_count", "pagination"},
	}
	notifFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("notification store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return notificationstore.NewPGNotificationStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "notification", notifMetadata, notifFactory); err != nil {
		gl.Errorf("Failed to register notification store: %v", err)
	}

	// Registra NotificationPrefStore (preferências Calm UX — mute-by-default)
	notifPrefMetadata := &t.StoreMetadata{
		Name:         "notification_pref",
		Type:         reflect.TypeFor[notificationstore.NotificationPref](),
		DriverName:   driverName,
		Description:  "Notification preferences store (mute-by-default) — Colonial NetImóveis",
		Version:      "1.0.0",
		Capabilities: []string{"upsert", "list_by_user"},
	}
	notifPrefFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("notification pref store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return notificationstore.NewPGNotificationPrefStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "notification_pref", notifPrefMetadata, notifPrefFactory); err != nil {
		gl.Errorf("Failed to register notification pref store: %v", err)
	}

	// Registra TransactionStore (negociações imobiliárias — Colonial NetImóveis)
	txMetadata := &t.StoreMetadata{
		Name:         "transaction",
		Type:         reflect.TypeFor[transactionstore.Transaction](),
		DriverName:   driverName,
		Description:  "Transaction store (Colonial NetImóveis) — negociações imobiliárias",
		Version:      "1.0.0",
		Capabilities: []string{"crud", "pagination", "filtering", "status_lifecycle"},
	}
	txFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("transaction store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return transactionstore.NewPGTransactionStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "transaction", txMetadata, txFactory); err != nil {
		gl.Errorf("Failed to register transaction store: %v", err)
	}

	// Registra TransactionStepStore (checklist de etapas — Colonial NetImóveis)
	stepMetadata := &t.StoreMetadata{
		Name:         "transaction_step",
		Type:         reflect.TypeFor[transactionstore.TransactionStep](),
		DriverName:   driverName,
		Description:  "Transaction step checklist store (Colonial NetImóveis)",
		Version:      "1.0.0",
		Capabilities: []string{"list", "update", "seed_defaults"},
	}
	stepFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("transaction step store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return transactionstore.NewPGTransactionStepStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "transaction_step", stepMetadata, stepFactory); err != nil {
		gl.Errorf("Failed to register transaction step store: %v", err)
	}

	// Registra TransactionDocStore (documentos do processo — Colonial NetImóveis)
	docMetadata := &t.StoreMetadata{
		Name:         "transaction_doc",
		Type:         reflect.TypeFor[transactionstore.TransactionDoc](),
		DriverName:   driverName,
		Description:  "Transaction document store (Colonial NetImóveis) — upload + LLM extraction",
		Version:      "1.0.0",
		Capabilities: []string{"append", "list", "llm_extraction"},
	}
	docFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("transaction doc store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return transactionstore.NewPGTransactionDocStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "transaction_doc", docMetadata, docFactory); err != nil {
		gl.Errorf("Failed to register transaction doc store: %v", err)
	}

	// Registra VisitEventStore (agenda de visitas — Colonial NetImóveis)
	visitMetadata := &t.StoreMetadata{
		Name:         "visit_event",
		Type:         reflect.TypeFor[visiteventstore.VisitEvent](),
		DriverName:   driverName,
		Description:  "Visit agenda store (Colonial NetImóveis) — agendamento + roteiro",
		Version:      "1.0.0",
		Capabilities: []string{"crud", "pagination", "filtering", "list_by_day"},
	}
	visitFactory := func(ctx context.Context, drv t.Driver) (t.StoreType, error) {
		exec, err := drv.Executor(ctx)
		if err != nil {
			return nil, err
		}
		if exec.Kind() != execution.BackendPostgres {
			return nil, gl.Errorf("visit event store requires Postgres, got %s", exec.Kind())
		}
		pgExec := exec.PG()
		if pgExec == nil {
			return nil, gl.Error("PGExecutor is nil")
		}
		return visiteventstore.NewPGVisitEventStore(pgExec), nil
	}
	if err := RegisterStore(driverName, "visit_event", visitMetadata, visitFactory); err != nil {
		gl.Errorf("Failed to register visit event store: %v", err)
	}

	gl.Noticef("Registered %d stores for driver: %s", len(ListStoresByDriver(driverName)), driverName)
}

// registerMongoStores registra stores para MongoDB (placeholder).
func registerMongoStores() {
	// TODO: Implementar quando houver stores MongoDB
	// Exemplo:
	// RegisterStore("mongo", "user", metadata, factory)
}

// registerRedisStores registra stores para Redis (placeholder).
func registerRedisStores() {
	// TODO: Implementar quando houver stores Redis
	// Exemplo:
	// RegisterStore("redis", "cache", metadata, factory)
}

// registerRabbitStores registra stores para RabbitMQ (placeholder).
func registerRabbitStores() {
	// TODO: Implementar quando houver stores RabbitMQ
	// Exemplo:
	// RegisterStore("rabbit", "queue", metadata, factory)
}
