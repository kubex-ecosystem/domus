// Package tenantsubscriptions provides basic data modeling management tools
package tenantsubscriptions

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*TenantSubscriptions, error)
	Create(tenantSubscription *TenantSubscriptions) error
	Update(tenantSubscription *TenantSubscriptions) error
	Delete(id string) error
}

type TenantSubscriptionsRepository[T TenantSubscriptions] struct {
	db *gorm.DB
}

func NewRepository[T TenantSubscriptions](db *gorm.DB) ORMRepository[T] {
	return &TenantSubscriptionsRepository[T]{db: db}
}

func (r *TenantSubscriptionsRepository[T]) GetAll() ([]T, error) {
	var tenantSubscriptions []T
	if err := r.db.Find(&tenantSubscriptions).Error; err != nil {
		return nil, err
	}
	return tenantSubscriptions, nil
}

func (r *TenantSubscriptionsRepository[T]) GetByID(id string) (*TenantSubscriptions, error) {
	var tenantSubscription TenantSubscriptions
	if err := r.db.First(&tenantSubscription, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &tenantSubscription, nil
}

func (r *TenantSubscriptionsRepository[T]) Create(tenantSubscription *TenantSubscriptions) error {
	if err := r.db.Create(tenantSubscription).Error; err != nil {
		return err
	}
	return nil
}

func (r *TenantSubscriptionsRepository[T]) Update(tenantSubscription *TenantSubscriptions) error {
	if err := r.db.Model(tenantSubscription).Omit("ID").Updates(tenantSubscription).Error; err != nil {
		return err
	}
	return nil
}

func (r *TenantSubscriptionsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&TenantSubscriptions{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
