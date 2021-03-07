package model

import (
	"time"

	"github.com/emvi/hide"
	"gorm.io/gorm"
)

type LineItem struct {
	ID           hide.ID `gorm:"type: bigserial;primary_key" json:"id"` // int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	InvoiceID    hide.ID
	Name         string
	Description  string
	UnitCost     float64
	TaxRate      *float64
	Quantity     float64
	TaxInclusive bool
}
