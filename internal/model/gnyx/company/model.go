// Package company contém o modelo da entidade Empresa.
package company

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

// SubscriptionPlan expressa o plano vinculado à empresa.
type SubscriptionPlan string

const (
	SubscriptionPlanStarter    SubscriptionPlan = "starter"
	SubscriptionPlanPro        SubscriptionPlan = "pro"
	SubscriptionPlanEnterprise SubscriptionPlan = "enterprise"
)

// Company espelha a tabela companies.
type Company struct {
	ID            gnyx.UUID         `json:"id" db:"id"`
	Name          string            `json:"name" db:"name"`
	Slug          string            `json:"slug" db:"slug"`
	CreatedAt     *gnyx.Timestamp   `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt     *gnyx.Timestamp   `json:"updated_at,omitempty" db:"updated_at"`
	Plan          *SubscriptionPlan `json:"plan,omitempty" db:"plan"`
	PlanExpiresAt *gnyx.Timestamp   `json:"plan_expires_at,omitempty" db:"plan_expires_at"`
	IsTrial       *bool             `json:"is_trial,omitempty" db:"is_trial"`
	IsActive      *bool             `json:"is_active,omitempty" db:"is_active"`
	Domain        *string           `json:"domain,omitempty" db:"domain"`
	Phone         *string           `json:"phone,omitempty" db:"phone"`
	Address       *string           `json:"address,omitempty" db:"address"`
}
