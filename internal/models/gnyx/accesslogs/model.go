// Package accesslogs defines the data model for access logs in the Kubex system.
package accesslogs

import (
	"time"
)

type AccessLogs struct {
	ID        *string   `json:"id,omitempty" gorm:"column:id;primaryKey"`
	UserID    *string   `json:"user_id,omitempty" gorm:"column:user_id"`
	Action    string    `json:"action" gorm:"column:action"`
	IPAddress *string   `json:"ip_address,omitempty" gorm:"column:ip_address"`
	UserAgent *string   `json:"user_agent,omitempty" gorm:"column:user_agent"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

func (AccessLogs) TableName() string {
	return "gnyx.access_logs"
}
