package model

import (
	"time"

	"github.com/emvi/hide"
	"github.com/shopspring/decimal"
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
	UnitCost     decimal.Decimal
	TaxRate      decimal.Decimal
	Quantity     float64
	TaxInclusive bool
}

func (l *LineItem) Tax() decimal.Decimal {
	tax := l.UnitCost.Mul(decimal.NewFromFloat(l.Quantity))
	if l.TaxInclusive {
		tax = tax.Sub(tax.Div(l.TaxRatePercent().Add(decimal.NewFromInt(1))))
	} else {
		tax = tax.Mul(l.TaxRatePercent())
	}
	return tax
}

func (l *LineItem) Total() decimal.Decimal {
	total := l.UnitCost.Mul(decimal.NewFromFloat(l.Quantity))
	if !l.TaxInclusive {
		// add tax
		total = total.Add(total.Mul(l.TaxRatePercent()))
	}
	return total
}

func (l *LineItem) TaxRatePercent() decimal.Decimal {
	return l.TaxRate.Div(decimal.NewFromInt(100))
}
