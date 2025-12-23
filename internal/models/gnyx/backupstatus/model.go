package backupstatus

import (
	"time"
)

type BackupStatus struct {
	ID           string     `json:"id" gorm:"column:id;primaryKey"`
	BackupType   string     `json:"backup_type" gorm:"column:backup_type"`
	Status       string     `json:"status" gorm:"column:status"`
	FilePath     *string    `json:"file_path,omitempty" gorm:"column:file_path"`
	ErrorMessage *string    `json:"error_message,omitempty" gorm:"column:error_message"`
	StartedAt    *time.Time `json:"started_at,omitempty" gorm:"column:started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" gorm:"column:completed_at"`
	CreatedAt    *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
}

func (BackupStatus) TableName() string {
	return "gnyx.backup_status"
}
