package pipelines

import (
	"time"
)

type Pipelines struct {
	ID          string     `json:"id" gorm:"column:id;primaryKey"`
	Name        string     `json:"name" gorm:"column:name"`
	Description *string    `json:"description,omitempty" gorm:"column:description"`
	IsActive    *bool      `json:"is_active,omitempty" gorm:"column:is_active"`
	CreatedAt   *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

func (Pipelines) TableName() string {
	return "gnyx.pipelines"
}
