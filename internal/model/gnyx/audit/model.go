package audit

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

// AuditAction mirrors CREATE TYPE audit_action.
type AuditAction string

const (
	AuditActionInsert AuditAction = "INSERT"
	AuditActionUpdate AuditAction = "UPDATE"
	AuditActionDelete AuditAction = "DELETE"
)

// AuditLog representa a tabela audit_logs.
type AuditLog struct {
	ID            gnyx.UUID      `json:"id" db:"id"`
	TableName     string         `json:"table_name" db:"table_name"`
	RecordID      gnyx.UUID      `json:"record_id" db:"record_id"`
	Action        AuditAction    `json:"action" db:"action"`
	UserID        *gnyx.UUID     `json:"user_id,omitempty" db:"user_id"`
	UserRole      *string        `json:"user_role,omitempty" db:"user_role"`
	OldValues     gnyx.JSONValue `json:"old_values,omitempty" db:"old_values"`
	NewValues     gnyx.JSONValue `json:"new_values,omitempty" db:"new_values"`
	ChangedFields []string       `json:"changed_fields,omitempty" db:"changed_fields"`
	IPAddress     *gnyx.Inet     `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string        `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt     gnyx.Timestamp `json:"created_at" db:"created_at"`
}
