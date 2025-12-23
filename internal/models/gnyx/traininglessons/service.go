package traininglessons

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingLessons, error)
	Create(trainingLesson *TrainingLessons) error
	Update(trainingLesson *TrainingLessons) error
	Delete(id string) error
}

type TrainingLessonsService[T TrainingLessons] struct {
	repo ORMRepository[T]
}

func NewService[T TrainingLessons](repo ORMRepository[T]) Service[T] {
	return &TrainingLessonsService[T]{repo: repo}
}

func (s *TrainingLessonsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *TrainingLessonsService[T]) GetByID(id string) (*TrainingLessons, error) {
	return s.repo.GetByID(id)
}

func (s *TrainingLessonsService[T]) Create(trainingLesson *TrainingLessons) error {
	return s.repo.Create(trainingLesson)
}

func (s *TrainingLessonsService[T]) Update(trainingLesson *TrainingLessons) error {
	return s.repo.Update(trainingLesson)
}

func (s *TrainingLessonsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
