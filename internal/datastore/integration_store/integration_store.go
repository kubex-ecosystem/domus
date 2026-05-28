// Package integrationstore provides a PostgreSQL implementation of the IntegrationStore interface with dynamic encryption of sensitive fields in JSONB.
package integrationstore

import (
	"context"
	"encoding/json"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

// IntegrationType define o tipo de integração (Sankhya, FastChannel, IMAP, etc)
type IntegrationType string

const (
	TypeMSSQLErp IntegrationType = "MSSQL_ERP"
	TypeRestAPI  IntegrationType = "REST_API"
	TypeIMAP     IntegrationType = "IMAP"
)

// IntegrationConfig é o cofre da integração (IPs, portas, credenciais)
type IntegrationConfig struct {
	ID        string          `json:"id"`
	TenantID  string          `json:"tenant_id"`
	PartnerID *string         `json:"partner_id,omitempty"`
	Type      IntegrationType `json:"type"`
	Name      string          `json:"name"`
	Settings  json.RawMessage `json:"settings"` // O JSONB brilha aqui!
	IsActive  bool            `json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
}

// SyncJob é a agenda do motor GNyx
type SyncJob struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	ConfigID       string     `json:"config_id"`
	TaskName       string     `json:"task_name"`
	CronExpression *string    `json:"cron_expression,omitempty"`
	IsActive       bool       `json:"is_active"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

// IntegrationStore define o contrato para o motor de integração do Kubex
type IntegrationStore interface {
	t.StoreType

	// Configs
	CreateConfig(ctx context.Context, input *IntegrationConfig) (*IntegrationConfig, error)
	GetConfigByID(ctx context.Context, id string) (*IntegrationConfig, error)
	ListConfigsByTenant(ctx context.Context, tenantID string) ([]IntegrationConfig, error)

	// Jobs
	CreateJob(ctx context.Context, input *SyncJob) (*SyncJob, error)
	ListActiveJobs(ctx context.Context, tenantID string) ([]SyncJob, error)
	UpdateJobLastSync(ctx context.Context, jobID string, syncTime time.Time) error
}
