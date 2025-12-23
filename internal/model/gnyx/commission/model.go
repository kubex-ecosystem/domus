// Package commission contém os modelos relacionados a comissões.
package commission

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

// CommissionType controla o tipo de pagamento de comissão.
type CommissionType string

const (
	CommissionTypeOneTime   CommissionType = "one-time"
	CommissionTypeRecurring CommissionType = "recurring"
)

// CommissionStatus rastreia o ciclo de vida da comissão.
type CommissionStatus string

const (
	CommissionStatusPending  CommissionStatus = "pending"
	CommissionStatusApproved CommissionStatus = "approved"
	CommissionStatusPaid     CommissionStatus = "paid"
	CommissionStatusClawback CommissionStatus = "clawback"
)

// CommissionRule espelha a tabela commission_rules.
type CommissionRule struct {
	ID                      gnyx.UUID       `json:"id" db:"id"`
	SetupCommissionRate     *float64        `json:"setup_commission_rate,omitempty" db:"setup_commission_rate"`
	RecurringCommissionRate *float64        `json:"recurring_commission_rate,omitempty" db:"recurring_commission_rate"`
	FirstPaymentOnly        *bool           `json:"first_payment_only,omitempty" db:"first_payment_only"`
	MinDealValue            *float64        `json:"min_deal_value,omitempty" db:"min_deal_value"`
	MaxCommissionAmount     *float64        `json:"max_commission_amount,omitempty" db:"max_commission_amount"`
	CreatedBy               *gnyx.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt               *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt               *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
}

// Commission representa a tabela commissions.
type Commission struct {
	ID               gnyx.UUID        `json:"id" db:"id"`
	LeadID           *gnyx.UUID       `json:"lead_id,omitempty" db:"lead_id"`
	Amount           float64          `json:"amount" db:"amount"`
	Type             CommissionType   `json:"type" db:"type"`
	Status           CommissionStatus `json:"status" db:"status"`
	Month            string           `json:"month" db:"month"`
	DealValue        *float64         `json:"deal_value,omitempty" db:"deal_value"`
	CommissionRate   *float64         `json:"commission_rate,omitempty" db:"commission_rate"`
	Notes            *string          `json:"notes,omitempty" db:"notes"`
	CreatedAt        *gnyx.Timestamp  `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt        *gnyx.Timestamp  `json:"updated_at,omitempty" db:"updated_at"`
	PartnerProfileID *gnyx.UUID       `json:"partner_profile_id,omitempty" db:"partner_profile_id"`
	CancelledReason  *string          `json:"cancelled_reason,omitempty" db:"cancelled_reason"`
	CancelledAt      *gnyx.Timestamp  `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancelledBy      *gnyx.UUID       `json:"cancelled_by,omitempty" db:"cancelled_by"`
}
