// Package pipelinestages provides basic data modeling management tools
package pipelinestages

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*PipelineStages, error)
	Create(pipelineStage *PipelineStages) error
	Update(pipelineStage *PipelineStages) error
	Delete(id string) error
}

type PipelineStagesRepository[T PipelineStages] struct {
	db *gorm.DB
}

func NewRepository[T PipelineStages](db *gorm.DB) ORMRepository[T] {
	return &PipelineStagesRepository[T]{db: db}
}

func (r *PipelineStagesRepository[T]) GetAll() ([]T, error) {
	var pipelineStages []T
	if err := r.db.Find(&pipelineStages).Error; err != nil {
		return nil, err
	}
	return pipelineStages, nil
}

func (r *PipelineStagesRepository[T]) GetByID(id string) (*PipelineStages, error) {
	var pipelineStage PipelineStages
	if err := r.db.First(&pipelineStage, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &pipelineStage, nil
}

func (r *PipelineStagesRepository[T]) Create(pipelineStage *PipelineStages) error {
	if err := r.db.Create(pipelineStage).Error; err != nil {
		return err
	}
	return nil
}

func (r *PipelineStagesRepository[T]) Update(pipelineStage *PipelineStages) error {
	if err := r.db.Model(pipelineStage).Omit("ID").Updates(pipelineStage).Error; err != nil {
		return err
	}
	return nil
}

func (r *PipelineStagesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&PipelineStages{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
