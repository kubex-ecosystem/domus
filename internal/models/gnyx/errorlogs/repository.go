// Package errorlogs provides basic data modeling management tools
package errorlogs

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*ErrorLogs, error)
	Create(errorLog *ErrorLogs) error
	Update(errorLog *ErrorLogs) error
	Delete(id string) error
}

type ErrorLogsRepository[T ErrorLogs] struct {
	db *gorm.DB
}

func NewRepository[T ErrorLogs](db *gorm.DB) ORMRepository[T] {
	return &ErrorLogsRepository[T]{db: db}
}

func (r *ErrorLogsRepository[T]) GetAll() ([]T, error) {
	var errorLogs []T
	if err := r.db.Find(&errorLogs).Error; err != nil {
		return nil, err
	}
	return errorLogs, nil
}

func (r *ErrorLogsRepository[T]) GetByID(id string) (*ErrorLogs, error) {
	var errorLog ErrorLogs
	if err := r.db.First(&errorLog, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &errorLog, nil
}

func (r *ErrorLogsRepository[T]) Create(errorLog *ErrorLogs) error {
	if err := r.db.Create(errorLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *ErrorLogsRepository[T]) Update(errorLog *ErrorLogs) error {
	if err := r.db.Model(errorLog).Omit("ID").Updates(errorLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *ErrorLogsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&ErrorLogs{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
