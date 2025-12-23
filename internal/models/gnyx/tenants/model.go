package tenants

import (
	"time"
)

type Tenants struct {
	ID         *string   `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Alias      string    `json:"alias" gorm:"column:alias"`
	Name       string    `json:"name" gorm:"column:name"`
	Email      string    `json:"email" gorm:"column:email"`
	Phone      *string   `json:"phone,omitempty" gorm:"column:phone"`
	Domain     string    `json:"domain" gorm:"column:domain"`
	TaxID      string    `json:"tax_id" gorm:"column:tax_id"`
	AddressIds []string  `json:"address_ids" gorm:"column:address_ids"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
	CreatedBy  *string   `json:"created_by,omitempty" gorm:"column:created_by"`
}

func (Tenants) TableName() string {
	return "gnyx.tenants"
}
