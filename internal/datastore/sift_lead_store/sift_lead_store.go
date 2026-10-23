// Package siftleadstore guarda os leads do funil de captação do Sift.
//
// Duas coisas aqui não existem em nenhuma outra tabela de lead do ecossistema,
// e são a razão de o Sift ter tabela própria em vez de reusar a do Colonial:
//
//   - CapturedBy — quem captou. É dele que sai a atribuição por pessoa, que é
//     uma das três dores que o produto existe para resolver.
//   - Source — a origem, opcional de propósito (ver o comentário no campo).
package siftleadstore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type Lead struct {
	ID       string  `json:"id"`
	TenantID string  `json:"tenant_id"`
	StageID  string  `json:"stage_id"`
	TurmaID  *string `json:"turma_id,omitempty"`

	Name     string `json:"name"`
	IGHandle string `json:"ig_handle"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`

	// Source é NULLABLE por decisão de produto, não por descuido: a ferramenta
	// suporta a possibilidade de duas origens, não exige as duas. Obrigar
	// classificação em toda entrada é atrito, e atrito derruba a adesão — que é
	// a régua honesta do piloto (VISION §8).
	Source       *string `json:"source,omitempty"`
	SourceDetail string  `json:"source_detail"`

	CapturedBy *string `json:"captured_by,omitempty"`
	AssignedTo *string `json:"assigned_to,omitempty"`

	FitScore   *int    `json:"fit_score,omitempty"`
	FitReason  string  `json:"fit_reason"`
	LostReason *string `json:"lost_reason,omitempty"`
	Notes      string  `json:"notes"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type CreateLeadInput struct {
	TenantID     string
	StageID      string
	TurmaID      *string
	Name         string
	IGHandle     string
	Phone        string
	Email        string
	Source       *string
	SourceDetail string
	CapturedBy   *string
	AssignedTo   *string
	Notes        string
}

type LeadFilters struct {
	TenantID string
	StageID  *string
	TurmaID  *string
	Source   *string
	Assigned *string
	Page     int
	Limit    int
}

type SiftLeadStore interface {
	t.StoreType

	Create(ctx context.Context, input *CreateLeadInput) (*Lead, error)
	GetByID(ctx context.Context, tenantID, id string) (*Lead, error)
	List(ctx context.Context, filters *LeadFilters) (*t.PaginatedResult[Lead], error)

	// FindDuplicate procura um lead já existente pelo handle (case-insensitive,
	// como o Instagram trata) ou pelo telefone. Existe porque a captura pública
	// é reentrante por natureza: a mesma pessoa preenche o formulário duas
	// vezes, ou chega pela landing depois de já ter sido cadastrada no DM
	// manual. Sem esta consulta o INSERT bateria nos índices únicos e a rota
	// devolveria erro no lugar de reconhecer quem já está no funil.
	//
	// (nil,nil) quando não há duplicata.
	FindDuplicate(ctx context.Context, tenantID, igHandle, phone string) (*Lead, error)

	// MoveToStage muda o estágio do lead. Não grava evento — quem orquestra é a
	// camada de rotas, que também escreve o lead_event correspondente; o store
	// não decide política.
	MoveToStage(ctx context.Context, tenantID, id, stageID string) (*Lead, error)
}
