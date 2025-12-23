// Package trainingcourses provides basic data modeling management tools
package trainingcourses

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TrainingCourses, error)
	Create(trainingCourse *TrainingCourses) error
	Update(trainingCourse *TrainingCourses) error
	Delete(id string) error
}

type TrainingCoursesRepository[T TrainingCourses] struct {
	db *gorm.DB
}

func NewRepository[T TrainingCourses](db *gorm.DB) ORMRepository[T] {
	return &TrainingCoursesRepository[T]{db: db}
}

func (r *TrainingCoursesRepository[T]) GetAll() ([]T, error) {
	var trainingCourses []T
	if err := r.db.Find(&trainingCourses).Error; err != nil {
		return nil, err
	}
	return trainingCourses, nil
}

func (r *TrainingCoursesRepository[T]) GetByID(id string) (*TrainingCourses, error) {
	var trainingCourse TrainingCourses
	if err := r.db.First(&trainingCourse, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &trainingCourse, nil
}

func (r *TrainingCoursesRepository[T]) Create(trainingCourse *TrainingCourses) error {
	if err := r.db.Create(trainingCourse).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingCoursesRepository[T]) Update(trainingCourse *TrainingCourses) error {
	if err := r.db.Model(trainingCourse).Omit("ID").Updates(trainingCourse).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrainingCoursesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&TrainingCourses{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
