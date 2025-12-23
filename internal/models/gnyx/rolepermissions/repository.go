// Package rolepermissions provides basic data modeling management tools
package rolepermissions

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*RolePermissions, error)
	Create(rolePermission *RolePermissions) error
	Update(rolePermission *RolePermissions) error
	Delete(id string) error
}

type RolePermissionsRepository[T RolePermissions] struct {
	db *gorm.DB
}

func NewRepository[T RolePermissions](db *gorm.DB) ORMRepository[T] {
	return &RolePermissionsRepository[T]{db: db}
}

func (r *RolePermissionsRepository[T]) GetAll() ([]T, error) {
	var rolePermissions []T
	if err := r.db.Find(&rolePermissions).Error; err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (r *RolePermissionsRepository[T]) GetByID(id string) (*RolePermissions, error) {
	var rolePermission RolePermissions
	if err := r.db.First(&rolePermission, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &rolePermission, nil
}

func (r *RolePermissionsRepository[T]) Create(rolePermission *RolePermissions) error {
	if err := r.db.Create(rolePermission).Error; err != nil {
		return err
	}
	return nil
}

func (r *RolePermissionsRepository[T]) Update(rolePermission *RolePermissions) error {
	if err := r.db.Model(rolePermission).Omit("ID").Updates(rolePermission).Error; err != nil {
		return err
	}
	return nil
}

func (r *RolePermissionsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&RolePermissions{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
