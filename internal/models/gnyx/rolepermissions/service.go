package rolepermissions

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*RolePermissions, error)
	Create(rolePermission *RolePermissions) error
	Update(rolePermission *RolePermissions) error
	Delete(id string) error
}

type RolePermissionsService[T RolePermissions] struct {
	repo ORMRepository[T]
}

func NewService[T RolePermissions](repo ORMRepository[T]) Service[T] {
	return &RolePermissionsService[T]{repo: repo}
}

func (s *RolePermissionsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *RolePermissionsService[T]) GetByID(id string) (*RolePermissions, error) {
	return s.repo.GetByID(id)
}

func (s *RolePermissionsService[T]) Create(rolePermission *RolePermissions) error {
	return s.repo.Create(rolePermission)
}

func (s *RolePermissionsService[T]) Update(rolePermission *RolePermissions) error {
	return s.repo.Update(rolePermission)
}

func (s *RolePermissionsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
