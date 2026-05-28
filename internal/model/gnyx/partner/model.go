// Package partner defines the data structures and types related to partner profiles within the Kubex system.
package partner

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

// PartnerStatus indicates whether a partner profile remains active.
type PartnerStatus string

const (
	PartnerStatusActive   PartnerStatus = "active"
	PartnerStatusInactive PartnerStatus = "inactive"
)

// Role captura a taxonomia detalhada de permissões internas.
type Role string

const (
	RoleAdmin              Role = "admin"
	RoleManager            Role = "manager"
	RoleSalesRep           Role = "sales_rep"
	RolePartnerReferral    Role = "partner_referral"
	RolePartnerPartnership Role = "partner_partnership"
	RoleSdrReferral        Role = "sdr_referral"
	RoleCloserReferral     Role = "closer_referral"
	RoleSdrPartnership     Role = "sdr_partnership"
	RoleCloserPartnership  Role = "closer_partnership"
	RoleKubexStaff         Role = "kubex_staff"
	RoleBackoffice         Role = "backoffice"
	RoleCs                 Role = "cs"
	RoleChannelManager     Role = "channel_manager"
	RoleDistributor        Role = "distributor"
	RoleCoSeller           Role = "co_seller"
	RoleAffiliate          Role = "affiliate"
	RolePartnerReseller    Role = "partner_reseller"
	RolePartner            Role = "partner"
)

// Profile espelha a tabela profiles.
type Profile struct {
	ID                 gnyx.UUID       `json:"id" db:"id"`
	Email              string          `json:"email" db:"email"`
	Name               string          `json:"name" db:"name"`
	Role               *string         `json:"role,omitempty" db:"role"`
	CompanyID          *gnyx.UUID      `json:"company_id,omitempty" db:"company_id"`
	CreatedAt          *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt          *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
	AvatarURL          *string         `json:"avatar_url,omitempty" db:"avatar_url"`
	Company            *string         `json:"company,omitempty" db:"company"`
	Phone              *string         `json:"phone,omitempty" db:"phone"`
	Status             *PartnerStatus  `json:"status,omitempty" db:"status"`
	LastName           *string         `json:"last_name,omitempty" db:"last_name"`
	ForcePasswordReset *bool           `json:"force_password_reset,omitempty" db:"force_password_reset"`
	ManagedBy          *gnyx.UUID      `json:"managed_by,omitempty" db:"managed_by"`
}
