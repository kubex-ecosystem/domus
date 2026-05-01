package transactionstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/faelmori/domus-pulse/internal/execution"
	t "github.com/faelmori/domus-pulse/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	txTable   = "limalar.transactions"
	stepTable = "limalar.transaction_steps"
	docTable  = "limalar.transaction_docs"

	txCols   = "id, tenant_id, property_id, buyer_lead_id, status, commission_pct, closing_date, access_token, created_at, updated_at"
	stepCols = "id, transaction_id, step_code, title, status, due_date, completed_at, assignee_id, created_at"
	docCols  = "id, transaction_id, file_name, file_url, doc_type, extracted_data, created_at"
)

// ── pgTransactionStore ────────────────────────────────────────────────────────

type pgTransactionStore struct {
	exec execution.PGExecutor
}

func NewPGTransactionStore(exec execution.PGExecutor) TransactionStore {
	return &pgTransactionStore{exec: exec}
}

func (s *pgTransactionStore) Create(ctx context.Context, input *CreateTransactionInput) (*Transaction, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	status := input.Status
	if status == "" {
		status = "negotiation"
	}
	var tx Transaction
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (tenant_id, property_id, buyer_lead_id, status, commission_pct, closing_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING %s
	`, txTable, txCols),
		input.TenantID, input.PropertyID, input.BuyerLeadID, status, input.CommissionPct, input.ClosingDate,
	).Scan(&tx.ID, &tx.TenantID, &tx.PropertyID, &tx.BuyerLeadID,
		&tx.Status, &tx.CommissionPct, &tx.ClosingDate, &tx.AccessToken, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	return &tx, nil
}

func (s *pgTransactionStore) GetByID(ctx context.Context, tenantID, id string) (*Transaction, error) {
	var tx Transaction
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s FROM %s WHERE id = $1 AND tenant_id = $2
	`, txCols, txTable), id, tenantID).Scan(
		&tx.ID, &tx.TenantID, &tx.PropertyID, &tx.BuyerLeadID,
		&tx.Status, &tx.CommissionPct, &tx.ClosingDate, &tx.CreatedAt, &tx.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("transaction not found: %v", t.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return &tx, nil
}

func (s *pgTransactionStore) List(ctx context.Context, filters *TransactionFilters) (*t.PaginatedResult[Transaction], error) {
	if filters == nil {
		filters = &TransactionFilters{}
	}
	page := normalizePage(filters.Page)
	limit := normalizeLimit(filters.Limit, 20, 100)
	offset := (page - 1) * limit

	where, args := buildTxFilters(filters)
	q := fmt.Sprintf(`SELECT %s FROM %s %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		txCols, txTable, where, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.exec.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	txs := []Transaction{}
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.TenantID, &tx.PropertyID, &tx.BuyerLeadID,
			&tx.Status, &tx.CommissionPct, &tx.ClosingDate, &tx.AccessToken, &tx.CreatedAt, &tx.UpdatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	cWhere, cArgs := buildTxFilters(filters)
	var total int64
	_ = s.exec.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, txTable, cWhere), cArgs...).Scan(&total)

	tp := int(total) / limit
	if int(total)%limit != 0 {
		tp++
	}
	return &t.PaginatedResult[Transaction]{Data: txs, Total: total, Page: page, Limit: limit, TotalPages: tp}, nil
}

func (s *pgTransactionStore) GetByToken(ctx context.Context, token string) (*Transaction, error) {
	var tx Transaction
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s FROM %s WHERE access_token = $1
	`, txCols, txTable), token).Scan(
		&tx.ID, &tx.TenantID, &tx.PropertyID, &tx.BuyerLeadID,
		&tx.Status, &tx.CommissionPct, &tx.ClosingDate, &tx.AccessToken, &tx.CreatedAt, &tx.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("transaction not found: %v", t.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by token: %w", err)
	}
	return &tx, nil
}

func (s *pgTransactionStore) RegenerateToken(ctx context.Context, tenantID, id string) (string, error) {
	var newToken string
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		UPDATE %s SET access_token = gen_random_uuid(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING access_token
	`, txTable), id, tenantID).Scan(&newToken)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("transaction not found: %v", t.ErrNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("failed to regenerate token: %w", err)
	}
	return newToken, nil
}

func (s *pgTransactionStore) UpdateStatus(ctx context.Context, tenantID, id, status string) error {
	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3
	`, txTable), status, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found: %v", t.ErrNotFound)
	}
	return nil
}

func (s *pgTransactionStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[Transaction](), "pg_transaction_store", nil
}
func (s *pgTransactionStore) GetName() string { return "pg_transaction_store" }
func (s *pgTransactionStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgTransactionStore) Close() error { return nil }

// ── pgTransactionStepStore ────────────────────────────────────────────────────

type pgTransactionStepStore struct {
	exec execution.PGExecutor
}

func NewPGTransactionStepStore(exec execution.PGExecutor) TransactionStepStore {
	return &pgTransactionStepStore{exec: exec}
}

func (s *pgTransactionStepStore) ListByTransaction(ctx context.Context, tenantID, transactionID string) ([]TransactionStep, error) {
	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT ts.%s
		FROM %s ts
		JOIN %s tx ON tx.id = ts.transaction_id
		WHERE ts.transaction_id = $1 AND tx.tenant_id = $2
		ORDER BY ts.created_at
	`, strings.ReplaceAll(stepCols, ", ", ", ts."), stepTable, txTable), transactionID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transaction steps: %w", err)
	}
	defer rows.Close()
	var steps []TransactionStep
	for rows.Next() {
		var step TransactionStep
		if err := rows.Scan(&step.ID, &step.TransactionID, &step.StepCode, &step.Title,
			&step.Status, &step.DueDate, &step.CompletedAt, &step.AssigneeID, &step.CreatedAt); err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

func (s *pgTransactionStepStore) UpdateStep(ctx context.Context, tenantID, stepID string, input *UpdateStepInput) error {
	if input == nil {
		return fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	completedAt := input.CompletedAt
	if input.Status == "done" && completedAt == nil {
		now := time.Now()
		completedAt = &now
	}
	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s ts SET status = $1, due_date = $2, assignee_id = $3, completed_at = $4
		FROM %s tx
		WHERE ts.id = $5 AND ts.transaction_id = tx.id AND tx.tenant_id = $6
	`, stepTable, txTable), input.Status, input.DueDate, input.AssigneeID, completedAt, stepID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update transaction step: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("step not found: %v", t.ErrNotFound)
	}
	return nil
}

func (s *pgTransactionStepStore) SeedDefaultSteps(ctx context.Context, transactionID string) error {
	for _, step := range DefaultSteps {
		_, err := s.exec.Exec(ctx, fmt.Sprintf(`
			INSERT INTO %s (transaction_id, step_code, title, status)
			VALUES ($1, $2, $3, 'pending')
			ON CONFLICT DO NOTHING
		`, stepTable), transactionID, step.Code, step.Title)
		if err != nil {
			return fmt.Errorf("failed to seed step %s: %w", step.Code, err)
		}
	}
	return nil
}

func (s *pgTransactionStepStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[TransactionStep](), "pg_transaction_step_store", nil
}
func (s *pgTransactionStepStore) GetName() string { return "pg_transaction_step_store" }
func (s *pgTransactionStepStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgTransactionStepStore) Close() error { return nil }

// ── pgTransactionDocStore ─────────────────────────────────────────────────────

type pgTransactionDocStore struct {
	exec execution.PGExecutor
}

func NewPGTransactionDocStore(exec execution.PGExecutor) TransactionDocStore {
	return &pgTransactionDocStore{exec: exec}
}

func (s *pgTransactionDocStore) AddDoc(ctx context.Context, tenantID string, input *AddDocInput) (*TransactionDoc, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required: %v", t.ErrInvalidInput)
	}
	docType := input.DocType
	if docType == "" {
		docType = "other"
	}
	extracted := input.ExtractedData
	if len(extracted) == 0 {
		extracted = json.RawMessage(`{}`)
	}
	var doc TransactionDoc
	err := s.exec.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (transaction_id, file_name, file_url, doc_type, extracted_data)
		SELECT $1, $2, $3, $4, $5
		FROM %s tx WHERE tx.id = $1 AND tx.tenant_id = $6
		RETURNING %s
	`, docTable, txTable, docCols),
		input.TransactionID, input.FileName, input.FileURL, docType, extracted, tenantID,
	).Scan(&doc.ID, &doc.TransactionID, &doc.FileName, &doc.FileURL, &doc.DocType, &doc.ExtractedData, &doc.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("transaction not found or access denied: %v", t.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to add document: %w", err)
	}
	return &doc, nil
}

func (s *pgTransactionDocStore) ListByTransaction(ctx context.Context, tenantID, transactionID string) ([]TransactionDoc, error) {
	rows, err := s.exec.Query(ctx, fmt.Sprintf(`
		SELECT d.%s FROM %s d
		JOIN %s tx ON tx.id = d.transaction_id
		WHERE d.transaction_id = $1 AND tx.tenant_id = $2
		ORDER BY d.created_at
	`, strings.ReplaceAll(docCols, ", ", ", d."), docTable, txTable), transactionID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transaction docs: %w", err)
	}
	defer rows.Close()
	var docs []TransactionDoc
	for rows.Next() {
		var doc TransactionDoc
		if err := rows.Scan(&doc.ID, &doc.TransactionID, &doc.FileName, &doc.FileURL,
			&doc.DocType, &doc.ExtractedData, &doc.CreatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, rows.Err()
}

func (s *pgTransactionDocStore) UpdateExtractedData(ctx context.Context, tenantID, docID string, data json.RawMessage) error {
	tag, err := s.exec.Exec(ctx, fmt.Sprintf(`
		UPDATE %s d SET extracted_data = $1
		FROM %s tx WHERE d.id = $2 AND d.transaction_id = tx.id AND tx.tenant_id = $3
	`, docTable, txTable), data, docID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update extracted data: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("document not found: %v", t.ErrNotFound)
	}
	return nil
}

func (s *pgTransactionDocStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[TransactionDoc](), "pg_transaction_doc_store", nil
}
func (s *pgTransactionDocStore) GetName() string { return "pg_transaction_doc_store" }
func (s *pgTransactionDocStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgTransactionDocStore) Close() error { return nil }

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildTxFilters(f *TransactionFilters) (string, []any) {
	var where []string
	var args []any
	if f.TenantID != "" {
		args = append(args, f.TenantID)
		where = append(where, fmt.Sprintf("tenant_id = $%d", len(args)))
	}
	if f.PropertyID != nil {
		args = append(args, *f.PropertyID)
		where = append(where, fmt.Sprintf("property_id = $%d", len(args)))
	}
	if f.Status != nil {
		args = append(args, *f.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func normalizePage(p int) int {
	if p < 1 {
		return 1
	}
	return p
}

func normalizeLimit(l, def, max int) int {
	if l <= 0 {
		return def
	}
	if l > max {
		return max
	}
	return l
}

// suppress unused import
var _ = time.Now
