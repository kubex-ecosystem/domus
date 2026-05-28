// package userstore fornece o store de domínio para usuários.
package userstore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// User representa a entidade de usuário conforme schema do banco.
type User struct {
	ID                 string     `json:"id"`
	Email              string     `json:"email"`
	Name               *string    `json:"name,omitempty"`
	LastName           *string    `json:"last_name,omitempty"`
	PasswordHash       *string    `json:"password_hash,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	AvatarURL          *string    `json:"avatar_url,omitempty"`
	Status             *string    `json:"status,omitempty"`
	ForcePasswordReset bool       `json:"force_password_reset"`
	LastLogin          *time.Time `json:"last_login,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty"`
}

// CreateUserInput define dados para criação de usuário.
type CreateUserInput struct {
	Email              string
	Name               *string
	LastName           *string
	PasswordHash       *string
	Phone              *string
	AvatarURL          *string
	Status             *string
	ForcePasswordReset bool
}

// UpdateUserInput define campos atualizáveis.
type UpdateUserInput struct {
	ID                 string
	Name               *string
	Email              *string
	LastName           *string
	PasswordHash       *string
	Phone              *string
	AvatarURL          *string
	Status             *string
	ForcePasswordReset *bool
}

// UserFilters define filtros para listagem de usuários.
type UserFilters struct {
	Email  *string
	Status *string
	Page   int
	Limit  int
}

// UserStore define operações específicas de persistência de usuários.
type UserStore interface {
	t.StoreType

	// Create cria um novo usuário e retorna o ID gerado.
	Create(ctx context.Context, input *CreateUserInput) (*User, error)

	// GetByID busca usuário por ID.
	// Retorna (nil, nil) se não encontrado.
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByEmail busca usuário por email (case-insensitive).
	// Retorna (nil, nil) se não encontrado.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update atualiza dados do usuário.
	// Apenas campos não-nil em UpdateUserInput são atualizados.
	Update(ctx context.Context, input *UpdateUserInput) (*User, error)

	// UpdatePassword atualiza apenas o hash de senha.
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error

	// UpdateLastLogin registra timestamp do último login.
	UpdateLastLogin(ctx context.Context, userID string) error

	// Delete remove o usuário (hard delete).
	Delete(ctx context.Context, id string) error

	// List retorna usuários paginados com filtros opcionais.
	List(ctx context.Context, filters *UserFilters) (*t.PaginatedResult[User], error)

	// Count retorna o total de usuários (com filtros opcionais).
	Count(ctx context.Context, filters *UserFilters) (int64, error)
}
