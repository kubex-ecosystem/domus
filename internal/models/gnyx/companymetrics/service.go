package companymetrics

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*CompanyMetrics, error)
	Create(companyMetric *CompanyMetrics) error
	Update(companyMetric *CompanyMetrics) error
	Delete(id string) error
}

type CompanyMetricsService[T CompanyMetrics] struct {
	repo ORMRepository[T]
}

func NewService[T CompanyMetrics](repo ORMRepository[T]) Service[T] {
	return &CompanyMetricsService[T]{repo: repo}
}

func (s *CompanyMetricsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *CompanyMetricsService[T]) GetByID(id string) (*CompanyMetrics, error) {
	return s.repo.GetByID(id)
}

func (s *CompanyMetricsService[T]) Create(companyMetric *CompanyMetrics) error {
	return s.repo.Create(companyMetric)
}

func (s *CompanyMetricsService[T]) Update(companyMetric *CompanyMetrics) error {
	return s.repo.Update(companyMetric)
}

func (s *CompanyMetricsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
