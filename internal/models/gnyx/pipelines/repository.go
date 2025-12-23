// Package pipelines provides basic data modeling management tools
package pipelines

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Pipelines, error)
	Create(pipeline *Pipelines) error
	Update(pipeline *Pipelines) error
	Delete(id string) error
}

type PipelinesRepository[T Pipelines] struct {
	db *gorm.DB
}

func NewRepository[T Pipelines](db *gorm.DB) ORMRepository[T] {
	return &PipelinesRepository[T]{db: db}
}

func (r *PipelinesRepository[T]) GetAll() ([]T, error) {
	var pipelines []T
	if err := r.db.Find(&pipelines).Error; err != nil {
		return nil, err
	}
	return pipelines, nil
}

func (r *PipelinesRepository[T]) GetByID(id string) (*Pipelines, error) {
	var pipeline Pipelines
	if err := r.db.First(&pipeline, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func (r *PipelinesRepository[T]) Create(pipeline *Pipelines) error {
	if err := r.db.Create(pipeline).Error; err != nil {
		return err
	}
	return nil
}

func (r *PipelinesRepository[T]) Update(pipeline *Pipelines) error {
	if err := r.db.Model(pipeline).Omit("ID").Updates(pipeline).Error; err != nil {
		return err
	}
	return nil
}

func (r *PipelinesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&Pipelines{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
