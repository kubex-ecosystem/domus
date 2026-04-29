package invitations

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

var (
	// ErrInvitationNotFound is returned when an invitation is not found
	ErrInvitationNotFound = errors.New("invitation not found")
	// ErrInvitationExpired is returned when an invitation has expired
	ErrInvitationExpired = errors.New("invitation has expired")
	// ErrInvalidInvitationStatus is returned when invitation status is invalid
	ErrInvalidInvitationStatus = errors.New("invalid invitation status")
	// ErrTokenAlreadyExists is returned when trying to create an invitation with a duplicate token
	ErrTokenAlreadyExists = errors.New("token already exists")
)

// Repository defines the interface for invitation repository operations
type Repository interface {
	// Partner Invitation methods
	CreatePartner(ctx context.Context, invitation *PartnerInvitation) error
	GetPartnerByID(ctx context.Context, id string) (*PartnerInvitation, error)
	GetPartnerByToken(ctx context.Context, token string) (*PartnerInvitation, error)
	UpdatePartner(ctx context.Context, invitation *PartnerInvitation) error
	UpdatePartnerStatus(ctx context.Context, id string, status InvitationStatus) error
	DeletePartner(ctx context.Context, id string) error
	ListPartners(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error)

	// Internal Invitation methods
	CreateInternal(ctx context.Context, invitation *InternalInvitation) error
	GetInternalByID(ctx context.Context, id string) (*InternalInvitation, error)
	GetInternalByToken(ctx context.Context, token string) (*InternalInvitation, error)
	UpdateInternal(ctx context.Context, invitation *InternalInvitation) error
	UpdateInternalStatus(ctx context.Context, id string, status InvitationStatus) error
	DeleteInternal(ctx context.Context, id string) error
	ListInternals(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error)

	// Generic methods (searches both tables)
	GetByToken(ctx context.Context, token string) (*GenericInvitation, error)
	ListAll(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error)
}

// invitationRepository is the concrete implementation of Repository
type invitationRepository struct {
	db *gorm.DB
}

// NewRepository creates a new invitation repository
func NewRepository(db *gorm.DB) Repository {
	return &invitationRepository{db: db}
}

// ===========================================================================
// PARTNER INVITATION METHODS
// ===========================================================================

// CreatePartner creates a new partner invitation
func (r *invitationRepository) CreatePartner(ctx context.Context, invitation *PartnerInvitation) error {
	if err := r.db.WithContext(ctx).Create(invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrTokenAlreadyExists
		}
		return fmt.Errorf("erro ao criar convite de parceiro: %v", err)
	}
	return nil
}

// GetPartnerByID finds a partner invitation by ID
func (r *invitationRepository) GetPartnerByID(ctx context.Context, id string) (*PartnerInvitation, error) {
	var invitation PartnerInvitation
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, fmt.Errorf("erro ao buscar convite de parceiro: %v", err)
	}
	return &invitation, nil
}

// GetPartnerByToken finds a partner invitation by token
func (r *invitationRepository) GetPartnerByToken(ctx context.Context, token string) (*PartnerInvitation, error) {
	var invitation PartnerInvitation
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, fmt.Errorf("erro ao buscar convite de parceiro por token: %v", err)
	}

	// Check if expired
	if time.Now().After(invitation.ExpiresAt) && invitation.Status == StatusPending {
		return nil, ErrInvitationExpired
	}

	// Check if status is valid (must be pending)
	if invitation.Status != StatusPending {
		return nil, ErrInvalidInvitationStatus
	}

	return &invitation, nil
}

// UpdatePartner updates an existing partner invitation
func (r *invitationRepository) UpdatePartner(ctx context.Context, invitation *PartnerInvitation) error {
	if err := r.db.WithContext(ctx).Save(invitation).Error; err != nil {
		return fmt.Errorf("erro ao atualizar convite de parceiro: %v", err)
	}
	return nil
}

// UpdatePartnerStatus updates only the status of a partner invitation
func (r *invitationRepository) UpdatePartnerStatus(ctx context.Context, id string, status InvitationStatus) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}

	// If accepting, set accepted_at
	if status == StatusAccepted {
		updates["accepted_at"] = now
	}

	if err := r.db.WithContext(ctx).
		Model(&PartnerInvitation{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("erro ao atualizar status do convite de parceiro: %v", err)
	}
	return nil
}

// DeletePartner deletes a partner invitation by ID
func (r *invitationRepository) DeletePartner(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&PartnerInvitation{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("erro ao deletar convite de parceiro: %v", err)
	}
	return nil
}

// ListPartners lists partner invitations with filtering and pagination
func (r *invitationRepository) ListPartners(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error) {
	var invitations []PartnerInvitation
	var total int64

	query := r.db.WithContext(ctx).Model(&PartnerInvitation{})

	// Apply filters
	if filters != nil {
		if filters.Email != nil && *filters.Email != "" {
			query = query.Where("partner_email ILIKE ?", "%"+*filters.Email+"%")
		}
		if filters.TenantID != nil {
			query = query.Where("tenant_id = ?", *filters.TenantID)
		}
		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
		}
		if filters.InvitedBy != nil {
			query = query.Where("invited_by = ?", *filters.InvitedBy)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("erro ao contar convites de parceiros: %v", err)
	}

	// Pagination
	page, limit := getPaginationParams(filters)
	offset := (page - 1) * limit

	// Fetch invitations
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("erro ao listar convites de parceiros: %v", err)
	}

	// Convert to generic
	genericInvitations := make([]GenericInvitation, len(invitations))
	for i, inv := range invitations {
		genericInvitations[i] = *inv.ToGeneric()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &PaginatedInvitationResult{
		Data:       genericInvitations,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// ===========================================================================
// INTERNAL INVITATION METHODS
// ===========================================================================

// CreateInternal creates a new internal invitation
func (r *invitationRepository) CreateInternal(ctx context.Context, invitation *InternalInvitation) error {
	if err := r.db.WithContext(ctx).Create(invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrTokenAlreadyExists
		}
		return fmt.Errorf("erro ao criar convite interno: %v", err)
	}
	return nil
}

// GetInternalByID finds an internal invitation by ID
func (r *invitationRepository) GetInternalByID(ctx context.Context, id string) (*InternalInvitation, error) {
	var invitation InternalInvitation
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, fmt.Errorf("erro ao buscar convite interno: %v", err)
	}
	return &invitation, nil
}

// GetInternalByToken finds an internal invitation by token
func (r *invitationRepository) GetInternalByToken(ctx context.Context, token string) (*InternalInvitation, error) {
	var invitation InternalInvitation
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, fmt.Errorf("erro ao buscar convite interno por token: %v", err)
	}

	// Check if expired
	if time.Now().After(invitation.ExpiresAt) && invitation.Status == StatusPending {
		return nil, ErrInvitationExpired
	}

	// Check if status is valid (must be pending)
	if invitation.Status != StatusPending {
		return nil, ErrInvalidInvitationStatus
	}

	return &invitation, nil
}

// UpdateInternal updates an existing internal invitation
func (r *invitationRepository) UpdateInternal(ctx context.Context, invitation *InternalInvitation) error {
	if err := r.db.WithContext(ctx).Save(invitation).Error; err != nil {
		return fmt.Errorf("erro ao atualizar convite interno: %v", err)
	}
	return nil
}

// UpdateInternalStatus updates only the status of an internal invitation
func (r *invitationRepository) UpdateInternalStatus(ctx context.Context, id string, status InvitationStatus) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}

	// If accepting, set accepted_at
	if status == StatusAccepted {
		updates["accepted_at"] = now
	}

	if err := r.db.WithContext(ctx).
		Model(&InternalInvitation{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("erro ao atualizar status do convite interno: %v", err)
	}
	return nil
}

// DeleteInternal deletes an internal invitation by ID
func (r *invitationRepository) DeleteInternal(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&InternalInvitation{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("erro ao deletar convite interno: %v", err)
	}
	return nil
}

// ListInternals lists internal invitations with filtering and pagination
func (r *invitationRepository) ListInternals(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error) {
	var invitations []InternalInvitation
	var total int64

	query := r.db.WithContext(ctx).Model(&InternalInvitation{})

	// Apply filters
	if filters != nil {
		if filters.Email != nil && *filters.Email != "" {
			query = query.Where("invitee_email ILIKE ?", "%"+*filters.Email+"%")
		}
		if filters.TenantID != nil {
			query = query.Where("tenant_id = ?", *filters.TenantID)
		}
		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
		}
		if filters.InvitedBy != nil {
			query = query.Where("invited_by = ?", *filters.InvitedBy)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("erro ao contar convites internos: %v", err)
	}

	// Pagination
	page, limit := getPaginationParams(filters)
	offset := (page - 1) * limit

	// Fetch invitations
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("erro ao listar convites internos: %v", err)
	}

	// Convert to generic
	genericInvitations := make([]GenericInvitation, len(invitations))
	for i, inv := range invitations {
		genericInvitations[i] = *inv.ToGeneric()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &PaginatedInvitationResult{
		Data:       genericInvitations,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// ===========================================================================
// GENERIC METHODS (searches both tables)
// ===========================================================================

// GetByToken searches for an invitation by token in both tables
func (r *invitationRepository) GetByToken(ctx context.Context, token string) (*GenericInvitation, error) {
	// Try partner invitation first
	partnerInv, err := r.GetPartnerByToken(ctx, token)
	if err == nil {
		return partnerInv.ToGeneric(), nil
	}
	if err != ErrInvitationNotFound {
		// Some other error occurred
		return nil, err
	}

	// Try internal invitation
	internalInv, err := r.GetInternalByToken(ctx, token)
	if err == nil {
		return internalInv.ToGeneric(), nil
	}
	if err != ErrInvitationNotFound {
		// Some other error occurred
		return nil, err
	}

	// Not found in either table
	return nil, ErrInvitationNotFound
}

// ListAll lists all invitations (both partner and internal) with filtering and pagination
func (r *invitationRepository) ListAll(ctx context.Context, filters *InvitationFilterParams) (*PaginatedInvitationResult, error) {
	// If type filter is specified, route to specific method
	if filters != nil && filters.Type != nil {
		if *filters.Type == TypePartner {
			return r.ListPartners(ctx, filters)
		}
		if *filters.Type == TypeInternal {
			return r.ListInternals(ctx, filters)
		}
	}

	// Otherwise, combine both
	partnerResult, err := r.ListPartners(ctx, filters)
	if err != nil {
		return nil, err
	}

	internalResult, err := r.ListInternals(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Combine results (simple approach - merge and re-paginate)
	combined := append(partnerResult.Data, internalResult.Data...)
	total := partnerResult.Total + internalResult.Total

	page, limit := getPaginationParams(filters)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &PaginatedInvitationResult{
		Data:       combined,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// ===========================================================================
// HELPER FUNCTIONS
// ===========================================================================

// getPaginationParams extracts and validates pagination parameters
func getPaginationParams(filters *InvitationFilterParams) (page int, limit int) {
	page = 1
	limit = 10

	if filters != nil {
		if filters.Page > 0 {
			page = filters.Page
		}
		if filters.Limit > 0 {
			limit = filters.Limit
		}
		if limit > 100 {
			limit = 100
		}
	}

	return page, limit
}
