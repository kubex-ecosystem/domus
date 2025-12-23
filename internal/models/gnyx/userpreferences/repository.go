// Package userpreferences provides basic data modeling management tools
package userpreferences

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*UserPreferences, error)
	Create(userPreference *UserPreferences) error
	Update(userPreference *UserPreferences) error
	Delete(id string) error
}

type UserPreferencesRepository[T UserPreferences] struct {
	db *gorm.DB
}

func NewRepository[T UserPreferences](db *gorm.DB) ORMRepository[T] {
	return &UserPreferencesRepository[T]{db: db}
}

func (r *UserPreferencesRepository[T]) GetAll() ([]T, error) {
	var userPreferences []T
	if err := r.db.Find(&userPreferences).Error; err != nil {
		return nil, err
	}
	return userPreferences, nil
}

func (r *UserPreferencesRepository[T]) GetByID(id string) (*UserPreferences, error) {
	var userPreference UserPreferences
	if err := r.db.First(&userPreference, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &userPreference, nil
}

func (r *UserPreferencesRepository[T]) Create(userPreference *UserPreferences) error {
	if err := r.db.Create(userPreference).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserPreferencesRepository[T]) Update(userPreference *UserPreferences) error {
	if err := r.db.Model(userPreference).Omit("ID").Updates(userPreference).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserPreferencesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&UserPreferences{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
