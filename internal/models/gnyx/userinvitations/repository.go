// Package userinvitations provides basic data modeling management tools
package userinvitations

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*UserInvitations, error)
	Create(userInvitation *UserInvitations) error
	Update(userInvitation *UserInvitations) error
	Delete(id string) error
}

type UserInvitationsRepository[T UserInvitations] struct {
	db *gorm.DB
}

func NewRepository[T UserInvitations](db *gorm.DB) ORMRepository[T] {
	return &UserInvitationsRepository[T]{db: db}
}

func (r *UserInvitationsRepository[T]) GetAll() ([]T, error) {
	var userInvitations []T
	if err := r.db.Find(&userInvitations).Error; err != nil {
		return nil, err
	}
	return userInvitations, nil
}

func (r *UserInvitationsRepository[T]) GetByID(id string) (*UserInvitations, error) {
	var userInvitation UserInvitations
	if err := r.db.First(&userInvitation, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &userInvitation, nil
}

func (r *UserInvitationsRepository[T]) Create(userInvitation *UserInvitations) error {
	if err := r.db.Create(userInvitation).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserInvitationsRepository[T]) Update(userInvitation *UserInvitations) error {
	if err := r.db.Model(userInvitation).Omit("ID").Updates(userInvitation).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserInvitationsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&UserInvitations{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
