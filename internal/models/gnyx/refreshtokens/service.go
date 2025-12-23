package refreshtokens

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id int64) (*RefreshTokens, error)
	Create(refreshToken *RefreshTokens) error
	Update(refreshToken *RefreshTokens) error
	Delete(id int64) error
}

type RefreshTokensService[T RefreshTokens] struct {
	repo ORMRepository[T]
}

func NewService[T RefreshTokens](repo ORMRepository[T]) Service[T] {
	return &RefreshTokensService[T]{repo: repo}
}

func (s *RefreshTokensService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *RefreshTokensService[T]) GetByID(id int64) (*RefreshTokens, error) {
	return s.repo.GetByID(id)
}

func (s *RefreshTokensService[T]) Create(refreshToken *RefreshTokens) error {
	return s.repo.Create(refreshToken)
}

func (s *RefreshTokensService[T]) Update(refreshToken *RefreshTokens) error {
	return s.repo.Update(refreshToken)
}

func (s *RefreshTokensService[T]) Delete(id int64) error {
	return s.repo.Delete(id)
}
