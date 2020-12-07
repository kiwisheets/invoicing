package model

import (
	"github.com/emvi/hide"
	orm "github.com/kiwisheets/orm/model"
)

type LineItem struct {
	orm.Model
	InvoiceID   hide.ID
	Name        string
	Description string
	UnitCost    float64
	TaxRate     *float64
	Quantity    float64
}
