// Package paymentmethods provides basic data modeling management tools
package paymentmethods

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*PaymentMethods, error)
	Create(paymentMethod *PaymentMethods) error
	Update(paymentMethod *PaymentMethods) error
	Delete(id string) error
}

type PaymentMethodsRepository[T PaymentMethods] struct {
	db *gorm.DB
}

func NewRepository[T PaymentMethods](db *gorm.DB) ORMRepository[T] {
	return &PaymentMethodsRepository[T]{db: db}
}

func (r *PaymentMethodsRepository[T]) GetAll() ([]T, error) {
	var paymentMethods []T
	if err := r.db.Find(&paymentMethods).Error; err != nil {
		return nil, err
	}
	return paymentMethods, nil
}

func (r *PaymentMethodsRepository[T]) GetByID(id string) (*PaymentMethods, error) {
	var paymentMethod PaymentMethods
	if err := r.db.First(&paymentMethod, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &paymentMethod, nil
}

func (r *PaymentMethodsRepository[T]) Create(paymentMethod *PaymentMethods) error {
	if err := r.db.Create(paymentMethod).Error; err != nil {
		return err
	}
	return nil
}

func (r *PaymentMethodsRepository[T]) Update(paymentMethod *PaymentMethods) error {
	if err := r.db.Model(paymentMethod).Omit("ID").Updates(paymentMethod).Error; err != nil {
		return err
	}
	return nil
}

func (r *PaymentMethodsRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&PaymentMethods{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
