package flavors

import (
	"context"
	"fmt"
	"time"

	"github.com/kubex-ecosystem/domus/internal/execution"
	"github.com/kubex-ecosystem/domus/internal/types"
	logz "github.com/kubex-ecosystem/logz"
	rds "github.com/redis/go-redis/v9"
)

type RedisDriver struct {
	logger   *logz.LoggerZ
	client   *rds.Client
	executor execution.Executor
}

func NewRedisDriver(logger *logz.LoggerZ) types.Driver {
	return &RedisDriver{
		logger: logger,
	}
}

func (d *RedisDriver) Connect(ctx context.Context, cfg *types.DBConfig) error {
	if cfg == nil {
		return fmt.Errorf("redis: nil config")
	}
	if cfg.DSN == "" {
		return fmt.Errorf("redis: empty DSN")
	}

	opt, err := rds.ParseURL(cfg.DSN)
	if err != nil {
		return fmt.Errorf("redis: parse dsn: %v", err)
	}

	client := rds.NewClient(opt)

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctxPing).Err(); err != nil {
		return fmt.Errorf("redis: ping fail: %v", err)
	}

	d.client = client
	return nil
}

func (d *RedisDriver) Ping(ctx context.Context) bool {
	if d.client == nil {
		return false
	}

	ctxPing, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return d.client.Ping(ctxPing).Err() == nil
}

func (d *RedisDriver) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

func (d *RedisDriver) Name() string {
	return "redis"
}

// Executor integration exported method to get executor pool
func (d *RedisDriver) Executor(ctx context.Context) (execution.Executor, error) {
	if d.executor != nil {
		return d.executor, nil
	}
	if d.client == nil {
		return nil, fmt.Errorf("executor requested but client is not initialized")
	}
	if !d.Ping(ctx) {
		return nil, fmt.Errorf("executor requested but ping failed")
	}

	// // Cria o RedisExecutor a partir do client atual
	// redisExec := execution.NewRedisExecutor(d.client)
	// d.executor = execution.NewExecutor(
	// 	execution.WithKind(execution.BackendRedis),
	// 	execution.WithRedis(redisExec),
	// )

	// return d.executor, nil
	return nil, fmt.Errorf("redis executor not implemented yet")
}
