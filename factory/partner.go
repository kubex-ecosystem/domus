package factory

// import (
// 	prm "github.com/kubex-ecosystem/domus/internal/models/gnyx"
// 	"gorm.io/gorm"
// )

// ==========================================
// Partner Service Public Aliases
// ==========================================

// IPartnerService is a public alias for the internal interface.
// type IPartnerService = prm.IPartnerService

// ==========================================
// Repository Public Aliases
// ==========================================

// IPartnerRepo is a public alias for the internal interface.
// type IPartnerRepo = prm.IPartnerRepo

// ==========================================
// Model Public Aliases
// ==========================================

// Partner is a public alias for the internal model.
// // type Partner = prm.Partner

// // CreatePartnerDTO is a public alias for the internal DTO.
// type CreatePartnerDTO = prm.CreatePartnerDTO

// // UpdatePartnerDTO is a public alias for the internal DTO.
// type UpdatePartnerDTO = prm.UpdatePartnerDTO

// // PartnerFilterParams is a public alias for the internal filter params.
// type PartnerFilterParams = prm.PartnerFilterParams

// // PaginatedPartnerResult is a public alias for the internal paginated result.
// type PaginatedPartnerResult = prm.PaginatedPartnerResult

// // PartnerStatus is a public alias for the internal status type.
// type PartnerStatus = prm.PartnerStatus

// // Partner status constants
// const (
// 	PartnerStatusActive   = prm.PartnerStatusActive
// 	PartnerStatusInactive = prm.PartnerStatusInactive
// 	PartnerStatusBlocked  = prm.PartnerStatusBlocked
// 	PartnerStatusPending  = prm.PartnerStatusPending
// )

// // PartnerRole is a public alias for the internal role type.
// type PartnerRole = prm.PartnerRole

// // Partner role constants
// const (
// 	RolePartner = prm.RolePartner
// 	RoleAdmin   = prm.RoleAdmin
// 	RoleManager = prm.RoleManager
// 	RoleUser    = prm.RoleUser
// )

// // ==========================================
// // Public Constructors
// // ==========================================

// // NewPartnerService is a public constructor for the partner service.
// func NewPartnerService(db *gorm.DB) IPartnerService {
// 	partnerRepo := prm.NewPartnerRepo(db)
// 	return prm.NewPartnerService(partnerRepo)
// }

// // NewPartnerRepo is a public constructor for the partner repository.
// func NewPartnerRepo(db *gorm.DB) IPartnerRepo {
// 	return prm.NewPartnerRepo(db)
// }
