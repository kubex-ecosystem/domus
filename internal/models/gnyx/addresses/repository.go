// Package addresses provides basic data modeling management tools
package addresses

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Addresses, error)
	Create(address *Addresses) error
	Update(address *Addresses) error
	Delete(id string) error
}

type AddressesRepository[T Addresses] struct {
	db *gorm.DB
}

func NewRepository[T Addresses](db *gorm.DB) ORMRepository[T] {
	return &AddressesRepository[T]{db: db}
}

func (r *AddressesRepository[T]) GetAll() ([]T, error) {
	var addresses []T
	if err := r.db.Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *AddressesRepository[T]) GetByID(id string) (*Addresses, error) {
	var address Addresses
	if err := r.db.First(&address, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &address, nil
}

func (r *AddressesRepository[T]) Create(address *Addresses) error {
	if err := r.db.Create(address).Error; err != nil {
		return err
	}
	return nil
}

func (r *AddressesRepository[T]) Update(address *Addresses) error {
	if err := r.db.Model(address).Omit("ID").Updates(address).Error; err != nil {
		return err
	}
	return nil
}

func (r *AddressesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&Addresses{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
