package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	m "github.com/kubex-ecosystem/domus/internal/models/users"
	"github.com/kubex-ecosystem/domus/internal/services/adaptive/backend"
)

type AdaptiveBackend = backend.Backend
type UserModelType = m.UserModel
type UserModel = m.IUser
type UserService = m.IUserService
type UserRepo = m.IUserRepo

func NewUserService(userRepo UserRepo) UserService {
	return m.NewUserService(userRepo)
}

func NewUserRepo(ctx context.Context, dbService backend.Backend) UserRepo {
	return m.NewUserRepo(dbService)
}

func NewUserModel(username, name, email string) UserModel {
	return &m.UserModel{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}
}
