// Package siftturmastore guarda as turmas do Sift.
//
// Tabela fina e representativa de propósito: serve de FK do lead e de exibição
// na landing. Não há gestão de matrícula, nota ou presença — isso é Fase 2, se
// houver (VISION §5).
//
// `turma` é a única palavra em português na estrutura, e é deliberado: "class"
// colide com palavra reservada e com classe-de-objeto, e "cohort" é jargão de
// outro domínio.
package siftturmastore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type Turma struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	Name         string     `json:"name"`
	PeriodStart  *time.Time `json:"period_start,omitempty"`
	PeriodEnd    *time.Time `json:"period_end,omitempty"`
	MonthlyPrice *float64   `json:"monthly_price,omitempty"`
	Status       string     `json:"status"`
	LandingSlug  string     `json:"landing_slug"`
	// LandingTemplate escolhe o template da landing. É coluna, e não decisão de
	// código, porque a escolha depende do acervo disponível na hora: turma sem
	// foto pronta fica melhor no template textual do que num de vitrine com
	// buraco no lugar da imagem.
	LandingTemplate string     `json:"landing_template"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type CreateTurmaInput struct {
	TenantID        string
	Name            string
	PeriodStart     *time.Time
	PeriodEnd       *time.Time
	MonthlyPrice    *float64
	Status          string
	LandingSlug     string
	LandingTemplate string
}

type SiftTurmaStore interface {
	t.StoreType

	// GetBySlug NÃO recebe tenant, e isso é a característica central da tabela.
	// A landing em /t/:slug é anônima: o slug é a única informação que o
	// servidor tem, e é dele que sai o tenant. Por isso o índice é UNIQUE
	// global. (nil,nil) quando não existe.
	GetBySlug(ctx context.Context, slug string) (*Turma, error)

	GetByID(ctx context.Context, tenantID, id string) (*Turma, error)
	List(ctx context.Context, tenantID string) ([]Turma, error)
	Create(ctx context.Context, input *CreateTurmaInput) (*Turma, error)
}
