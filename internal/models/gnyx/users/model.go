package users

import (
	"time"
)

type Users struct {
	// Identificadores

	ID       *string `json:"id,omitempty" gorm:"column:id;primaryKey"`
	TenantID *string `json:"tenant_id,omitempty" gorm:"column:tenant_id"`

	// Dados e referenciais textuais

	FirstName string  `json:"first_name" gorm:"column:first_name"`
	LastName  *string `json:"last_name,omitempty" gorm:"column:last_name"`
	Username  string  `json:"username" gorm:"column:username"`
	Email     string  `json:"email" gorm:"column:email"`
	Phone     *string `json:"phone,omitempty" gorm:"column:phone"`
	Document  *string `json:"document,omitempty" gorm:"column:document"`

	// Segurança e controle

	Hash               string  `json:"hash" gorm:"column:hash"`
	ForcePasswordReset bool    `json:"force_password_reset" gorm:"column:force_password_reset"`
	Role               *string `json:"role,omitempty" gorm:"column:role"`

	// Log e status

	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty" gorm:"column:last_login"`
	Active    bool       `json:"active" gorm:"column:active"`

	// Metadados adicionais

	Status *string `json:"status,omitempty" gorm:"column:status"`
	Avatar *string `json:"avatar,omitempty" gorm:"column:avatar"`
}

func (Users) TableName() string {
	return "gnyx.users"
}
