package userinvitations

import (
	"time"
)

type UserInvitations struct {
	ID         string     `json:"id" gorm:"column:id;primaryKey"`
	Email      string     `json:"email" gorm:"column:email"`
	Role       string     `json:"role" gorm:"column:role"`
	Token      string     `json:"token" gorm:"column:token"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" gorm:"column:expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty" gorm:"column:accepted_at"`
	CreatedAt  *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
}

func (UserInvitations) TableName() string {
	return "gnyx.user_invitations"
}
