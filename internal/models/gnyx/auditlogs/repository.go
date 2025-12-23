// Package auditlogs provides basic data modeling management tools
package auditlogs

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*AuditLogs, error)
	Create(auditLog *AuditLogs) error
	Update(auditLog *AuditLogs) error
	Delete(id string) error
}

type AuditLogsRepository[T AuditLogs] struct {
	db *gorm.DB
}

func NewRepository[T AuditLogs](db *gorm.DB) ORMRepository[T] {
	return &AuditLogsRepository[T]{db: db}
}

func (r *AuditLogsRepository[T]) GetAll() ([]T, error) {
	var auditLogs []T
	if err := r.db.Find(&auditLogs).Error; err != nil {
		return nil, err
	}
	return auditLogs, nil
}

func (r *AuditLogsRepository[T]) GetByID(id string) (*AuditLogs, error) {
	var auditLog AuditLogs
	if err := r.db.First(&auditLog, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &auditLog, nil
}

func (r *AuditLogsRepository[T]) Create(auditLog *AuditLogs) error {
	if err := r.db.Create(auditLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *AuditLogsRepository[T]) Update(auditLog *AuditLogs) error {
	if err := r.db.Model(auditLog).Omit("ID").Updates(auditLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *AuditLogsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&AuditLogs{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
