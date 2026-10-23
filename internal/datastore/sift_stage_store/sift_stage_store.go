// Package siftstagestore guarda os estágios do pipeline do Sift.
//
// Estágio é DADO, não código: a operação reordena, renomeia e insere sem
// deploy. Por isso não há enum em Go espelhando as chaves — um `switch` sobre
// estágio no gateway anularia a decisão na primeira vez que a Growth quisesse
// uma coluna nova. Quem precisa distinguir "acabou" de "em andamento" usa
// `IsTerminal`, não o nome.
package siftstagestore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type Stage struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	Key        string     `json:"key"`
	Name       string     `json:"name"`
	OrderIndex int        `json:"order_index"`
	Color      string     `json:"color"`
	IsTerminal bool       `json:"is_terminal"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

// UpsertStageInput é a forma de um estágio no seed. `Key` é a chave natural
// dentro do tenant — é por ela que o upsert reconhece o que já existe.
type UpsertStageInput struct {
	Key        string
	Name       string
	OrderIndex int
	Color      string
	IsTerminal bool
}

type SiftStageStore interface {
	t.StoreType

	// List devolve os estágios do tenant em ordem de exibição.
	List(ctx context.Context, tenantID string) ([]Stage, error)

	// GetByKey busca pela chave natural. (nil,nil) quando não existe —
	// ausência não é erro.
	GetByKey(ctx context.Context, tenantID, key string) (*Stage, error)

	// Upsert semeia ou atualiza os estágios do tenant, casando por
	// (tenant_id, key). Idempotente: rodar duas vezes não duplica nem
	// reordena. É o caminho do bootstrap de tenant, que não cabe na DDL —
	// no momento da migração nenhum tenant real existe ainda.
	//
	// Não remove estágio ausente da entrada: apagar um estágio que a operação
	// criou por conta própria seria destruir configuração dela, e há leads
	// apontando para ele.
	Upsert(ctx context.Context, tenantID string, inputs []UpsertStageInput) ([]Stage, error)
}
