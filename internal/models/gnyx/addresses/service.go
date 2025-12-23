package addresses

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Addresses, error)
	Create(address *Addresses) error
	Update(address *Addresses) error
	Delete(id string) error
}

type AddressesService[T Addresses] struct {
	repo ORMRepository[T]
}

func NewService[T Addresses](repo ORMRepository[T]) Service[T] {
	return &AddressesService[T]{repo: repo}
}

func (s *AddressesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *AddressesService[T]) GetByID(id string) (*Addresses, error) {
	return s.repo.GetByID(id)
}

func (s *AddressesService[T]) Create(address *Addresses) error {
	return s.repo.Create(address)
}

func (s *AddressesService[T]) Update(address *Addresses) error {
	return s.repo.Update(address)
}

func (s *AddressesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
