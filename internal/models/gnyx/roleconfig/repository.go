// Package roleconfig provides basic data modeling management tools
package roleconfig

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id int64) (*RoleConfig, error)
	Create(roleConfig *RoleConfig) error
	Update(roleConfig *RoleConfig) error
	Delete(id int64) error
}

type RoleConfigRepository[T RoleConfig] struct {
	db *gorm.DB
}

func NewRepository[T RoleConfig](db *gorm.DB) ORMRepository[T] {
	return &RoleConfigRepository[T]{db: db}
}

func (r *RoleConfigRepository[T]) GetAll() ([]T, error) {
	var roleConfigs []T
	if err := r.db.Find(&roleConfigs).Error; err != nil {
		return nil, err
	}
	return roleConfigs, nil
}

func (r *RoleConfigRepository[T]) GetByID(id int64) (*RoleConfig, error) {
	var roleConfig RoleConfig
	if err := r.db.First(&roleConfig, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &roleConfig, nil
}

func (r *RoleConfigRepository[T]) Create(roleConfig *RoleConfig) error {
	if err := r.db.Create(roleConfig).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleConfigRepository[T]) Update(roleConfig *RoleConfig) error {
	if err := r.db.Model(roleConfig).Omit("ID").Updates(roleConfig).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleConfigRepository[T]) Delete(id int64) error {
	if err := r.db.Delete(&RoleConfig{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
