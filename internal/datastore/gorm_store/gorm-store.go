// Package gormstore provides a GORM-backed implementation of the generic Repository interface.
package gormstore

import (
	"context"
	"errors"
	"reflect"

	"github.com/kubex-ecosystem/domus/internal/datastore"
	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

// GormRepository is the single generic GORM-backed repository implementation.
// It uses execution.GormExecutor to avoid leaking gorm.DB to consumers.
type GormRepository[T any] struct {
	exec execution.GormExecutor
}

// NewGormRepository creates a new generic GORM repository.
// It extracts the GormExecutor from the provided Executor.
func NewGormRepository[T any](exec execution.Executor) datastore.Repository[T] {
	return &GormRepository[T]{exec: exec.Gorm()}
}

// NewGormRepositoryFromGormExecutor creates a new generic GORM repository
// directly from a GormExecutor (useful for testing or when you already have one).
func NewGormRepositoryFromGormExecutor[T any](gormExec execution.GormExecutor) datastore.Repository[T] {
	return &GormRepository[T]{exec: gormExec}
}

// extractID extracts the ID field from the entity.
// This is intentionally minimal and fails fast.
func extractID[T any](entity *T) (string, error) {
	v := reflect.ValueOf(entity)
	if v.Kind() != reflect.Pointer {
		return "", errors.New("entity must be a pointer")
	}

	elem := v.Elem()
	if !elem.IsValid() {
		return "", errors.New("invalid entity")
	}

	id := elem.FieldByName("ID")
	if !id.IsValid() || id.Kind() != reflect.String {
		return "", errors.New("entity has no string ID field")
	}

	return id.String(), nil
}

// Create inserts a new entity and returns its ID.
func (r *GormRepository[T]) Create(ctx context.Context, entity *T) (string, error) {
	result := r.exec.WithContext(ctx).Create(entity)
	if err := result.Error(); err != nil {
		return "", err
	}
	return extractID(entity)
}

// GetByID returns an entity by ID or (nil, nil) if not found.
func (r *GormRepository[T]) GetByID(ctx context.Context, id string) (*T, error) {
	var entity T
	result := r.exec.WithContext(ctx).First(&entity, "id = ?", id)

	if result.IsNotFound() {
		return nil, nil
	}

	if err := result.Error(); err != nil {
		return nil, err
	}

	return &entity, nil
}

// Update updates an existing entity.
func (r *GormRepository[T]) Update(ctx context.Context, entity *T) error {
	result := r.exec.WithContext(ctx).Model(entity).Updates(entity)
	return result.Error()
}

// Delete removes an entity by ID.
func (r *GormRepository[T]) Delete(ctx context.Context, id string) error {
	var entity T
	result := r.exec.WithContext(ctx).Delete(&entity, "id = ?", id)
	return result.Error()
}

// List returns a paginated result set.
// Pagination and advanced filtering are intentionally minimal for now.
func (r *GormRepository[T]) List(
	ctx context.Context,
	filters map[string]any,
) (*t.PaginatedResult[T], error) {
	var entities []T

	query := r.exec.WithContext(ctx)

	for k, v := range filters {
		query = query.Where(k+" = ?", v)
	}

	result := query.Find(&entities)
	if err := result.Error(); err != nil {
		return nil, err
	}

	return &t.PaginatedResult[T]{
		Data:  entities,
		Total: int64(len(entities)),
	}, nil
}
