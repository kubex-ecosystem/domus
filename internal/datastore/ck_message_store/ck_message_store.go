package ckmessagestore

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/domus/internal/types"
)

type CKMessage struct {
	ID         string     `json:"id"`
	CompanyID  string     `json:"company_id"`
	OrderID    *string    `json:"order_id,omitempty"`
	Channel    string     `json:"channel"`
	ClientName string     `json:"client_name"`
	Content    string     `json:"content"`
	Direction  string     `json:"direction"`
	Read       bool       `json:"read"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
}

type CreateCKMessageInput struct {
	CompanyID  string
	OrderID    *string
	Channel    string
	ClientName string
	Content    string
	Direction  string
	Read       bool
}

type CKMessageFilters struct {
	CompanyID *string
	OrderID   *string
	Channel   *string
	Direction *string
	Unread    *bool
	Page      int
	Limit     int
}

type CKMessageStore interface {
	t.StoreType

	Create(ctx context.Context, input *CreateCKMessageInput) (*CKMessage, error)
	GetByID(ctx context.Context, id string) (*CKMessage, error)
	List(ctx context.Context, filters *CKMessageFilters) (*t.PaginatedResult[CKMessage], error)
	MarkRead(ctx context.Context, id string) error
	// MarkConversationRead marca como lidas todas as mensagens inbound de uma conversa.
	// Se orderID != nil, agrupa por order; senão, por companyID+channel+clientName.
	MarkConversationRead(ctx context.Context, companyID, channel, clientName string, orderID *string) error
}
