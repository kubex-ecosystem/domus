package permissions

import (
	"time"
)

type Permissions struct {
	ID          *string   `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name        string    `json:"name" gorm:"column:name"`
	Description *string   `json:"description,omitempty" gorm:"column:description"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Permissions) TableName() string {
	return "gnyx.permissions"
}
