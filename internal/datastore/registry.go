package datastore

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	t "github.com/kubex-ecosystem/domus/internal/types"
	gl "github.com/kubex-ecosystem/logz"
)

// globalRegistry é o registro singleton de stores.
var globalRegistry = newStoreRegistry()

// storeRegistry implementa types.StoreRegistry.
type storeRegistry struct {
	mu        sync.RWMutex
	stores    map[string]map[string]*storeEntry // [driverName][storeName]
	factories map[string]map[string]storeFactoryFunc
}

type storeEntry struct {
	metadata *t.StoreMetadata
	factory  storeFactoryFunc
}

type storeFactoryFunc func(context.Context, t.Driver) (t.StoreType, error)

func newStoreRegistry() *storeRegistry {
	return &storeRegistry{
		stores:    make(map[string]map[string]*storeEntry),
		factories: make(map[string]map[string]storeFactoryFunc),
	}
}

// Register registra um store para um driver específico.
func (r *storeRegistry) Register(driverName, storeName string, metadata *t.StoreMetadata, factory storeFactoryFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if driverName == "" || storeName == "" {
		return fmt.Errorf("driver name and store name are required")
	}

	if factory == nil {
		return fmt.Errorf("factory function is required")
	}

	// Inicializa mapas se necessário
	if r.stores[driverName] == nil {
		r.stores[driverName] = make(map[string]*storeEntry)
	}
	if r.factories[driverName] == nil {
		r.factories[driverName] = make(map[string]storeFactoryFunc)
	}

	// Verifica duplicação
	if _, exists := r.stores[driverName][storeName]; exists {
		gl.Warn("Store %s already registered for driver %s, overwriting", storeName, driverName)
	}

	// Usa metadata fornecido ou cria um básico
	if metadata == nil {
		metadata = &t.StoreMetadata{
			Name:       storeName,
			DriverName: driverName,
		}
	} else {
		metadata.Name = storeName
		metadata.DriverName = driverName
	}

	r.stores[driverName][storeName] = &storeEntry{
		metadata: metadata,
		factory:  factory,
	}
	r.factories[driverName][storeName] = factory

	return nil
}

// Get retorna a factory function para um store específico.
func (r *storeRegistry) Get(driverName, storeName string) (storeFactoryFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if drivers, ok := r.factories[driverName]; ok {
		factory, exists := drivers[storeName]
		return factory, exists
	}
	return nil, false
}

// ListByDriver retorna todos os stores registrados para um driver.
func (r *storeRegistry) ListByDriver(driverName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stores := r.stores[driverName]
	if stores == nil {
		return []string{}
	}

	names := make([]string, 0, len(stores))
	for name := range stores {
		names = append(names, name)
	}
	return names
}

// GetMetadata retorna metadados de um store.
func (r *storeRegistry) GetMetadata(driverName, storeName string) (*t.StoreMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if drivers, ok := r.stores[driverName]; ok {
		if entry, exists := drivers[storeName]; exists {
			return entry.metadata, nil
		}
	}
	return nil, fmt.Errorf("store %s not found for driver %s", storeName, driverName)
}

// AllDrivers retorna lista de drivers com stores registrados.
func (r *storeRegistry) AllDrivers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	drivers := make([]string, 0, len(r.stores))
	for driver := range r.stores {
		drivers = append(drivers, driver)
	}
	return drivers
}

// AllStores retorna mapa completo de stores por driver.
func (r *storeRegistry) AllStores() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]string)
	for driver, stores := range r.stores {
		names := make([]string, 0, len(stores))
		for name := range stores {
			names = append(names, name)
		}
		result[driver] = names
	}
	return result
}

// AllStoreTypes retorna mapa completo de reflect.Type por driver e store.
func (r *storeRegistry) AllStoreTypes() map[string]map[string]reflect.Type {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]map[string]reflect.Type)
	for driver, stores := range r.stores {
		result[driver] = make(map[string]reflect.Type)
		for name, entry := range stores {
			if entry.metadata != nil {
				result[driver][name] = entry.metadata.Type
			}
		}
	}
	return result
}

// Funções públicas para acessar o registro global

// RegisterStore registra um store no registro global.
func RegisterStore(driverName, storeName string, metadata *t.StoreMetadata, factory func(context.Context, t.Driver) (t.StoreType, error)) error {
	return globalRegistry.Register(driverName, storeName, metadata, factory)
}

// GetStoreFactory retorna a factory function para um store.
func GetStoreFactory(driverName, storeName string) (func(context.Context, t.Driver) (t.StoreType, error), bool) {
	return globalRegistry.Get(driverName, storeName)
}

// ListStoresByDriver retorna todos os stores de um driver.
func ListStoresByDriver(driverName string) []string {
	return globalRegistry.ListByDriver(driverName)
}

// GetStoreMetadata retorna metadados de um store.
func GetStoreMetadata(driverName, storeName string) (*t.StoreMetadata, error) {
	return globalRegistry.GetMetadata(driverName, storeName)
}

// AllDrivers retorna todos os drivers registrados.
func AllDrivers() []string {
	return globalRegistry.AllDrivers()
}

// AllStores retorna mapa de stores por driver.
func AllStores() map[string][]string {
	return globalRegistry.AllStores()
}

// AllStoreTypes retorna mapa de reflect.Type por driver e store.
func AllStoreTypes() map[string]map[string]reflect.Type {
	return globalRegistry.AllStoreTypes()
}
