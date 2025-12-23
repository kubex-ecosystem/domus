// Package pipeline contém as definições de modelos relacionados a pipelines.
package pipeline

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

// Pipeline espelha a tabela pipelines.
type Pipeline struct {
	ID          gnyx.UUID       `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Description *string         `json:"description,omitempty" db:"description"`
	CompanyID   *gnyx.UUID      `json:"company_id,omitempty" db:"company_id"`
	IsActive    *bool           `json:"is_active,omitempty" db:"is_active"`
	CreatedAt   *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt   *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
}

// PipelineStage representa pipeline_stages.
type PipelineStage struct {
	ID         gnyx.UUID       `json:"id" db:"id"`
	PipelineID *gnyx.UUID      `json:"pipeline_id,omitempty" db:"pipeline_id"`
	Name       string          `json:"name" db:"name"`
	StageOrder int64           `json:"stage_order" db:"stage_order"`
	Color      *string         `json:"color,omitempty" db:"color"`
	IsActive   *bool           `json:"is_active,omitempty" db:"is_active"`
	CreatedAt  *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
}
