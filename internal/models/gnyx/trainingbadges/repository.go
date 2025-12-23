// Package trainingbadges provides basic data modeling management tools
package trainingbadges

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingBadges, error)
	Create(trainingBadge *TrainingBadges) error
	Update(trainingBadge *TrainingBadges) error
	Delete(id string) error
}

type TrainingBadgesRepository[T TrainingBadges] struct {
	db *gorm.DB
}

func NewRepository[T TrainingBadges](db *gorm.DB) ORMRepository[T] {
	return &TrainingBadgesRepository[T]{db: db}
}

func (r *TrainingBadgesRepository[T]) GetAll() ([]T, error) {
	var trainingBadges []T
	if err := r.db.Find(&trainingBadges).Error; err != nil {
		return nil, err
	}
	return trainingBadges, nil
}

func (r *TrainingBadgesRepository[T]) GetByID(id string) (*TrainingBadges, error) {
	var trainingBadge TrainingBadges
	if err := r.db.First(&trainingBadge, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &trainingBadge, nil
}

func (r *TrainingBadgesRepository[T]) Create(trainingBadge *TrainingBadges) error {
	if err := r.db.Create(trainingBadge).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingBadgesRepository[T]) Update(trainingBadge *TrainingBadges) error {
	if err := r.db.Model(trainingBadge).Omit("ID").Updates(trainingBadge).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingBadgesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&TrainingBadges{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
