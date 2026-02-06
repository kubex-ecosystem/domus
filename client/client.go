// Package client provides structures and types for database client configuration and management.
// It defines configurations for multiple backend databases and the main DS client.
package client

import (
	"context"
	"os"

	"github.com/kubex-ecosystem/domus/internal/adapter"
	st "github.com/kubex-ecosystem/domus/internal/datastore"
	company "github.com/kubex-ecosystem/domus/internal/datastore/company_store"
	invite "github.com/kubex-ecosystem/domus/internal/datastore/invite_store"
	pendingaccess "github.com/kubex-ecosystem/domus/internal/datastore/pending_access_store"
	user "github.com/kubex-ecosystem/domus/internal/datastore/user_store"
	en "github.com/kubex-ecosystem/domus/internal/engine"
	kbxMod "github.com/kubex-ecosystem/domus/internal/module/kbx"
	t "github.com/kubex-ecosystem/domus/internal/types"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

// StoreType é um alias para a interface base de stores.
type StoreType = t.StoreType

// Type aliases for specific stores (re-export from stores.go)
type (
	UserStore          = user.UserStore
	InviteStore        = invite.InviteStore
	CompanyStore       = company.CompanyStore
	PendingAccessStore = pendingaccess.PendingAccessStore
)

// BackendConnections is an alias for the database manager type.
type BackendConnections = en.DatabaseManager

// BackendConnection is an alias for the database connection type.
type BackendConnection = t.DBConnection

// BackendCurrentDriver is an alias for the current database driver type.
type BackendCurrentDriver = t.Driver

// BackendDSConfigMap holds the configuration for the DS client managing multiple backends.
type BackendDSConfigMap = map[string]t.DBConfig

// DSClient defines the interface for the data service client managing multiple backend connections.
type DSClient interface {
	Init(ctx context.Context) error
	GetConn(ctx context.Context, name string) (*BackendConnection, error)
	Config(ctx context.Context) *DSClientConfig
	ConfigPath(ctx context.Context) string
	GetReference(ctx context.Context) kbxMod.Reference
	Close(ctx context.Context) error

	// Driver access
	GetDriver(ctx context.Context, name string) (BackendCurrentDriver, error)

	// Generic store access (by name)
	GetStore(ctx context.Context, dbName, storeName string) (StoreType, error)

	// Typed store helpers
	GetUserStore(ctx context.Context, dbName string) (UserStore, error)
	GetInviteStore(ctx context.Context, dbName string) (InviteStore, error)
	GetCompanyStore(ctx context.Context, dbName string) (CompanyStore, error)
	GetPendingAccessStore(ctx context.Context, dbName string) (PendingAccessStore, error)

	// Adapter factory methods
	NewAdapterFactory(ctx context.Context, dbName string, db *BackendConnection, config *adapter.RepositoryConfig) (*adapter.AdapterFactory, error)
	NewStoreOnlyAdapterFactory(ctx context.Context, dbName string) (*adapter.AdapterFactory, error)
	NewORMOnlyAdapterFactory(db *BackendConnection, config *adapter.RepositoryConfig) (*adapter.AdapterFactory, error)
}



// DSClientImpl represents the data service client managing multiple backend connections.
type DSClientImpl struct {
	// Logger for logging purposes.
	logger *gl.LoggerZ

	// Reference is the reference information for the DS client.
	kbxMod.Reference `yaml:",inline" json:",inline" mapstructure:",squash"`

	// Backends holds the configurations for multiple backends.
	dsConfig *DSClientConfig `yaml:"-" json:"-" mapstructure:"-"`

	// mgr is the database manager for handling multiple connections.
	mgr *BackendConnections `yaml:"-" json:"-" mapstructure:"-"`
}

// NewDSClientImpl creates a new DSClientImpl instance.
func NewDSClientImpl(ctx context.Context, configPath string, dsConfig *DSClientConfig, logger *gl.LoggerZ) *DSClientImpl {

	// Create the database manager.
	mgr := en.NewDatabaseManager(logger)

	// Return the new DSClientImpl instance with the initialized manager.
	return &DSClientImpl{
		Reference: kbxMod.NewReference(configPath),
		logger:    logger,
		mgr:       mgr,
		dsConfig:  dsConfig,
	}
}

// NewDSClient creates a new DSClient instance.
func NewDSClient(ctx context.Context, configPath string, dsConfig *DSClientConfig, logger *gl.LoggerZ) DSClient {
	return NewDSClientImpl(ctx, configPath, dsConfig, logger)
}

// Init initializes the DS client by loading the configuration and setting up the database manager.
func (c *DSClientImpl) Init(ctx context.Context) error {
	// Validate basic fields.
	if err := c.validateBasics(); err != nil {
		return err
	}

	if c.dsConfig.FilePath == "" {
		gl.Debug("DS config path not set, using default path from env or default constant")
		c.dsConfig.FilePath = os.ExpandEnv(kbxGet.EnvOr("KUBEX_DS_CONFIG_PATH", kbxMod.DefaultKubexDomusConfigPath))
	}

	// Load the configuration for the DS client.
	rootConfig, err := en.LoadRootConfig(c.dsConfig.FilePath)
	if err != nil {
		// c.logger.Fatalf("Failed to load config (%s): %v", c.dsConfig.FilePath, err)
		return gl.Errorf("Failed to load DS config (%s): %v", c.dsConfig.FilePath, err)
	}

	// Initialize the manager with the loaded configuration.
	if err := c.mgr.InitFromRootConfig(ctx, &rootConfig); err != nil {
		// c.logger.Fatalf("Failed to initialize database manager (%s): %v", c.dsConfig.FilePath, err)
		return gl.Errorf("Failed to initialize DS database manager (%s): %v", c.dsConfig.FilePath, err)
	}

	return nil
}

// GetConn retrieves a backend connection by its name.
func (c *DSClientImpl) GetConn(ctx context.Context, name string) (*BackendConnection, error) {
	// Validate basic fields.
	if err := c.validateBasics(); err != nil {
		return nil, err
	}
	return c.mgr.SecureConn(ctx, name)
}

// MustGetConn retrieves a backend connection by its name and panics if an error occurs.
func (c *DSClientImpl) MustGetConn(name string) *BackendConnection {
	// Validate basic fields.
	if err := c.validateBasics(); err != nil {
		return nil
	}
	conn, err := c.GetConn(context.TODO(), name)
	if err != nil {
		panic(err)
	}
	return conn
}

// Config returns the DS client configuration.
func (c *DSClientImpl) Config(ctx context.Context) *DSClientConfig {
	return c.dsConfig
}

// GetReference returns the reference information for the DS client.
func (c *DSClientImpl) GetReference(ctx context.Context) kbxMod.Reference {
	return c.Reference
}

// Close closes all backend connections managed by the DS client.
func (c *DSClientImpl) Close(ctx context.Context) error {
	return c.mgr.Shutdown(ctx)
}

// GetStore obtém um store por nome para o banco de dados especificado.
// Usa o registry interno para lookup dinâmico e criação via factory.
func (c *DSClientImpl) GetStore(ctx context.Context, dbName, storeName string) (StoreType, error) {
	// Validate basic fields.
	if err := c.validateBasics(); err != nil {
		return nil, err
	}

	// Obtém conexão
	conn, err := c.mgr.SecureConn(ctx, dbName)
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return nil, gl.Error("Database connection is nil for db: %s", dbName)
	}

	// Cria factory para o driver
	factory := st.NewStoreFactory(conn.Driver)

	// Cria store usando factory
	store, err := factory.Create(ctx, storeName)
	if err != nil {
		return nil, gl.Error("Failed to create store %s: %v", storeName, err)
	}

	// Valida store antes de retornar
	if err := c.validateBasics(); err != nil {
		return nil, gl.Error("Store validation failed: %v", err)
	}

	return store, nil
}

func (c *DSClientImpl) ConfigPath(ctx context.Context) string { return c.dsConfig.FilePath }

func (c *DSClientImpl) GetDriver(ctx context.Context, name string) (BackendCurrentDriver, error) {
	// Validate basic fields.
	if err := c.validateBasics(); err != nil {
		return nil, err
	}
	drv, err := c.mgr.SecureConn(ctx, name)
	if err != nil {
		return nil, err
	}
	if drv == nil {
		return nil, gl.Error("Database connection is nil for db: %s", name)
	}
	return drv.Driver, nil
}

func (c *DSClientImpl) validateBasics() error {
	// Validate essential fields.
	if c.Reference == (kbxMod.Reference{}) {
		c.logger.Fatal("DS client reference is nil")
	}
	if c.dsConfig == nil {
		c.logger.Fatal("DS client configuration is nil")
	}
	if c.mgr == nil {
		c.logger.Fatal("Database manager is nil")
	}
	return nil
}

// NewAdapterFactory cria uma nova AdapterFactory com driver e opcionalmente GORM DB.
//
// Esta factory permite criar DSRepository[T] que podem usar tanto Store (PGExecutor)
// quanto ORM (GORM) de forma transparente, com fallback automático.
//
// Parâmetros:
// - ctx: contexto para operações
// - dbName: nome do banco de dados configurado no DSClient
// - db: instância *gorm.DB (opcional, pode ser nil para usar apenas Store)
// - config: configuração de política de uso (opcional, usa DefaultConfig se nil)
//
// Retorna erro se dbName não existir ou driver não estiver disponível.
func (c *DSClientImpl) NewAdapterFactory(ctx context.Context, dbName string, db *BackendConnection, config *adapter.RepositoryConfig) (*adapter.AdapterFactory, error) {
	if err := c.validateBasics(); err != nil {
		return nil, err
	}

	// Obtém driver para o dbName
	driver, err := c.GetDriver(ctx, dbName)
	if err != nil {
		return nil, gl.Error("Failed to get driver for db %s: %v", dbName, err)
	}

	// Cria factory com driver e opcionalmente db
	return adapter.NewAdapterFactory(driver, db, config)
}

// NewStoreOnlyAdapterFactory cria uma AdapterFactory que usa APENAS stores (sem ORM).
//
// Ideal para novos serviços que querem usar exclusivamente a camada de Store
// com PGExecutor e queries manuais.
//
// Parâmetros:
// - ctx: contexto para operações
// - dbName: nome do banco de dados configurado no DSClient
//
// Retorna erro se dbName não existir ou driver não estiver disponível.
func (c *DSClientImpl) NewStoreOnlyAdapterFactory(ctx context.Context, dbName string) (*adapter.AdapterFactory, error) {
	if err := c.validateBasics(); err != nil {
		return nil, err
	}

	driver, err := c.GetDriver(ctx, dbName)
	if err != nil {
		return nil, gl.Error("Failed to get driver for db %s: %v", dbName, err)
	}

	return adapter.NewStoreOnlyFactory(driver)
}

// NewORMOnlyAdapterFactory cria uma AdapterFactory que usa APENAS ORM (sem stores).
//
// Útil para código legado que ainda usa GORM e precisa gradualmente
// migrar para a arquitetura de adapters sem quebrar funcionalidade existente.
//
// Parâmetros:
// - db: instância *gorm.DB configurada
// - config: configuração de política (opcional, usa config ForceORM se nil)
//
// Retorna erro se db for nil.
func (c *DSClientImpl) NewORMOnlyAdapterFactory(db *BackendConnection, config *adapter.RepositoryConfig) (*adapter.AdapterFactory, error) {
	if err := c.validateBasics(); err != nil {
		return nil, err
	}

	return adapter.NewORMOnlyFactory(db)
}
