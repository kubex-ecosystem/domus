// Package companymetrics provides basic data modeling management tools
package companymetrics

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*CompanyMetrics, error)
	Create(companyMetric *CompanyMetrics) error
	Update(companyMetric *CompanyMetrics) error
	Delete(id string) error
}

type CompanyMetricsRepository[T CompanyMetrics] struct {
	db *gorm.DB
}

func NewRepository[T CompanyMetrics](db *gorm.DB) ORMRepository[T] {
	return &CompanyMetricsRepository[T]{db: db}
}

func (r *CompanyMetricsRepository[T]) GetAll() ([]T, error) {
	var companyMetrics []T
	if err := r.db.Find(&companyMetrics).Error; err != nil {
		return nil, err
	}
	return companyMetrics, nil
}

func (r *CompanyMetricsRepository[T]) GetByID(id string) (*CompanyMetrics, error) {
	var companyMetric CompanyMetrics
	if err := r.db.First(&companyMetric, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &companyMetric, nil
}

func (r *CompanyMetricsRepository[T]) Create(companyMetric *CompanyMetrics) error {
	if err := r.db.Create(companyMetric).Error; err != nil {
		return err
	}
	return nil
}

func (r *CompanyMetricsRepository[T]) Update(companyMetric *CompanyMetrics) error {
	if err := r.db.Model(companyMetric).Omit("ID").Updates(companyMetric).Error; err != nil {
		return err
	}
	return nil
}

func (r *CompanyMetricsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&CompanyMetrics{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
