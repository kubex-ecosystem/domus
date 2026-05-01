// Package transactionstore fornece os stores de transações imobiliárias (Colonial NetImóveis).
// Três stores coexistem no mesmo pacote por serem domínio estreitamente acoplado:
//   - TransactionStore     : CRUD de negociações + status lifecycle
//   - TransactionStepStore : checklist de etapas do fechamento
//   - TransactionDocStore  : documentos do processo (upload + extração LLM)
package transactionstore

import (
	"context"
	"encoding/json"
	"time"

	t "github.com/faelmori/domus-pulse/internal/types"
)

// ── Transaction ───────────────────────────────────────────────────────────────

// Transaction representa uma negociação imobiliária em andamento.
type Transaction struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	PropertyID    *string    `json:"property_id,omitempty"`
	BuyerLeadID   *string    `json:"buyer_lead_id,omitempty"`
	Status        string     `json:"status"`         // negotiation|contract_signed|due_diligence|financing|closing|completed|cancelled
	CommissionPct float64    `json:"commission_pct"` // ex: 6.00
	ClosingDate   *time.Time `json:"closing_date,omitempty"`
	AccessToken   *string    `json:"access_token,omitempty"` // token para portal público do cliente
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateTransactionInput define dados para criar uma transação.
type CreateTransactionInput struct {
	TenantID      string
	PropertyID    *string
	BuyerLeadID   *string
	Status        string
	CommissionPct float64
	ClosingDate   *time.Time
}

// TransactionFilters define filtros para listagem de transações.
type TransactionFilters struct {
	TenantID   string
	PropertyID *string
	Status     *string
	Page       int
	Limit      int
}

// TransactionStore define operações de persistência de transações.
type TransactionStore interface {
	t.StoreType

	// Create registra uma nova transação e cria as etapas padrão.
	Create(ctx context.Context, input *CreateTransactionInput) (*Transaction, error)

	// GetByID retorna uma transação pelo ID, dentro do tenant.
	GetByID(ctx context.Context, tenantID, id string) (*Transaction, error)

	// List retorna transações paginadas com filtros opcionais.
	List(ctx context.Context, filters *TransactionFilters) (*t.PaginatedResult[Transaction], error)

	// UpdateStatus atualiza o status de uma transação.
	UpdateStatus(ctx context.Context, tenantID, id, status string) error

	// GetByToken retorna uma transação pelo access_token público (sem tenant check).
	// Usado pelo portal do cliente — não expõe dados sensíveis do tenant.
	GetByToken(ctx context.Context, token string) (*Transaction, error)

	// RegenerateToken gera um novo access_token para a transação e retorna o novo valor.
	RegenerateToken(ctx context.Context, tenantID, id string) (string, error)
}

// ── TransactionStep ───────────────────────────────────────────────────────────

// TransactionStep representa uma etapa do checklist de fechamento.
type TransactionStep struct {
	ID            string     `json:"id"`
	TransactionID string     `json:"transaction_id"`
	StepCode      string     `json:"step_code"`
	Title         string     `json:"title"`
	Status        string     `json:"status"` // pending|in_progress|done|skipped
	DueDate       *time.Time `json:"due_date,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	AssigneeID    *string    `json:"assignee_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// UpdateStepInput define dados para atualizar o status de uma etapa.
type UpdateStepInput struct {
	Status      string
	DueDate     *time.Time
	AssigneeID  *string
	CompletedAt *time.Time
}

// TransactionStepStore define operações de persistência de etapas.
type TransactionStepStore interface {
	t.StoreType

	// ListByTransaction retorna as etapas de uma transação ordenadas por created_at.
	ListByTransaction(ctx context.Context, tenantID, transactionID string) ([]TransactionStep, error)

	// UpdateStep atualiza o status/due_date/assignee de uma etapa específica.
	UpdateStep(ctx context.Context, tenantID, stepID string, input *UpdateStepInput) error

	// SeedDefaultSteps insere as etapas padrão para uma nova transação.
	SeedDefaultSteps(ctx context.Context, transactionID string) error
}

// DefaultSteps são as etapas padrão do processo de fechamento imobiliário no Brasil.
var DefaultSteps = []struct {
	Code  string
	Title string
}{
	{"proposta", "Proposta de compra"},
	{"contrato_promessa", "Contrato de promessa de compra e venda"},
	{"due_diligence", "Due diligence (matrícula, certidões)"},
	{"financiamento", "Aprovação de financiamento"},
	{"avaliacao_banco", "Avaliação do imóvel pelo banco"},
	{"escritura", "Lavratura da escritura"},
	{"registro", "Registro no cartório de imóveis"},
	{"entrega_chaves", "Entrega das chaves"},
}

// ── TransactionDoc ────────────────────────────────────────────────────────────

// TransactionDoc representa um documento do processo de transação.
type TransactionDoc struct {
	ID            string          `json:"id"`
	TransactionID string          `json:"transaction_id"`
	FileName      string          `json:"file_name"`
	FileURL       string          `json:"file_url"`
	DocType       string          `json:"doc_type"` // contract|deed|financing_proposal|id_document|other
	ExtractedData json.RawMessage `json:"extracted_data"`
	CreatedAt     time.Time       `json:"created_at"`
}

// AddDocInput define dados para adicionar um documento.
type AddDocInput struct {
	TransactionID string
	FileName      string
	FileURL       string
	DocType       string
	ExtractedData json.RawMessage
}

// TransactionDocStore define operações de persistência de documentos.
type TransactionDocStore interface {
	t.StoreType

	// AddDoc registra um documento em uma transação.
	AddDoc(ctx context.Context, tenantID string, input *AddDocInput) (*TransactionDoc, error)

	// ListByTransaction retorna os documentos de uma transação.
	ListByTransaction(ctx context.Context, tenantID, transactionID string) ([]TransactionDoc, error)

	// UpdateExtractedData atualiza os dados extraídos (LLM analysis) de um documento.
	UpdateExtractedData(ctx context.Context, tenantID, docID string, data json.RawMessage) error
}
