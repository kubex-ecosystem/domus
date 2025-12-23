package trainingcourses

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingCourses, error)
	Create(trainingCourse *TrainingCourses) error
	Update(trainingCourse *TrainingCourses) error
	Delete(id string) error
}

type TrainingCoursesService[T TrainingCourses] struct {
	repo ORMRepository[T]
}

func NewService[T TrainingCourses](repo ORMRepository[T]) Service[T] {
	return &TrainingCoursesService[T]{repo: repo}
}

func (s *TrainingCoursesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *TrainingCoursesService[T]) GetByID(id string) (*TrainingCourses, error) {
	return s.repo.GetByID(id)
}

func (s *TrainingCoursesService[T]) Create(trainingCourse *TrainingCourses) error {
	return s.repo.Create(trainingCourse)
}

func (s *TrainingCoursesService[T]) Update(trainingCourse *TrainingCourses) error {
	return s.repo.Update(trainingCourse)
}

func (s *TrainingCoursesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
