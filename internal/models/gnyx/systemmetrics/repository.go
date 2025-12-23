// Package systemmetrics provides basic data modeling management tools
package systemmetrics

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*SystemMetrics, error)
	Create(systemMetric *SystemMetrics) error
	Update(systemMetric *SystemMetrics) error
	Delete(id string) error
}

type SystemMetricsRepository[T SystemMetrics] struct {
	db *gorm.DB
}

func NewRepository[T SystemMetrics](db *gorm.DB) ORMRepository[T] {
	return &SystemMetricsRepository[T]{db: db}
}

func (r *SystemMetricsRepository[T]) GetAll() ([]T, error) {
	var systemMetrics []T
	if err := r.db.Find(&systemMetrics).Error; err != nil {
		return nil, err
	}
	return systemMetrics, nil
}

func (r *SystemMetricsRepository[T]) GetByID(id string) (*SystemMetrics, error) {
	var systemMetric SystemMetrics
	if err := r.db.First(&systemMetric, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &systemMetric, nil
}

func (r *SystemMetricsRepository[T]) Create(systemMetric *SystemMetrics) error {
	if err := r.db.Create(systemMetric).Error; err != nil {
		return err
	}
	return nil
}

func (r *SystemMetricsRepository[T]) Update(systemMetric *SystemMetrics) error {
	if err := r.db.Model(systemMetric).Omit("ID").Updates(systemMetric).Error; err != nil {
		return err
	}
	return nil
}

func (r *SystemMetricsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&SystemMetrics{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
