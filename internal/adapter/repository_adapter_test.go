package adapter

import (
	"context"
	"errors"
	"testing"

	store "github.com/kubex-ecosystem/domus/internal/datastore"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

// TestEntity é uma entidade simples para testes
type TestEntity struct {
	ID    string
	Name  string
	Email string
}

// mockStoreRepo é um mock de Repository[T] para testes
type mockStoreRepo struct {
	createFunc  func(ctx context.Context, entity *TestEntity) (string, error)
	getByIDFunc func(ctx context.Context, id string) (*TestEntity, error)
	updateFunc  func(ctx context.Context, entity *TestEntity) error
	deleteFunc  func(ctx context.Context, id string) error
	listFunc    func(ctx context.Context, filters map[string]any) (*t.PaginatedResult[TestEntity], error)
}

func (m *mockStoreRepo) Create(ctx context.Context, entity *TestEntity) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, entity)
	}
	return "store-id-123", nil
}

func (m *mockStoreRepo) GetByID(ctx context.Context, id string) (*TestEntity, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &TestEntity{ID: id, Name: "Store User"}, nil
}

func (m *mockStoreRepo) Update(ctx context.Context, entity *TestEntity) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, entity)
	}
	return nil
}

func (m *mockStoreRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockStoreRepo) List(ctx context.Context, filters map[string]any) (*t.PaginatedResult[TestEntity], error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filters)
	}
	return &t.PaginatedResult[TestEntity]{
		Data:  []TestEntity{{ID: "1", Name: "Store User 1"}},
		Total: 1,
	}, nil
}

// mockORMRepo é um mock de ORMRepository[T] para testes
type mockORMRepo struct {
	createFunc  func(entity *TestEntity) error
	getByIDFunc func(id string) (*TestEntity, error)
	updateFunc  func(entity *TestEntity) error
	deleteFunc  func(id string) error
	getAllFunc  func() ([]TestEntity, error)
}

func (m *mockORMRepo) Create(entity *TestEntity) error {
	if m.createFunc != nil {
		return m.createFunc(entity)
	}
	entity.ID = "orm-id-456" // Simula set do ID pelo GORM
	return nil
}

func (m *mockORMRepo) GetByID(id string) (*TestEntity, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(id)
	}
	return &TestEntity{ID: id, Name: "ORM User"}, nil
}

func (m *mockORMRepo) Update(entity *TestEntity) error {
	if m.updateFunc != nil {
		return m.updateFunc(entity)
	}
	return nil
}

func (m *mockORMRepo) Delete(id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(id)
	}
	return nil
}

func (m *mockORMRepo) GetAll() ([]TestEntity, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc()
	}
	return []TestEntity{{ID: "1", Name: "ORM User 1"}}, nil
}

// TestNewDSRepository_ValidateConfig testa validação de configuração
func TestNewDSRepository_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		storeRepo   store.Repository[TestEntity]
		ormRepo     ORMRepository[TestEntity]
		config      *RepositoryConfig
		expectError bool
	}{
		{
			name:      "ForceStore sem storeRepo deve falhar",
			storeRepo: nil,
			ormRepo:   &mockORMRepo{},
			config: &RepositoryConfig{
				ForceStore: true,
			},
			expectError: true,
		},
		{
			name:      "ForceORM sem ormRepo deve falhar",
			storeRepo: &mockStoreRepo{},
			ormRepo:   nil,
			config: &RepositoryConfig{
				ForceORM: true,
			},
			expectError: true,
		},
		{
			name:      "ForceStore E ForceORM deve falhar",
			storeRepo: &mockStoreRepo{},
			ormRepo:   &mockORMRepo{},
			config: &RepositoryConfig{
				ForceStore: true,
				ForceORM:   true,
			},
			expectError: true,
		},
		{
			name:        "Ambos nil deve falhar",
			storeRepo:   nil,
			ormRepo:     nil,
			config:      DefaultConfig(),
			expectError: true,
		},
		{
			name:        "Apenas Store deve funcionar",
			storeRepo:   &mockStoreRepo{},
			ormRepo:     nil,
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name:        "Apenas ORM deve funcionar",
			storeRepo:   nil,
			ormRepo:     &mockORMRepo{},
			config:      DefaultConfig(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDSRepository[TestEntity](tt.storeRepo, tt.ormRepo, tt.config)
			if tt.expectError && err == nil {
				t.Errorf("esperava erro, mas obteve nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("não esperava erro, mas obteve: %v", err)
			}
		})
	}
}

// TestDSRepository_Create_PreferStore testa Create com preferência por Store
func TestDSRepository_Create_PreferStore(t *testing.T) {
	ctx := context.Background()

	storeRepo := &mockStoreRepo{
		createFunc: func(ctx context.Context, entity *TestEntity) (string, error) {
			return "store-created-123", nil
		},
	}

	ormRepo := &mockORMRepo{
		createFunc: func(entity *TestEntity) error {
			t.Fatal("ORM não deveria ser chamado quando Store funciona")
			return nil
		},
	}

	repo, err := NewDSRepository[TestEntity](storeRepo, ormRepo, DefaultConfig())
	if err != nil {
		t.Fatalf("falha ao criar adapter: %v", err)
	}

	entity := &TestEntity{Name: "Test"}
	id, err := repo.Create(ctx, entity)
	if err != nil {
		t.Fatalf("Create falhou: %v", err)
	}

	if id != "store-created-123" {
		t.Errorf("ID esperado 'store-created-123', obteve '%s'", id)
	}
}

// TestDSRepository_Create_FallbackToORM testa fallback para ORM quando Store falha
func TestDSRepository_Create_FallbackToORM(t *testing.T) {
	ctx := context.Background()

	storeRepo := &mockStoreRepo{
		createFunc: func(ctx context.Context, entity *TestEntity) (string, error) {
			return "", errors.New("store falhou")
		},
	}

	ormCalled := false
	ormRepo := &mockORMRepo{
		createFunc: func(entity *TestEntity) error {
			ormCalled = true
			entity.ID = "orm-fallback-456"
			return nil
		},
	}

	repo, err := NewDSRepository[TestEntity](storeRepo, ormRepo, DefaultConfig())
	if err != nil {
		t.Fatalf("falha ao criar adapter: %v", err)
	}

	entity := &TestEntity{Name: "Test"}
	id, err := repo.Create(ctx, entity)
	if err != nil {
		t.Fatalf("Create falhou: %v", err)
	}

	if !ormCalled {
		t.Error("ORM deveria ter sido chamado em fallback")
	}

	if id != "orm-fallback-456" {
		t.Errorf("ID esperado 'orm-fallback-456', obteve '%s'", id)
	}
}

// TestDSRepository_Create_ForceStore testa ForceStore sem fallback
func TestDSRepository_Create_ForceStore(t *testing.T) {
	ctx := context.Background()

	storeRepo := &mockStoreRepo{
		createFunc: func(ctx context.Context, entity *TestEntity) (string, error) {
			return "", errors.New("store falhou")
		},
	}

	ormRepo := &mockORMRepo{}

	config := &RepositoryConfig{
		PreferStore:   true,
		FallbackToORM: false,
		ForceStore:    true,
	}

	repo, err := NewDSRepository[TestEntity](storeRepo, ormRepo, config)
	if err != nil {
		t.Fatalf("falha ao criar adapter: %v", err)
	}

	entity := &TestEntity{Name: "Test"}
	_, err = repo.Create(ctx, entity)
	if err == nil {
		t.Error("esperava erro quando Store falha com ForceStore, mas obteve nil")
	}
}

// TestDSRepository_GetByID_StoreReturnsNil testa (nil, nil) do Store
func TestDSRepository_GetByID_StoreReturnsNil(t *testing.T) {
	ctx := context.Background()

	storeRepo := &mockStoreRepo{
		getByIDFunc: func(ctx context.Context, id string) (*TestEntity, error) {
			return nil, nil // Convenção Kubex: não encontrado
		},
	}

	ormRepo := &mockORMRepo{}

	repo, err := NewDSRepository[TestEntity](storeRepo, ormRepo, DefaultConfig())
	if err != nil {
		t.Fatalf("falha ao criar adapter: %v", err)
	}

	result, err := repo.GetByID(ctx, "123")
	if err != nil {
		t.Fatalf("GetByID falhou: %v", err)
	}

	if result != nil {
		t.Errorf("esperava nil quando Store retorna (nil, nil), obteve %+v", result)
	}
}

// TestDSRepository_List_ORMFallback testa conversão GetAll -> List
func TestDSRepository_List_ORMFallback(t *testing.T) {
	ctx := context.Background()

	ormRepo := &mockORMRepo{
		getAllFunc: func() ([]TestEntity, error) {
			return []TestEntity{
				{ID: "1", Name: "User 1"},
				{ID: "2", Name: "User 2"},
			}, nil
		},
	}

	repo, err := NewDSRepository[TestEntity](nil, ormRepo, DefaultConfig())
	if err != nil {
		t.Fatalf("falha ao criar adapter: %v", err)
	}

	result, err := repo.List(ctx, nil)
	if err != nil {
		t.Fatalf("List falhou: %v", err)
	}

	if len(result.Data) != 2 {
		t.Errorf("esperava 2 itens, obteve %d", len(result.Data))
	}

	if result.Total != 2 {
		t.Errorf("esperava Total=2, obteve %d", result.Total)
	}
}

// TestDSRepository_HasStore_HasORM testa métodos de introspecção
func TestDSRepository_HasStore_HasORM(t *testing.T) {
	storeRepo := &mockStoreRepo{}
	ormRepo := &mockORMRepo{}

	tests := []struct {
		name           string
		storeRepo      store.Repository[TestEntity]
		ormRepo        ORMRepository[TestEntity]
		expectHasStore bool
		expectHasORM   bool
	}{
		{
			name:           "Ambos disponíveis",
			storeRepo:      storeRepo,
			ormRepo:        ormRepo,
			expectHasStore: true,
			expectHasORM:   true,
		},
		{
			name:           "Apenas Store",
			storeRepo:      storeRepo,
			ormRepo:        nil,
			expectHasStore: true,
			expectHasORM:   false,
		},
		{
			name:           "Apenas ORM",
			storeRepo:      nil,
			ormRepo:        ormRepo,
			expectHasStore: false,
			expectHasORM:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewDSRepository[TestEntity](tt.storeRepo, tt.ormRepo, DefaultConfig())
			if err != nil {
				t.Fatalf("falha ao criar adapter: %v", err)
			}

			if repo.HasStore() != tt.expectHasStore {
				t.Errorf("HasStore: esperava %v, obteve %v", tt.expectHasStore, repo.HasStore())
			}

			if repo.HasORM() != tt.expectHasORM {
				t.Errorf("HasORM: esperava %v, obteve %v", tt.expectHasORM, repo.HasORM())
			}
		})
	}
}
