// Package execution provides abstractions for executing database operations.
package execution

import "context"

// BackendKind representa o tipo lógico de backend por trás do executor.
// Isso não precisa necessariamente refletir o driver físico; é uma
// classificação de "flavor" para consumo de camada superior.
type BackendKind string

const (
	// BackendUnknown é o tipo de backend padrão, sem sabor.
	BackendUnknown BackendKind = "unknown"

	// BackendPostgres é o tipo de backend para PostgreSQL.
	BackendPostgres BackendKind = "postgres"

	// BackendMongo é o tipo de backend para MongoDB.
	BackendMongo BackendKind = "mongo"

	// BackendRedis é o tipo de backend para Redis.
	BackendRedis BackendKind = "redis"

	// BackendHTTP é o tipo de backend para HTTP.
	BackendHTTP BackendKind = "http"
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

	// Mongo retorna o executor baseado em MongoDB (se disponível).
	Mongo() MongoExecutor

	// Redis retorna o executor baseado em Redis (se disponível).
	Redis() RedisExecutor

	// HTTP retorna o executor baseado em HTTP (se disponível).
	HTTP() HTTPExecutor

	// Generic retorna o executor genérico que trabalha com QueryModel.
	Generic() GenericExecutor
}

// executorImpl é a implementação concreta padrão.
// Ela é simples: guarda referências opcionais para cada executor específico.
type executorImpl struct {
	kind BackendKind

	pg    PGExecutor
	gen   GenericExecutor
	gorm  GormExecutor
	mongo MongoExecutor
	redis RedisExecutor
	http  HTTPExecutor
}

// ExecutorOption permite configurar a implementação do Executor
// sem precisar criar múltiplos construtores.
type ExecutorOption func(Executor)

// WithKind define o tipo principal do backend.
func WithKind(kind BackendKind) ExecutorOption {
	return func(e Executor) {
		if a, ok := e.(*executorImpl); ok {
			a.kind = kind
		}
	}
}

// WithPG define o executor específico de Postgres.
func WithPG(pg PGExecutor) ExecutorOption {
	return func(e Executor) {
		if a, ok := e.(*executorImpl); ok {
			a.pg = pg
		}
	}
}

// WithGeneric define o executor genérico.
func WithGeneric(gen GenericExecutor) ExecutorOption {
	return func(e Executor) {
		if a, ok := e.(*executorImpl); ok {
			a.gen = gen
		}
	}
}

// WithGorm define o executor baseado em GORM.
func WithGorm(gorm GormExecutor) ExecutorOption {
	return func(e Executor) {
		if a, ok := e.(*executorImpl); ok {
			a.gorm = gorm
		}
	}
}

// WithMongo define o executor baseado em MongoDB.
func WithMongo(mongo MongoExecutor) ExecutorOption {
	return func(e Executor) {
		if a, ok := e.(*executorImpl); ok {
			a.mongo = mongo
		}
	}
}

// WithRedis define o executor baseado em Redis.
func WithRedis(redis RedisExecutor) ExecutorOption {
	return func(e Executor) {
		if a, ok := e.(*executorImpl); ok {
			a.redis = redis
		}
	}
}

// WithHTTP define o executor baseado em HTTP.
func WithHTTP(http HTTPExecutor) ExecutorOption {
	return func(e Executor) {
		if a, ok := e.(*executorImpl); ok {
			a.http = http
		}
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

// Kind retorna o tipo principal do backend.
func (e *executorImpl) Kind() BackendKind {
	return e.kind
}

// PG retorna o executor específico de Postgres.
func (e *executorImpl) PG() PGExecutor {
	return e.pg
}

// Gorm retorna o executor baseado em GORM.
func (e *executorImpl) Gorm() GormExecutor {
	return e.gorm
}

// Mongo retorna o executor baseado em MongoDB.
func (e *executorImpl) Mongo() MongoExecutor {
	return e.mongo
}

// Redis retorna o executor baseado em Redis.
func (e *executorImpl) Redis() RedisExecutor {
	return e.redis
}

// HTTP retorna o executor baseado em HTTP.
func (e *executorImpl) HTTP() HTTPExecutor {
	return e.http
}

// Generic retorna o executor genérico que trabalha com QueryModel.
func (e *executorImpl) Generic() GenericExecutor {
	return e.gen
}

// NewPureGenericExecutor cria um Executor totalmente genérico, sem backend específico.
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
