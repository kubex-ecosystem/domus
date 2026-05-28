package schemasstore

import (
	"context"

	"github.com/kubex-ecosystem/domus/internal/execution"
)

// PersistentSchemaStore is a persistent implementation of SchemaStore.
// It uses GormExecutor for database operations.
type PersistentSchemaStore struct {
	exec execution.GormExecutor
}

// NewPersistentSchemaStore creates a new persistent schema store.
func NewPersistentSchemaStore(exec execution.Executor) *PersistentSchemaStore {
	return &PersistentSchemaStore{
		exec: exec.Gorm(),
	}
}

// NewPersistentSchemaStoreFromGorm creates a new persistent schema store from GormExecutor.
func NewPersistentSchemaStoreFromGorm(gormExec execution.GormExecutor) *PersistentSchemaStore {
	return &PersistentSchemaStore{
		exec: gormExec,
	}
}

func (p *PersistentSchemaStore) Register(ctx context.Context, schema *Schema) error {
	if err := schema.Validate(); err != nil {
		return err
	}

	existing, _ := p.Get(ctx, schema.Key)
	if existing != nil {
		return ErrSchemaAlreadyExists
	}

	result := p.exec.WithContext(ctx).Create(schema)
	return result.Error()
}

func (p *PersistentSchemaStore) Get(ctx context.Context, key string) (*Schema, error) {
	var schema Schema
	result := p.exec.WithContext(ctx).First(&schema, "key = ?", key)

	if result.IsNotFound() {
		return nil, ErrSchemaNotFound
	}

	if err := result.Error(); err != nil {
		return nil, err
	}

	return &schema, nil
}

func (p *PersistentSchemaStore) GetOrDefault(ctx context.Context, key string, defaultSchema *Schema) *Schema {
	schema, err := p.Get(ctx, key)
	if err != nil {
		return defaultSchema
	}
	return schema
}

func (p *PersistentSchemaStore) Update(ctx context.Context, schema *Schema) error {
	if err := schema.Validate(); err != nil {
		return err
	}

	existing, _ := p.Get(ctx, schema.Key)
	if existing == nil {
		return ErrSchemaNotFound
	}

	result := p.exec.WithContext(ctx).Model(&Schema{}).Where("key = ?", schema.Key).Updates(schema)
	return result.Error()
}

func (p *PersistentSchemaStore) Upsert(ctx context.Context, schema *Schema) error {
	if err := schema.Validate(); err != nil {
		return err
	}

	result := p.exec.WithContext(ctx).Save(schema)
	return result.Error()
}

func (p *PersistentSchemaStore) Delete(ctx context.Context, key string) error {
	existing, _ := p.Get(ctx, key)
	if existing == nil {
		return ErrSchemaNotFound
	}

	result := p.exec.WithContext(ctx).Delete(&Schema{}, "key = ?", key)
	return result.Error()
}

func (p *PersistentSchemaStore) Exists(ctx context.Context, key string) bool {
	var count int64
	p.exec.WithContext(ctx).Model(&Schema{}).Where("key = ?", key).Count(&count)
	return count > 0
}

func (p *PersistentSchemaStore) List(ctx context.Context) ([]*Schema, error) {
	var schemas []*Schema
	result := p.exec.WithContext(ctx).Find(&schemas)
	if err := result.Error(); err != nil {
		return nil, err
	}
	return schemas, nil
}

func (p *PersistentSchemaStore) Keys(ctx context.Context) ([]string, error) {
	var keys []string
	result := p.exec.WithContext(ctx).Model(&Schema{}).Pluck("key", &keys)
	if err := result.Error(); err != nil {
		return nil, err
	}
	return keys, nil
}

func (p *PersistentSchemaStore) Count(ctx context.Context) (int, error) {
	var count int64
	result := p.exec.WithContext(ctx).Model(&Schema{}).Count(&count)
	if err := result.Error(); err != nil {
		return 0, err
	}
	return int(count), nil
}

func (p *PersistentSchemaStore) Clear(ctx context.Context) error {
	result := p.exec.WithContext(ctx).Where("1 = 1").Delete(&Schema{})
	return result.Error()
}

// Migrate creates the schema table if it doesn't exist.
func (p *PersistentSchemaStore) Migrate() error {
	return p.exec.AutoMigrate(&Schema{})
}

// Compile-time check.
var _ SchemaStore = (*PersistentSchemaStore)(nil)
