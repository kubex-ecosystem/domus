package sessionstore

import (
	"context"
	"time"

	"github.com/google/uuid"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

// Session representa uma sessão de refresh token persistida no banco.
type Session struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	RefreshTokenHash string     `json:"-"` // hash do refresh token
	UserAgent        string     `json:"user_agent,omitempty"`
	IP               string     `json:"ip,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// SessionStore define as operações para gerenciar sessões no banco de dados.
type SessionStore interface {
	t.StoreType

	Create(ctx context.Context, s *Session) error
	FindByRefreshHash(ctx context.Context, hash string) (*Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeByRefreshHash(ctx context.Context, hash string) error
	RevokeByUser(ctx context.Context, userID uuid.UUID) error
}
