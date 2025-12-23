// Package activities provides basic data modeling management tools
package activities

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Activities, error)
	Create(activity *Activities) error
	Update(activity *Activities) error
	Delete(id string) error
}

type ActivitiesRepository[T Activities] struct {
	db *gorm.DB
}

func NewRepository[T Activities](db *gorm.DB) ORMRepository[T] {
	return &ActivitiesRepository[T]{db: db}
}

func (r *ActivitiesRepository[T]) GetAll() ([]T, error) {
	var activities []T
	if err := r.db.Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *ActivitiesRepository[T]) GetByID(id string) (*Activities, error) {
	var activity Activities
	if err := r.db.First(&activity, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivitiesRepository[T]) Create(activity *Activities) error {
	if err := r.db.Create(activity).Error; err != nil {
		return err
	}
	return nil
}

func (r *ActivitiesRepository[T]) Update(activity *Activities) error {
	if err := r.db.Model(activity).Omit("ID").Updates(activity).Error; err != nil {
		return err
	}
	return nil
}

func (r *ActivitiesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&Activities{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
