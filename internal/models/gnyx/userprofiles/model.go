package userprofiles

import (
	"time"
)

type UserProfiles struct {
	ID                 *string   `json:"id,omitempty" gorm:"column:id;primaryKey"`
	UserID             string    `json:"user_id" gorm:"column:user_id"`
	AvatarURL          *string   `json:"avatar_url,omitempty" gorm:"column:avatar_url"`
	Company            *string   `json:"company,omitempty" gorm:"column:company"`
	CompanyID          *string   `json:"company_id,omitempty" gorm:"column:company_id"`
	ForcePasswordReset bool      `json:"force_password_reset" gorm:"column:force_password_reset"`
	Role               *string   `json:"role,omitempty" gorm:"column:role"`
	Status             *string   `json:"status,omitempty" gorm:"column:status"`
	Phone              *string   `json:"phone,omitempty" gorm:"column:phone"`
	LastName           *string   `json:"last_name,omitempty" gorm:"column:last_name"`
	CreatedAt          time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (UserProfiles) TableName() string {
	return "gnyx.user_profiles"
}
