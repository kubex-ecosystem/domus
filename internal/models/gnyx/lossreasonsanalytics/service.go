package lossreasonsanalytics

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*LossReasonsAnalytics, error)
	Create(lossReasonsAnalytic *LossReasonsAnalytics) error
	Update(lossReasonsAnalytic *LossReasonsAnalytics) error
	Delete(id string) error
}

type LossReasonsAnalyticsService[T LossReasonsAnalytics] struct {
	repo ORMRepository[T]
}

func NewService[T LossReasonsAnalytics](repo ORMRepository[T]) Service[T] {
	return &LossReasonsAnalyticsService[T]{repo: repo}
}

func (s *LossReasonsAnalyticsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *LossReasonsAnalyticsService[T]) GetByID(id string) (*LossReasonsAnalytics, error) {
	return s.repo.GetByID(id)
}

func (s *LossReasonsAnalyticsService[T]) Create(lossReasonsAnalytic *LossReasonsAnalytics) error {
	return s.repo.Create(lossReasonsAnalytic)
}

func (s *LossReasonsAnalyticsService[T]) Update(lossReasonsAnalytic *LossReasonsAnalytics) error {
	return s.repo.Update(lossReasonsAnalytic)
}

func (s *LossReasonsAnalyticsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
