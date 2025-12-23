// Package flavors implements database drivers and validators for different flavors.
package flavors

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kubex-ecosystem/domus/internal/execution"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/types"
	logz "github.com/kubex-ecosystem/logz"
)

// PostgresValidator garante um mínimo de sanidade na config.
type PostgresValidator struct{}

func (v *PostgresValidator) Validate(cfg *kbx.DBConfig) error {
	if cfg.Host == "" {
		return logz.GetLoggerZ("postgres-driver").Errorf("host is required")
	}
	if cfg.Port == "" {
		return logz.GetLoggerZ("postgres-driver").Errorf("port is required")
	}
	if cfg.User == "" {
		return logz.GetLoggerZ("postgres-driver").Errorf("user is required")
	}
	if cfg.DBName == "" {
		return logz.GetLoggerZ("postgres-driver").Errorf("db_name is required")
	}
	return nil
}

// PostgresDriver implementa Driver usando pgxpool.
type PostgresDriver struct {
	logger   *logz.LoggerZ
	pool     *pgxpool.Pool
	executor execution.Executor
}

func NewPostgresDriver(logger *logz.LoggerZ) types.Driver {
	if logger == nil {
		logger = logz.GetLoggerZ("postgres-driver")
	}
	return &PostgresDriver{
		logger: logger,
	}
}

func (d *PostgresDriver) Connect(ctx context.Context, cfg *types.DBConfig) error {
	connStr := cfg.DSN
	if connStr == "" {
		sslmode := "disable"
		if v, ok := cfg.Options["sslmode"].(string); ok && v != "" {
			sslmode = v
		}
		connStr = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.User,
			cfg.Pass,
			cfg.Host,
			cfg.Port,
			cfg.DBName,
			sslmode,
		)
	}

	conf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return d.logger.Errorf("pgxpool.ParseConfig: %v", err)
	}

	// Pool tuning a partir de Options (opcional)
	if v, ok := cfg.Options["max_connections"].(int); ok && v > 0 {
		conf.MaxConns = int32(v)
	}
	if v, ok := cfg.Options["pool_max_lifetime"].(string); ok && v != "" {
		if dur, err := time.ParseDuration(v); err == nil {
			conf.MaxConnLifetime = dur
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return d.logger.Errorf("pgxpool.NewWithConfig: %v", err)
	}

	// Teste rápido de conexão
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return d.logger.Errorf("postgres ping failed: %v", err)
	}

	d.logger.Debugf("connected to postgres: conn-name=%s db=%s host=%s port=%s", cfg.Name, cfg.DBName, cfg.Host, cfg.Port)
	d.pool = pool
	return nil
}

func (d *PostgresDriver) Ping(ctx context.Context) bool {
	if d.pool == nil {
		return false
	}
	ctxPing, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := d.pool.Ping(ctxPing); err != nil {
		d.logger.Warnf("postgres ping failed: %v", err)
		return false
	}
	return true
}

func (d *PostgresDriver) Close() error {
	if d.pool != nil {
		d.pool.Close()
		d.pool = nil
	}
	return nil
}

func (d *PostgresDriver) Name() string {
	return "postgres"
}

// Executor integration exported method to get executor pool
func (d *PostgresDriver) Executor(ctx context.Context) (execution.Executor, error) {
	if d.executor != nil {
		return d.executor, nil
	}
	if d.pool == nil {
		return nil, d.logger.Errorf("executor requested but pool is not initialized")
	}
	if !d.Ping(ctx) {
		return nil, d.logger.Errorf("executor requested but ping failed")
	}

	// Cria o PGExecutor a partir do pool atual
	pgExec := execution.NewPGExecutor(d.pool)
	d.executor = execution.NewExecutor(
		execution.WithKind(execution.BackendPostgres),
		execution.WithPG(pgExec),
	)

	return d.executor, nil
}
