package userpreferences

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*UserPreferences, error)
	Create(userPreference *UserPreferences) error
	Update(userPreference *UserPreferences) error
	Delete(id string) error
}

type UserPreferencesService[T UserPreferences] struct {
	repo ORMRepository[T]
}

func NewService[T UserPreferences](repo ORMRepository[T]) Service[T] {
	return &UserPreferencesService[T]{repo: repo}
}

func (s *UserPreferencesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *UserPreferencesService[T]) GetByID(id string) (*UserPreferences, error) {
	return s.repo.GetByID(id)
}

func (s *UserPreferencesService[T]) Create(userPreference *UserPreferences) error {
	return s.repo.Create(userPreference)
}

func (s *UserPreferencesService[T]) Update(userPreference *UserPreferences) error {
	return s.repo.Update(userPreference)
}

func (s *UserPreferencesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
