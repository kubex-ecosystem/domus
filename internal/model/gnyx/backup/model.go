// Package backup defines data models related to backup operations.
package backup

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

// BackupType enumerates supported backup modes.
type BackupType string

const (
	BackupTypeFull         BackupType = "full"
	BackupTypeIncremental  BackupType = "incremental"
	BackupTypeDifferential BackupType = "differential"
)

// JobStatus expresses job execution state.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// BackupStatus representa a tabela backup_status.
type BackupStatus struct {
	ID           gnyx.UUID       `json:"id" db:"id"`
	BackupType   BackupType      `json:"backup_type" db:"backup_type"`
	Status       JobStatus       `json:"status" db:"status"`
	FileSize     *int64          `json:"file_size,omitempty" db:"file_size"`
	FilePath     *string         `json:"file_path,omitempty" db:"file_path"`
	ErrorMessage *string         `json:"error_message,omitempty" db:"error_message"`
	StartedAt    *gnyx.Timestamp `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *gnyx.Timestamp `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt    *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
}
