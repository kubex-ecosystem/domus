package pipelines

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Pipelines, error)
	Create(pipeline *Pipelines) error
	Update(pipeline *Pipelines) error
	Delete(id string) error
}

type PipelinesService[T Pipelines] struct {
	repo ORMRepository[T]
}

func NewService[T Pipelines](repo ORMRepository[T]) Service[T] {
	return &PipelinesService[T]{repo: repo}
}

func (s *PipelinesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *PipelinesService[T]) GetByID(id string) (*Pipelines, error) {
	return s.repo.GetByID(id)
}

func (s *PipelinesService[T]) Create(pipeline *Pipelines) error {
	return s.repo.Create(pipeline)
}

func (s *PipelinesService[T]) Update(pipeline *Pipelines) error {
	return s.repo.Update(pipeline)
}

func (s *PipelinesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
