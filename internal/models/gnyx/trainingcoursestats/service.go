package trainingcoursestats

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingCourseStats, error)
	Create(trainingCourseStat *TrainingCourseStats) error
	Update(trainingCourseStat *TrainingCourseStats) error
	Delete(id string) error
}

type TrainingCourseStatsService[T TrainingCourseStats] struct {
	repo ORMRepository[T]
}

func NewService[T TrainingCourseStats](repo ORMRepository[T]) Service[T] {
	return &TrainingCourseStatsService[T]{repo: repo}
}

func (s *TrainingCourseStatsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *TrainingCourseStatsService[T]) GetByID(id string) (*TrainingCourseStats, error) {
	return s.repo.GetByID(id)
}

func (s *TrainingCourseStatsService[T]) Create(trainingCourseStat *TrainingCourseStats) error {
	return s.repo.Create(trainingCourseStat)
}

func (s *TrainingCourseStatsService[T]) Update(trainingCourseStat *TrainingCourseStats) error {
	return s.repo.Update(trainingCourseStat)
}

func (s *TrainingCourseStatsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
