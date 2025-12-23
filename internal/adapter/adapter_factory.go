// Package adapter fornece factories para criação de adapters unificados.
package adapter

import (
	"context"
	"fmt"

	store "github.com/kubex-ecosystem/domus/internal/datastore"
	"github.com/kubex-ecosystem/domus/internal/execution"
	"github.com/kubex-ecosystem/domus/internal/types"
)

// AdapterFactory cria adapters DSRepository[T] a partir de Driver ou GORM.
type AdapterFactory struct {
	// driver é o Driver do DSClient (para criar stores)
	driver types.Driver

	// db é o *gorm.DB (para criar ORM repositories)
	db *types.DBConnection

	// config é a configuração padrão para novos adapters
	config *RepositoryConfig
}

// NewAdapterFactory cria uma nova factory.
//
// Pode receber:
// - Apenas driver (db = nil) → cria adapters apenas com Store
// - Apenas db (driver = nil) → cria adapters apenas com ORM
// - Ambos → cria adapters com fallback automático
//
// Se config for nil, usa DefaultConfig().
func NewAdapterFactory(driver types.Driver, db *types.DBConnection, config *RepositoryConfig) (*AdapterFactory, error) {
	if driver == nil && db == nil {
		return nil, fmt.Errorf("ao menos um backend (driver ou db) deve ser fornecido")
	}

	if config == nil {
		config = DefaultConfig()
	}

	return &AdapterFactory{
		driver: driver,
		db:     db,
		config: config,
	}, nil
}

// NewStoreOnlyFactory cria uma factory que usa APENAS stores (sem ORM).
func NewStoreOnlyFactory(driver types.Driver) (*AdapterFactory, error) {
	if driver == nil {
		return nil, fmt.Errorf("driver não pode ser nil")
	}

	config := &RepositoryConfig{
		PreferStore:   true,
		FallbackToORM: false,
		ForceStore:    true,
		ForceORM:      false,
	}

	return &AdapterFactory{
		driver: driver,
		db:     nil,
		config: config,
	}, nil
}

// NewORMOnlyFactory cria uma factory que usa APENAS ORM (sem stores).
func NewORMOnlyFactory(db *types.DBConnection) (*AdapterFactory, error) {
	if db == nil {
		return nil, fmt.Errorf("db não pode ser nil")
	}

	config := &RepositoryConfig{
		PreferStore:   false,
		FallbackToORM: false,
		ForceStore:    false,
		ForceORM:      true,
	}

	return &AdapterFactory{
		driver: nil,
		db:     db,
		config: config,
	}, nil
}

// HasDriver retorna true se a factory possui driver configurado.
func (f *AdapterFactory) HasDriver() bool {
	return f.driver != nil
}

// HasDB retorna true se a factory possui db configurado.
func (f *AdapterFactory) HasDB() bool {
	return f.db != nil
}

// Config retorna a configuração padrão usada pela factory.
func (f *AdapterFactory) Config() *RepositoryConfig {
	return f.config
}

// SetConfig atualiza a configuração padrão da factory.
func (f *AdapterFactory) SetConfig(config *RepositoryConfig) {
	if config != nil {
		f.config = config
	}
}

// GetDriver retorna o driver configurado (pode ser nil).
func (f *AdapterFactory) GetDriver() types.Driver {
	return f.driver
}

// GetDB retorna o *gorm.DB configurado (pode ser nil).
func (f *AdapterFactory) GetDB(ctx context.Context) (execution.Executor, error) {
	return f.db.Driver.Executor(ctx)
}

// CreateAdapter cria um DSRepository[T] genérico.
//
// Parâmetros:
// - ctx: contexto para operações
// - storeName: nome do store no registry (ex: "user", "invite")
// - ormRepoFactory: função que cria ORMRepository[T] a partir de *gorm.DB
//
// Se ormRepoFactory for nil, cria adapter apenas com Store.
// Se storeName estiver vazio e driver disponível, tenta criar adapter apenas com ORM.
func CreateAdapter[T any](
	f *AdapterFactory,
	ctx context.Context,
	storeName string,
	ormRepoFactory func(execution.Executor, error) ORMRepository[T],
) (*DSRepository[T], error) {
	var storeRepo store.Repository[T]
	var ormRepo ORMRepository[T]

	// Tenta criar store repository se driver disponível e storeName fornecido
	if f.driver != nil && storeName != "" {
		factory := store.NewStoreFactory(f.driver)
		storeInterface, err := factory.Create(ctx, storeName)
		if err != nil {
			// Se ForceStore, retorna erro
			if f.config.ForceStore {
				return nil, fmt.Errorf("falha ao criar store %s (ForceStore=true): %v", storeName, err)
			}
			// Caso contrário, apenas loga e continua sem store
		} else {
			// Type assertion para Repository[T]
			var ok bool
			storeRepo, ok = storeInterface.(store.Repository[T])
			if !ok {
				return nil, fmt.Errorf("store %s não implementa Repository[%T]", storeName, *new(T))
			}
		}
	}

	// Tenta criar ORM repository se db disponível e factory fornecida
	if f.db != nil && ormRepoFactory != nil {
		ormRepo = ormRepoFactory(f.db.Driver.Executor(ctx))
	}

	// Cria adapter com os repositories disponíveis
	return NewDSRepository[T](storeRepo, ormRepo, f.config)
}

// CreateStoreAdapter cria um adapter que usa APENAS store (ignora ORM).
func CreateStoreAdapter[T any](f *AdapterFactory, ctx context.Context, storeName string) (*DSRepository[T], error) {
	if f.driver == nil {
		return nil, fmt.Errorf("driver não disponível para criar store adapter")
	}

	factory := store.NewStoreFactory(f.driver)
	storeInterface, err := factory.Create(ctx, storeName)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar store %s: %v", storeName, err)
	}

	storeRepo, ok := storeInterface.(store.Repository[T])
	if !ok {
		return nil, fmt.Errorf("store %s não implementa Repository[%T]", storeName, *new(T))
	}

	config := &RepositoryConfig{
		PreferStore:   true,
		FallbackToORM: false,
		ForceStore:    true,
		ForceORM:      false,
	}

	return NewDSRepository[T](storeRepo, nil, config)
}

// CreateORMAdapter cria um adapter que usa APENAS ORM (ignora Store).
func CreateORMAdapter[T any](f *AdapterFactory, ormRepoFactory func(execution.Executor, error) ORMRepository[T]) (*DSRepository[T], error) {
	if f.db == nil {
		return nil, fmt.Errorf("db não disponível para criar ORM adapter")
	}

	if ormRepoFactory == nil {
		return nil, fmt.Errorf("ormRepoFactory não pode ser nil")
	}

	ormRepo := ormRepoFactory(f.db.Driver.Executor(context.Background()))

	config := &RepositoryConfig{
		PreferStore:   false,
		FallbackToORM: false,
		ForceStore:    false,
		ForceORM:      true,
	}

	return NewDSRepository[T](nil, ormRepo, config)
}
