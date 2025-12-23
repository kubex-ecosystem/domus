package user

import (
	"time"
)

// IUser interface for abstraction and encapsulation
type IUser interface {
	TableName() string
	GetID() string
	SetID(id string)
	GetName() string
	SetName(name string)
	GetEmail() string
	SetEmail(email string)
	GetUserObj() *UserModel
	GetCreatedAt() time.Time
	SetCreatedAt(createdAt time.Time)
}

type UserModel struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Email     string    `gorm:"type:varchar(255);unique;not null" json:"email"`
	CreatedAt time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
}

func (um *UserModel) TableName() string {
	return "user"
}
func (um *UserModel) SetID(id string) {
	um.ID = id
}
func (um *UserModel) SetName(name string) {
	um.Name = name
}
func (um *UserModel) SetEmail(email string) {
	um.Email = email
}
func (um *UserModel) GetID() string {
	return um.ID
}
func (um *UserModel) GetName() string {
	return um.Name
}
func (um *UserModel) GetEmail() string {
	return um.Email
}
func (um *UserModel) GetUserObj() *UserModel {
	return um
}
func (um *UserModel) GetCreatedAt() time.Time {
	return um.CreatedAt
}
func (um *UserModel) SetCreatedAt(createdAt time.Time) {
	um.CreatedAt = createdAt
}
