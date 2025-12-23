package errorlogs

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*ErrorLogs, error)
	Create(errorLog *ErrorLogs) error
	Update(errorLog *ErrorLogs) error
	Delete(id string) error
}

type ErrorLogsService[T ErrorLogs] struct {
	repo ORMRepository[T]
}

func NewService[T ErrorLogs](repo ORMRepository[T]) Service[T] {
	return &ErrorLogsService[T]{repo: repo}
}

func (s *ErrorLogsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *ErrorLogsService[T]) GetByID(id string) (*ErrorLogs, error) {
	return s.repo.GetByID(id)
}

func (s *ErrorLogsService[T]) Create(errorLog *ErrorLogs) error {
	return s.repo.Create(errorLog)
}

func (s *ErrorLogsService[T]) Update(errorLog *ErrorLogs) error {
	return s.repo.Update(errorLog)
}

func (s *ErrorLogsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
