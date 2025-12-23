package refreshtokens

import (
	"time"
)

type RefreshTokens struct {
	ID        *int64    `json:"id,omitempty" gorm:"column:id;primaryKey"`
	UserID    string    `json:"user_id" gorm:"column:user_id"`
	TokenID   string    `json:"token_id" gorm:"column:token_id"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (RefreshTokens) TableName() string {
	return "gnyx.refresh_tokens"
}
