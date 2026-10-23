// Package siftleadeventstore guarda a trilha de eventos do lead.
//
// Toda transição vira evento com timestamp. É o que dá tempo-até-escalar e
// tempo-em-estágio sem cálculo posterior — e é o que faz o handoff parar de
// vazar: quem recebe o lead abre e vê o que aconteceu, em vez de o prospect
// repetir a história.
//
// A lista de tipos é fechada de propósito (CHECK no banco, constantes aqui):
// tipo de evento é vocabulário do código, não configuração da operação. Tipo
// novo entra por migração aditiva — ao contrário de stage.key, que a operação
// muda sem deploy.
package siftleadeventstore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

const (
	TypeCreated      = "created"
	TypeStageChanged = "stage_changed"
	TypeMessageIn    = "message_in"
	TypeMessageOut   = "message_out"
	TypeFitScored    = "fit_scored"
	TypeEscalated    = "escalated"
	TypeTurmaLinked  = "turma_linked"
	TypeLost         = "lost"
	TypeNoteAdded    = "note_added"
)

type LeadEvent struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenant_id"`
	LeadID      string         `json:"lead_id"`
	Type        string         `json:"type"`
	Body        string         `json:"body"`
	ActorUserID *string        `json:"actor_user_id,omitempty"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   *time.Time     `json:"created_at,omitempty"`
}

type CreateLeadEventInput struct {
	TenantID    string
	LeadID      string
	Type        string
	Body        string
	ActorUserID *string
	Metadata    map[string]any
}

type SiftLeadEventStore interface {
	t.StoreType

	Create(ctx context.Context, input *CreateLeadEventInput) (*LeadEvent, error)

	// ListByLead devolve a trilha em ordem cronológica — do mais antigo para o
	// mais novo, que é como se lê uma história. O índice é DESC porque a
	// consulta mais frequente é "o que aconteceu por último".
	ListByLead(ctx context.Context, tenantID, leadID string) ([]LeadEvent, error)
}
