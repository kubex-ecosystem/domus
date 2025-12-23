// Package roles provides basic data modeling management tools
package roles

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Roles, error)
	Create(role *Roles) error
	Update(role *Roles) error
	Delete(id string) error
}

type RolesRepository[T Roles] struct {
	db *gorm.DB
}

func NewRepository[T Roles](db *gorm.DB) ORMRepository[T] {
	return &RolesRepository[T]{db: db}
}

func (r *RolesRepository[T]) GetAll() ([]T, error) {
	var roles []T
	if err := r.db.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RolesRepository[T]) GetByID(id string) (*Roles, error) {
	var role Roles
	if err := r.db.First(&role, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RolesRepository[T]) Create(role *Roles) error {
	if err := r.db.Create(role).Error; err != nil {
		return err
	}
	return nil
}

func (r *RolesRepository[T]) Update(role *Roles) error {
	if err := r.db.Model(role).Omit("ID").Updates(role).Error; err != nil {
		return err
	}
	return nil
}

func (r *RolesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&Roles{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
