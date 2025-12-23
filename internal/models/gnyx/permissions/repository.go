// Package permissions provides basic data modeling management tools
package permissions

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Permissions, error)
	Create(permission *Permissions) error
	Update(permission *Permissions) error
	Delete(id string) error
}

type PermissionsRepository[T Permissions] struct {
	db *gorm.DB
}

func NewRepository[T Permissions](db *gorm.DB) ORMRepository[T] {
	return &PermissionsRepository[T]{db: db}
}

func (r *PermissionsRepository[T]) GetAll() ([]T, error) {
	var permissions []T
	if err := r.db.Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionsRepository[T]) GetByID(id string) (*Permissions, error) {
	var permission Permissions
	if err := r.db.First(&permission, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionsRepository[T]) Create(permission *Permissions) error {
	if err := r.db.Create(permission).Error; err != nil {
		return err
	}
	return nil
}

func (r *PermissionsRepository[T]) Update(permission *Permissions) error {
	if err := r.db.Model(permission).Omit("ID").Updates(permission).Error; err != nil {
		return err
	}
	return nil
}

func (r *PermissionsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&Permissions{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
