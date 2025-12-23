package client

import (
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/accesslogs"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/activities"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/activityparticipants"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/addresses"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/auditlogs"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/backupstatus"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/companies"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/companymetrics"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/errorlogs"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/lossreasonsanalytics"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/paymentmethods"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/permissions"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/pipelines"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/pipelinestages"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/profiles"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/refreshtokens"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/roleconfig"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/rolepermissions"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/roles"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/subscriptionplans"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/systemlogs"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/systemmetrics"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/tenants"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/tenantsubscriptions"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/trainingbadges"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/trainingcourses"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/trainingcoursestats"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/traininglessons"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/trainingprogress"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/userinvitations"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/userpreferences"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/userprofiles"
	"github.com/kubex-ecosystem/domus/internal/models/gnyx/users"
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
