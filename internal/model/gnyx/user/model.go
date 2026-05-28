// Package user contém definições de modelos relacionados a usuários.
package user

import (
	"github.com/google/uuid"
	"github.com/kubex-ecosystem/domus/internal/model/gnyx"
)

// UserRole é utilizado em convites e regras de acesso mais amplas.
type UserRole string

const (
	UserRoleAdmin             UserRole = "admin"
	UserRoleManager           UserRole = "manager"
	UserRolePartner           UserRole = "partner"
	UserRoleCS                UserRole = "cs"
	UserRoleBackoffice        UserRole = "backoffice"
	UserRoleKubexStaff        UserRole = "kubex_staff"
	UserRolePartnerReferral   UserRole = "partner_referral"
	UserRolePartnerReseller   UserRole = "partner_reseller"
	UserRoleAffiliate         UserRole = "affiliate"
	UserRoleCoSeller          UserRole = "co_seller"
	UserRoleDistributor       UserRole = "distributor"
	UserRoleChannelManager    UserRole = "channel_manager"
	UserRoleCSManager         UserRole = "cs_manager"
	UserRoleSdr               UserRole = "sdr"
	UserRoleCloser            UserRole = "closer"
	UserRoleSdrReferral       UserRole = "sdr_referral"
	UserRoleCloserReferral    UserRole = "closer_referral"
	UserRoleSdrPartnership    UserRole = "sdr_partnership"
	UserRoleCloserPartnership UserRole = "closer_partnership"
)

// UserInvitation representa a tabela user_invitations.
type UserInvitation struct {
	ID         gnyx.UUID       `json:"id" db:"id"`
	Email      string          `json:"email" db:"email"`
	Role       *string         `json:"role,omitempty" db:"role"`
	CompanyID  *gnyx.UUID      `json:"company_id,omitempty" db:"company_id"`
	InvitedBy  *gnyx.UUID      `json:"invited_by,omitempty" db:"invited_by"`
	Token      string          `json:"token" db:"token"`
	ExpiresAt  *gnyx.Timestamp `json:"expires_at,omitempty" db:"expires_at"`
	AcceptedAt *gnyx.Timestamp `json:"accepted_at,omitempty" db:"accepted_at"`
	CreatedAt  *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
}

// User representa a tabela users.
type User struct {
	ID         gnyx.UUID       `json:"id" db:"id"`
	Email      string          `json:"email" db:"email"`
	Name       string          `json:"name" db:"name"`
	Role       *string         `json:"role,omitempty" db:"role"`
	CompanyID  *gnyx.UUID      `json:"company_id,omitempty" db:"company_id"`
	Password   string          `json:"-" db:"password"`
	CreatedAt  *gnyx.Timestamp `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt  *gnyx.Timestamp `json:"updated_at,omitempty" db:"updated_at"`
	LastLogin  *gnyx.Timestamp `json:"last_login,omitempty" db:"last_login"`
	IsActive   bool            `json:"is_active" db:"is_active"`
	InvitedBy  *gnyx.UUID      `json:"invited_by,omitempty" db:"invited_by"`
	AcceptedAt *gnyx.Timestamp `json:"accepted_at,omitempty" db:"accepted_at"`
}

// CurrentUser retrieves the current system user.
func CurrentUser() (*User, error) {
	role := string(UserRoleAdmin)
	// Implementação fictícia para exemplo.
	return &User{
		ID:    gnyx.UUID(uuid.New()),
		Email: "user@example.com",
		Name:  "John Doe",
		Role:  &role,
	}, nil
}
