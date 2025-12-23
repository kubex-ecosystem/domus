// Package traininglessons provides basic data modeling management tools
package traininglessons

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingLessons, error)
	Create(trainingLesson *TrainingLessons) error
	Update(trainingLesson *TrainingLessons) error
	Delete(id string) error
}

type TrainingLessonsRepository[T TrainingLessons] struct {
	db *gorm.DB
}

func NewRepository[T TrainingLessons](db *gorm.DB) ORMRepository[T] {
	return &TrainingLessonsRepository[T]{db: db}
}

func (r *TrainingLessonsRepository[T]) GetAll() ([]T, error) {
	var trainingLessons []T
	if err := r.db.Find(&trainingLessons).Error; err != nil {
		return nil, err
	}
	return trainingLessons, nil
}

func (r *TrainingLessonsRepository[T]) GetByID(id string) (*TrainingLessons, error) {
	var trainingLesson TrainingLessons
	if err := r.db.First(&trainingLesson, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &trainingLesson, nil
}

func (r *TrainingLessonsRepository[T]) Create(trainingLesson *TrainingLessons) error {
	if err := r.db.Create(trainingLesson).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingLessonsRepository[T]) Update(trainingLesson *TrainingLessons) error {
	if err := r.db.Model(trainingLesson).Omit("ID").Updates(trainingLesson).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingLessonsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&TrainingLessons{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
