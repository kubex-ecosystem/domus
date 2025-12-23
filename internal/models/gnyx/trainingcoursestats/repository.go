// Package trainingcoursestats provides basic data modeling management tools
package trainingcoursestats

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingCourseStats, error)
	Create(trainingCourseStat *TrainingCourseStats) error
	Update(trainingCourseStat *TrainingCourseStats) error
	Delete(id string) error
}

type TrainingCourseStatsRepository[T TrainingCourseStats] struct {
	db *gorm.DB
}

func NewRepository[T TrainingCourseStats](db *gorm.DB) ORMRepository[T] {
	return &TrainingCourseStatsRepository[T]{db: db}
}

func (r *TrainingCourseStatsRepository[T]) GetAll() ([]T, error) {
	var trainingCourseStats []T
	if err := r.db.Find(&trainingCourseStats).Error; err != nil {
		return nil, err
	}
	return trainingCourseStats, nil
}

func (r *TrainingCourseStatsRepository[T]) GetByID(id string) (*TrainingCourseStats, error) {
	var trainingCourseStat TrainingCourseStats
	if err := r.db.First(&trainingCourseStat, "CourseID = ?", id).Error; err != nil {
		return nil, err
	}
	return &trainingCourseStat, nil
}

func (r *TrainingCourseStatsRepository[T]) Create(trainingCourseStat *TrainingCourseStats) error {
	if err := r.db.Create(trainingCourseStat).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingCourseStatsRepository[T]) Update(trainingCourseStat *TrainingCourseStats) error {
	if err := r.db.Model(trainingCourseStat).Omit("CourseID").Updates(trainingCourseStat).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingCourseStatsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&TrainingCourseStats{}, "CourseID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
