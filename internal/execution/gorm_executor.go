// Package execution provides abstractions for executing database operations.
package execution

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// GormExecutor encapsula *gorm.DB sem vazar o tipo para consumidores externos.
// Esta interface expõe apenas operações necessárias de forma controlada.
//
// IMPORTANTE: Esta interface NÃO deve vazar gorm.DB, gorm.Error ou qualquer
// tipo do pacote gorm para fora do domus.
type GormExecutor interface {
	// WithContext retorna um novo executor com o contexto aplicado.
	WithContext(ctx context.Context) GormExecutor

	// --- CRUD Básico ---

	// Create insere um novo registro.
	Create(value any) GormExecutor

	// Save salva (insert ou update) um registro.
	Save(value any) GormExecutor

	// First retorna o primeiro registro que corresponde às condições.
	First(dest any, conds ...any) GormExecutor

	// Take retorna um registro que corresponde às condições (sem ordenação).
	Take(dest any, conds ...any) GormExecutor

	// Last retorna o último registro que corresponde às condições.
	Last(dest any, conds ...any) GormExecutor

	// Find retorna todos os registros que correspondem às condições.
	Find(dest any, conds ...any) GormExecutor

	// Delete remove registros que correspondem às condições.
	Delete(value any, conds ...any) GormExecutor

	// --- Query Building ---

	// Model especifica o modelo para operações.
	Model(value any) GormExecutor

	// Table especifica a tabela para operações.
	Table(name string, args ...any) GormExecutor

	// Select especifica campos para consulta.
	Select(query any, args ...any) GormExecutor

	// Where adiciona condições de filtro.
	Where(query any, args ...any) GormExecutor

	// Or adiciona condição OR.
	Or(query any, args ...any) GormExecutor

	// Not adiciona condição NOT.
	Not(query any, args ...any) GormExecutor

	// Order especifica ordenação.
	Order(value any) GormExecutor

	// Limit define limite de registros.
	Limit(limit int) GormExecutor

	// Offset define offset para paginação.
	Offset(offset int) GormExecutor

	// Group especifica GROUP BY.
	Group(name string) GormExecutor

	// Having especifica HAVING.
	Having(query any, args ...any) GormExecutor

	// Joins adiciona JOIN.
	Joins(query string, args ...any) GormExecutor

	// Preload carrega associações.
	Preload(query string, args ...any) GormExecutor

	// --- Updates ---

	// Update atualiza uma coluna específica.
	Update(column string, value any) GormExecutor

	// Updates atualiza múltiplas colunas.
	Updates(values any) GormExecutor

	// UpdateColumn atualiza uma coluna sem hooks.
	UpdateColumn(column string, value any) GormExecutor

	// UpdateColumns atualiza múltiplas colunas sem hooks.
	UpdateColumns(values any) GormExecutor

	// --- Raw Queries ---

	// Raw executa SQL raw.
	Raw(sql string, values ...any) GormExecutor

	// Exec executa SQL de modificação.
	Exec(sql string, values ...any) GormExecutor

	// Scan escaneia resultado para destino.
	Scan(dest any) GormExecutor

	// --- Aggregation ---

	// Count retorna contagem de registros.
	Count(count *int64) GormExecutor

	// Pluck extrai uma coluna para slice.
	Pluck(column string, dest any) GormExecutor

	// --- Transactions ---

	// Begin inicia uma transação.
	Begin() GormExecutor

	// Commit confirma a transação.
	Commit() GormExecutor

	// Rollback reverte a transação.
	Rollback() GormExecutor

	// Transaction executa função em transação com commit/rollback automático.
	Transaction(fc func(tx GormExecutor) error) error

	// --- Session & Config ---

	// Session cria nova sessão com configurações.
	Session(config *SessionConfig) GormExecutor

	// Debug ativa modo debug.
	Debug() GormExecutor

	// Unscoped desativa soft delete.
	Unscoped() GormExecutor

	// --- Result Info ---

	// Error retorna erro da última operação.
	Error() error

	// RowsAffected retorna número de linhas afetadas.
	RowsAffected() int64

	// --- Helpers ---

	// IsNotFound verifica se o erro é "record not found".
	IsNotFound() bool

	// AutoMigrate executa migração automática.
	AutoMigrate(dst ...any) error
}

// SessionConfig contém configurações para Session().
// Evita expor gorm.Session diretamente.
type SessionConfig struct {
	DryRun                   bool
	PrepareStmt              bool
	NewDB                    bool
	Initialized              bool
	SkipHooks                bool
	SkipDefaultTransaction   bool
	DisableNestedTransaction bool
	AllowGlobalUpdate        bool
	FullSaveAssociations     bool
	QueryFields              bool
	CreateBatchSize          int
}

// gormExecutor é a implementação concreta baseada em *gorm.DB.
type gormExecutor struct {
	db *gorm.DB
}

// NewGormExecutor cria um GormExecutor a partir de um *gorm.DB.
// Esta é a ÚNICA forma de criar um GormExecutor, e deve ser usada
// apenas dentro do pacote domus (drivers, etc).
func NewGormExecutor(db *gorm.DB) GormExecutor {
	if db == nil {
		return nil
	}
	return &gormExecutor{db: db}
}

// wrap cria um novo executor a partir de um *gorm.DB resultante.
func (g *gormExecutor) wrap(db *gorm.DB) GormExecutor {
	return &gormExecutor{db: db}
}

// --- Implementação ---

func (g *gormExecutor) WithContext(ctx context.Context) GormExecutor {
	return g.wrap(g.db.WithContext(ctx))
}

func (g *gormExecutor) Create(value any) GormExecutor {
	return g.wrap(g.db.Create(value))
}

func (g *gormExecutor) Save(value any) GormExecutor {
	return g.wrap(g.db.Save(value))
}

func (g *gormExecutor) First(dest any, conds ...any) GormExecutor {
	return g.wrap(g.db.First(dest, conds...))
}

func (g *gormExecutor) Take(dest any, conds ...any) GormExecutor {
	return g.wrap(g.db.Take(dest, conds...))
}

func (g *gormExecutor) Last(dest any, conds ...any) GormExecutor {
	return g.wrap(g.db.Last(dest, conds...))
}

func (g *gormExecutor) Find(dest any, conds ...any) GormExecutor {
	return g.wrap(g.db.Find(dest, conds...))
}

func (g *gormExecutor) Delete(value any, conds ...any) GormExecutor {
	return g.wrap(g.db.Delete(value, conds...))
}

func (g *gormExecutor) Model(value any) GormExecutor {
	return g.wrap(g.db.Model(value))
}

func (g *gormExecutor) Table(name string, args ...any) GormExecutor {
	return g.wrap(g.db.Table(name, args...))
}

func (g *gormExecutor) Select(query any, args ...any) GormExecutor {
	return g.wrap(g.db.Select(query, args...))
}

func (g *gormExecutor) Where(query any, args ...any) GormExecutor {
	return g.wrap(g.db.Where(query, args...))
}

func (g *gormExecutor) Or(query any, args ...any) GormExecutor {
	return g.wrap(g.db.Or(query, args...))
}

func (g *gormExecutor) Not(query any, args ...any) GormExecutor {
	return g.wrap(g.db.Not(query, args...))
}

func (g *gormExecutor) Order(value any) GormExecutor {
	return g.wrap(g.db.Order(value))
}

func (g *gormExecutor) Limit(limit int) GormExecutor {
	return g.wrap(g.db.Limit(limit))
}

func (g *gormExecutor) Offset(offset int) GormExecutor {
	return g.wrap(g.db.Offset(offset))
}

func (g *gormExecutor) Group(name string) GormExecutor {
	return g.wrap(g.db.Group(name))
}

func (g *gormExecutor) Having(query any, args ...any) GormExecutor {
	return g.wrap(g.db.Having(query, args...))
}

func (g *gormExecutor) Joins(query string, args ...any) GormExecutor {
	return g.wrap(g.db.Joins(query, args...))
}

func (g *gormExecutor) Preload(query string, args ...any) GormExecutor {
	return g.wrap(g.db.Preload(query, args...))
}

func (g *gormExecutor) Update(column string, value any) GormExecutor {
	return g.wrap(g.db.Update(column, value))
}

func (g *gormExecutor) Updates(values any) GormExecutor {
	return g.wrap(g.db.Updates(values))
}

func (g *gormExecutor) UpdateColumn(column string, value any) GormExecutor {
	return g.wrap(g.db.UpdateColumn(column, value))
}

func (g *gormExecutor) UpdateColumns(values any) GormExecutor {
	return g.wrap(g.db.UpdateColumns(values))
}

func (g *gormExecutor) Raw(sql string, values ...any) GormExecutor {
	return g.wrap(g.db.Raw(sql, values...))
}

func (g *gormExecutor) Exec(sql string, values ...any) GormExecutor {
	return g.wrap(g.db.Exec(sql, values...))
}

func (g *gormExecutor) Scan(dest any) GormExecutor {
	return g.wrap(g.db.Scan(dest))
}

func (g *gormExecutor) Count(count *int64) GormExecutor {
	return g.wrap(g.db.Count(count))
}

func (g *gormExecutor) Pluck(column string, dest any) GormExecutor {
	return g.wrap(g.db.Pluck(column, dest))
}

func (g *gormExecutor) Begin() GormExecutor {
	return g.wrap(g.db.Begin())
}

func (g *gormExecutor) Commit() GormExecutor {
	return g.wrap(g.db.Commit())
}

func (g *gormExecutor) Rollback() GormExecutor {
	return g.wrap(g.db.Rollback())
}

func (g *gormExecutor) Transaction(fc func(tx GormExecutor) error) error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		return fc(&gormExecutor{db: tx})
	})
}

func (g *gormExecutor) Session(config *SessionConfig) GormExecutor {
	if config == nil {
		return g.wrap(g.db.Session(&gorm.Session{}))
	}
	return g.wrap(g.db.Session(&gorm.Session{
		DryRun:                   config.DryRun,
		PrepareStmt:              config.PrepareStmt,
		NewDB:                    config.NewDB,
		Initialized:              config.Initialized,
		SkipHooks:                config.SkipHooks,
		SkipDefaultTransaction:   config.SkipDefaultTransaction,
		DisableNestedTransaction: config.DisableNestedTransaction,
		AllowGlobalUpdate:        config.AllowGlobalUpdate,
		FullSaveAssociations:     config.FullSaveAssociations,
		QueryFields:              config.QueryFields,
		CreateBatchSize:          config.CreateBatchSize,
	}))
}

func (g *gormExecutor) Debug() GormExecutor {
	return g.wrap(g.db.Debug())
}

func (g *gormExecutor) Unscoped() GormExecutor {
	return g.wrap(g.db.Unscoped())
}

func (g *gormExecutor) Error() error {
	return g.db.Error
}

func (g *gormExecutor) RowsAffected() int64 {
	return g.db.RowsAffected
}

func (g *gormExecutor) IsNotFound() bool {
	return errors.Is(g.db.Error, gorm.ErrRecordNotFound)
}

func (g *gormExecutor) AutoMigrate(dst ...any) error {
	return g.db.AutoMigrate(dst...)
}

// Compile-time check que gormExecutor implementa GormExecutor.
var _ GormExecutor = (*gormExecutor)(nil)
