// Package accesslogs provides basic data modeling management tools
package accesslogs

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*AccessLogs, error)
	Create(accessLog *AccessLogs) error
	Update(accessLog *AccessLogs) error
	Delete(id string) error
}

type AccessLogsRepository[T AccessLogs] struct {
	db *gorm.DB
}

func NewRepository[T AccessLogs](db *gorm.DB) ORMRepository[T] {
	return &AccessLogsRepository[T]{db: db}
}

func (r *AccessLogsRepository[T]) GetAll() ([]T, error) {
	var accessLogs []T
	if err := r.db.Find(&accessLogs).Error; err != nil {
		return nil, err
	}
	return accessLogs, nil
}

func (r *AccessLogsRepository[T]) GetByID(id string) (*AccessLogs, error) {
	var accessLog AccessLogs
	if err := r.db.First(&accessLog, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &accessLog, nil
}

func (r *AccessLogsRepository[T]) Create(accessLog *AccessLogs) error {
	if err := r.db.Create(accessLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *AccessLogsRepository[T]) Update(accessLog *AccessLogs) error {
	if err := r.db.Model(accessLog).Omit("ID").Updates(accessLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *AccessLogsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&AccessLogs{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
