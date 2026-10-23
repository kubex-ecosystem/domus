package propertystore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

const tableName = "limalar.properties"

// pgPropertyStore implementa PropertyStore usando PGExecutor.
type pgPropertyStore struct {
	exec execution.PGExecutor
}

// NewPGPropertyStore cria uma instância de PropertyStore para Postgres.
func NewPGPropertyStore(exec execution.PGExecutor) PropertyStore {
	return &pgPropertyStore{exec: exec}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (s *pgPropertyStore) Create(ctx context.Context, input *CreatePropertyInput) (*Property, error) {
	if input == nil {
		return nil, fmt.Errorf("create input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.TenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %v", t.ErrInvalidInput)
	}

	images := input.Images
	if images == nil {
		images = []PropertyImage{}
	}
	imagesJSON, err := json.Marshal(images)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal images: %v", err)
	}

	now := time.Now().UTC()
	q := fmt.Sprintf(`
		INSERT INTO %s (
			tenant_id, type, transaction, neighborhood, address,
			price, area, bedrooms, suites, bathrooms, parking,
			description, highlights, contact_phone, images,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$16)
		RETURNING id, tenant_id, type, transaction, neighborhood, address,
		          price, area, bedrooms, suites, bathrooms, parking,
		          description, highlights, contact_phone, images,
		          created_at, updated_at
	`, tableName)

	row := s.exec.QueryRow(ctx, q,
		input.TenantID,
		input.Type,
		input.Transaction,
		input.Neighborhood,
		input.Address,
		input.Price,
		input.Area,
		input.Bedrooms,
		input.Suites,
		input.Bathrooms,
		input.Parking,
		input.Description,
		input.Highlights,
		input.ContactPhone,
		imagesJSON,
		now,
	)

	return scanProperty(row)
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func (s *pgPropertyStore) GetByID(ctx context.Context, tenantID, id string) (*Property, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(tenantID) == "" {
		return nil, nil
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, type, transaction, neighborhood, address,
		       price, area, bedrooms, suites, bathrooms, parking,
		       description, highlights, contact_phone, images,
		       created_at, updated_at
		FROM %s
		WHERE id = $1 AND tenant_id = $2
	`, tableName)

	row := s.exec.QueryRow(ctx, q, id, tenantID)
	prop, err := scanProperty(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get property by id: %v", err)
	}
	return prop, nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (s *pgPropertyStore) Update(ctx context.Context, input *UpdatePropertyInput) (*Property, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" {
		return nil, fmt.Errorf("update input with ID is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.TenantID) == "" {
		return nil, fmt.Errorf("tenant_id is required: %v", t.ErrInvalidInput)
	}

	setClauses := []string{}
	args := []any{}
	idx := 1

	addField := func(col string, val any) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, val)
		idx++
	}

	if input.Type != nil {
		addField("type", *input.Type)
	}
	if input.Transaction != nil {
		addField("transaction", *input.Transaction)
	}
	if input.Neighborhood != nil {
		addField("neighborhood", *input.Neighborhood)
	}
	if input.Address != nil {
		addField("address", *input.Address)
	}
	if input.Price != nil {
		addField("price", *input.Price)
	}
	if input.Area != nil {
		addField("area", *input.Area)
	}
	if input.Bedrooms != nil {
		addField("bedrooms", *input.Bedrooms)
	}
	if input.Suites != nil {
		addField("suites", *input.Suites)
	}
	if input.Bathrooms != nil {
		addField("bathrooms", *input.Bathrooms)
	}
	if input.Parking != nil {
		addField("parking", *input.Parking)
	}
	if input.Description != nil {
		addField("description", *input.Description)
	}
	if input.Highlights != nil {
		addField("highlights", *input.Highlights)
	}
	if input.ContactPhone != nil {
		addField("contact_phone", *input.ContactPhone)
	}

	if len(setClauses) == 0 {
		return s.GetByID(ctx, input.TenantID, input.ID)
	}

	// Sempre atualiza updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now().UTC())
	idx++

	// WHERE id + tenant_id
	args = append(args, input.ID, input.TenantID)

	q := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE id = $%d AND tenant_id = $%d
		RETURNING id, tenant_id, type, transaction, neighborhood, address,
		          price, area, bedrooms, suites, bathrooms, parking,
		          description, highlights, contact_phone, images,
		          created_at, updated_at
	`, tableName, strings.Join(setClauses, ", "), idx, idx+1)

	row := s.exec.QueryRow(ctx, q, args...)
	prop, err := scanProperty(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("property not found: %v", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to update property: %v", err)
	}
	return prop, nil
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (s *pgPropertyStore) Delete(ctx context.Context, tenantID, id string) error {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(tenantID) == "" {
		return fmt.Errorf("id and tenant_id are required: %v", t.ErrInvalidInput)
	}

	q := fmt.Sprintf(`DELETE FROM %s WHERE id = $1 AND tenant_id = $2`, tableName)
	tag, err := s.exec.Exec(ctx, q, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete property: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("property not found: %v", t.ErrNotFound)
	}
	return nil
}

// ── List ─────────────────────────────────────────────────────────────────────

func (s *pgPropertyStore) List(ctx context.Context, filters *PropertyFilters) (*t.PaginatedResult[Property], error) {
	if filters == nil {
		filters = &PropertyFilters{}
	}

	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := calcOffset(page, limit)

	where, args := buildPropertyFilters(filters)

	q := fmt.Sprintf(`
		SELECT id, tenant_id, type, transaction, neighborhood, address,
		       price, area, bedrooms, suites, bathrooms, parking,
		       description, highlights, contact_phone, images,
		       created_at, updated_at
		FROM %s
		%s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d
	`, tableName, where, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list properties: %v", err)
	}
	defer rows.Close()

	properties := []Property{}
	for rows.Next() {
		prop, err := scanProperty(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan property: %v", err)
		}
		properties = append(properties, *prop)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	// Contagem total com os mesmos filtros
	countWhere, countArgs := buildPropertyFilters(filters)
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, tableName, countWhere)
	var total int64
	if err := s.exec.QueryRow(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count properties: %v", err)
	}

	return &t.PaginatedResult[Property]{
		Data:       properties,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: calcTotalPages(total, limit),
	}, nil
}

// ── AddImages ─────────────────────────────────────────────────────────────────

func (s *pgPropertyStore) AddImages(ctx context.Context, tenantID, propertyID string, images []PropertyImage) ([]PropertyImage, error) {
	if strings.TrimSpace(propertyID) == "" || strings.TrimSpace(tenantID) == "" {
		return nil, fmt.Errorf("propertyID and tenantID are required: %v", t.ErrInvalidInput)
	}

	// Busca imagens existentes
	q := fmt.Sprintf(`SELECT images FROM %s WHERE id = $1 AND tenant_id = $2`, tableName)
	var existing []byte
	if err := s.exec.QueryRow(ctx, q, propertyID, tenantID).Scan(&existing); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("property not found: %v", t.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to fetch existing images: %v", err)
	}

	current := []PropertyImage{}
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &current); err != nil {
			return nil, fmt.Errorf("failed to unmarshal existing images: %v", err)
		}
	}

	// Adiciona IDs às novas imagens e anexa
	for i := range images {
		if images[i].ID == "" {
			images[i].ID = uuid.New().String()
		}
		current = append(current, images[i])
	}

	updated, err := json.Marshal(current)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated images: %v", err)
	}

	uq := fmt.Sprintf(`
		UPDATE %s
		SET images = $1, updated_at = $2
		WHERE id = $3 AND tenant_id = $4
	`, tableName)
	tag, err := s.exec.Exec(ctx, uq, updated, time.Now().UTC(), propertyID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to update images: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("property not found: %v", t.ErrNotFound)
	}

	return current, nil
}

// ── DeleteImage ───────────────────────────────────────────────────────────────

func (s *pgPropertyStore) DeleteImage(ctx context.Context, tenantID, propertyID, imageID string) error {
	if strings.TrimSpace(propertyID) == "" || strings.TrimSpace(imageID) == "" || strings.TrimSpace(tenantID) == "" {
		return fmt.Errorf("propertyID, tenantID and imageID are required: %v", t.ErrInvalidInput)
	}

	q := fmt.Sprintf(`SELECT images FROM %s WHERE id = $1 AND tenant_id = $2`, tableName)
	var raw []byte
	if err := s.exec.QueryRow(ctx, q, propertyID, tenantID).Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("property not found: %v", t.ErrNotFound)
		}
		return fmt.Errorf("failed to fetch images: %v", err)
	}

	current := []PropertyImage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &current); err != nil {
			return fmt.Errorf("failed to unmarshal images: %v", err)
		}
	}

	filtered := make([]PropertyImage, 0, len(current))
	found := false
	for _, img := range current {
		if img.ID == imageID {
			found = true
			continue
		}
		filtered = append(filtered, img)
	}
	if !found {
		return fmt.Errorf("image not found: %v", t.ErrNotFound)
	}

	updated, err := json.Marshal(filtered)
	if err != nil {
		return fmt.Errorf("failed to marshal updated images: %v", err)
	}

	uq := fmt.Sprintf(`
		UPDATE %s
		SET images = $1, updated_at = $2
		WHERE id = $3 AND tenant_id = $4
	`, tableName)
	_, err = s.exec.Exec(ctx, uq, updated, time.Now().UTC(), propertyID, tenantID)
	return err
}

// ── StoreType interface ───────────────────────────────────────────────────────

func (s *pgPropertyStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Property](), "pg_property_store", nil
}

func (s *pgPropertyStore) GetName() string {
	return "pg_property_store"
}

func (s *pgPropertyStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}

func (s *pgPropertyStore) Close() error {
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func scanProperty(row pgx.Row) (*Property, error) {
	var p Property
	var imagesJSON []byte
	err := row.Scan(
		&p.ID,
		&p.TenantID,
		&p.Type,
		&p.Transaction,
		&p.Neighborhood,
		&p.Address,
		&p.Price,
		&p.Area,
		&p.Bedrooms,
		&p.Suites,
		&p.Bathrooms,
		&p.Parking,
		&p.Description,
		&p.Highlights,
		&p.ContactPhone,
		&imagesJSON,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	p.Images = []PropertyImage{}
	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, fmt.Errorf("failed to unmarshal images: %v", err)
		}
	}
	return &p, nil
}

func buildPropertyFilters(filters *PropertyFilters) (string, []any) {
	where := []string{}
	args := []any{}

	if filters.TenantID != "" {
		where = append(where, fmt.Sprintf("tenant_id = $%d", len(args)+1))
		args = append(args, filters.TenantID)
	}
	if filters.Transaction != nil && *filters.Transaction != "" {
		where = append(where, fmt.Sprintf("transaction = $%d", len(args)+1))
		args = append(args, *filters.Transaction)
	}
	if filters.Neighborhood != nil && *filters.Neighborhood != "" {
		where = append(where, fmt.Sprintf("neighborhood = $%d", len(args)+1))
		args = append(args, *filters.Neighborhood)
	}
	if filters.Type != nil && *filters.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", len(args)+1))
		args = append(args, *filters.Type)
	}

	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeLimit(limit, defaultLimit, maxLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func calcOffset(page, limit int) int {
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
	if pages == 0 && total > 0 {
		return 1
	}
	return pages
}
