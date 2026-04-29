package invitations

import (
	"time"
)

// InvitationStatus represents the status of an invitation
type InvitationStatus string

const (
	StatusPending  InvitationStatus = "pending"
	StatusAccepted InvitationStatus = "accepted"
	StatusExpired  InvitationStatus = "expired"
	StatusRevoked  InvitationStatus = "revoked"
)

// InvitationType represents the type of invitation
type InvitationType string

const (
	TypePartner  InvitationType = "partner"
	TypeInternal InvitationType = "internal"
)

// PartnerInvitation represents an invitation for a partner
// Maps to 'partner_invitation' table in the database
type PartnerInvitation struct {
	ID           string           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Token        string           `gorm:"type:varchar(255);unique;not null" json:"token"`
	PartnerEmail string           `gorm:"type:text;not null" json:"partner_email"`
	PartnerName  *string          `gorm:"type:text" json:"partner_name,omitempty"`
	CompanyName  *string          `gorm:"type:text" json:"company_name,omitempty"`
	Role         string           `gorm:"type:text" json:"role"` // Flexible: UUID or code
	Status       InvitationStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ExpiresAt    time.Time        `gorm:"type:timestamptz;not null" json:"expires_at"`
	AcceptedAt   *time.Time       `gorm:"type:timestamptz" json:"accepted_at,omitempty"`
	TenantID     string           `gorm:"type:uuid;not null" json:"tenant_id"`
	InvitedBy    string           `gorm:"type:uuid;not null" json:"invited_by"`
	Metadata     *string          `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time        `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    *time.Time       `gorm:"type:timestamptz" json:"updated_at,omitempty"`
}

// TableName specifies the table name for GORM
func (PartnerInvitation) TableName() string {
	return "partner_invitation"
}

// InternalInvitation represents an invitation for an internal user
// Maps to 'internal_invitation' table in the database
type InternalInvitation struct {
	ID           string           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Token        string           `gorm:"type:varchar(255);unique;not null" json:"token"`
	InviteeEmail string           `gorm:"type:text;not null" json:"invitee_email"`
	InviteeName  *string          `gorm:"type:text" json:"invitee_name,omitempty"`
	Role         string           `gorm:"type:text" json:"role"` // Flexible: UUID or code
	TeamID       *string          `gorm:"type:uuid" json:"team_id,omitempty"`
	Status       InvitationStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ExpiresAt    time.Time        `gorm:"type:timestamptz;not null" json:"expires_at"`
	AcceptedAt   *time.Time       `gorm:"type:timestamptz" json:"accepted_at,omitempty"`
	TenantID     string           `gorm:"type:uuid;not null" json:"tenant_id"`
	InvitedBy    string           `gorm:"type:uuid;not null" json:"invited_by"`
	Metadata     *string          `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time        `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    *time.Time       `gorm:"type:timestamptz" json:"updated_at,omitempty"`
}

// TableName specifies the table name for GORM
func (InternalInvitation) TableName() string {
	return "internal_invitation"
}

// GenericInvitation represents a unified view of any invitation type
// Used for API responses and cross-type operations
type GenericInvitation struct {
	ID         string           `json:"id"`
	Token      string           `json:"token"`
	Email      string           `json:"email"`
	Name       *string          `json:"name,omitempty"`
	Role       string           `json:"role"`
	Type       InvitationType   `json:"type"`
	Status     InvitationStatus `json:"status"`
	ExpiresAt  time.Time        `json:"expires_at"`
	AcceptedAt *time.Time       `json:"accepted_at,omitempty"`
	TenantID   string           `json:"tenant_id"`
	InvitedBy  string           `json:"invited_by"`
	TeamID     *string          `json:"team_id,omitempty"` // Only for internal
	Company    *string          `json:"company,omitempty"` // Only for partner
	Metadata   *string          `json:"metadata,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  *time.Time       `json:"updated_at,omitempty"`
}

// CreatePartnerInvitationDTO represents the data for creating a partner invitation
type CreatePartnerInvitationDTO struct {
	PartnerEmail string     `json:"partner_email" binding:"required,email"`
	PartnerName  *string    `json:"partner_name,omitempty"`
	CompanyName  *string    `json:"company_name,omitempty"`
	Role         string     `json:"role" binding:"required"`
	TenantID     string     `json:"tenant_id" binding:"required"`
	InvitedBy    string     `json:"invited_by" binding:"required"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"` // Optional, defaults to now + 7 days
	Metadata     *string    `json:"metadata,omitempty"`
}

// CreateInternalInvitationDTO represents the data for creating an internal invitation
type CreateInternalInvitationDTO struct {
	InviteeEmail string     `json:"invitee_email" binding:"required,email"`
	InviteeName  *string    `json:"invitee_name,omitempty"`
	Role         string     `json:"role" binding:"required"`
	TeamID       *string    `json:"team_id,omitempty"`
	TenantID     string     `json:"tenant_id" binding:"required"`
	InvitedBy    string     `json:"invited_by" binding:"required"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"` // Optional, defaults to now + 7 days
	Metadata     *string    `json:"metadata,omitempty"`
}

// UpdateInvitationDTO represents the data for updating an invitation
type UpdateInvitationDTO struct {
	Status     *InvitationStatus `json:"status,omitempty"`
	AcceptedAt *time.Time        `json:"accepted_at,omitempty"`
	ExpiresAt  *time.Time        `json:"expires_at,omitempty"`
	Metadata   *string           `json:"metadata,omitempty"`
}

// InvitationFilterParams represents the parameters for filtering invitations
type InvitationFilterParams struct {
	Email     *string           `json:"email,omitempty"`
	TenantID  *string           `json:"tenant_id,omitempty"`
	Status    *InvitationStatus `json:"status,omitempty"`
	InvitedBy *string           `json:"invited_by,omitempty"`
	Type      *InvitationType   `json:"type,omitempty"`
	Page      int               `json:"page" binding:"min=1"`
	Limit     int               `json:"limit" binding:"min=1,max=100"`
}

// PaginatedInvitationResult represents a paginated result of invitations
type PaginatedInvitationResult struct {
	Data       []GenericInvitation `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
}

// ToGeneric converts a PartnerInvitation to GenericInvitation
func (p *PartnerInvitation) ToGeneric() *GenericInvitation {
	return &GenericInvitation{
		ID:         p.ID,
		Token:      p.Token,
		Email:      p.PartnerEmail,
		Name:       p.PartnerName,
		Role:       p.Role,
		Type:       TypePartner,
		Status:     p.Status,
		ExpiresAt:  p.ExpiresAt,
		AcceptedAt: p.AcceptedAt,
		TenantID:   p.TenantID,
		InvitedBy:  p.InvitedBy,
		Company:    p.CompanyName,
		Metadata:   p.Metadata,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

// ToGeneric converts an InternalInvitation to GenericInvitation
func (i *InternalInvitation) ToGeneric() *GenericInvitation {
	return &GenericInvitation{
		ID:         i.ID,
		Token:      i.Token,
		Email:      i.InviteeEmail,
		Name:       i.InviteeName,
		Role:       i.Role,
		Type:       TypeInternal,
		Status:     i.Status,
		ExpiresAt:  i.ExpiresAt,
		AcceptedAt: i.AcceptedAt,
		TenantID:   i.TenantID,
		InvitedBy:  i.InvitedBy,
		TeamID:     i.TeamID,
		Metadata:   i.Metadata,
		CreatedAt:  i.CreatedAt,
		UpdatedAt:  i.UpdatedAt,
	}
}
