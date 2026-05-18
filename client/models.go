package client

import (
	"github.com/faelmori/domus-pulse/internal/models/gnyx/accesslogs"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/activities"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/activityparticipants"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/addresses"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/auditlogs"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/backupstatus"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/companies"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/companymetrics"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/errorlogs"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/lossreasonsanalytics"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/paymentmethods"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/permissions"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/pipelines"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/pipelinestages"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/profiles"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/refreshtokens"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/roleconfig"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/rolepermissions"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/roles"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/subscriptionplans"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/sysdashboards"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/sysdatasources"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/systemlogs"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/systemmetrics"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/syswidgets"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/tenants"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/tenantsubscriptions"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/trainingbadges"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/trainingcourses"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/trainingcoursestats"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/traininglessons"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/trainingprogress"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/userinvitations"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/userpreferences"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/userprofiles"
	"github.com/faelmori/domus-pulse/internal/models/gnyx/users"
	"github.com/faelmori/domus-pulse/internal/models/invitations"
)

type SubscriptionPlans = subscriptionplans.SubscriptionPlans
type LossReasonsAnalytics = lossreasonsanalytics.LossReasonsAnalytics
type ActivityParticipants = activityparticipants.ActivityParticipants
type Companies = companies.Companies
type CompanyMetrics = companymetrics.CompanyMetrics
type ErrorLogs = errorlogs.ErrorLogs
type UserInvitations = userinvitations.UserInvitations
type PipelineStages = pipelinestages.PipelineStages
type TrainingBadges = trainingbadges.TrainingBadges
type RolePermissions = rolepermissions.RolePermissions
type TrainingLessons = traininglessons.TrainingLessons
type SystemLogs = systemlogs.SystemLogs
type Pipelines = pipelines.Pipelines
type RoleConfig = roleconfig.RoleConfig
type UserProfiles = userprofiles.UserProfiles
type TrainingProgress = trainingprogress.TrainingProgress
type UserPreferences = userpreferences.UserPreferences
type Users = users.Users
type TrainingCourseStats = trainingcoursestats.TrainingCourseStats
type PaymentMethods = paymentmethods.PaymentMethods
type Profiles = profiles.Profiles
type AccessLogs = accesslogs.AccessLogs
type TrainingCourses = trainingcourses.TrainingCourses
type AuditLogs = auditlogs.AuditLogs
type Addresses = addresses.Addresses
type Roles = roles.Roles
type TenantSubscriptions = tenantsubscriptions.TenantSubscriptions
type RefreshTokens = refreshtokens.RefreshTokens
type Activities = activities.Activities
type BackupStatus = backupstatus.BackupStatus
type Permissions = permissions.Permissions
type Tenants = tenants.Tenants
type SystemMetrics = systemmetrics.SystemMetrics
type SysDashboards = sysdashboards.SysDashboards
type SysWidgets = syswidgets.SysWidgets
type SysDataSources = sysdatasources.SysDataSources

const (
	InvitationTypePartner  = invitations.TypePartner
	InvitationTypeInternal = invitations.TypeInternal
	InvitationTypeUser     = invitations.TypeUser

	InvitationStatusPending  = invitations.StatusPending
	InvitationStatusAccepted = invitations.StatusAccepted
	InvitationStatusRejected = invitations.StatusRejected
	InvitationStatusRevoked  = invitations.StatusRevoked
	InvitationStatusExpired  = invitations.StatusExpired
)
