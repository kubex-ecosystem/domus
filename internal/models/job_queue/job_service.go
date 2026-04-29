// Package jobqueue provides the service for managing job queues in the MCP system.
package jobqueue

// import (
// 	"context"
// 	"errors"
// 	"time"

// 	"github.com/google/uuid"
// )

// type IJobQueueService interface {
// 	CreateJob(ctx context.Context, job *JobQueue) (*JobQueue, error)
// 	GetJobByID(ctx context.Context, id uuid.UUID) (*JobQueue, error)
// 	ListJobs(ctx context.Context) ([]*JobQueue, error)
// 	UpdateJob(ctx context.Context, job *JobQueue) (*JobQueue, error)
// 	DeleteJob(ctx context.Context, id uuid.UUID) error
// 	ListJobsByStatus(ctx context.Context, status string) ([]*JobQueue, error)
// 	ListJobsByUserID(ctx context.Context, userID uuid.UUID) ([]*JobQueue, error)
// 	ExecuteJobManually(ctx context.Context, id uuid.UUID) error
// 	ValidateJobSchedule(ctx context.Context, schedule string) error
// 	ListJobsByType(ctx context.Context, jobType string) ([]*JobQueue, error)
// 	ListJobsByCreatedAt(ctx context.Context, createdAt time.Time) ([]*JobQueue, error)
// }
// type JobQueueService struct{ Repo IJobQueueRepo }

// func NewJobQueueService(repo IJobQueueRepo) IJobQueueService { return &JobQueueService{Repo: repo} }

// func (s *JobQueueService[T]) CreateJob(ctx context.Context, job *JobQueue) (*JobQueue, error) {
// 	if job == nil {
// 		return nil, errors.New("job cannot be nil")
// 	}
// 	if job.ID == uuid.Nil {
// 		return nil, errors.New("job ID cannot be empty")
// 	}
// 	if job.UserID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	if job.Status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	return s.Repo.Create(ctx, job)
// }
// func (s *JobQueueService[T]) GetJobByID(ctx context.Context, id uuid.UUID) (*JobQueue, error) {
// 	if id == uuid.Nil {
// 		return nil, errors.New("job ID cannot be empty")
// 	}
// 	return s.Repo.FindByID(ctx, id)
// }
// func (s *JobQueueService[T]) ListJobs(ctx context.Context) ([]*JobQueue, error) {
// 	return s.Repo.FindAll(ctx)
// }

// func (s *JobQueueService[T]) UpdateJob(ctx context.Context, job *JobQueue) (*JobQueue, error) {
// 	if job == nil {
// 		return nil, errors.New("job cannot be nil")
// 	}
// 	if job.ID == uuid.Nil {
// 		return nil, errors.New("job ID cannot be empty")
// 	}
// 	return s.Repo.Update(ctx, job)
// }
// func (s *JobQueueService[T]) DeleteJob(ctx context.Context, id uuid.UUID) error {
// 	if id == uuid.Nil {
// 		return errors.New("job ID cannot be empty")
// 	}
// 	return s.Repo.Delete(ctx, id)
// }
// func (s *JobQueueService[T]) ListJobsByStatus(ctx context.Context, status string) ([]*JobQueue, error) {
// 	if status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	return s.Repo.FindByStatus(ctx, status)
// }
// func (s *JobQueueService[T]) ListJobsByUserID(ctx context.Context, userID uuid.UUID) ([]*JobQueue, error) {
// 	if userID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	return s.Repo.FindByUserID(ctx, userID)
// }
// func (s *JobQueueService[T]) ExecuteJobManually(ctx context.Context, id uuid.UUID) error {
// 	if id == uuid.Nil {
// 		return errors.New("job ID cannot be empty")
// 	}
// 	return s.Repo.ExecuteJobManually(ctx, id)
// }
// func (s *JobQueueService[T]) RescheduleJob(ctx context.Context, id uuid.UUID, newSchedule time.Time) error {
// 	if id == uuid.Nil {
// 		return errors.New("job ID cannot be empty")
// 	}
// 	return s.Repo.RescheduleJob(ctx, id, newSchedule)
// }
// func (s *JobQueueService[T]) ValidateJobSchedule(ctx context.Context, schedule string) error {
// 	if schedule == "" {
// 		return errors.New("schedule cannot be empty")
// 	}
// 	// Add your validation logic here
// 	return nil
// }
// func (s *JobQueueService[T]) ListJobsByType(ctx context.Context, jobType string) ([]*JobQueue, error) {
// 	if jobType == "" {
// 		return nil, errors.New("job type cannot be empty")
// 	}
// 	return s.Repo.FindByType(ctx, jobType)
// }
// func (s *JobQueueService[T]) ListJobsByCreatedAt(ctx context.Context, createdAt time.Time) ([]*JobQueue, error) {
// 	if createdAt.IsZero() {
// 		return nil, errors.New("created at cannot be empty")
// 	}
// 	return s.Repo.FindByCreatedAt(ctx, createdAt)
// }
// func (s *JobQueueService[T]) ListJobsByCreatedBy(ctx context.Context, createdBy uuid.UUID) ([]*JobQueue, error) {
// 	if createdBy == uuid.Nil {
// 		return nil, errors.New("created by cannot be empty")
// 	}
// 	return s.Repo.FindByCreatedBy(ctx, createdBy)
// }
// func (s *JobQueueService[T]) ListJobsByUpdatedAt(ctx context.Context, updatedAt time.Time) ([]*JobQueue, error) {
// 	if updatedAt.IsZero() {
// 		return nil, errors.New("updated at cannot be empty")
// 	}
// 	return s.Repo.FindByUpdatedAt(ctx, updatedAt)
// }
// func (s *JobQueueService[T]) ListJobsByUpdatedBy(ctx context.Context, updatedBy uuid.UUID) ([]*JobQueue, error) {
// 	if updatedBy == uuid.Nil {
// 		return nil, errors.New("updated by cannot be empty")
// 	}
// 	return s.Repo.FindByUpdatedBy(ctx, updatedBy)
// }
// func (s *JobQueueService[T]) ListJobsByLastExecutedAt(ctx context.Context, lastExecutedAt time.Time) ([]*JobQueue, error) {
// 	if lastExecutedAt.IsZero() {
// 		return nil, errors.New("last executed at cannot be empty")
// 	}
// 	return s.Repo.FindByLastExecutedAt(ctx, lastExecutedAt)
// }
// func (s *JobQueueService[T]) ListJobsByLastExecutedBy(ctx context.Context, lastExecutedBy uuid.UUID) ([]*JobQueue, error) {
// 	if lastExecutedBy == uuid.Nil {
// 		return nil, errors.New("last executed by cannot be empty")
// 	}
// 	return s.Repo.FindByLastExecutedBy(ctx, lastExecutedBy)
// }
// func (s *JobQueueService[T]) ListJobsByStatusAndUserID(ctx context.Context, status string, userID uuid.UUID) ([]*JobQueue, error) {
// 	if status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	if userID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	return s.Repo.FindByStatusAndUserID(ctx, status, userID)
// }
// func (s *JobQueueService[T]) ListJobsByStatusAndType(ctx context.Context, status string, jobType string) ([]*JobQueue, error) {
// 	if status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	if jobType == "" {
// 		return nil, errors.New("job type cannot be empty")
// 	}
// 	return s.Repo.FindByStatusAndType(ctx, status, jobType)
// }
// func (s *JobQueueService[T]) ListJobsByStatusAndCreatedAt(ctx context.Context, status string, createdAt time.Time) ([]*JobQueue, error) {
// 	if status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	if createdAt.IsZero() {
// 		return nil, errors.New("created at cannot be empty")
// 	}
// 	return s.Repo.FindByStatusAndCreatedAt(ctx, status, createdAt)
// }
// func (s *JobQueueService[T]) ListJobsByStatusAndCreatedBy(ctx context.Context, status string, createdBy uuid.UUID) ([]*JobQueue, error) {
// 	if status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	if createdBy == uuid.Nil {
// 		return nil, errors.New("created by cannot be empty")
// 	}
// 	return s.Repo.FindByStatusAndCreatedBy(ctx, status, createdBy)
// }
// func (s *JobQueueService[T]) ListJobsByStatusAndUpdatedAt(ctx context.Context, status string, updatedAt time.Time) ([]*JobQueue, error) {
// 	if status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	if updatedAt.IsZero() {
// 		return nil, errors.New("updated at cannot be empty")
// 	}
// 	return s.Repo.FindByStatusAndUpdatedAt(ctx, status, updatedAt)
// }
// func (s *JobQueueService[T]) ListJobsByStatusAndUpdatedBy(ctx context.Context, status string, updatedBy uuid.UUID) ([]*JobQueue, error) {
// 	if status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	if updatedBy == uuid.Nil {
// 		return nil, errors.New("updated by cannot be empty")
// 	}
// 	return s.Repo.FindByStatusAndUpdatedBy(ctx, status, updatedBy)
// }
// func (s *JobQueueService[T]) ListJobsByStatusAndLastExecutedAt(ctx context.Context, status string, lastExecutedAt time.Time) ([]*JobQueue, error) {
// 	if status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	if lastExecutedAt.IsZero() {
// 		return nil, errors.New("last executed at cannot be empty")
// 	}
// 	return s.Repo.FindByStatusAndLastExecutedAt(ctx, status, lastExecutedAt)
// }
// func (s *JobQueueService[T]) ListJobsByStatusAndLastExecutedBy(ctx context.Context, status string, lastExecutedBy uuid.UUID) ([]*JobQueue, error) {
// 	if status == "" {
// 		return nil, errors.New("job status cannot be empty")
// 	}
// 	if lastExecutedBy == uuid.Nil {
// 		return nil, errors.New("last executed by cannot be empty")
// 	}
// 	return s.Repo.FindByStatusAndLastExecutedBy(ctx, status, lastExecutedBy)
// }
// func (s *JobQueueService[T]) ListJobsByUserIDAndType(ctx context.Context, userID uuid.UUID, jobType string) ([]*JobQueue, error) {
// 	if userID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	if jobType == "" {
// 		return nil, errors.New("job type cannot be empty")
// 	}
// 	return s.Repo.FindByUserIDAndType(ctx, userID, jobType)
// }
// func (s *JobQueueService[T]) ListJobsByUserIDAndCreatedAt(ctx context.Context, userID uuid.UUID, createdAt time.Time) ([]*JobQueue, error) {
// 	if userID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	if createdAt.IsZero() {
// 		return nil, errors.New("created at cannot be empty")
// 	}
// 	return s.Repo.FindByUserIDAndCreatedAt(ctx, userID, createdAt)
// }
// func (s *JobQueueService[T]) ListJobsByUserIDAndCreatedBy(ctx context.Context, userID uuid.UUID, createdBy uuid.UUID) ([]*JobQueue, error) {
// 	if userID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	if createdBy == uuid.Nil {
// 		return nil, errors.New("created by cannot be empty")
// 	}
// 	return s.Repo.FindByUserIDAndCreatedBy(ctx, userID, createdBy)
// }
// func (s *JobQueueService[T]) ListJobsByUserIDAndUpdatedAt(ctx context.Context, userID uuid.UUID, updatedAt time.Time) ([]*JobQueue, error) {
// 	if userID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	if updatedAt.IsZero() {
// 		return nil, errors.New("updated at cannot be empty")
// 	}
// 	return s.Repo.FindByUserIDAndUpdatedAt(ctx, userID, updatedAt)
// }
// func (s *JobQueueService[T]) ListJobsByUserIDAndUpdatedBy(ctx context.Context, userID uuid.UUID, updatedBy uuid.UUID) ([]*JobQueue, error) {
// 	if userID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	if updatedBy == uuid.Nil {
// 		return nil, errors.New("updated by cannot be empty")
// 	}
// 	return s.Repo.FindByUserIDAndUpdatedBy(ctx, userID, updatedBy)
// }
// func (s *JobQueueService[T]) ListJobsByUserIDAndLastExecutedAt(ctx context.Context, userID uuid.UUID, lastExecutedAt time.Time) ([]*JobQueue, error) {
// 	if userID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	if lastExecutedAt.IsZero() {
// 		return nil, errors.New("last executed at cannot be empty")
// 	}
// 	return s.Repo.FindByUserIDAndLastExecutedAt(ctx, userID, lastExecutedAt)
// }
// func (s *JobQueueService[T]) ListJobsByUserIDAndLastExecutedBy(ctx context.Context, userID uuid.UUID, lastExecutedBy uuid.UUID) ([]*JobQueue, error) {
// 	if userID == uuid.Nil {
// 		return nil, errors.New("user ID cannot be empty")
// 	}
// 	if lastExecutedBy == uuid.Nil {
// 		return nil, errors.New("last executed by cannot be empty")
// 	}
// 	return s.Repo.FindByUserIDAndLastExecutedBy(ctx, userID, lastExecutedBy)
// }
