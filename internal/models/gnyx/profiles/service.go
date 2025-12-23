package profiles

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Profiles, error)
	Create(profile *Profiles) error
	Update(profile *Profiles) error
	Delete(id string) error
}

type ProfilesService[T Profiles] struct {
	repo ORMRepository[T]
}

func NewService[T Profiles](repo ORMRepository[T]) Service[T] {
	return &ProfilesService[T]{repo: repo}
}

func (s *ProfilesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *ProfilesService[T]) GetByID(id string) (*Profiles, error) {
	return s.repo.GetByID(id)
}

func (s *ProfilesService[T]) Create(profile *Profiles) error {
	return s.repo.Create(profile)
}

func (s *ProfilesService[T]) Update(profile *Profiles) error {
	return s.repo.Update(profile)
}

func (s *ProfilesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
