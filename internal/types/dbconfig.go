package types

import (
	"context"

	"github.com/kubex-ecosystem/domus/internal/execution"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
)

// DBConfigRT é a “view runtime”: agrega coisas de runtime sem poluir o modelo puro.
type DBConfigRT struct {
	Config DBConfig `json:"-" yaml:"-" mapstructure:"-"`

	Mapper *Mapper[DBConfig] `json:"-" yaml:"-" mapstructure:"-"`

	*Mutexes `json:"-" yaml:"-" mapstructure:"-"`
}

type MigrationInfo struct {
	// MigrationPath é o caminho para os arquivos de migração.
	MigrationPath string `json:"migration_path,omitempty" yaml:"migration_path,omitempty" mapstructure:"migration_path,omitempty"`
	// Enabled indica se a migração está habilitada.
	Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled,omitempty" default:"false"`
	// Options são opções adicionais para a migração.
	Options map[string]any `json:"options,omitempty" yaml:"options,omitempty" mapstructure:"options,omitempty"`
	// Auto indica se a migração deve ser executada automaticamente.
	Auto bool `json:"auto,omitempty" yaml:"auto,omitempty" mapstructure:"auto,omitempty" default:"false"`
	// Version indica a versão da migração.
	Version string `json:"version,omitempty" yaml:"version,omitempty" mapstructure:"version,omitempty"`
	// Bootstrap indica se a migração deve ser inicializada.
	Bootstrap bool `json:"bootstrap,omitempty" yaml:"bootstrap,omitempty" mapstructure:"bootstrap,omitempty" default:"false"`
}

// DBConfig is the pure, serializable configuration for a single database.
type DBConfig = kbx.DBConfig

// Validator garante que a configuração para um dado tipo de DB é coerente.
type Validator interface {
	Validate(cfg DBConfig) error
}

// Migrator é plugável por tipo de banco (Pg, Mongo, etc).
type Migrator interface {
	Migrate(ctx context.Context, drv Driver, cfg DBConfig) error
}

// DBConnection agrega config runtime + driver vivo.
type DBConnection struct {
	Config DBConfigRT `json:"config" yaml:"config"`
	Driver Driver     `json:"-" yaml:"-"`
}

type Driver interface {
	Name() string
	Connect(ctx context.Context, cfg *kbx.DBConfig) error
	Ping(ctx context.Context) bool
	Close() error
	Executor(ctx context.Context) (execution.Executor, error)
}
