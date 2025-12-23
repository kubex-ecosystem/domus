package types

import (
	"context"

	"github.com/kubex-ecosystem/domus/internal/engine"
	it "github.com/kubex-ecosystem/domus/internal/interfaces"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	svc "github.com/kubex-ecosystem/domus/internal/types"

	logz "github.com/kubex-ecosystem/logz"
)

const (
	// DefaultConfigDir is the default directory for configuration files
	DefaultConfigDir = kbx.DefaultConfigDir

	// DefaultConfigFile is the default configuration file path
	DefaultConfigFile     = kbx.DefaultConfigFile
	DefaultVolumesDir     = kbx.DefaultVolumesDir
	DefaultRedisVolume    = kbx.DefaultRedisVolume
	DefaultPostgresVolume = kbx.DefaultPostgresVolume
	DefaultMongoVolume    = kbx.DefaultMongoVolume
	DefaultRabbitMQVolume = kbx.DefaultRabbitMQVolume
)

type DBConfig = svc.DBConfig
type DBConnection = svc.DBConnection
type EnvironmentType = it.IEnvironment

func NewDBConnection(ctx context.Context, name, filePath string, enabled bool, logger *logz.LoggerZ, debug bool) (DBConnection, bool) {
	mgr := engine.NewDatabaseManager(logger)
	rootCfg, err := mgr.LoadOrBootstrap(filePath)
	if err != nil {
		logger.Error("Failed to load or bootstrap database config: %v", err)
		return DBConnection{}, false
	}
	if err := mgr.InitFromRootConfig(ctx, &rootCfg); err != nil {
		logger.Error("Failed to initialize database manager from root config: %v", err)
		return DBConnection{}, false
	}
	return mgr.GetDefault()
}
