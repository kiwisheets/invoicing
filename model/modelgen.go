// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/emvi/hide"
)

type InvoiceInput struct {
	ClientID hide.ID          `json:"clientID"`
	Items    []*LineItemInput `json:"items"`
}

type LineItemInput struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	UnitCost     float64  `json:"unitCost"`
	TaxRate      *float64 `json:"taxRate"`
	Quantity     float64  `json:"quantity"`
	TaxInclusive *bool    `json:"taxInclusive"`
}

type User struct {
	ID hide.ID `json:"id"`
}

func (User) IsEntity() {}
