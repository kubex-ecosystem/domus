// Package lossreasonsanalytics provides basic data modeling management tools
package lossreasonsanalytics

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*LossReasonsAnalytics, error)
	Create(lossReasonsAnalytic *LossReasonsAnalytics) error
	Update(lossReasonsAnalytic *LossReasonsAnalytics) error
	Delete(id string) error
}

type LossReasonsAnalyticsRepository[T LossReasonsAnalytics] struct {
	db *gorm.DB
}

func NewRepository[T LossReasonsAnalytics](db *gorm.DB) ORMRepository[T] {
	return &LossReasonsAnalyticsRepository[T]{db: db}
}

func (r *LossReasonsAnalyticsRepository[T]) GetAll() ([]T, error) {
	var lossReasonsAnalytics []T
	if err := r.db.Find(&lossReasonsAnalytics).Error; err != nil {
		return nil, err
	}
	return lossReasonsAnalytics, nil
}

func (r *LossReasonsAnalyticsRepository[T]) GetByID(id string) (*LossReasonsAnalytics, error) {
	var lossReasonsAnalytic LossReasonsAnalytics
	if err := r.db.First(&lossReasonsAnalytic, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &lossReasonsAnalytic, nil
}

func (r *LossReasonsAnalyticsRepository[T]) Create(lossReasonsAnalytic *LossReasonsAnalytics) error {
	if err := r.db.Create(lossReasonsAnalytic).Error; err != nil {
		return err
	}
	return nil
}

func (r *LossReasonsAnalyticsRepository[T]) Update(lossReasonsAnalytic *LossReasonsAnalytics) error {
	if err := r.db.Model(lossReasonsAnalytic).Omit("ID").Updates(lossReasonsAnalytic).Error; err != nil {
		return err
	}
	return nil
}

func (r *LossReasonsAnalyticsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&LossReasonsAnalytics{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
