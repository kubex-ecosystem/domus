package users

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Users, error)
	Create(user *Users) error
	Update(user *Users) error
	Delete(id string) error
}

type UsersService[T Users] struct {
	repo ORMRepository[T]
}

func NewService[T Users](repo ORMRepository[T]) Service[T] {
	return &UsersService[T]{repo: repo}
}

func (s *UsersService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *UsersService[T]) GetByID(id string) (*Users, error) {
	return s.repo.GetByID(id)
}

func (s *UsersService[T]) Create(user *Users) error {
	return s.repo.Create(user)
}

func (s *UsersService[T]) Update(user *Users) error {
	return s.repo.Update(user)
}

func (s *UsersService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
