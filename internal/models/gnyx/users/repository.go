// Package users provides basic data modeling management tools
package users

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Users, error)
	Create(user *Users) error
	Update(user *Users) error
	Delete(id string) error
}

type UsersRepository[T Users] struct {
	db *gorm.DB
}

func NewRepository[T Users](db *gorm.DB) ORMRepository[T] {
	return &UsersRepository[T]{db: db}
}

func (r *UsersRepository[T]) GetAll() ([]T, error) {
	var users []T
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UsersRepository[T]) GetByID(id string) (*Users, error) {
	var user Users
	if err := r.db.First(&user, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UsersRepository[T]) Create(user *Users) error {
	if err := r.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *UsersRepository[T]) Update(user *Users) error {
	if err := r.db.Model(user).Omit("ID").Updates(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *UsersRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&Users{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
