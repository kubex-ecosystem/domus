// Package backupstatus provides basic data modeling management tools
package backupstatus

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*BackupStatus, error)
	Create(backupStatus *BackupStatus) error
	Update(backupStatus *BackupStatus) error
	Delete(id string) error
}

type BackupStatusRepository[T BackupStatus] struct {
	db *gorm.DB
}

func NewRepository[T BackupStatus](db *gorm.DB) ORMRepository[T] {
	return &BackupStatusRepository[T]{db: db}
}

func (r *BackupStatusRepository[T]) GetAll() ([]T, error) {
	var backupStatuses []T
	if err := r.db.Find(&backupStatuses).Error; err != nil {
		return nil, err
	}
	return backupStatuses, nil
}

func (r *BackupStatusRepository[T]) GetByID(id string) (*BackupStatus, error) {
	var backupStatus BackupStatus
	if err := r.db.First(&backupStatus, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &backupStatus, nil
}

func (r *BackupStatusRepository[T]) Create(backupStatus *BackupStatus) error {
	if err := r.db.Create(backupStatus).Error; err != nil {
		return err
	}
	return nil
}

func (r *BackupStatusRepository[T]) Update(backupStatus *BackupStatus) error {
	if err := r.db.Model(backupStatus).Omit("ID").Updates(backupStatus).Error; err != nil {
		return err
	}
	return nil
}

func (r *BackupStatusRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&BackupStatus{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
