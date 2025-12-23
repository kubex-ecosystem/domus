package accesslogs

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*AccessLogs, error)
	Create(accessLog *AccessLogs) error
	Update(accessLog *AccessLogs) error
	Delete(id string) error
}

type AccessLogsService[T AccessLogs] struct {
	repo ORMRepository[T]
}

func NewService[T AccessLogs](repo ORMRepository[T]) Service[T] {
	return &AccessLogsService[T]{repo: repo}
}

func (s *AccessLogsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *AccessLogsService[T]) GetByID(id string) (*AccessLogs, error) {
	return s.repo.GetByID(id)
}

func (s *AccessLogsService[T]) Create(accessLog *AccessLogs) error {
	return s.repo.Create(accessLog)
}

func (s *AccessLogsService[T]) Update(accessLog *AccessLogs) error {
	return s.repo.Update(accessLog)
}

func (s *AccessLogsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
