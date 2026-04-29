package rolepermissions

import (
	"time"
)

type RolePermissions struct {
	ID           *string   `json:"id,omitempty" gorm:"column:id;primaryKey"`
	RoleID       string    `json:"role_id" gorm:"column:role_id"`
	PermissionID string    `json:"permission_id" gorm:"column:permission_id"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (RolePermissions) TableName() string {
	return "gnyx.role_permissions"
}
