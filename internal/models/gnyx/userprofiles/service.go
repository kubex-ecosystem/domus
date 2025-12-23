package userprofiles

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*UserProfiles, error)
	Create(userProfile *UserProfiles) error
	Update(userProfile *UserProfiles) error
	Delete(id string) error
}

type UserProfilesService[T UserProfiles] struct {
	repo ORMRepository[T]
}

func NewService[T UserProfiles](repo ORMRepository[T]) Service[T] {
	return &UserProfilesService[T]{repo: repo}
}

func (s *UserProfilesService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *UserProfilesService[T]) GetByID(id string) (*UserProfiles, error) {
	return s.repo.GetByID(id)
}

func (s *UserProfilesService[T]) Create(userProfile *UserProfiles) error {
	return s.repo.Create(userProfile)
}

func (s *UserProfilesService[T]) Update(userProfile *UserProfiles) error {
	return s.repo.Update(userProfile)
}

func (s *UserProfilesService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
