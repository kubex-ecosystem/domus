// Package tenants provides basic data modeling management tools
package tenants

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Tenants, error)
	Create(tenant *Tenants) error
	Update(tenant *Tenants) error
	Delete(id string) error
}

type TenantsRepository[T Tenants] struct {
	db *gorm.DB
}

func NewRepository[T Tenants](db *gorm.DB) ORMRepository[T] {
	return &TenantsRepository[T]{db: db}
}

func (r *TenantsRepository[T]) GetAll() ([]T, error) {
	var tenants []T
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *TenantsRepository[T]) GetByID(id string) (*Tenants, error) {
	var tenant Tenants
	if err := r.db.First(&tenant, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantsRepository[T]) Create(tenant *Tenants) error {
	if err := r.db.Create(tenant).Error; err != nil {
		return err
	}
	return nil
}

func (r *TenantsRepository[T]) Update(tenant *Tenants) error {
	if err := r.db.Model(tenant).Omit("ID").Updates(tenant).Error; err != nil {
		return err
	}
	return nil
}

func (r *TenantsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&Tenants{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
