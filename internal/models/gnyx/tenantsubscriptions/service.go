package tenantsubscriptions

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TenantSubscriptions, error)
	Create(tenantSubscription *TenantSubscriptions) error
	Update(tenantSubscription *TenantSubscriptions) error
	Delete(id string) error
}

type TenantSubscriptionsService[T TenantSubscriptions] struct {
	repo ORMRepository[T]
}

func NewService[T TenantSubscriptions](repo ORMRepository[T]) Service[T] {
	return &TenantSubscriptionsService[T]{repo: repo}
}

func (s *TenantSubscriptionsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *TenantSubscriptionsService[T]) GetByID(id string) (*TenantSubscriptions, error) {
	return s.repo.GetByID(id)
}

func (s *TenantSubscriptionsService[T]) Create(tenantSubscription *TenantSubscriptions) error {
	return s.repo.Create(tenantSubscription)
}

func (s *TenantSubscriptionsService[T]) Update(tenantSubscription *TenantSubscriptions) error {
	return s.repo.Update(tenantSubscription)
}

func (s *TenantSubscriptionsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
