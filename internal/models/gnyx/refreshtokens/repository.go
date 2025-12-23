// Package refreshtokens provides basic data modeling management tools
package refreshtokens

import (
	"gorm.io/gorm"
)

type ORMRepository[T any] interface {
	GetAll() ([]T, error)
	GetByID(id int64) (*RefreshTokens, error)
	Create(refreshToken *RefreshTokens) error
	Update(refreshToken *RefreshTokens) error
	Delete(id int64) error
}

type RefreshTokensRepository[T RefreshTokens] struct {
	db *gorm.DB
}

func NewRepository[T RefreshTokens](db *gorm.DB) ORMRepository[T] {
	return &RefreshTokensRepository[T]{db: db}
}

func (r *RefreshTokensRepository[T]) GetAll() ([]T, error) {
	var refreshTokens []T
	if err := r.db.Find(&refreshTokens).Error; err != nil {
		return nil, err
	}
	return refreshTokens, nil
}

func (r *RefreshTokensRepository[T]) GetByID(id int64) (*RefreshTokens, error) {
	var refreshToken RefreshTokens
	if err := r.db.First(&refreshToken, "ID = ?", id).Error; err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *RefreshTokensRepository[T]) Create(refreshToken *RefreshTokens) error {
	if err := r.db.Create(refreshToken).Error; err != nil {
		return err
	}
	return nil
}

func (r *RefreshTokensRepository[T]) Update(refreshToken *RefreshTokens) error {
	if err := r.db.Model(refreshToken).Omit("ID").Updates(refreshToken).Error; err != nil {
		return err
	}
	return nil
}

func (r *RefreshTokensRepository[T]) Delete(id int64) error {
	if err := r.db.Delete(&RefreshTokens{}, "ID = ?", id).Error; err != nil {
		return err
	}
	return nil
}
