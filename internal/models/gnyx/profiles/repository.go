// Package profiles provides basic data modeling management tools
package profiles

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Profiles, error)
	Create(profile *Profiles) error
	Update(profile *Profiles) error
	Delete(id string) error
}

type ProfilesRepository[T Profiles] struct {
	db *gorm.DB
}

func NewRepository[T Profiles](db *gorm.DB) ORMRepository[T] {
	return &ProfilesRepository[T]{db: db}
}

func (r *ProfilesRepository[T]) GetAll() ([]T, error) {
	var profiles []T
	if err := r.db.Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

func (r *ProfilesRepository[T]) GetByID(id string) (*Profiles, error) {
	var profile Profiles
	if err := r.db.First(&profile, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ProfilesRepository[T]) Create(profile *Profiles) error {
	if err := r.db.Create(profile).Error; err != nil {
		return err
	}
	return nil
}

func (r *ProfilesRepository[T]) Update(profile *Profiles) error {
	if err := r.db.Model(profile).Omit("ID").Updates(profile).Error; err != nil {
		return err
	}
	return nil
}

func (r *ProfilesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&Profiles{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
