package model

import (
	"time"

	"github.com/emvi/hide"
)

type Address struct {
	ID         hide.ID `gorm:"type: bigserial;primary_key" json:"id"` // int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	Name       string
	Street1    string
	Street2    *string
	City       string
	State      *string
	PostalCode int
	Country    string

	AddresseeID   hide.ID
	AddresseeType string
}
