package externalmetadatastore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

type pgExternalMetadataStore struct {
	exec execution.PGExecutor
}

func NewPGExternalMetadataStore(exec execution.PGExecutor) ExternalMetadataStore {
	return &pgExternalMetadataStore{exec: exec}
}

func (s *pgExternalMetadataStore) Upsert(ctx context.Context, input *UpsertExternalMetadataInput) (*ExternalMetadataRecord, error) {
	if input == nil {
		return nil, fmt.Errorf("upsert input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.SourceSystem) == "" || strings.TrimSpace(input.Domain) == "" || strings.TrimSpace(input.SchemaName) == "" || strings.TrimSpace(input.DatasetName) == "" || strings.TrimSpace(input.TableName) == "" {
		return nil, fmt.Errorf("source_system, domain, schema_name, dataset_name and table_name are required: %v", t.ErrInvalidInput)
	}

	loadMode := "full_refresh"
	if input.LoadMode != nil && strings.TrimSpace(*input.LoadMode) != "" {
		loadMode = strings.TrimSpace(*input.LoadMode)
	}
	status := "ready"
	if input.Status != nil && strings.TrimSpace(*input.Status) != "" {
		status = strings.TrimSpace(*input.Status)
	}

	const q = `
        INSERT INTO external_metadata_registry (
            source_system, domain, schema_name, dataset_name, table_name,
            manifest, row_count, last_loaded_at, load_mode, status, notes, created_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
        ON CONFLICT (source_system, domain, schema_name, dataset_name)
        DO UPDATE SET
            table_name = EXCLUDED.table_name,
            manifest = EXCLUDED.manifest,
            row_count = EXCLUDED.row_count,
            last_loaded_at = EXCLUDED.last_loaded_at,
            load_mode = EXCLUDED.load_mode,
            status = EXCLUDED.status,
            notes = EXCLUDED.notes,
            updated_at = now()
        RETURNING id, source_system, domain, schema_name, dataset_name, table_name,
                  manifest, row_count, last_loaded_at, load_mode, status, notes,
                  created_at, updated_at
    `

	row := s.exec.QueryRow(ctx, q,
		strings.TrimSpace(input.SourceSystem),
		strings.TrimSpace(input.Domain),
		strings.TrimSpace(input.SchemaName),
		strings.TrimSpace(input.DatasetName),
		strings.TrimSpace(input.TableName),
		normalizeManifest(input.Manifest),
		input.RowCount,
		input.LastLoadedAt,
		loadMode,
		status,
		input.Notes,
		time.Now().UTC(),
	)

	record, err := scanExternalMetadataRecord(row)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert external metadata registry record: %v", err)
	}
	return record, nil
}

func (s *pgExternalMetadataStore) GetByDataset(ctx context.Context, sourceSystem string, domain string, schemaName string, datasetName string) (*ExternalMetadataRecord, error) {
	if strings.TrimSpace(sourceSystem) == "" || strings.TrimSpace(domain) == "" || strings.TrimSpace(schemaName) == "" || strings.TrimSpace(datasetName) == "" {
		return nil, nil
	}

	const q = `
        SELECT id, source_system, domain, schema_name, dataset_name, table_name,
               manifest, row_count, last_loaded_at, load_mode, status, notes,
               created_at, updated_at
        FROM external_metadata_registry
        WHERE source_system = $1 AND domain = $2 AND schema_name = $3 AND dataset_name = $4
    `

	row := s.exec.QueryRow(ctx, q,
		strings.TrimSpace(sourceSystem),
		strings.TrimSpace(domain),
		strings.TrimSpace(schemaName),
		strings.TrimSpace(datasetName),
	)

	record, err := scanExternalMetadataRecord(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get external metadata registry record: %v", err)
	}
	return record, nil
}

func (s *pgExternalMetadataStore) List(ctx context.Context, filters *ExternalMetadataFilters) (*t.PaginatedResult[ExternalMetadataRecord], error) {
	if filters == nil {
		filters = &ExternalMetadataFilters{}
	}

	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 200)
	offset := calcOffset(page, limit)

	where, args := buildFilters(filters)
	query := fmt.Sprintf(`
        SELECT id, source_system, domain, schema_name, dataset_name, table_name,
               manifest, row_count, last_loaded_at, load_mode, status, notes,
               created_at, updated_at
        FROM external_metadata_registry
        %s
        ORDER BY source_system, domain, schema_name, dataset_name
        LIMIT $%d OFFSET $%d
    `, where, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list external metadata registry records: %v", err)
	}
	defer rows.Close()

	records := []ExternalMetadataRecord{}
	for rows.Next() {
		record, err := scanExternalMetadataRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan external metadata registry record: %v", err)
		}
		records = append(records, *record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	total, err := s.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &t.PaginatedResult[ExternalMetadataRecord]{
		Data:       records,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: calcTotalPages(total, limit),
	}, nil
}

func (s *pgExternalMetadataStore) Count(ctx context.Context, filters *ExternalMetadataFilters) (int64, error) {
	if filters == nil {
		filters = &ExternalMetadataFilters{}
	}

	where, args := buildFilters(filters)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM external_metadata_registry %s`, where)

	var total int64
	if err := s.exec.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to count external metadata registry records: %v", err)
	}
	return total, nil
}

func (s *pgExternalMetadataStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[ExternalMetadataRecord](), "pg_external_metadata_store", nil
}

func (s *pgExternalMetadataStore) GetName() string {
	return "pg_external_metadata_store"
}

func (s *pgExternalMetadataStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgExternalMetadataStore) Close() error {
	return nil
}

func scanExternalMetadataRecord(row pgx.Row) (*ExternalMetadataRecord, error) {
	var record ExternalMetadataRecord
	var manifest []byte
	if err := row.Scan(
		&record.ID,
		&record.SourceSystem,
		&record.Domain,
		&record.SchemaName,
		&record.DatasetName,
		&record.TableName,
		&manifest,
		&record.RowCount,
		&record.LastLoadedAt,
		&record.LoadMode,
		&record.Status,
		&record.Notes,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return nil, err
	}
	record.Manifest = manifest
	return &record, nil
}

func buildFilters(filters *ExternalMetadataFilters) (string, []any) {
	where := []string{}
	args := []any{}

	add := func(column string, value string) {
		where = append(where, fmt.Sprintf("%s = $%d", column, len(args)+1))
		args = append(args, strings.TrimSpace(value))
	}

	if filters.SourceSystem != nil && strings.TrimSpace(*filters.SourceSystem) != "" {
		add("source_system", *filters.SourceSystem)
	}
	if filters.Domain != nil && strings.TrimSpace(*filters.Domain) != "" {
		add("domain", *filters.Domain)
	}
	if filters.SchemaName != nil && strings.TrimSpace(*filters.SchemaName) != "" {
		add("schema_name", *filters.SchemaName)
	}
	if filters.DatasetName != nil && strings.TrimSpace(*filters.DatasetName) != "" {
		add("dataset_name", *filters.DatasetName)
	}
	if filters.Status != nil && strings.TrimSpace(*filters.Status) != "" {
		add("status", *filters.Status)
	}

	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func normalizeManifest(raw []byte) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizeLimit(limit int, defaultLimit int, maxLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func calcOffset(page int, limit int) int {
	return (page - 1) * limit
}

func calcTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := int(total) / limit
	if int(total)%limit != 0 {
		pages++
	}
	if pages == 0 {
		pages = 1
	}
	return pages
}
