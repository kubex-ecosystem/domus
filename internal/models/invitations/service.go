// Package invitations provides services for managing invitations (partner and internal).
package invitations

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

const (
	// DefaultTokenLength is the default length for invitation tokens
	DefaultTokenLength = 32
	// DefaultExpirationDays is the default expiration time in days
	DefaultExpirationDays = 7
)

// Service defines the interface for invitation service operations
type Service interface {
	// Partner Invitation methods
	CreatePartnerInvitation(ctx context.Context, dto *CreatePartnerInvitationDTO) (*PartnerInvitation, error)
	GetPartnerInvitation(ctx context.Context, id string) (*PartnerInvitation, error)
	GetPartnerInvitationByToken(ctx context.Context, token string) (*PartnerInvitation, error)
	UpdatePartnerInvitation(ctx context.Context, id string, dto *UpdateInvitationDTO) (*PartnerInvitation, error)
	RevokePartnerInvitation(ctx context.Context, id string) error
	AcceptPartnerInvitation(ctx context.Context, token string) (*PartnerInvitation, error)
	DeletePartnerInvitation(ctx context.Context, id string) error
	ListPartnerInvitations(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error)

	// Internal Invitation methods
	CreateInternalInvitation(ctx context.Context, dto *CreateInternalInvitationDTO) (*InternalInvitation, error)
	GetInternalInvitation(ctx context.Context, id string) (*InternalInvitation, error)
	GetInternalInvitationByToken(ctx context.Context, token string) (*InternalInvitation, error)
	UpdateInternalInvitation(ctx context.Context, id string, dto *UpdateInvitationDTO) (*InternalInvitation, error)
	RevokeInternalInvitation(ctx context.Context, id string) error
	AcceptInternalInvitation(ctx context.Context, token string) (*InternalInvitation, error)
	DeleteInternalInvitation(ctx context.Context, id string) error
	ListInternalInvitations(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error)

	// Generic methods
	GetInvitationByToken(ctx context.Context, token string) (*GenericInvitation, error)
	ListAllInvitations(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error)
}

// invitationService is the concrete implementation of Service
type invitationService struct {
	repo Repository
}

// NewService creates a new invitation service
func NewService(repo Repository) Service {
	return &invitationService{repo: repo}
}

// ===========================================================================
// PARTNER INVITATION METHODS
// ===========================================================================

// CreatePartnerInvitation creates a new partner invitation with a generated token
func (s *invitationService) CreatePartnerInvitation(ctx context.Context, dto *CreatePartnerInvitationDTO) (*PartnerInvitation, error) {
	// Generate unique token
	token, err := generateToken(DefaultTokenLength)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar token: %v", err)
	}

	// Set expiration (default: now + 7 days)
	expiresAt := time.Now().Add(DefaultExpirationDays * 24 * time.Hour)
	if dto.ExpiresAt != nil {
		expiresAt = *dto.ExpiresAt
	}

	invitation := &PartnerInvitation{
		Token:        token,
		PartnerEmail: dto.PartnerEmail,
		PartnerName:  dto.PartnerName,
		CompanyName:  dto.CompanyName,
		Role:         dto.Role,
		Status:       StatusPending,
		ExpiresAt:    expiresAt,
		TenantID:     dto.TenantID,
		InvitedBy:    dto.InvitedBy,
		Metadata:     dto.Metadata,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreatePartner(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

// GetPartnerInvitation retrieves a partner invitation by ID
func (s *invitationService) GetPartnerInvitation(ctx context.Context, id string) (*PartnerInvitation, error) {
	return s.repo.GetPartnerByID(ctx, id)
}

// GetPartnerInvitationByToken retrieves a partner invitation by token
func (s *invitationService) GetPartnerInvitationByToken(ctx context.Context, token string) (*PartnerInvitation, error) {
	return s.repo.GetPartnerByToken(ctx, token)
}

// UpdatePartnerInvitation updates a partner invitation
func (s *invitationService) UpdatePartnerInvitation(ctx context.Context, id string, dto *UpdateInvitationDTO) (*PartnerInvitation, error) {
	invitation, err := s.repo.GetPartnerByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if dto.Status != nil {
		invitation.Status = *dto.Status
	}
	if dto.AcceptedAt != nil {
		invitation.AcceptedAt = dto.AcceptedAt
	}
	if dto.ExpiresAt != nil {
		invitation.ExpiresAt = *dto.ExpiresAt
	}
	if dto.Metadata != nil {
		invitation.Metadata = dto.Metadata
	}

	now := time.Now()
	invitation.UpdatedAt = &now

	if err := s.repo.UpdatePartner(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

// RevokePartnerInvitation revokes a partner invitation
func (s *invitationService) RevokePartnerInvitation(ctx context.Context, id string) error {
	return s.repo.UpdatePartnerStatus(ctx, id, StatusRevoked)
}

// AcceptPartnerInvitation accepts a partner invitation by token
func (s *invitationService) AcceptPartnerInvitation(ctx context.Context, token string) (*PartnerInvitation, error) {
	invitation, err := s.repo.GetPartnerByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Update status to accepted
	if err := s.repo.UpdatePartnerStatus(ctx, invitation.ID, StatusAccepted); err != nil {
		return nil, err
	}

	// Fetch updated invitation
	return s.repo.GetPartnerByID(ctx, invitation.ID)
}

// DeletePartnerInvitation deletes a partner invitation
func (s *invitationService) DeletePartnerInvitation(ctx context.Context, id string) error {
	return s.repo.DeletePartner(ctx, id)
}

// ListPartnerInvitations lists partner invitations with filtering
func (s *invitationService) ListPartnerInvitations(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error) {
	return s.repo.ListPartners(ctx, filters)
}

// ===========================================================================
// INTERNAL INVITATION METHODS
// ===========================================================================

// CreateInternalInvitation creates a new internal invitation with a generated token
func (s *invitationService) CreateInternalInvitation(ctx context.Context, dto *CreateInternalInvitationDTO) (*InternalInvitation, error) {
	// Generate unique token
	token, err := generateToken(DefaultTokenLength)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar token: %v", err)
	}

	// Set expiration (default: now + 7 days)
	expiresAt := time.Now().Add(DefaultExpirationDays * 24 * time.Hour)
	if dto.ExpiresAt != nil {
		expiresAt = *dto.ExpiresAt
	}

	invitation := &InternalInvitation{
		Token:        token,
		InviteeEmail: dto.InviteeEmail,
		InviteeName:  dto.InviteeName,
		Role:         dto.Role,
		TeamID:       dto.TeamID,
		Status:       StatusPending,
		ExpiresAt:    expiresAt,
		TenantID:     dto.TenantID,
		InvitedBy:    dto.InvitedBy,
		Metadata:     dto.Metadata,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateInternal(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

// GetInternalInvitation retrieves an internal invitation by ID
func (s *invitationService) GetInternalInvitation(ctx context.Context, id string) (*InternalInvitation, error) {
	return s.repo.GetInternalByID(ctx, id)
}

// GetInternalInvitationByToken retrieves an internal invitation by token
func (s *invitationService) GetInternalInvitationByToken(ctx context.Context, token string) (*InternalInvitation, error) {
	return s.repo.GetInternalByToken(ctx, token)
}

// UpdateInternalInvitation updates an internal invitation
func (s *invitationService) UpdateInternalInvitation(ctx context.Context, id string, dto *UpdateInvitationDTO) (*InternalInvitation, error) {
	invitation, err := s.repo.GetInternalByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if dto.Status != nil {
		invitation.Status = *dto.Status
	}
	if dto.AcceptedAt != nil {
		invitation.AcceptedAt = dto.AcceptedAt
	}
	if dto.ExpiresAt != nil {
		invitation.ExpiresAt = *dto.ExpiresAt
	}
	if dto.Metadata != nil {
		invitation.Metadata = dto.Metadata
	}

	now := time.Now()
	invitation.UpdatedAt = &now

	if err := s.repo.UpdateInternal(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

// RevokeInternalInvitation revokes an internal invitation
func (s *invitationService) RevokeInternalInvitation(ctx context.Context, id string) error {
	return s.repo.UpdateInternalStatus(ctx, id, StatusRevoked)
}

// AcceptInternalInvitation accepts an internal invitation by token
func (s *invitationService) AcceptInternalInvitation(ctx context.Context, token string) (*InternalInvitation, error) {
	invitation, err := s.repo.GetInternalByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Update status to accepted
	if err := s.repo.UpdateInternalStatus(ctx, invitation.ID, StatusAccepted); err != nil {
		return nil, err
	}

	// Fetch updated invitation
	return s.repo.GetInternalByID(ctx, invitation.ID)
}

// DeleteInternalInvitation deletes an internal invitation
func (s *invitationService) DeleteInternalInvitation(ctx context.Context, id string) error {
	return s.repo.DeleteInternal(ctx, id)
}

// ListInternalInvitations lists internal invitations with filtering
func (s *invitationService) ListInternalInvitations(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error) {
	return s.repo.ListInternals(ctx, filters)
}

// ===========================================================================
// GENERIC METHODS
// ===========================================================================

// GetInvitationByToken searches for an invitation by token in both tables
func (s *invitationService) GetInvitationByToken(ctx context.Context, token string) (*GenericInvitation, error) {
	return s.repo.GetByToken(ctx, token)
}

// ListAllInvitations lists all invitations (both partner and internal)
func (s *invitationService) ListAllInvitations(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error) {
	return s.repo.ListAll(ctx, filters)
}

// ===========================================================================
// HELPER FUNCTIONS
// ===========================================================================

// generateToken generates a cryptographically secure random token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("erro ao gerar bytes aleatórios: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
