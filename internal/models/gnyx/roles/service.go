package roles

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Roles, error)
	Create(role *Roles) error
	Update(role *Roles) error
	Delete(id string) error
}

type RolesService[T Roles] struct {
	repo ORMRepository[T]
}

func NewService[T Roles](repo ORMRepository[T]) Service[T] {
	return &RolesService[T]{repo: repo}
}

func (s *RolesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *RolesService[T]) GetByID(id string) (*Roles, error) {
	return s.repo.GetByID(id)
}

func (s *RolesService[T]) Create(role *Roles) error {
	return s.repo.Create(role)
}

func (s *RolesService[T]) Update(role *Roles) error {
	return s.repo.Update(role)
}

func (s *RolesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
