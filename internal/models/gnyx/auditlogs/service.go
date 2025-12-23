package auditlogs

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*AuditLogs, error)
	Create(auditLog *AuditLogs) error
	Update(auditLog *AuditLogs) error
	Delete(id string) error
}

type AuditLogsService[T AuditLogs] struct {
	repo ORMRepository[T]
}

func NewService[T AuditLogs](repo ORMRepository[T]) Service[T] {
	return &AuditLogsService[T]{repo: repo}
}

func (s *AuditLogsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *AuditLogsService[T]) GetByID(id string) (*AuditLogs, error) {
	return s.repo.GetByID(id)
}

func (s *AuditLogsService[T]) Create(auditLog *AuditLogs) error {
	return s.repo.Create(auditLog)
}

func (s *AuditLogsService[T]) Update(auditLog *AuditLogs) error {
	return s.repo.Update(auditLog)
}

func (s *AuditLogsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
