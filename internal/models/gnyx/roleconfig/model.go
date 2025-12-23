package roleconfig

import (
	"time"
)

type RoleConfig struct {
	ID           *int64     `json:"id,omitempty" gorm:"column:id;primaryKey"`
	UserRole     string     `json:"user_role" gorm:"column:user_role"`
	DefaultRoute string     `json:"default_route" gorm:"column:default_route"`
	IsActive     *bool      `json:"is_active,omitempty" gorm:"column:is_active"`
	CreatedAt    *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

func (RoleConfig) TableName() string {
	return "gnyx.role_config"
}
