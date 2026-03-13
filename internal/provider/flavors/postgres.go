// Package flavors implements database drivers and validators for different flavors.
package flavors

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kubex-ecosystem/domus/internal/execution"
	"github.com/kubex-ecosystem/domus/internal/types"

	kbxMod "github.com/kubex-ecosystem/domus/internal/module/kbx"
	gl "github.com/kubex-ecosystem/logz"
)

// PostgresValidator garante um mínimo de sanidade na config.
type PostgresValidator struct{}

func (v *PostgresValidator) Validate(cfg *kbxMod.DBConfig) error {
	if cfg.Host == "" {
		return gl.GetLoggerZ("postgres-driver").Errorf("host is required")
	}
	if cfg.Port == "" {
		return gl.GetLoggerZ("postgres-driver").Errorf("port is required")
	}
	if cfg.User == "" {
		return gl.GetLoggerZ("postgres-driver").Errorf("user is required")
	}
	if cfg.DBName == "" {
		return gl.GetLoggerZ("postgres-driver").Errorf("db_name is required")
	}
	return nil
}

// PostgresDriver implementa Driver usando pgxpool.
type PostgresDriver struct {
	logger   *gl.LoggerZ
	pool     *pgxpool.Pool
	executor execution.Executor
}

func NewPostgresDriver(logger *gl.LoggerZ) types.Driver {
	if logger == nil {
		logger = gl.GetLoggerZ("postgres-driver")
	}
	return &PostgresDriver{
		logger: logger,
	}
}

func (d *PostgresDriver) Name() string { return "postgres" }

func (d *PostgresDriver) Connect(ctx context.Context, cfg *kbxMod.DBConfig) error {
	dsn := types.NewDSNFromDBConfig[*PostgresDriver](*cfg)
	if err := dsn.Validate(); err != nil {
		if len(cfg.DSN) > 0 {
			if err = dsn.Parse(cfg.DSN); err != nil {
				return d.logger.Errorf("dsn parse failed: %v", err)
			}
		} else {
			return d.logger.Errorf("dsn validation failed: %v", err)
		}
	}

	conf, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return d.logger.Errorf("pgxpool.ParseConfig: %v", err)
	}

	// Pool tuning a partir de Options (opcional)
	maxConn, ok := dsn.GetOption("max_connections")
	if ok {
		conf.MaxConns = maxConn.(int32)
	}
	poolMaxLifetime, ok := dsn.GetOption("pool_max_lifetime")
	if ok {
		if dur, err := time.ParseDuration(poolMaxLifetime.(string)); err == nil {
			conf.MaxConnLifetime = dur
		}
	}

	// Create pool
	d.pool, err = pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return d.logger.Errorf("pgxpool.NewWithConfig: %v", err)
	}

	// Check connection
	ctxPing, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := d.pool.Ping(ctxPing); err != nil {
		d.pool.Close()
		return d.logger.Errorf("postgres ping failed: %v", err)
	}

	// Log the successful connection
	d.logger.Debugf("connected to postgres: %s", dsn.Redacted())
	return nil
}

func (d *PostgresDriver) Ping(ctx context.Context) bool {
	if d.pool != nil {
		ctxPing, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := d.pool.Ping(ctxPing); err != nil {
			d.logger.Debugf("postgres ping failed: %v", err)
			return false
		}
		return true
	}
	d.logger.Debugf("postgres pool is not initialized")
	return false
}

func (d *PostgresDriver) Close() error {
	if d.pool != nil {
		d.pool.Close()
		d.pool = nil
	}
	return nil
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
