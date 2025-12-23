package permissions

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Permissions, error)
	Create(permission *Permissions) error
	Update(permission *Permissions) error
	Delete(id string) error
}

type PermissionsService[T Permissions] struct {
	repo ORMRepository[T]
}

func NewService[T Permissions](repo ORMRepository[T]) Service[T] {
	return &PermissionsService[T]{repo: repo}
}

func (s *PermissionsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *PermissionsService[T]) GetByID(id string) (*Permissions, error) {
	return s.repo.GetByID(id)
}

func (s *PermissionsService[T]) Create(permission *Permissions) error {
	return s.repo.Create(permission)
}

func (s *PermissionsService[T]) Update(permission *Permissions) error {
	return s.repo.Update(permission)
}

func (s *PermissionsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
