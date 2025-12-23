// Package userinvitations provides services for managing user invitations.
package userinvitations

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*UserInvitations, error)
	Create(userInvitation *UserInvitations) error
	Update(userInvitation *UserInvitations) error
	Delete(id string) error
}

type UserInvitationsService[T UserInvitations] struct {
	repo ORMRepository[T]
}

func NewService[T UserInvitations](repo ORMRepository[T]) Service[T] {
	return &UserInvitationsService[T]{repo: repo}
}

func (s *UserInvitationsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *UserInvitationsService[T]) GetByID(id string) (*UserInvitations, error) {
	return s.repo.GetByID(id)
}

func (s *UserInvitationsService[T]) Create(userInvitation *UserInvitations) error {
	return s.repo.Create(userInvitation)
}

func (s *UserInvitationsService[T]) Update(userInvitation *UserInvitations) error {
	return s.repo.Update(userInvitation)
}

func (s *UserInvitationsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
