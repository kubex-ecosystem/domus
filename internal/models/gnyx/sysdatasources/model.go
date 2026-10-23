package sysdatasources

import (
	"time"
)

type SysDataSources struct {
	ID       *string `json:"id,omitempty" gorm:"column:id;primaryKey"`
	TenantID *string `json:"tenant_id,omitempty" gorm:"column:tenant_id"`

	Name        string  `json:"name" gorm:"column:name"`
	Description *string `json:"description,omitempty" gorm:"column:description"`

	SQLQuery       string `json:"sql_query" gorm:"column:sql_query"`
	RequiredParams string `json:"required_params,omitempty" gorm:"column:required_params;type:jsonb"`
	CacheTTL       *int   `json:"cache_ttl,omitempty" gorm:"column:cache_ttl"`

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
	Active    bool      `json:"active" gorm:"column:active"`
}

func (SysDataSources) TableName() string {
	return "gnyx.sys_data_sources"
}
