package activities

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Activities, error)
	Create(activity *Activities) error
	Update(activity *Activities) error
	Delete(id string) error
}

type ActivitiesService[T Activities] struct {
	repo ORMRepository[T]
}

func NewService[T Activities](repo ORMRepository[T]) Service[T] {
	return &ActivitiesService[T]{repo: repo}
}

func (s *ActivitiesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *ActivitiesService[T]) GetByID(id string) (*Activities, error) {
	return s.repo.GetByID(id)
}

func (s *ActivitiesService[T]) Create(activity *Activities) error {
	return s.repo.Create(activity)
}

func (s *ActivitiesService[T]) Update(activity *Activities) error {
	return s.repo.Update(activity)
}

func (s *ActivitiesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
