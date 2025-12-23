package profiles

import (
	"time"
)

type Profiles struct {
	ID                 string     `json:"id" gorm:"column:id;primaryKey"`
	Email              string     `json:"email" gorm:"column:email"`
	Name               string     `json:"name" gorm:"column:name"`
	Role               string     `json:"role" gorm:"column:role"`
	CreatedAt          *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
	AvatarURL          *string    `json:"avatar_url,omitempty" gorm:"column:avatar_url"`
	Company            *string    `json:"company,omitempty" gorm:"column:company"`
	Phone              *string    `json:"phone,omitempty" gorm:"column:phone"`
	LastName           *string    `json:"last_name,omitempty" gorm:"column:last_name"`
	ForcePasswordReset *bool      `json:"force_password_reset,omitempty" gorm:"column:force_password_reset"`
}

func (Profiles) TableName() string {
	return "gnyx.profiles"
}
