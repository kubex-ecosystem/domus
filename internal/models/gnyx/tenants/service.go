package tenants

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Tenants, error)
	Create(tenant *Tenants) error
	Update(tenant *Tenants) error
	Delete(id string) error
}

type TenantsService[T Tenants] struct {
	repo ORMRepository[T]
}

func NewService[T Tenants](repo ORMRepository[T]) Service[T] {
	return &TenantsService[T]{repo: repo}
}

func (s *TenantsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *TenantsService[T]) GetByID(id string) (*Tenants, error) {
	return s.repo.GetByID(id)
}

func (s *TenantsService[T]) Create(tenant *Tenants) error {
	return s.repo.Create(tenant)
}

func (s *TenantsService[T]) Update(tenant *Tenants) error {
	return s.repo.Update(tenant)
}

func (s *TenantsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
