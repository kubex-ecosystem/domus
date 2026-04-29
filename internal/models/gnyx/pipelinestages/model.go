package pipelinestages

import (
	"time"
)

type PipelineStages struct {
	ID        string     `json:"id" gorm:"column:id;primaryKey"`
	Name      string     `json:"name" gorm:"column:name"`
	Color     *string    `json:"color,omitempty" gorm:"column:color"`
	IsActive  *bool      `json:"is_active,omitempty" gorm:"column:is_active"`
	CreatedAt *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
}

func (PipelineStages) TableName() string {
	return "gnyx.pipeline_stages"
}
