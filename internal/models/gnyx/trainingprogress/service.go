package trainingprogress

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingProgress, error)
	Create(trainingProgress *TrainingProgress) error
	Update(trainingProgress *TrainingProgress) error
	Delete(id string) error
}

type TrainingProgressService[T TrainingProgress] struct {
	repo ORMRepository[T]
}

func NewService[T TrainingProgress](repo ORMRepository[T]) Service[T] {
	return &TrainingProgressService[T]{repo: repo}
}

func (s *TrainingProgressService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *TrainingProgressService[T]) GetByID(id string) (*TrainingProgress, error) {
	return s.repo.GetByID(id)
}

func (s *TrainingProgressService[T]) Create(trainingProgress *TrainingProgress) error {
	return s.repo.Create(trainingProgress)
}

func (s *TrainingProgressService[T]) Update(trainingProgress *TrainingProgress) error {
	return s.repo.Update(trainingProgress)
}

func (s *TrainingProgressService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
