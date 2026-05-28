// Package datastore fornece abstrações e implementações de stores de domínio sobre a Execution Layer.
package datastore

import (
	"context"
	"reflect"

	s "github.com/kubex-ecosystem/domus/internal/datastore/schemas_store"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

// Repository define operações CRUD genéricas para qualquer entidade.
// Implementações concretas podem adicionar métodos específicos ao domínio.
type Repository[T any] interface {
	// Create insere uma nova entidade e retorna o ID gerado.
	// Retorna erro se a operação falhar.
	Create(ctx context.Context, entity *T) (string, error)

	// GetByID retorna a entidade pelo ID.
	// Retorna (nil, nil) se não encontrada, seguindo convenção Kubex.
	GetByID(ctx context.Context, id string) (*T, error)

	// Update atualiza a entidade existente.
	// Retorna erro se não encontrada ou se falhar.
	Update(ctx context.Context, entity *T) error

	// Delete remove a entidade pelo ID.
	// Retorna erro se não encontrada ou se falhar.
	Delete(ctx context.Context, id string) error

	// List retorna entidades filtradas e paginadas.
	// Filters é um map genérico para permitir flexibilidade por store.
	List(ctx context.Context, filters map[string]any) (*t.PaginatedResult[T], error)
}

// ListFilters representa parâmetros comuns de listagem.
type ListFilters struct {
	Page   int
	Limit  int
	SortBy string
	Order  string // "ASC" ou "DESC"
}

// storeFactoryMap mantém um mapeamento de tipos para fábricas de stores.
var storeFactoryMap = make(map[reflect.Type]t.StoreType)

var registry *s.SchemaRegistry
