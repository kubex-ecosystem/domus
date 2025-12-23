package systemmetrics

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*SystemMetrics, error)
	Create(systemMetric *SystemMetrics) error
	Update(systemMetric *SystemMetrics) error
	Delete(id string) error
}

type SystemMetricsService[T SystemMetrics] struct {
	repo ORMRepository[T]
}

func NewService[T SystemMetrics](repo ORMRepository[T]) Service[T] {
	return &SystemMetricsService[T]{repo: repo}
}

func (s *SystemMetricsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *SystemMetricsService[T]) GetByID(id string) (*SystemMetrics, error) {
	return s.repo.GetByID(id)
}

func (s *SystemMetricsService[T]) Create(systemMetric *SystemMetrics) error {
	return s.repo.Create(systemMetric)
}

func (s *SystemMetricsService[T]) Update(systemMetric *SystemMetrics) error {
	return s.repo.Update(systemMetric)
}

func (s *SystemMetricsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
