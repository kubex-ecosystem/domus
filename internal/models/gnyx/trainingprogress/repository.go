// Package trainingprogress provides basic data modeling management tools
package trainingprogress

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingProgress, error)
	Create(trainingProgress *TrainingProgress) error
	Update(trainingProgress *TrainingProgress) error
	Delete(id string) error
}

type TrainingProgressRepository[T TrainingProgress] struct {
	db *gorm.DB
}

func NewRepository[T TrainingProgress](db *gorm.DB) ORMRepository[T] {
	return &TrainingProgressRepository[T]{db: db}
}

func (r *TrainingProgressRepository[T]) GetAll() ([]T, error) {
	var trainingProgress []T
	if err := r.db.Find(&trainingProgress).Error; err != nil {
		return nil, err
	}
	return trainingProgress, nil
}

func (r *TrainingProgressRepository[T]) GetByID(id string) (*TrainingProgress, error) {
	var trainingProgress TrainingProgress
	if err := r.db.First(&trainingProgress, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &trainingProgress, nil
}

func (r *TrainingProgressRepository[T]) Create(trainingProgress *TrainingProgress) error {
	if err := r.db.Create(trainingProgress).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingProgressRepository[T]) Update(trainingProgress *TrainingProgress) error {
	if err := r.db.Model(trainingProgress).Omit("ID").Updates(trainingProgress).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingProgressRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&TrainingProgress{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
