package sysdashboards

import (
	"time"
)

type SysDashboards struct {
	ID          *string `json:"id,omitempty" gorm:"column:id;primaryKey"`
	TenantID    *string `json:"tenant_id,omitempty" gorm:"column:tenant_id"`

	Title       string  `json:"title" gorm:"column:title"`
	Description *string `json:"description,omitempty" gorm:"column:description"`

	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
	Active      bool      `json:"active" gorm:"column:active"`
}

func (SysDashboards) TableName() string {
	return "gnyx.sys_dashboards"
}
