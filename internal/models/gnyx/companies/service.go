package companies

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Companies, error)
	Create(company *Companies) error
	Update(company *Companies) error
	Delete(id string) error
}

type CompaniesService[T Companies] struct {
	repo ORMRepository[T]
}

func NewService[T Companies](repo ORMRepository[T]) Service[T] {
	return &CompaniesService[T]{repo: repo}
}

func (s *CompaniesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *CompaniesService[T]) GetByID(id string) (*Companies, error) {
	return s.repo.GetByID(id)
}

func (s *CompaniesService[T]) Create(company *Companies) error {
	return s.repo.Create(company)
}

func (s *CompaniesService[T]) Update(company *Companies) error {
	return s.repo.Update(company)
}

func (s *CompaniesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
