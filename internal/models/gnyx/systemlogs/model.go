package systemlogs

import (
	"time"
)

type SystemLogs struct {
	ID        string     `json:"id" gorm:"column:id;primaryKey"`
	Level     string     `json:"level" gorm:"column:level"`
	Message   string     `json:"message" gorm:"column:message"`
	Category  string     `json:"category" gorm:"column:category"`
	CreatedAt *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
}

func (SystemLogs) TableName() string {
	return "gnyx.system_logs"
}
