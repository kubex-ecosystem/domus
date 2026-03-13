// Package remotestack provides Docker-based backend implementation
package remotestack

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kubex-ecosystem/domus/internal/engine"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/provider"
	"github.com/kubex-ecosystem/domus/internal/types"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
	logz "github.com/kubex-ecosystem/logz"
)

// RemoteStackProvider wraps legacy Docker services into new Provider interface.
// It implements Provider, MigratableProvider, and RootConfigProvider interfaces.
type RemoteStackProvider struct {
	logger *logz.LoggerZ
}

// NewRemoteStackProvider creates a new RemoteStack provider.
func NewRemoteStackProvider() *RemoteStackProvider {
	return &RemoteStackProvider{
		logger: logz.NewLogger("domus"),
	}
}

// Name returns the provider name
func (p *RemoteStackProvider) Name() string {
	return "remotestack"
}

// Capabilities returns what this provider can do
func (p *RemoteStackProvider) Capabilities(ctx context.Context) (provider.Capabilities, error) {
	return provider.Capabilities{
		Managed: true, // Remote managed stack
		Notes: []string{
			"Remote stack using remote database connections",
			"Supports PostgreSQL, MongoDB, Redis, RabbitMQ",
		},
		Features: map[string]bool{
			"network.internal": false,
			"publish.ports":    false,
			"volumes.persist":  false,
			"migrations":       true,
		},
	}, nil
}

// Start provisions or attaches services and returns ready endpoints.
// This implements the Provider interface without handling migrations.
// Use StartServices() for complete orchestration including migrations.
func (p *RemoteStackProvider) Start(ctx context.Context, spec provider.StartSpec) (map[string]provider.Endpoint, error) {
	// 1. Convert provider.StartSpec to legacy DBConfig format
	cfg := p.ConvertSpecToManager(spec)

	// 2. Extract endpoints from running containers
	endpoints, err := p.ExtractEndpoints(&cfg)
	if err != nil {
		return nil, logz.Errorf("failed to extract endpoints: %v", err)
	}

	logz.Debugf("Checking options and capabilities for tasks")

	return endpoints, nil
}

// ConvertSpecToManager converts new StartSpec to DatabaseManager
func (p *RemoteStackProvider) ConvertSpecToManager(spec provider.StartSpec) engine.DatabaseManager {
	dbManager := engine.DatabaseManager{Conns: make(map[string]types.DBConnection)}

	for _, svc := range spec.Services {
		configs := provider.GetConfigListByService(spec, svc.Name)
		if len(configs) == 0 {
			continue
		}

		for _, dbConfig := range configs {
			fnDrvr, ok := engine.GetDriver(string(dbConfig.Protocol))
			if !ok {
				continue
			}
			dbManager.Conns[dbConfig.Name] = types.DBConnection{
				Config: types.DBConfigRT{
					Config:  dbConfig,
					Mutexes: types.NewMutexesType(),
				},
				Driver: fnDrvr(p.logger),
			}
		}
	}

	return dbManager
}

func (p *RemoteStackProvider) ConvertDBConfigToSpec(dbConfig *kbx.DBConfig) (*provider.StartSpec, error) {
	spec := &provider.StartSpec{
		Services: []provider.ServiceRef{
			{
				Name:   dbConfig.Name,
				Engine: provider.Engine(dbConfig.Protocol),
			},
		},
		PreferredPort: map[string]int{
			dbConfig.Name: func() int {
				port, err := strconv.Atoi(dbConfig.Port)
				if err != nil {
					return 0
				}
				return port
			}(),
		},
		Secrets: map[string]string{
			"pg_admin": dbConfig.Pass,
		},
		Labels: map[string]string{
			"app": dbConfig.Name,
		},
	}

	return spec, nil
}

// ExtractEndpoints converts legacy DBConfig to new Endpoint format
func (p *RemoteStackProvider) ExtractEndpoints(cfg *engine.DatabaseManager) (map[string]provider.Endpoint, error) {
	endpoints := make(map[string]provider.Endpoint)

	if cfg == nil {
		return nil, nil
	}
	if len(cfg.Conns) == 0 {
		cn, ok := cfg.GetDefault()
		if !ok {
			return nil, logz.Error("no default database connection found")
		}
		cfg.Conns[cn.Config.Config.Name] = cn
	}

	for _, db := range cfg.Conns {
		endpoints[db.Config.Config.Name] = provider.BuildEndpoint(&db.Config.Config)
	}

	return endpoints, nil
}

// Health verifies connectivity to all services
func (p *RemoteStackProvider) Health(ctx context.Context, eps map[string]provider.Endpoint) error {
	// TODO: Implement real health checks
	return nil
}

// Stop stops all managed containers
func (p *RemoteStackProvider) Stop(ctx context.Context, refs []provider.ServiceRef) error {
	// TODO: Call docker service stop methods
	return nil
}

func (p *RemoteStackProvider) PrepareMigrations(ctx context.Context, conn *types.DBConnection) error {
	if conn == nil {
		return logz.Error("invalid database connection")
	}
	if conn.Config.Config.Protocol != "postgresql" && conn.Config.Config.Protocol != "postgres" {
		return logz.Error("migrations only supported for PostgreSQL")
	}
	if !conn.Driver.Ping(ctx) {
		return logz.Error("database is not reachable")
	}
	if err := conn.Driver.Connect(ctx, &conn.Config.Config); err != nil {
		return logz.Errorf("failed to connect to database: %v", err)
	}

	migrationManager := NewMigrationManager(conn.Config.Config.DSN, p.logger)

	// Wait for PostgreSQL to be ready
	if err := migrationManager.WaitForPostgres(ctx, 30*time.Second); err != nil {
		return err
	}

	logz.Info("Validating PostgreSQL connection for migrations...")
	if err := migrationManager.ValidateConnection(); err != nil {
		return logz.Errorf("failed to validate connection: %v", err)
	}

	logz.Info("PostgreSQL migrations ready to be executed.")
	return nil
}

func (p *RemoteStackProvider) RunMigrations(ctx context.Context, conn *types.DBConnection, migrationInfo *kbx.MigrationInfo) error {
	if conn == nil {
		return logz.Error("invalid database connection")
	}
	if conn.Config.Config.Protocol != "postgresql" && conn.Config.Config.Protocol != "postgres" {
		return logz.Error("migrations only supported for PostgreSQL")
	}

	migrationManager := NewMigrationManager(conn.Config.Config.DSN, p.logger)
	// Wait for PostgreSQL to be ready
	if err := migrationManager.WaitForPostgres(ctx, 30*time.Second); err != nil {
		return err
	}

	results, err := migrationManager.RunMigrations(ctx, migrationInfo)
	if err != nil {
		return logz.Errorf("migrations failed: %v", err)
	}

	// Log final summary
	totalSuccess := 0
	totalFailed := 0
	for _, r := range results {
		totalSuccess += r.SuccessfulStmts
		totalFailed += r.FailedStmts
	}

	if totalFailed > 0 {
		logz.Warnf("Migration completed with partial success: %d succeeded, %d failed", totalSuccess, totalFailed)
		// Don't return error for partial failures - let the service continue
	} else {
		logz.Infof("All migrations completed successfully! (%d statements)", totalSuccess)
	}

	return nil
}

// StartServices implements RootConfigProvider interface.
// This is the complete orchestration flow that:
// 1. Starts Docker containers for all enabled databases
// 2. Waits for database readiness
// 3. Runs migrations (if auto-migrate is enabled)
// 4. Returns only when everything is ready
func (p *RemoteStackProvider) StartServices(ctx context.Context, rootConfig *kbx.RootConfig) error {
	// Validate inputs
	if rootConfig == nil {
		return logz.Error("rootConfig cannot be nil")
	}

	// ========== STEP 1: WAIT + MIGRATE FOR EACH DATABASE ==========
	for _, dbConf := range rootConfig.Databases {
		// Skip disabled databases
		if !kbx.DefaultFalse(dbConf.Enabled) {
			logz.Debugf("Skipping disabled database: %s", dbConf.Name)
			continue
		}

		// Build DSN if missing
		if dbConf.DSN == "" {
			dbConf.DSN = kbxGet.EnvOr("DATABASE_URL", dbConf.DSN)
		}
		if dbConf.DSN == "" {
			dbConf.DSN = p.buildDSN(&dbConf)
		}

		// Log database processing
		logz.Infof("Processing database: %s (%s)", dbConf.Name, dbConf.Protocol)

		// Wait for database readiness
		logz.Infof("Waiting for database readiness: %s", dbConf.DBName)
		mm := NewMigrationManager(dbConf.DSN, p.logger)

		if err := mm.WaitForPostgres(ctx, 30*time.Second); err != nil {
			return logz.Errorf("database %s not ready: %v", dbConf.DBName, err)
		}

		// Run migrations if enabled
		if dbConf.Migration != nil && kbx.DefaultFalse(dbConf.Migration.Auto) {
			logz.Infof("Running migrations for database: %s", dbConf.Name)

			// Check if schema already exists (skip if so)
			exists, err := mm.SchemaExists()
			if err != nil {
				logz.Warnf("Could not check schema existence: %v", err)
			}

			if exists {
				missingTables, err := mm.MissingTables("public", "external_metadata_registry")
				if err != nil {
					logz.Warnf("Could not check required tables for %s: %v", dbConf.Name, err)
				}
				if len(missingTables) == 0 {
					logz.Infof("Schema already exists for %s, skipping migrations", dbConf.Name)
					continue
				}

				logz.Infof("Schema already exists for %s, but required tables are missing (%s). Running idempotent migrations.", dbConf.Name, strings.Join(missingTables, ", "))
			}

			// Run migrations with error recovery
			results, err := mm.RunMigrations(ctx, dbConf.Migration)
			if err != nil {
				return logz.Errorf("migrations failed for %s: %v", dbConf.Name, err)
			}

			// Summary logging
			totalSuccess := 0
			totalFailed := 0
			for _, r := range results {
				totalSuccess += r.SuccessfulStmts
				totalFailed += r.FailedStmts
			}

			if totalFailed > 0 {
				logz.Warnf("%s: %d succeeded, %d failed",
					dbConf.Name, totalSuccess, totalFailed)
				// Don't return error - allow partial success (resilience)
			} else {
				logz.Successf("%s: all %d statements executed successfully",
					dbConf.Name, totalSuccess)
			}
		} else {
			logz.Infof("Skipping migrations for: %s (auto-migrate disabled)", dbConf.Name)
		}
	}

	logz.Success("All services started and migrated successfully")
	return nil
}

// buildDSN constructs a connection string from DBConfig.
// Helper method to avoid repeating DSN logic.
func (p *RemoteStackProvider) buildDSN(db *kbx.DBConfig) string {
	switch db.Protocol {
	case "postgresql", "postgres", "pg", "domus":
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			db.User, db.Pass, db.Host, db.Port, db.Name)
	case "mongodb", "mongo", "kubex_mdb":
		return fmt.Sprintf("mongodb://%s:%s@%s:%s",
			db.User, db.Pass, db.Host, db.Port)
	case "redis", "kubex_rdb":
		return fmt.Sprintf("redis://:%s@%s:%s",
			db.Pass, db.Host, db.Port)
	case "rabbitmq", "rabbit", "kubex_rmq":
		return fmt.Sprintf("amqp://%s:%s@%s:%s/",
			db.User, db.Pass, db.Host, db.Port)
	default:
		return ""
	}
}

// Note: EndpointRedacted is now available as provider.RedactDSN(dsn) utility function.

func init() {
	// Provider registration is now handled by CLI/main initialization
	// to allow proper dependency injection of dockerService.
	// See cmd/cli/migrate.go for usage pattern.
}
