// Package user contains the user repository implementation.
package user

import (
	"context"

	"github.com/kubex-ecosystem/domus/internal/datastore"
	"github.com/kubex-ecosystem/domus/internal/services/adaptive/backend"
	"github.com/kubex-ecosystem/domus/internal/types"
)

// PaginatedResult is a type alias for types.PaginatedResult.
type PaginatedResult[T any] = types.PaginatedResult[T]

// IUserRepo defines the interface for user repository operations.
type IUserRepo = datastore.Repository[UserModel]

// UserRepo is the concrete implementation of IUserRepo.
type UserRepo struct {
	IUserRepo //Constraint to implement the interface
	backend   backend.Backend
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(backend backend.Backend) IUserRepo {
	return &UserRepo{backend: backend}
}

// Create creates a new user.
func (r *UserRepo) Create(ctx context.Context, user *UserModel) (string, error) {
	err := r.backend.Create(user.GetUserObj())
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

// FindOne finds a single user by ID.
func (r *UserRepo) FindOne(ctx context.Context, id string) (*UserModel, error) {
	var user UserModel
	err := r.backend.GetByID(id, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user.
func (r *UserRepo) Update(ctx context.Context, user *UserModel) error {
	err := r.backend.Update(user.GetUserObj())
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a user.
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	return r.backend.Delete(id)
}

// FindAll finds all users.
func (r *UserRepo) List(ctx context.Context, filters map[string]any) (*PaginatedResult[UserModel], error) {
	var storageResult types.PaginatedResult[UserModel]
	err := r.backend.GetAll(&storageResult)
	if err != nil {
		return nil, err
	}

	// Map backend.UserObj to UserModel
	var users []UserModel
	for _, u := range storageResult.Data {
		user := u // create a new variable to avoid referencing the loop variable
		users = append(users, user)
	}

	return &PaginatedResult[UserModel]{
		Data:       users,
		Total:      storageResult.Total,
		Page:       storageResult.Page,
		TotalPages: storageResult.TotalPages,
		Limit:      storageResult.Limit,
	}, nil
}
