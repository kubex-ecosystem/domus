package auditlogs

import (
	"encoding/json"
	"time"
)

type AuditLogs struct {
	ID         *string         `json:"id,omitempty" gorm:"column:id;primaryKey"`
	UserID     *string         `json:"user_id,omitempty" gorm:"column:user_id"`
	Action     string          `json:"action" gorm:"column:action"`
	EntityType string          `json:"entity_type" gorm:"column:entity_type"`
	EntityID   string          `json:"entity_id" gorm:"column:entity_id"`
	Changes    json.RawMessage `json:"changes" gorm:"column:changes"`
	CreatedAt  time.Time       `json:"created_at" gorm:"column:created_at"`
}

func (AuditLogs) TableName() string {
	return "gnyx.audit_logs"
}
