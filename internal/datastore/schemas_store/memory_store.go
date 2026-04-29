package schemasstore

import (
	"context"
	"sync"
)

// MemorySchemaStore is an in-memory implementation of SchemaStore.
// It is thread-safe and suitable for use as a fallback or in tests.
type MemorySchemaStore struct {
	mu      sync.RWMutex
	schemas map[string]*Schema
}

// NewMemorySchemaStore creates a new in-memory schema store.
func NewMemorySchemaStore() *MemorySchemaStore {
	return &MemorySchemaStore{
		schemas: make(map[string]*Schema),
	}
}

// NewMemorySchemaStoreWith creates a new in-memory schema store pre-populated with schemas.
func NewMemorySchemaStoreWith(schemas ...*Schema) *MemorySchemaStore {
	store := NewMemorySchemaStore()
	for _, s := range schemas {
		if s != nil && s.Key != "" {
			store.schemas[s.Key] = s
		}
	}
	return store
}

func (m *MemorySchemaStore) Register(_ context.Context, schema *Schema) error {
	if err := schema.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.schemas[schema.Key]; exists {
		return ErrSchemaAlreadyExists
	}

	m.schemas[schema.Key] = schema
	return nil
}

func (m *MemorySchemaStore) Get(_ context.Context, key string) (*Schema, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schema, exists := m.schemas[key]
	if !exists {
		return nil, ErrSchemaNotFound
	}

	return schema, nil
}

func (m *MemorySchemaStore) GetOrDefault(_ context.Context, key string, defaultSchema *Schema) *Schema {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schema, exists := m.schemas[key]
	if !exists {
		return defaultSchema
	}
	return schema
}

func (m *MemorySchemaStore) Update(_ context.Context, schema *Schema) error {
	if err := schema.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.schemas[schema.Key]; !exists {
		return ErrSchemaNotFound
	}

	m.schemas[schema.Key] = schema
	return nil
}

func (m *MemorySchemaStore) Upsert(_ context.Context, schema *Schema) error {
	if err := schema.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.schemas[schema.Key] = schema
	return nil
}

func (m *MemorySchemaStore) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.schemas[key]; !exists {
		return ErrSchemaNotFound
	}

	delete(m.schemas, key)
	return nil
}

func (m *MemorySchemaStore) Exists(_ context.Context, key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.schemas[key]
	return exists
}

func (m *MemorySchemaStore) List(_ context.Context) ([]*Schema, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schemas := make([]*Schema, 0, len(m.schemas))
	for _, s := range m.schemas {
		schemas = append(schemas, s)
	}
	return schemas, nil
}

func (m *MemorySchemaStore) Keys(_ context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.schemas))
	for k := range m.schemas {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *MemorySchemaStore) Count(_ context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.schemas), nil
}

func (m *MemorySchemaStore) Clear(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.schemas = make(map[string]*Schema)
	return nil
}

// Compile-time check.
var _ SchemaStore = (*MemorySchemaStore)(nil)
