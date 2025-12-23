package user

import (
	"errors"
	"fmt"
)

type IUserService interface {
	CreateUser(user IUser) (IUser, error)
	GetUserByID(id string) (IUser, error)
	UpdateUser(user IUser) (IUser, error)
	DeleteUser(id string) error
	ListUsers() ([]IUser, error)
	GetUserByEmail(email string) (IUser, error)
}

type UserService struct {
	repo IUserRepo
}

func NewUserService(repo IUserRepo) IUserService {
	return &UserService{repo: repo}
}

func (us *UserService) CreateUser(user IUser) (IUser, error) {
	if user.GetEmail() == "" || user.GetName() == "" {
		return nil, errors.New("missing required fields")
	}
	createdUser, err := us.repo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}
	return createdUser, nil
}

func (us *UserService) GetUserByID(id string) (IUser, error) {
	user, err := us.repo.FindOne(id)
	if err != nil {
		return nil, fmt.Errorf("error fetching user: %v", err)
	}
	return user, nil
}

func (us *UserService) UpdateUser(user IUser) (IUser, error) {
	updatedUser, err := us.repo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %v", err)
	}
	return updatedUser, nil
}

func (us *UserService) DeleteUser(id string) error {
	err := us.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}
	return nil

}

func (us *UserService) ListUsers() ([]IUser, error) {
	users, err := us.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("error listing users: %v", err)
	}
	return users, nil
}

func (us *UserService) GetUserByEmail(email string) (IUser, error) {
	// This implementation is not efficient and is for demonstration purposes.
	// A proper implementation would have a FindByEmail method in the repo.
	users, err := us.repo.FindAll()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.GetEmail() == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}
