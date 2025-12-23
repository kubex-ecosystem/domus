// Package systemlogs provides basic data modeling management tools
package systemlogs

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*SystemLogs, error)
	Create(systemLog *SystemLogs) error
	Update(systemLog *SystemLogs) error
	Delete(id string) error
}

type SystemLogsRepository[T SystemLogs] struct {
	db *gorm.DB
}

func NewRepository[T SystemLogs](db *gorm.DB) ORMRepository[T] {
	return &SystemLogsRepository[T]{db: db}
}

func (r *SystemLogsRepository[T]) GetAll() ([]T, error) {
	var systemLogs []T
	if err := r.db.Find(&systemLogs).Error; err != nil {
		return nil, err
	}
	return systemLogs, nil
}

func (r *SystemLogsRepository[T]) GetByID(id string) (*SystemLogs, error) {
	var systemLog SystemLogs
	if err := r.db.First(&systemLog, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &systemLog, nil
}

func (r *SystemLogsRepository[T]) Create(systemLog *SystemLogs) error {
	if err := r.db.Create(systemLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *SystemLogsRepository[T]) Update(systemLog *SystemLogs) error {
	if err := r.db.Model(systemLog).Omit("ID").Updates(systemLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *SystemLogsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&SystemLogs{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
