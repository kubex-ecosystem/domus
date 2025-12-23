// Package companies provides basic data modeling management tools
package companies

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*Companies, error)
	Create(company *Companies) error
	Update(company *Companies) error
	Delete(id string) error
}

type CompaniesRepository[T Companies] struct {
	db *gorm.DB
}

func NewRepository[T Companies](db *gorm.DB) ORMRepository[T] {
	return &CompaniesRepository[T]{db: db}
}

func (r *CompaniesRepository[T]) GetAll() ([]T, error) {
	var companies []T
	if err := r.db.Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

func (r *CompaniesRepository[T]) GetByID(id string) (*Companies, error) {
	var company Companies
	if err := r.db.First(&company, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompaniesRepository[T]) Create(company *Companies) error {
	if err := r.db.Create(company).Error; err != nil {
		return err
	}
	return nil
}

func (r *CompaniesRepository[T]) Update(company *Companies) error {
	if err := r.db.Model(company).Omit("ID").Updates(company).Error; err != nil {
		return err
	}
	return nil
}

func (r *CompaniesRepository[T]) Delete(id string) error {
	if err := r.db.Delete(&Companies{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
