package pipelinestages

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*PipelineStages, error)
	Create(pipelineStage *PipelineStages) error
	Update(pipelineStage *PipelineStages) error
	Delete(id string) error
}

type PipelineStagesService[T PipelineStages] struct {
	repo ORMRepository[T]
}

func NewService[T PipelineStages](repo ORMRepository[T]) Service[T] {
	return &PipelineStagesService[T]{repo: repo}
}

func (s *PipelineStagesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *PipelineStagesService[T]) GetByID(id string) (*PipelineStages, error) {
	return s.repo.GetByID(id)
}

func (s *PipelineStagesService[T]) Create(pipelineStage *PipelineStages) error {
	return s.repo.Create(pipelineStage)
}

func (s *PipelineStagesService[T]) Update(pipelineStage *PipelineStages) error {
	return s.repo.Update(pipelineStage)
}

func (s *PipelineStagesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
