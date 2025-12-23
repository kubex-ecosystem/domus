package roleconfig

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id int64) (*RoleConfig, error)
	Create(roleConfig *RoleConfig) error
	Update(roleConfig *RoleConfig) error
	Delete(id int64) error
}

type RoleConfigService[T RoleConfig] struct {
	repo ORMRepository[T]
}

func NewService[T RoleConfig](repo ORMRepository[T]) Service[T] {
	return &RoleConfigService[T]{repo: repo}
}

func (s *RoleConfigService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *RoleConfigService[T]) GetByID(id int64) (*RoleConfig, error) {
	return s.repo.GetByID(id)
}

func (s *RoleConfigService[T]) Create(roleConfig *RoleConfig) error {
	return s.repo.Create(roleConfig)
}

func (s *RoleConfigService[T]) Update(roleConfig *RoleConfig) error {
	return s.repo.Update(roleConfig)
}

func (s *RoleConfigService[T]) Delete(id int64) error {
	return s.repo.Delete(id)
}
