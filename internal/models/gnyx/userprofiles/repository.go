// Package userprofiles provides basic data modeling management tools
package userprofiles

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*UserProfiles, error)
	Create(userProfile *UserProfiles) error
	Update(userProfile *UserProfiles) error
	Delete(id string) error
}

type UserProfilesRepository[T UserProfiles] struct {
	db *gorm.DB
}

func NewRepository[T UserProfiles](db *gorm.DB) ORMRepository[T] {
	return &UserProfilesRepository[T]{db: db}
}

func (r *UserProfilesRepository[T]) GetAll() ([]T, error) {
	var userProfiles []T
	if err := r.db.Find(&userProfiles).Error; err != nil {
		return nil, err
	}
	return userProfiles, nil
}

func (r *UserProfilesRepository[T]) GetByID(id string) (*UserProfiles, error) {
	var userProfile UserProfiles
	if err := r.db.First(&userProfile, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &userProfile, nil
}

func (r *UserProfilesRepository[T]) Create(userProfile *UserProfiles) error {
	if err := r.db.Create(userProfile).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserProfilesRepository[T]) Update(userProfile *UserProfiles) error {
	if err := r.db.Model(userProfile).Omit("ID").Updates(userProfile).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserProfilesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&UserProfiles{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
