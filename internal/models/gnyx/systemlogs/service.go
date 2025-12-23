package systemlogs

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*SystemLogs, error)
	Create(systemLog *SystemLogs) error
	Update(systemLog *SystemLogs) error
	Delete(id string) error
}

type SystemLogsService[T SystemLogs] struct {
	repo ORMRepository[T]
}

func NewService[T SystemLogs](repo ORMRepository[T]) Service[T] {
	return &SystemLogsService[T]{repo: repo}
}

func (s *SystemLogsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *SystemLogsService[T]) GetByID(id string) (*SystemLogs, error) {
	return s.repo.GetByID(id)
}

func (s *SystemLogsService[T]) Create(systemLog *SystemLogs) error {
	return s.repo.Create(systemLog)
}

func (s *SystemLogsService[T]) Update(systemLog *SystemLogs) error {
	return s.repo.Update(systemLog)
}

func (s *SystemLogsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
