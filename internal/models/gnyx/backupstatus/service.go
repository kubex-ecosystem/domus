package backupstatus

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*BackupStatus, error)
	Create(backupStatus *BackupStatus) error
	Update(backupStatus *BackupStatus) error
	Delete(id string) error
}

type BackupStatusService[T BackupStatus] struct {
	repo ORMRepository[T]
}

func NewService[T BackupStatus](repo ORMRepository[T]) Service[T] {
	return &BackupStatusService[T]{repo: repo}
}

func (s *BackupStatusService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *BackupStatusService[T]) GetByID(id string) (*BackupStatus, error) {
	return s.repo.GetByID(id)
}

func (s *BackupStatusService[T]) Create(backupStatus *BackupStatus) error {
	return s.repo.Create(backupStatus)
}

func (s *BackupStatusService[T]) Update(backupStatus *BackupStatus) error {
	return s.repo.Update(backupStatus)
}

func (s *BackupStatusService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
