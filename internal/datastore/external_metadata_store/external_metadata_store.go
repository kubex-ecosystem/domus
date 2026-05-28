package externalmetadatastore

import (
	"context"
	"encoding/json"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type ExternalMetadataRecord struct {
	ID           string          `json:"id"`
	SourceSystem string          `json:"source_system"`
	Domain       string          `json:"domain"`
	SchemaName   string          `json:"schema_name"`
	DatasetName  string          `json:"dataset_name"`
	TableName    string          `json:"table_name"`
	Manifest     json.RawMessage `json:"manifest,omitempty"`
	RowCount     *int64          `json:"row_count,omitempty"`
	LastLoadedAt *time.Time      `json:"last_loaded_at,omitempty"`
	LoadMode     string          `json:"load_mode"`
	Status       string          `json:"status"`
	Notes        *string         `json:"notes,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    *time.Time      `json:"updated_at,omitempty"`
}

type UpsertExternalMetadataInput struct {
	SourceSystem string
	Domain       string
	SchemaName   string
	DatasetName  string
	TableName    string
	Manifest     json.RawMessage
	RowCount     *int64
	LastLoadedAt *time.Time
	LoadMode     *string
	Status       *string
	Notes        *string
}

type ExternalMetadataFilters struct {
	SourceSystem *string
	Domain       *string
	SchemaName   *string
	DatasetName  *string
	Status       *string
	Page         int
	Limit        int
}

type ExternalMetadataStore interface {
	t.StoreType

	Upsert(ctx context.Context, input *UpsertExternalMetadataInput) (*ExternalMetadataRecord, error)
	GetByDataset(ctx context.Context, sourceSystem string, domain string, schemaName string, datasetName string) (*ExternalMetadataRecord, error)
	List(ctx context.Context, filters *ExternalMetadataFilters) (*t.PaginatedResult[ExternalMetadataRecord], error)
	Count(ctx context.Context, filters *ExternalMetadataFilters) (int64, error)
}
