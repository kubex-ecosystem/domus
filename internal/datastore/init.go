package datastore

import (
	"context"
	"reflect"

	company "github.com/kubex-ecosystem/domus/internal/datastore/company_store"
	externalmetadata "github.com/kubex-ecosystem/domus/internal/datastore/external_metadata_store"
	invite "github.com/kubex-ecosystem/domus/internal/datastore/invite_store"
	pendingaccess "github.com/kubex-ecosystem/domus/internal/datastore/pending_access_store"
	userstore "github.com/kubex-ecosystem/domus/internal/datastore/user_store"
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
