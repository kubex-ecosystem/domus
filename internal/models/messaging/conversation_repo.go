package messaging

// import (
// 	"context"
// 	"fmt"

// 	svc "github.com/kubex-ecosystem/domus/internal/services"
// 	gl "github.com/kubex-ecosystem/logz"

// 	"gorm.io/gorm"
// )

// type IConversationRepo = datastore.Repository[ConversationModel]
// 	Create(ctx context.Context, conversation *ConversationModel) (*ConversationModel, error)
// 	Update(ctx context.Context, conversation *ConversationModel) (*ConversationModel, error)
// 	Delete(ctx context.Context, id string) error
// 	FindByID(ctx context.Context, id string) (*ConversationModel, error)
// 	FindByPlatformAndConversationID(ctx context.Context, platform Platform, platformConversationID string) (*ConversationModel, error)
// 	FindByIntegrationID(ctx context.Context, integrationID string) ([]*ConversationModel, error)
// 	FindByPlatform(ctx context.Context, platform Platform) ([]*ConversationModel, error)
// 	FindByStatus(ctx context.Context, status ConversationStatus) ([]*ConversationModel, error)
// 	FindByConversationType(ctx context.Context, conversationType ConversationType) ([]*ConversationModel, error)
// 	FindByUserID(ctx context.Context, userID string) ([]*ConversationModel, error)
// 	FindActiveConversations(ctx context.Context) ([]*ConversationModel, error)
// 	FindRecentConversations(ctx context.Context, limit int) ([]*ConversationModel, error)
// 	FindAll(ctx context.Context) ([]*ConversationModel, error)
// 	Count(ctx context.Context) (int64, error)
// 	UpdateStatus(ctx context.Context, id string, status ConversationStatus) error
// 	UpdateLastMessage(ctx context.Context, id string, messageID string) error
// 	IncrementMessageCount(ctx context.Context, id string) error
// 	FindByTargetTaskID(ctx context.Context, targetTaskID string) ([]*ConversationModel, error)
// }

// type ConversationRepository[T Conversation] struct {
// 	db *gorm.DB
// }

// func NewConversationRepository(ctx context.Context, dbService *svc.DBServiceImpl) IConversationRepo {
// 	if dbService == nil {
// 		logz.Log("error", "Conversation repository: dbService is nil")
// 		return nil
// 	}
// 	db, err := dbService.GetDB(ctx)
// 	if err != nil {
// 		logz.Log("error", fmt.Sprintf("Conversation repository: failed to get DB from dbService: %v", err))
// 		return nil
// 	}
// 	return &ConversationRepository[Conversation]{db: db}
// }

// func (r *ConversationRepository[T]) Create(ctx context.Context, conversation *ConversationModel) (*ConversationModel, error) {
// 	if err := conversation.Validate(); err != nil {
// 		logz.Log("error", "Conversation validation failed", err)
// 		return nil, fmt.Errorf("validation failed: %v", err)
// 	}

// 	conversation.Sanitize()

// 	if err := r.db.WithContext(ctx).Create(conversation).Error; err != nil {
// 		logz.Log("error", "Failed to create conversation", err)
// 		return nil, fmt.Errorf("failed to create conversation: %v", err)
// 	}

// 	logz.Log("info", "Conversation created successfully", conversation.ID)
// 	return conversation, nil
// }

// func (r *ConversationRepository[T]) Update(ctx context.Context, conversation *ConversationModel) (*ConversationModel, error) {
// 	if err := conversation.Validate(); err != nil {
// 		logz.Log("error", "Conversation validation failed", err)
// 		return nil, fmt.Errorf("validation failed: %v", err)
// 	}

// 	conversation.Sanitize()

// 	if err := r.db.WithContext(ctx).Save(conversation).Error; err != nil {
// 		logz.Log("error", "Failed to update conversation", err)
// 		return nil, fmt.Errorf("failed to update conversation: %v", err)
// 	}

// 	logz.Log("info", "Conversation updated successfully", conversation.ID)
// 	return conversation, nil
// }

// func (r *ConversationRepository[T]) Delete(ctx context.Context, id string) error {
// 	result := r.db.WithContext(ctx).Delete(&ConversationModel{}, "id = ?", id)
// 	if result.Error != nil {
// 		logz.Log("error", "Failed to delete conversation", result.Error)
// 		return fmt.Errorf("failed to delete conversation: %v", result.Error)
// 	}

// 	if result.RowsAffected == 0 {
// 		logz.Log("warning", "Conversation not found for deletion", id)
// 		return fmt.Errorf("conversation not found")
// 	}

// 	logz.Log("info", "Conversation deleted successfully", id)
// 	return nil
// }

// func (r *ConversationRepository[T]) FindByID(ctx context.Context, id string) (*ConversationModel, error) {
// 	var conversation ConversationModel
// 	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&conversation).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, fmt.Errorf("conversation not found")
// 		}
// 		logz.Log("error", "Failed to find conversation by ID", err)
// 		return nil, fmt.Errorf("failed to find conversation: %v", err)
// 	}

// 	return &conversation, nil
// }

// func (r *ConversationRepository[T]) FindByPlatformAndConversationID(ctx context.Context, platform Platform, platformConversationID string) (*ConversationModel, error) {
// 	var conversation ConversationModel
// 	if err := r.db.WithContext(ctx).Where("platform = ? AND platform_conversation_id = ?", platform, platformConversationID).First(&conversation).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, fmt.Errorf("conversation not found for platform: %s, conversation ID: %s", platform, platformConversationID)
// 		}
// 		logz.Log("error", "Failed to find conversation by platform and conversation ID", err)
// 		return nil, fmt.Errorf("failed to find conversation: %v", err)
// 	}

// 	return &conversation, nil
// }

// func (r *ConversationRepository[T]) FindByIntegrationID(ctx context.Context, integrationID string) ([]*ConversationModel, error) {
// 	var conversations []*ConversationModel
// 	if err := r.db.WithContext(ctx).Where("integration_id = ?", integrationID).Find(&conversations).Error; err != nil {
// 		logz.Log("error", "Failed to find conversations by integration ID", err)
// 		return nil, fmt.Errorf("failed to find conversations: %v", err)
// 	}

// 	return conversations, nil
// }

// func (r *ConversationRepository[T]) FindByPlatform(ctx context.Context, platform Platform) ([]*ConversationModel, error) {
// 	var conversations []*ConversationModel
// 	if err := r.db.WithContext(ctx).Where("platform = ?", platform).Find(&conversations).Error; err != nil {
// 		logz.Log("error", "Failed to find conversations by platform", err)
// 		return nil, fmt.Errorf("failed to find conversations: %v", err)
// 	}

// 	return conversations, nil
// }

// func (r *ConversationRepository[T]) FindByStatus(ctx context.Context, status ConversationStatus) ([]*ConversationModel, error) {
// 	var conversations []*ConversationModel
// 	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&conversations).Error; err != nil {
// 		logz.Log("error", "Failed to find conversations by status", err)
// 		return nil, fmt.Errorf("failed to find conversations: %v", err)
// 	}

// 	return conversations, nil
// }

// func (r *ConversationRepository[T]) FindByConversationType(ctx context.Context, conversationType ConversationType) ([]*ConversationModel, error) {
// 	var conversations []*ConversationModel
// 	if err := r.db.WithContext(ctx).Where("conversation_type = ?", conversationType).Find(&conversations).Error; err != nil {
// 		logz.Log("error", "Failed to find conversations by conversation type", err)
// 		return nil, fmt.Errorf("failed to find conversations: %v", err)
// 	}

// 	return conversations, nil
// }

// func (r *ConversationRepository[T]) FindByUserID(ctx context.Context, userID string) ([]*ConversationModel, error) {
// 	var conversations []*ConversationModel
// 	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&conversations).Error; err != nil {
// 		logz.Log("error", "Failed to find conversations by user ID", err)
// 		return nil, fmt.Errorf("failed to find conversations: %v", err)
// 	}

// 	return conversations, nil
// }

// func (r *ConversationRepository[T]) FindActiveConversations(ctx context.Context) ([]*ConversationModel, error) {
// 	var conversations []*ConversationModel
// 	if err := r.db.WithContext(ctx).Where("status = ?", ConversationStatusActive).Find(&conversations).Error; err != nil {
// 		logz.Log("error", "Failed to find active conversations", err)
// 		return nil, fmt.Errorf("failed to find active conversations: %v", err)
// 	}

// 	return conversations, nil
// }

// func (r *ConversationRepository[T]) FindRecentConversations(ctx context.Context, limit int) ([]*ConversationModel, error) {
// 	var conversations []*ConversationModel
// 	query := r.db.WithContext(ctx).Order("last_message_at DESC")
// 	if limit > 0 {
// 		query = query.Limit(limit)
// 	}

// 	if err := query.Find(&conversations).Error; err != nil {
// 		logz.Log("error", "Failed to find recent conversations", err)
// 		return nil, fmt.Errorf("failed to find recent conversations: %v", err)
// 	}

// 	return conversations, nil
// }

// func (r *ConversationRepository[T]) FindAll(ctx context.Context) ([]*ConversationModel, error) {
// 	var conversations []*ConversationModel
// 	if err := r.db.WithContext(ctx).Find(&conversations).Error; err != nil {
// 		logz.Log("error", "Failed to find all conversations", err)
// 		return nil, fmt.Errorf("failed to find conversations: %v", err)
// 	}

// 	return conversations, nil
// }

// func (r *ConversationRepository[T]) Count(ctx context.Context) (int64, error) {
// 	var count int64
// 	if err := r.db.WithContext(ctx).Model(&ConversationModel{}).Count(&count).Error; err != nil {
// 		logz.Log("error", "Failed to count conversations", err)
// 		return 0, fmt.Errorf("failed to count conversations: %v", err)
// 	}

// 	return count, nil
// }

// func (r *ConversationRepository[T]) UpdateStatus(ctx context.Context, id string, status ConversationStatus) error {
// 	result := r.db.WithContext(ctx).Model(&ConversationModel{}).Where("id = ?", id).Update("status", status)
// 	if result.Error != nil {
// 		logz.Log("error", "Failed to update conversation status", result.Error)
// 		return fmt.Errorf("failed to update conversation status: %v", result.Error)
// 	}

// 	if result.RowsAffected == 0 {
// 		return fmt.Errorf("conversation not found")
// 	}

// 	logz.Log("info", "Conversation status updated", id, status)
// 	return nil
// }

// func (r *ConversationRepository[T]) UpdateLastMessage(ctx context.Context, id string, messageID string) error {
// 	updates := map[string]interface{}{
// 		"last_message_id": messageID,
// 		"last_message_at": "NOW()",
// 		"updated_at":      "NOW()",
// 	}

// 	result := r.db.WithContext(ctx).Model(&ConversationModel{}).Where("id = ?", id).Updates(updates)
// 	if result.Error != nil {
// 		logz.Log("error", "Failed to update conversation last message", result.Error)
// 		return fmt.Errorf("failed to update conversation last message: %v", result.Error)
// 	}

// 	if result.RowsAffected == 0 {
// 		return fmt.Errorf("conversation not found")
// 	}

// 	return nil
// }

// func (r *ConversationRepository[T]) IncrementMessageCount(ctx context.Context, id string) error {
// 	result := r.db.WithContext(ctx).Model(&ConversationModel{}).Where("id = ?", id).
// 		Updates(map[string]interface{}{
// 			"message_count": gorm.Expr("message_count + 1"),
// 			"updated_at":    "NOW()",
// 		})

// 	if result.Error != nil {
// 		logz.Log("error", "Failed to increment conversation message count", result.Error)
// 		return fmt.Errorf("failed to increment conversation message count: %v", result.Error)
// 	}

// 	if result.RowsAffected == 0 {
// 		return fmt.Errorf("conversation not found")
// 	}

// 	return nil
// }

// func (r *ConversationRepository[T]) FindByTargetTaskID(ctx context.Context, targetTaskID string) ([]*ConversationModel, error) {
// 	var conversations []*ConversationModel
// 	if err := r.db.WithContext(ctx).Where("target_task_id = ?", targetTaskID).Find(&conversations).Error; err != nil {
// 		logz.Log("error", "Failed to find conversations by target task ID", err)
// 		return nil, fmt.Errorf("failed to find conversations: %v", err)
// 	}

// 	return conversations, nil
// }
