package trainingbadges

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingBadges, error)
	Create(trainingBadge *TrainingBadges) error
	Update(trainingBadge *TrainingBadges) error
	Delete(id string) error
}

type TrainingBadgesService[T TrainingBadges] struct {
	repo ORMRepository[T]
}

func NewService[T TrainingBadges](repo ORMRepository[T]) Service[T] {
	return &TrainingBadgesService[T]{repo: repo}
}

func (s *TrainingBadgesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *TrainingBadgesService[T]) GetByID(id string) (*TrainingBadges, error) {
	return s.repo.GetByID(id)
}

func (s *TrainingBadgesService[T]) Create(trainingBadge *TrainingBadges) error {
	return s.repo.Create(trainingBadge)
}

func (s *TrainingBadgesService[T]) Update(trainingBadge *TrainingBadges) error {
	return s.repo.Update(trainingBadge)
}

func (s *TrainingBadgesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
