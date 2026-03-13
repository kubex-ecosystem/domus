package datastore

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kubex-ecosystem/domus/internal/engine"
	"github.com/kubex-ecosystem/domus/internal/execution"
	"github.com/kubex-ecosystem/domus/internal/provider/flavors"

	c "github.com/kubex-ecosystem/domus/internal/datastore/company_store"
	externalmetadata "github.com/kubex-ecosystem/domus/internal/datastore/external_metadata_store"
	integration "github.com/kubex-ecosystem/domus/internal/datastore/integration_store"
	i "github.com/kubex-ecosystem/domus/internal/datastore/invite_store"
	p "github.com/kubex-ecosystem/domus/internal/datastore/pending_access_store"
	s "github.com/kubex-ecosystem/domus/internal/datastore/schemas_store"
	u "github.com/kubex-ecosystem/domus/internal/datastore/user_store"
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

	mgr := engine.NewDatabaseManagerType(gl.GetLoggerZ("invite-store"))
	cnn, ok := mgr.GetDefault()
	if !ok {
		return nil, fmt.Errorf("default connection not found")
	}
	exec, err := cnn.Driver.Executor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get executor from driver: %v", err)
	}
	store, err := NewInviteStoreFromExecutor(exec)
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
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}

	if exec.Kind() != execution.BackendPostgres {
		return nil, fmt.Errorf("user store requires Postgres backend, got %s", exec.Kind())
	}

	pgExec := exec.PG()
	if pgExec == nil {
		return nil, fmt.Errorf("PGExecutor is nil")
	}

	return u.NewPGUserStore(pgExec), nil
}

// NewInviteStoreFromExecutor cria InviteStore diretamente de um Executor.
func NewInviteStoreFromExecutor(exec execution.Executor) (i.InviteStore, error) {
	if exec == nil {
		return nil, gl.Error("executor is nil (invite store from executor)")
	}

	if exec.Kind() != execution.BackendPostgres {
		return nil, gl.Error("invite store requires Postgres backend, got %s", exec.Kind())
	}

	pgExec := exec.PG()
	if pgExec == nil {
		return nil, gl.Error("PGExecutor is nil (invite store from executor)")
	}

	return i.NewPGInviteStore(pgExec), nil
}

// NewCompanyStoreFromExecutor cria CompanyStore diretamente de um Executor.
func NewCompanyStoreFromExecutor(exec execution.Executor) (c.CompanyStore, error) {
	if exec == nil {
		return nil, fmt.Errorf("executor is nil")
	}

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
