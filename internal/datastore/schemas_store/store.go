// Package schemasstore provides schema registration and retrieval functionality.
package schemasstore

import (
	"context"
	"errors"
	"sync"
)

// Common errors for SchemaStore operations.
var (
	ErrSchemaNotFound      = errors.New("schema not found")
	ErrSchemaAlreadyExists = errors.New("schema already exists")
	ErrInvalidSchema       = errors.New("invalid schema")
)

// Schema represents a registered schema entry.
// This is the data structure stored and retrieved by SchemaStore.
type Schema struct {
	// Key is the unique identifier for this schema.
	Key string `json:"key" gorm:"primaryKey"`

	// Name is a human-readable name for the schema.
	Name string `json:"name"`

	// Version tracks schema version for compatibility.
	Version string `json:"version,omitempty"`

	// Definition holds the schema definition (JSON, struct info, etc).
	Definition []byte `json:"definition,omitempty"`

	// Metadata holds optional key-value metadata.
	Metadata map[string]string `json:"metadata,omitempty" gorm:"serializer:json"`
}

// Validate checks if the schema is valid.
func (s *Schema) Validate() error {
	if s.Key == "" {
		return ErrInvalidSchema
	}
	return nil
}

// SchemaStore defines the contract for schema registration and retrieval.
// Implementations can be in-memory, persistent, or a combination.
type SchemaStore interface {
	// Register adds a new schema to the store.
	// Returns ErrSchemaAlreadyExists if key is already registered.
	Register(ctx context.Context, schema *Schema) error

	// Get retrieves a schema by key.
	// Returns ErrSchemaNotFound if not found.
	Get(ctx context.Context, key string) (*Schema, error)

	// GetOrDefault retrieves a schema by key, returning defaultSchema if not found.
	// Does not return ErrSchemaNotFound.
	GetOrDefault(ctx context.Context, key string, defaultSchema *Schema) *Schema

	// Update updates an existing schema.
	// Returns ErrSchemaNotFound if not found.
	Update(ctx context.Context, schema *Schema) error

	// Upsert registers or updates a schema.
	Upsert(ctx context.Context, schema *Schema) error

	// Delete removes a schema by key.
	// Returns ErrSchemaNotFound if not found.
	Delete(ctx context.Context, key string) error

	// Exists checks if a schema exists.
	Exists(ctx context.Context, key string) bool

	// List returns all registered schemas.
	List(ctx context.Context) ([]*Schema, error)

	// Keys returns all registered keys.
	Keys(ctx context.Context) ([]string, error)

	// Count returns the number of registered schemas.
	Count(ctx context.Context) (int, error)

	// Clear removes all schemas from the store.
	Clear(ctx context.Context) error
}

// FallbackSchemaStore wraps a primary store with a fallback store.
// Reads check primary first, then fallback.
// Writes go only to primary.
type FallbackSchemaStore struct {
	primary  SchemaStore
	fallback SchemaStore
	mu       sync.RWMutex
}

// NewFallbackSchemaStore creates a new FallbackSchemaStore.
func NewFallbackSchemaStore(primary, fallback SchemaStore) *FallbackSchemaStore {
	return &FallbackSchemaStore{
		primary:  primary,
		fallback: fallback,
	}
}

func (f *FallbackSchemaStore) Register(ctx context.Context, schema *Schema) error {
	return f.primary.Register(ctx, schema)
}

func (f *FallbackSchemaStore) Get(ctx context.Context, key string) (*Schema, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	schema, err := f.primary.Get(ctx, key)
	if err == nil {
		return schema, nil
	}

	if !errors.Is(err, ErrSchemaNotFound) {
		return nil, err
	}

	return f.fallback.Get(ctx, key)
}

func (f *FallbackSchemaStore) GetOrDefault(ctx context.Context, key string, defaultSchema *Schema) *Schema {
	schema, err := f.Get(ctx, key)
	if err != nil {
		return defaultSchema
	}
	return schema
}

func (f *FallbackSchemaStore) Update(ctx context.Context, schema *Schema) error {
	return f.primary.Update(ctx, schema)
}

func (f *FallbackSchemaStore) Upsert(ctx context.Context, schema *Schema) error {
	return f.primary.Upsert(ctx, schema)
}

func (f *FallbackSchemaStore) Delete(ctx context.Context, key string) error {
	return f.primary.Delete(ctx, key)
}

func (f *FallbackSchemaStore) Exists(ctx context.Context, key string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.primary.Exists(ctx, key) {
		return true
	}
	return f.fallback.Exists(ctx, key)
}

func (f *FallbackSchemaStore) List(ctx context.Context) ([]*Schema, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	primarySchemas, err := f.primary.List(ctx)
	if err != nil {
		return nil, err
	}

	fallbackSchemas, err := f.fallback.List(ctx)
	if err != nil {
		return primarySchemas, nil
	}

	seen := make(map[string]bool, len(primarySchemas))
	for _, s := range primarySchemas {
		seen[s.Key] = true
	}

	for _, s := range fallbackSchemas {
		if !seen[s.Key] {
			primarySchemas = append(primarySchemas, s)
		}
	}

	return primarySchemas, nil
}

func (f *FallbackSchemaStore) Keys(ctx context.Context) ([]string, error) {
	schemas, err := f.List(ctx)
	if err != nil {
		return nil, err
	}

	keys := make([]string, len(schemas))
	for i, s := range schemas {
		keys[i] = s.Key
	}
	return keys, nil
}

func (f *FallbackSchemaStore) Count(ctx context.Context) (int, error) {
	schemas, err := f.List(ctx)
	if err != nil {
		return 0, err
	}
	return len(schemas), nil
}

func (f *FallbackSchemaStore) Clear(ctx context.Context) error {
	return f.primary.Clear(ctx)
}

// Compile-time check.
var _ SchemaStore = (*FallbackSchemaStore)(nil)
