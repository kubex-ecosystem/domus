// Package lead defines the Lead model and related types.
package lead

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

// LeadStage enumerates the lifecycle stages for a lead.
type LeadStage string

const (
	LeadStageProspect      LeadStage = "prospect"
	LeadStageContacting    LeadStage = "contacting"
	LeadStageQualified     LeadStage = "qualified"
	LeadStageDemoScheduled LeadStage = "demo_scheduled"
	LeadStageProposal      LeadStage = "proposal"
	LeadStageNegotiation   LeadStage = "negotiation"
	LeadStageWon           LeadStage = "won"
	LeadStageLost          LeadStage = "lost"
)

// LeadStatus expresses the high-level state of a lead.
type LeadStatus string

const (
	LeadStatusActive LeadStatus = "active"
	LeadStatusWon    LeadStatus = "won"
	LeadStatusLost   LeadStatus = "lost"
)

// LeadType categorizes the lead origin.
type LeadType string

const (
	LeadTypeCustomer LeadType = "customer"
	LeadTypePartner  LeadType = "partner"
)

// Lead espelha a tabela leads.
type Lead struct {
	ID                         gnyx.UUID       `json:"id" db:"id"`
	Name                       string          `json:"name" db:"name"`
	Email                      string          `json:"email" db:"email"`
	Phone                      *string         `json:"phone,omitempty" db:"phone"`
	Company                    *string         `json:"company,omitempty" db:"company"`
	Value                      *float64        `json:"value,omitempty" db:"value"`
	Status                     *LeadStatus     `json:"status,omitempty" db:"status"`
	Stage                      *LeadStage      `json:"stage,omitempty" db:"stage"`
	CompanyID                  *gnyx.UUID      `json:"company_id,omitempty" db:"company_id"`
	Observations               *string         `json:"observations,omitempty" db:"observations"`
	CreatedAt                  *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt                  *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
	CreatedBy                  *gnyx.UUID      `json:"created_by,omitempty" db:"created_by"`
	AssignedTo                 *gnyx.UUID      `json:"assigned_to,omitempty" db:"assigned_to"`
	PipelineID                 *gnyx.UUID      `json:"pipeline_id,omitempty" db:"pipeline_id"`
	StageID                    *gnyx.UUID      `json:"stage_id,omitempty" db:"stage_id"`
	ReferredBy                 *gnyx.UUID      `json:"referred_by,omitempty" db:"referred_by"`
	ReopenedAt                 *gnyx.Timestamp `json:"reopened_at,omitempty" db:"reopened_at"`
	ReopenedBy                 *gnyx.UUID      `json:"reopened_by,omitempty" db:"reopened_by"`
	ClosedAt                   *gnyx.Timestamp `json:"closed_at,omitempty" db:"closed_at"`
	LeadType                   *LeadType       `json:"lead_type,omitempty" db:"lead_type"`
	PartnerClientPortfolioSize *float64        `json:"partner_client_portfolio_size,omitempty" db:"partner_client_portfolio_size"`
	PartnerMonthlyRevenue      *float64        `json:"partner_monthly_revenue,omitempty" db:"partner_monthly_revenue"`
	PartnerTargetType          *string         `json:"partner_target_type,omitempty" db:"partner_target_type"`
	PartnerExperienceYears     *int64          `json:"partner_experience_years,omitempty" db:"partner_experience_years"`
	PartnerMarketSegment       *string         `json:"partner_market_segment,omitempty" db:"partner_market_segment"`
}
