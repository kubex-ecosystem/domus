package addresses

import (
	"time"
)

type Addresses struct {
	ID              *string    `json:"id,omitempty" gorm:"column:id;primaryKey"`
	ExternalID      *string    `json:"external_id,omitempty" gorm:"column:external_id"`
	ExternalCode    *string    `json:"external_code,omitempty" gorm:"column:external_code"`
	Street          string     `json:"street" gorm:"column:street"`
	Number          string     `json:"number" gorm:"column:number"`
	Complement      *string    `json:"complement,omitempty" gorm:"column:complement"`
	District        *string    `json:"district,omitempty" gorm:"column:district"`
	City            string     `json:"city" gorm:"column:city"`
	State           string     `json:"state" gorm:"column:state"`
	Country         string     `json:"country" gorm:"column:country"`
	ZipCode         string     `json:"zip_code" gorm:"column:zip_code"`
	IsMain          *bool      `json:"is_main,omitempty" gorm:"column:is_main"`
	IsActive        bool       `json:"is_active" gorm:"column:is_active"`
	Notes           *string    `json:"notes,omitempty" gorm:"column:notes"`
	Latitude        *float64   `json:"latitude,omitempty" gorm:"column:latitude"`
	Longitude       *float64   `json:"longitude,omitempty" gorm:"column:longitude"`
	AddressType     *string    `json:"address_type,omitempty" gorm:"column:address_type"`
	AddressStatus   *string    `json:"address_status,omitempty" gorm:"column:address_status"`
	AddressCategory *string    `json:"address_category,omitempty" gorm:"column:address_category"`
	AddressTags     *string    `json:"address_tags,omitempty" gorm:"column:address_tags"`
	IsDefault       *bool      `json:"is_default,omitempty" gorm:"column:is_default"`
	CreatedAt       time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"column:updated_at"`
	LastSyncAt      *time.Time `json:"last_sync_at,omitempty" gorm:"column:last_sync_at"`
}

func (Addresses) TableName() string {
	return "gnyx.addresses"
}
