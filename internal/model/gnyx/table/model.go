// Package table contém definições relacionadas a nomes de tabelas no banco de dados.
package table

// TableName enumera os identificadores de tabela conhecidos.
type TableName string

const (
	TableBackupStatus         TableName = "backup_status"
	TableCommissionRules      TableName = "commission_rules"
	TableCompanies            TableName = "companies"
	TableSystemLogs           TableName = "system_logs"
	TableSystemMetrics        TableName = "system_metrics"
	TableCompanyMetrics       TableName = "company_metrics"
	TableLossReasonsAnalytics TableName = "loss_reasons_analytics"
	TablePipelines            TableName = "pipelines"
	TablePipelineStages       TableName = "pipeline_stages"
	TableActivities           TableName = "activities"
	TableActivityParticipants TableName = "activity_participants"
	TableAuditLogs            TableName = "audit_logs"
	TableCommissions          TableName = "commissions"
	TableLeads                TableName = "leads"
	TableProfiles             TableName = "profiles"
	TableTrainingCourseStats  TableName = "training_course_stats"
	TableTrainingCourses      TableName = "training_courses"
	TableTrainingLessons      TableName = "training_lessons"
	TableTrainingProgress     TableName = "training_progress"
	TableUserInvitations      TableName = "user_invitations"
	TableRealTimeMetrics      TableName = "real_time_metrics"
)
