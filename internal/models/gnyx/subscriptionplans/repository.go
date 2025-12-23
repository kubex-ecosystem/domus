// Package subscriptionplans provides basic data modeling management tools
package subscriptionplans

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*SubscriptionPlans, error)
	Create(subscriptionPlan *SubscriptionPlans) error
	Update(subscriptionPlan *SubscriptionPlans) error
	Delete(id string) error
}

type SubscriptionPlansRepository[T SubscriptionPlans] struct {
	db *gorm.DB
}

func NewRepository[T SubscriptionPlans](db *gorm.DB) ORMRepository[T] {
	return &SubscriptionPlansRepository[T]{db: db}
}

func (r *SubscriptionPlansRepository[T]) GetAll() ([]T, error) {
	var subscriptionPlans []T
	if err := r.db.Find(&subscriptionPlans).Error; err != nil {
		return nil, err
	}
	return subscriptionPlans, nil
}

func (r *SubscriptionPlansRepository[T]) GetByID(id string) (*SubscriptionPlans, error) {
	var subscriptionPlan SubscriptionPlans
	if err := r.db.First(&subscriptionPlan, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &subscriptionPlan, nil
}

func (r *SubscriptionPlansRepository[T]) Create(subscriptionPlan *SubscriptionPlans) error {
	if err := r.db.Create(subscriptionPlan).Error; err != nil {
		return err
	}
	return nil
}

func (r *SubscriptionPlansRepository[T]) Update(subscriptionPlan *SubscriptionPlans) error {
	if err := r.db.Model(subscriptionPlan).Omit("ID").Updates(subscriptionPlan).Error; err != nil {
		return err
	}
	return nil
}

func (r *SubscriptionPlansRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&SubscriptionPlans{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
