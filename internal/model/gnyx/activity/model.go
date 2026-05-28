// Package activity defines the data structures and types for managing activities
package activity

import "github.com/kubex-ecosystem/domus/internal/model/gnyx"

// ActivityType enumerates supported activity records.
type ActivityType string

const (
	ActivityTypeCall     ActivityType = "call"
	ActivityTypeEmail    ActivityType = "email"
	ActivityTypeMeeting  ActivityType = "meeting"
	ActivityTypeTask     ActivityType = "task"
	ActivityTypeFollowUp ActivityType = "follow_up"
)

// ActivityStatus captura o estado de workflow das atividades.
type ActivityStatus string

const (
	ActivityStatusPending   ActivityStatus = "pending"
	ActivityStatusCompleted ActivityStatus = "completed"
	ActivityStatusCancelled ActivityStatus = "cancelled"
	ActivityStatusPostponed ActivityStatus = "postponed"
)

// ActivityPriority ranks urgency for tasks.
type ActivityPriority string

const (
	ActivityPriorityLow    ActivityPriority = "low"
	ActivityPriorityNormal ActivityPriority = "normal"
	ActivityPriorityHigh   ActivityPriority = "high"
	ActivityPriorityUrgent ActivityPriority = "urgent"
)

// ParticipantRole defines involvement inside an activity.
type ParticipantRole string

const (
	ParticipantRoleOrganizer   ParticipantRole = "organizer"
	ParticipantRoleParticipant ParticipantRole = "participant"
	ParticipantRoleOptional    ParticipantRole = "optional"
)

// ParticipantStatus mirrors RSVP state.
type ParticipantStatus string

const (
	ParticipantStatusPending  ParticipantStatus = "pending"
	ParticipantStatusAccepted ParticipantStatus = "accepted"
	ParticipantStatusDeclined ParticipantStatus = "declined"
)

// Activity representa a tabela activities.
type Activity struct {
	ID               gnyx.UUID         `json:"id" db:"id"`
	Title            string            `json:"title" db:"title"`
	Description      *string           `json:"description,omitempty" db:"description"`
	Type             ActivityType      `json:"type" db:"type"`
	LeadID           *gnyx.UUID        `json:"lead_id,omitempty" db:"lead_id"`
	AssignedTo       gnyx.UUID         `json:"assigned_to" db:"assigned_to"`
	CreatedBy        gnyx.UUID         `json:"created_by" db:"created_by"`
	CompanyID        gnyx.UUID         `json:"company_id" db:"company_id"`
	StartTime        gnyx.Timestamp    `json:"start_time" db:"start_time"`
	EndTime          *gnyx.Timestamp   `json:"end_time,omitempty" db:"end_time"`
	DurationMinutes  *int64            `json:"duration_minutes,omitempty" db:"duration_minutes"`
	Status           *ActivityStatus   `json:"status,omitempty" db:"status"`
	Priority         *ActivityPriority `json:"priority,omitempty" db:"priority"`
	Location         *string           `json:"location,omitempty" db:"location"`
	Notes            *string           `json:"notes,omitempty" db:"notes"`
	ReminderMinutes  []int64           `json:"reminder_minutes,omitempty" db:"reminder_minutes"`
	IsRecurring      *bool             `json:"is_recurring,omitempty" db:"is_recurring"`
	RecurrenceRule   gnyx.JSONValue    `json:"recurrence_rule,omitempty" db:"recurrence_rule"`
	ParentActivityID *gnyx.UUID        `json:"parent_activity_id,omitempty" db:"parent_activity_id"`
	CompletedAt      *gnyx.Timestamp   `json:"completed_at,omitempty" db:"completed_at"`
	PostponedFrom    *gnyx.Timestamp   `json:"postponed_from,omitempty" db:"postponed_from"`
	PostponedReason  *string           `json:"postponed_reason,omitempty" db:"postponed_reason"`
	CreatedAt        *gnyx.Timestamp   `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt        *gnyx.Timestamp   `json:"updated_at,omitempty" db:"updated_at"`
}

// ActivityParticipant representa activity_participants.
type ActivityParticipant struct {
	ID            gnyx.UUID          `json:"id" db:"id"`
	ActivityID    *gnyx.UUID         `json:"activity_id,omitempty" db:"activity_id"`
	ParticipantID *gnyx.UUID         `json:"participant_id,omitempty" db:"participant_id"`
	Role          *ParticipantRole   `json:"role,omitempty" db:"role"`
	Status        *ParticipantStatus `json:"status,omitempty" db:"status"`
	CreatedAt     *gnyx.Timestamp    `json:"created_at,omitempty" db:"created_at"`
}
