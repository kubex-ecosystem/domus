// Package activityparticipants provides basic data modeling management tools
package activityparticipants

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*ActivityParticipants, error)
	Create(activityParticipant *ActivityParticipants) error
	Update(activityParticipant *ActivityParticipants) error
	Delete(id string) error
}

type ActivityParticipantsRepository[T ActivityParticipants] struct {
	db *gorm.DB
}

func NewRepository[T ActivityParticipants](db *gorm.DB) ORMRepository[T] {
	return &ActivityParticipantsRepository[T]{db: db}
}

func (r *ActivityParticipantsRepository[T]) GetAll() ([]T, error) {
	var activityParticipants []T
	if err := r.db.Find(&activityParticipants).Error; err != nil {
		return nil, err
	}
	return activityParticipants, nil
}

func (r *ActivityParticipantsRepository[T]) GetByID(id string) (*ActivityParticipants, error) {
	var activityParticipant ActivityParticipants
	if err := r.db.First(&activityParticipant, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &activityParticipant, nil
}

func (r *ActivityParticipantsRepository[T]) Create(activityParticipant *ActivityParticipants) error {
	if err := r.db.Create(activityParticipant).Error; err != nil {
		return err
	}
	return nil
}

func (r *ActivityParticipantsRepository[T]) Update(activityParticipant *ActivityParticipants) error {
	if err := r.db.Model(activityParticipant).Omit("ID").Updates(activityParticipant).Error; err != nil {
		return err
	}
	return nil
}

func (r *ActivityParticipantsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&ActivityParticipants{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
