// Package adapter fornece adaptadores unificados para diferentes implementações de repositórios.
//
// DSRepository[T] é um adapter que unifica:
// - Repository[T] (baseado em PGExecutor/DSStore com context.Context)
// - ORMRepository[T] (baseado em GORM sem context.Context)
//
// Permite uso transparente pelo Service/Controller sem saber qual backend está sendo usado.
package adapter

import (
	"context"
	"fmt"
	"reflect"

	store "github.com/kubex-ecosystem/domus/internal/datastore"
	t "github.com/kubex-ecosystem/domus/internal/types"
	"github.com/kubex-ecosystem/logz"
)

// ORMRepository define a interface genérica dos repositórios GORM.
// Esta é a mesma interface definida nos 97 modelos importados.
type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*T, error)
	Create(entity *T) error
	Update(entity *T) error
	Delete(id string) error
}

// DSRepository é o adapter unificado que implementa Repository[T].
// Internamente pode usar StoreRepository[T] (PGExecutor) ou ORMRepository[T] (GORM).
type DSRepository[T any] struct {
	// storeRepo é o repository baseado em PGExecutor/DSStore (com context)
	storeRepo store.Repository[T]

	// ormRepo é o repository baseado em GORM (sem context)
	ormRepo ORMRepository[T]

	// config define a política de uso
	config *RepositoryConfig

	// entityType armazena o tipo da entidade para logs
	entityType reflect.Type
}

// RepositoryConfig define a política de uso do adapter.
type RepositoryConfig struct {
	// PreferStore define se deve tentar usar storeRepo primeiro
	// Default: true (usar DSStore quando disponível)
	PreferStore bool

	// FallbackToORM permite fallback automático para ORM se store falhar
	// Default: true
	FallbackToORM bool

	// ForceStore força uso exclusivo de storeRepo (erro se não disponível)
	// Default: false
	ForceStore bool

	// ForceORM força uso exclusivo de ormRepo (erro se não disponível)
	// Default: false
	ForceORM bool
}

// DefaultConfig retorna a configuração padrão do adapter.
func DefaultConfig() *RepositoryConfig {
	return &RepositoryConfig{
		PreferStore:   true,
		FallbackToORM: true,
		ForceStore:    false,
		ForceORM:      false,
	}
}

// NewDSRepository cria um novo adapter unificado.
//
// Pode receber:
// - Apenas storeRepo (ormRepo = nil) → usa apenas Store
// - Apenas ormRepo (storeRepo = nil) → usa apenas ORM
// - Ambos → usa política de fallback definida em config
//
// Se config for nil, usa DefaultConfig().
func NewDSRepository[T any](
	storeRepo store.Repository[T],
	ormRepo ORMRepository[T],
	config *RepositoryConfig,
) (*DSRepository[T], error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validação de configuração
	if config.ForceStore && storeRepo == nil {
		return nil, fmt.Errorf("ForceStore=true mas storeRepo é nil")
	}
	if config.ForceORM && ormRepo == nil {
		return nil, fmt.Errorf("ForceORM=true mas ormRepo é nil")
	}
	if config.ForceStore && config.ForceORM {
		return nil, fmt.Errorf("ForceStore e ForceORM não podem ser ambos true")
	}

	// Ao menos um repository deve existir
	if storeRepo == nil && ormRepo == nil {
		return nil, fmt.Errorf("ao menos um repository (store ou orm) deve ser fornecido")
	}

	return &DSRepository[T]{
		storeRepo:  storeRepo,
		ormRepo:    ormRepo,
		config:     config,
		entityType: reflect.TypeFor[T](),
	}, nil
}

// Create insere uma nova entidade e retorna o ID gerado.
func (r *DSRepository[T]) Create(ctx context.Context, entity *T) (string, error) {
	// Política: tenta store primeiro se disponível e preferido
	if r.shouldUseStore("Create") {
		id, err := r.storeRepo.Create(ctx, entity)
		if err == nil {
			return id, nil
		}

		// Se não deve fazer fallback, retorna erro
		if !r.config.FallbackToORM {
			return "", err
		}

		// Log do fallback
		logz.Warn("DSRepository[%s].Create: store falhou (%v), tentando ORM fallback",
			r.entityType.Name(), err)
	}

	// Usa ORM
	if r.ormRepo == nil {
		return "", fmt.Errorf("ORM repository não disponível para %s", r.entityType.Name())
	}

	if err := r.ormRepo.Create(entity); err != nil {
		return "", err
	}

	// ORMRepository não retorna ID, tenta extrair do campo ID da entidade
	id, err := r.extractID(entity)
	if err != nil {
		logz.Warn("DSRepository[%s].Create: não foi possível extrair ID após ORM Create: %v",
			r.entityType.Name(), err)
		return "", nil // Retorna vazio mas sem erro (entity foi criada)
	}

	return id, nil
}

// GetByID retorna a entidade pelo ID.
// Retorna (nil, nil) se não encontrada, seguindo convenção Kubex.
func (r *DSRepository[T]) GetByID(ctx context.Context, id string) (*T, error) {
	if r.shouldUseStore("GetByID") {
		entity, err := r.storeRepo.GetByID(ctx, id)
		if err == nil || entity != nil {
			return entity, err
		}

		if !r.config.FallbackToORM {
			return entity, err
		}

		logz.Warn("DSRepository[%s].GetByID: store falhou ou não encontrou (%v), tentando ORM fallback",
			r.entityType.Name(), err)
	}

	if r.ormRepo == nil {
		return nil, fmt.Errorf("ORM repository não disponível para %s", r.entityType.Name())
	}

	entity, err := r.ormRepo.GetByID(id)
	if err != nil {
		// GORM retorna erro se não encontrado, normaliza para (nil, nil)
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return entity, nil
}

// Update atualiza a entidade existente.
func (r *DSRepository[T]) Update(ctx context.Context, entity *T) error {
	if r.shouldUseStore("Update") {
		err := r.storeRepo.Update(ctx, entity)
		if err == nil {
			return nil
		}

		if !r.config.FallbackToORM {
			return err
		}

		logz.Warn("DSRepository[%s].Update: store falhou (%v), tentando ORM fallback",
			r.entityType.Name(), err)
	}

	if r.ormRepo == nil {
		return fmt.Errorf("ORM repository não disponível para %s", r.entityType.Name())
	}

	return r.ormRepo.Update(entity)
}

// Delete remove a entidade pelo ID.
func (r *DSRepository[T]) Delete(ctx context.Context, id string) error {
	if r.shouldUseStore("Delete") {
		err := r.storeRepo.Delete(ctx, id)
		if err == nil {
			return nil
		}

		if !r.config.FallbackToORM {
			return err
		}

		logz.Warn("DSRepository[%s].Delete: store falhou (%v), tentando ORM fallback",
			r.entityType.Name(), err)
	}

	if r.ormRepo == nil {
		return fmt.Errorf("ORM repository não disponível para %s", r.entityType.Name())
	}

	return r.ormRepo.Delete(id)
}

// List retorna entidades filtradas e paginadas.
func (r *DSRepository[T]) List(ctx context.Context, filters map[string]any) (*t.PaginatedResult[T], error) {
	if r.shouldUseStore("List") {
		result, err := r.storeRepo.List(ctx, filters)
		if err == nil {
			return result, nil
		}

		if !r.config.FallbackToORM {
			return nil, err
		}

		logz.Warn("DSRepository[%s].List: store falhou (%v), tentando ORM fallback",
			r.entityType.Name(), err)
	}

	if r.ormRepo == nil {
		return nil, fmt.Errorf("ORM repository não disponível para %s", r.entityType.Name())
	}

	// ORM só tem GetAll(), sem filtros nem paginação
	// Converte GetAll() para PaginatedResult
	all, err := r.ormRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// Cria resultado paginado simples (sem paginação real)
	return &t.PaginatedResult[T]{
		Data:       all,
		Total:      int64(len(all)),
		Page:       1,
		Limit:      len(all),
		TotalPages: 1,
	}, nil
}

// shouldUseStore determina se deve usar storeRepo baseado na política.
func (r *DSRepository[T]) shouldUseStore(operation string) bool {
	// ForceORM tem prioridade
	if r.config.ForceORM {
		return false
	}

	// ForceStore tem prioridade
	if r.config.ForceStore {
		if r.storeRepo == nil {
			panic(fmt.Sprintf("ForceStore=true mas storeRepo é nil para operação %s", operation))
		}
		return true
	}

	// Política padrão: usar store se disponível e preferido
	return r.storeRepo != nil && r.config.PreferStore
}

// extractID tenta extrair o campo ID da entidade usando reflection.
// Procura por campos: ID, Id, id
func (r *DSRepository[T]) extractID(entity *T) (string, error) {
	if entity == nil {
		return "", fmt.Errorf("entity é nil")
	}

	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return "", fmt.Errorf("entity não é struct")
	}

	// Tenta campos comuns de ID
	for _, fieldName := range []string{"ID", "Id", "id"} {
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			continue
		}

		// Converte para string se possível
		switch field.Kind() {
		case reflect.String:
			return field.String(), nil
		case reflect.Int, reflect.Int64, reflect.Int32:
			return fmt.Sprintf("%d", field.Int()), nil
		case reflect.Uint, reflect.Uint64, reflect.Uint32:
			return fmt.Sprintf("%d", field.Uint()), nil
		}
	}

	return "", fmt.Errorf("campo ID não encontrado ou tipo não suportado")
}

// isNotFoundError verifica se o erro é de "registro não encontrado".
// Compatível com GORM e outros ORMs.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return errMsg == "record not found" ||
		errMsg == "not found" ||
		errMsg == "no rows in result set"
}

// Config retorna a configuração atual do adapter.
func (r *DSRepository[T]) Config() *RepositoryConfig {
	return r.config
}

// HasStore retorna true se o adapter possui storeRepo configurado.
func (r *DSRepository[T]) HasStore() bool {
	return r.storeRepo != nil
}

// HasORM retorna true se o adapter possui ormRepo configurado.
func (r *DSRepository[T]) HasORM() bool {
	return r.ormRepo != nil
}

// EntityType retorna o tipo da entidade gerenciada.
func (r *DSRepository[T]) EntityType() reflect.Type {
	return r.entityType
}
