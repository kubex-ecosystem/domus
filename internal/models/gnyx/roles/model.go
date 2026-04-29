package roles

import (
	"time"
)

type Roles struct {
	ID           *string   `json:"id,omitempty" gorm:"column:id;primaryKey"`
	ParentRoleID *string   `json:"parent_role_id,omitempty" gorm:"column:parent_role_id"`
	Name         string    `json:"name" gorm:"column:name"`
	Description  *string   `json:"description,omitempty" gorm:"column:description"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Roles) TableName() string {
	return "gnyx.roles"
}
