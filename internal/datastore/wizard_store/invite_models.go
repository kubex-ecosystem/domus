// Package wizardstore contém os modelos de banco de dados e funções relacionadas a convites.
package wizardstore

import (
	"strings"
	"time"

	invitestore "github.com/kubex-ecosystem/domus/internal/datastore/invite_store"
	"gorm.io/gorm"
)

type partnerInvitationModel struct {
	ID         string                       `gorm:"column:id"`
	Token      string                       `gorm:"column:token"`
	Name       string                       `gorm:"column:name"`
	Email      string                       `gorm:"column:email"`
	Role       string                       `gorm:"column:role"`
	Status     invitestore.InvitationStatus `gorm:"column:status"`
	ExpiresAt  time.Time                    `gorm:"column:expires_at"`
	AcceptedAt *time.Time                   `gorm:"column:accepted_at"`
	TenantID   string                       `gorm:"column:tenant_id"`
	TeamID     *string                      `gorm:"column:team_id"`
	InvitedBy  string                       `gorm:"column:invited_by"`
	Metadata   *string                      `gorm:"column:metadata"`
	CreatedAt  time.Time                    `gorm:"column:created_at"`
	UpdatedAt  *time.Time                   `gorm:"column:updated_at"`
}

func (partnerInvitationModel) TableName() string { return "partner_invitation" }

func (m *partnerInvitationModel) toEntity() *invitestore.Invitation {
	if m == nil {
		return nil
	}
	return &invitestore.Invitation{
		ID:         m.ID,
		Token:      m.Token,
		Email:      m.Email,
		Name:       m.Name,
		Role:       m.Role,
		Type:       invitestore.TypePartner,
		Status:     m.Status,
		ExpiresAt:  m.ExpiresAt,
		AcceptedAt: m.AcceptedAt,
		TenantID:   m.TenantID,
		TeamID:     m.TeamID,
		InvitedBy:  m.InvitedBy,
		// Metadata:   m.Metadata,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type internalInvitationModel struct {
	ID         string                       `gorm:"column:id"`
	Token      string                       `gorm:"column:token"`
	Name       string                       `gorm:"column:name"`
	Email      string                       `gorm:"column:email"`
	Role       string                       `gorm:"column:role"`
	TeamID     *string                      `gorm:"column:team_id"`
	Status     invitestore.InvitationStatus `gorm:"column:status"`
	ExpiresAt  time.Time                    `gorm:"column:expires_at"`
	AcceptedAt *time.Time                   `gorm:"column:accepted_at"`
	TenantID   string                       `gorm:"column:tenant_id"`
	InvitedBy  string                       `gorm:"column:invited_by"`
	Metadata   *string                      `gorm:"column:metadata"`
	CreatedAt  time.Time                    `gorm:"column:created_at"`
	UpdatedAt  *time.Time                   `gorm:"column:updated_at"`
}

func (internalInvitationModel) TableName() string { return "internal_invitation" }

func (m *internalInvitationModel) toEntity() *invitestore.Invitation {
	if m == nil {
		return nil
	}
	return &invitestore.Invitation{
		ID:         m.ID,
		Token:      m.Token,
		Email:      m.Email,
		Name:       m.Name,
		Role:       m.Role,
		Type:       invitestore.TypeInternal,
		Status:     m.Status,
		ExpiresAt:  m.ExpiresAt,
		AcceptedAt: m.AcceptedAt,
		TenantID:   m.TenantID,
		InvitedBy:  m.InvitedBy,
		TeamID:     m.TeamID,
		// Metadata:   m.Metadata,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func applyPartnerFilters(db *gorm.DB, filters *invitestore.InvitationFilters) *gorm.DB {
	if filters == nil {
		return db
	}
	if filters.Email != nil && strings.TrimSpace(*filters.Email) != "" {
		db = db.Where("LOWER(partner_email) = ?", strings.ToLower(*filters.Email))
	}
	if filters.TenantID != nil {
		db = db.Where("tenant_id = ?", *filters.TenantID)
	}
	if filters.Status != nil {
		db = db.Where("status = ?", *filters.Status)
	}
	if filters.InvitedBy != nil {
		db = db.Where("invited_by = ?", *filters.InvitedBy)
	}
	return db
}

func applyInternalFilters(db *gorm.DB, filters *invitestore.InvitationFilters) *gorm.DB {
	if filters == nil {
		return db
	}
	if filters.Email != nil && strings.TrimSpace(*filters.Email) != "" {
		db = db.Where("LOWER(invitee_email) = ?", strings.ToLower(*filters.Email))
	}
	if filters.TenantID != nil {
		db = db.Where("tenant_id = ?", *filters.TenantID)
	}
	if filters.Status != nil {
		db = db.Where("status = ?", *filters.Status)
	}
	if filters.InvitedBy != nil {
		db = db.Where("invited_by = ?", *filters.InvitedBy)
	}
	return db
}

func paginate(filters *invitestore.InvitationFilters) func(*gorm.DB) *gorm.DB {
	page := max(filters.Page, 1)
	limit := max(filters.Limit, 20)
	offset := (page - 1) * limit
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(limit)
	}
}

func countTotal(db *gorm.DB) (int64, error) {
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
