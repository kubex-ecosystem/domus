// Package execution provides abstractions for executing database operations.
package execution

import "context"

// BackendKind representa o tipo lógico de backend por trás do executor.
// Isso não precisa necessariamente refletir o driver físico; é uma
// classificação de "flavor" para consumo de camada superior.
type BackendKind string

const (
	BackendUnknown  BackendKind = "unknown"
	BackendPostgres BackendKind = "postgres"
	BackendMongo    BackendKind = "mongo"
	BackendRedis    BackendKind = "redis"
	BackendHTTP     BackendKind = "http"
)

// Executor é a interface unificada que um Store ou Service deve receber.
//
// A ideia do "Dual Mode" é:
//
//   - Quando existir um executor nativo (ex: PGExecutor), você usa ele;
//   - Quando não existir / não fizer sentido, cai para o GenericExecutor.
//
// Isso permite:
//   - Stores com SQL explícito, forte, rápido (PG);
//   - Stores polimórficos / experimentais usando QueryModel (Generic);
//   - Stores ORM-based usando GORM (Gorm).
type Executor interface {
	// Kind retorna o "flavor" principal do backend.
	Kind() BackendKind

	// PG retorna o executor específico de Postgres (se disponível).
	PG() PGExecutor

	// Gorm retorna o executor baseado em GORM (se disponível).
	Gorm() GormExecutor

	// Futuro: extensões para Mongo, Redis, etc.
	// Mongo() MongoExecutor
	// Redis() RedisExecutor
	// HTTP() HTTPExecutor

	// Generic retorna o executor genérico que trabalha com QueryModel.
	Generic() GenericExecutor
}

// executorImpl é a implementação concreta padrão.
// Ela é simples: guarda referências opcionais para cada executor específico.
type executorImpl struct {
	kind BackendKind

	pg   PGExecutor
	gen  GenericExecutor
	gorm GormExecutor

	// Futuro:
	// mongo MongoExecutor
	// redis RedisExecutor
	// http  HTTPExecutor
}

// ExecutorOption permite configurar a implementação do Executor
// sem precisar criar múltiplos construtores.
type ExecutorOption func(*executorImpl)

// WithKind define o tipo principal do backend.
func WithKind(kind BackendKind) ExecutorOption {
	return func(e *executorImpl) {
		e.kind = kind
	}
}

// WithPG define o executor específico de Postgres.
func WithPG(pg PGExecutor) ExecutorOption {
	return func(e *executorImpl) {
		e.pg = pg
	}
}

// WithGeneric define o executor genérico.
func WithGeneric(gen GenericExecutor) ExecutorOption {
	return func(e *executorImpl) {
		e.gen = gen
	}
}

// WithGorm define o executor baseado em GORM.
func WithGorm(gorm GormExecutor) ExecutorOption {
	return func(e *executorImpl) {
		e.gorm = gorm
	}
}

// NewExecutor cria um Executor dual-mode.
//
// Exemplo de uso:
//
//	ex := NewExecutor(
//	    WithKind(BackendPostgres),
//	    WithPG(pgExec),
//	    WithGorm(gormExec),
//	    WithGeneric(NewGenericExecutor()),
//	)
func NewExecutor(opts ...ExecutorOption) Executor {
	e := &executorImpl{
		kind: BackendUnknown,
		gen:  NewGenericExecutor(), // fallback default
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Implementação dos métodos da interface Executor.

func (e *executorImpl) Kind() BackendKind {
	return e.kind
}

func (e *executorImpl) PG() PGExecutor {
	return e.pg
}

func (e *executorImpl) Gorm() GormExecutor {
	return e.gorm
}

// Futuro:
// func (e *executorImpl) Mongo() MongoExecutor { return e.mongo }
// func (e *executorImpl) Redis() RedisExecutor { return e.redis }
// func (e *executorImpl) HTTP() HTTPExecutor   { return e.http }

func (e *executorImpl) Generic() GenericExecutor {
	return e.gen
}

// Helper que pode ser útil em testes ou casos pontuais: cria um Executor
// totalmente genérico, sem backend específico.

func NewPureGenericExecutor() Executor {
	return &executorImpl{
		kind: BackendUnknown,
		gen:  NewGenericExecutor(),
	}
}

// No futuro, se quiser um atalho "from context":
// func FromContext(ctx context.Context) Executor { ... }

// Mantive um import de context só para sugerir evoluções futuras.
// Se não usar, pode remover o import.
var _ = context.Background
