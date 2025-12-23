package paymentmethods

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*PaymentMethods, error)
	Create(paymentMethod *PaymentMethods) error
	Update(paymentMethod *PaymentMethods) error
	Delete(id string) error
}

type PaymentMethodsService[T PaymentMethods] struct {
	repo ORMRepository[T]
}

func NewService[T PaymentMethods](repo ORMRepository[T]) Service[T] {
	return &PaymentMethodsService[T]{repo: repo}
}

func (s *PaymentMethodsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *PaymentMethodsService[T]) GetByID(id string) (*PaymentMethods, error) {
	return s.repo.GetByID(id)
}

func (s *PaymentMethodsService[T]) Create(paymentMethod *PaymentMethods) error {
	return s.repo.Create(paymentMethod)
}

func (s *PaymentMethodsService[T]) Update(paymentMethod *PaymentMethods) error {
	return s.repo.Update(paymentMethod)
}

func (s *PaymentMethodsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
