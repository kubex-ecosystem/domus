// Package repo provides data access for user-related operations.
package repo

import (
	"context"

	"gorm.io/gorm"
)

// Repo provides data access for user-related operations.
type Repo[M any] struct {
	db *gorm.DB
}

// NewRepo creates a new Repo instance.
func NewRepo[M any](db *gorm.DB, model *M) *Repo[M] {
	db.Model(new(M)).AutoMigrate(model)
	return &Repo[M]{
		db: db,
	}
}

// Create creates a new entity in the database.
func (r *Repo[M]) Create(u *M) error {
	return r.db.Create(u).Error
}

// GetByID retrieves an entity by its ID.
func (r *Repo[M]) GetByID(id uint) (*M, error) {
	var u M
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// Update updates an existing entity in the database.
func (r *Repo[M]) Update(u *M) error {
	return r.db.Save(u).Error
}

// Delete deletes an entity from the database by its ID.
func (r *Repo[M]) Delete(id uint) error {
	return r.db.Delete(new(M), id).Error
}

// List retrieves all entities from the database.
func (r *Repo[M]) List() ([]M, error) {
	var entities []M
	if err := r.db.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// GetCurrent retrieves the current system entity.
func GetCurrent[M any](ctx context.Context) (*M, error) {
	v := ctx.Value(new(M))
	if v == nil {
		return nil, nil
	}
	return v.(*M), nil
}
