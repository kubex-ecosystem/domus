package errorlogs

import (
	"time"
)

type ErrorLogs struct {
	ID           *string   `json:"id,omitempty" gorm:"column:id;primaryKey"`
	ErrorMessage string    `json:"error_message" gorm:"column:error_message"`
	StackTrace   *string   `json:"stack_trace,omitempty" gorm:"column:stack_trace"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
}

func (ErrorLogs) TableName() string {
	return "gnyx.error_logs"
}
