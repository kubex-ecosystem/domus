package syswidgets

import (
	"time"
)

type SysWidgets struct {
	ID          *string `json:"id,omitempty" gorm:"column:id;primaryKey"`
	DashboardID *string `json:"dashboard_id,omitempty" gorm:"column:dashboard_id"`

	Type        string  `json:"type" gorm:"column:type"`
	Size        string  `json:"size" gorm:"column:size;type:jsonb"` // Store as JSONB string representation in DB e.g. {"cols":12,"rows":4}
	Config      string  `json:"config" gorm:"column:config;type:jsonb"` // Store as JSONB string representation in DB, e.g. {"title": "xyz", ...}

	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
	Active      bool      `json:"active" gorm:"column:active"`
}

func (SysWidgets) TableName() string {
	return "gnyx.sys_widgets"
}
