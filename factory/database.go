package factory

import (
	"context"
	"embed"

	ci "github.com/kubex-ecosystem/domus/internal/interfaces"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/services/docker"
	"github.com/kubex-ecosystem/domus/internal/types"

	logz "github.com/kubex-ecosystem/logz"
)

var migrationFiles embed.FS

type DBConnection = types.DBConnection

func NewDatabaseConnectionB(ctx context.Context, cfg *types.DBConfig, logger *logz.LoggerZ) (*DBConnection, error) {
	// return svc.NewDatabaseServiceImpl(ctx, cfg, logger)
	return nil, nil
}
func NewDatabaseConnection(ctx context.Context, cfg *types.DBConfig, logger *logz.LoggerZ) (*DBConnection, error) {
	// return svc.NewDatabaseService(ctx, cfg, logger)
	return nil, nil
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}
type DirectDatabase interface {
	Query(context.Context, string, ...interface{}) (any, error)
}

type DBConfig = *types.DBConfig
type IDBConfig interface {
	*types.DBConfig
}
type DBConfigImpl = types.DBConfig

func NewDBConfigWithArgs(ctx context.Context, dbName, dbConfigFilePath string, autoMigrate bool, logger *logz.LoggerZ, debug bool) *DBConfigImpl {
	// return types.NewDBConfigWithArgs(ctx, dbName, dbConfigFilePath, autoMigrate, logger, debug)
	return nil
}
func NewDBConfigFromFile(ctx context.Context, dbConfigFilePath string, autoMigrate bool, logger *logz.LoggerZ, debug bool) (*DBConfigImpl, error) {
	// return svc.NewDBConfigFromFile(ctx, dbConfigFilePath, autoMigrate, logger, debug)
	return nil, nil
}
func SetupDatabaseServices(ctx context.Context, d ci.IDockerService, cfg *types.DBConfig) error {
	// return svc.SetupDatabaseServices(ctx, d, cfg)
	return nil
}

func SetMigrationFiles(mf embed.FS) {
	migrationFiles = mf
}
func GetMigrationFiles() embed.FS {
	return migrationFiles
}

type Environment = ci.IEnvironment
type EnvironmentType = types.Environment

func NewEnvironment(configFile string, isConfidential bool, logger *logz.LoggerZ) (*EnvironmentType, error) {
	return types.NewEnvironmentType(configFile, isConfidential, logger)
}

type IDockerService = ci.IDockerService
type DockerService = docker.DockerService

type JSONB = ci.IJSONB
type IJSONB interface {
	ci.IJSONB
}
type JSONBImpl = types.JSONBImpl

func NewJSONB() IJSONB {
	return &types.JSONBImpl{}
}

type JSONBData = types.JSONBData
type IJSONBData interface{ ci.IJSONB }

func NewJSONBData() JSONBData { return types.NewJSONBData() }

type JWT = types.JWT
type JWTImpl = types.JWT

func NewJWT() *JWTImpl {
	return &types.JWT{}
}

type Reference = kbx.Reference
type IReference interface {
	ci.IReference
}
type ReferenceImpl = kbx.Reference

func NewReference(name string) Reference {
	return kbx.NewReference(name)
}
