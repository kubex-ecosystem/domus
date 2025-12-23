package subscriptionplans

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*SubscriptionPlans, error)
	Create(subscriptionPlan *SubscriptionPlans) error
	Update(subscriptionPlan *SubscriptionPlans) error
	Delete(id string) error
}

type SubscriptionPlansService[T SubscriptionPlans] struct {
	repo ORMRepository[T]
}

func NewService[T SubscriptionPlans](repo ORMRepository[T]) Service[T] {
	return &SubscriptionPlansService[T]{repo: repo}
}

func (s *SubscriptionPlansService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *SubscriptionPlansService[T]) GetByID(id string) (*SubscriptionPlans, error) {
	return s.repo.GetByID(id)
}

func (s *SubscriptionPlansService[T]) Create(subscriptionPlan *SubscriptionPlans) error {
	return s.repo.Create(subscriptionPlan)
}

func (s *SubscriptionPlansService[T]) Update(subscriptionPlan *SubscriptionPlans) error {
	return s.repo.Update(subscriptionPlan)
}

func (s *SubscriptionPlansService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
