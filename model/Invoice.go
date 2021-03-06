package model

import (
	"time"

	"github.com/emvi/hide"
	"github.com/leekchan/accounting"
)

type Invoice struct {
	ID        hide.ID `gorm:"type: bigserial;primary_key" json:"Id"` // int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Number    InvoiceNumber `gorm:"default:1"`
	CompanyID hide.ID
	CreatedBy hide.ID
	ClientID  hide.ID
	DateDue   time.Time

	LineItems []LineItem `json:"Items"`
}

type InvoiceRenderMQ struct {
	Invoice      Invoice
	NotifyConfig Notify
}

type InvoiceTemplateData struct {
	Number  int
	Client  Client
	Company Company
	Items   []*LineItemInput
}

func InvoiceTotalHelper(invoice *InvoiceTemplateData) string {
	total := 0.0
	for _, item := range invoice.Items {
		total = total + (item.Quantity * item.UnitCost)
	}

	ac := accounting.DefaultAccounting("$", 2)
	return ac.FormatMoneyFloat64(total)
}

func InvoiceItemCostHelper(item *LineItemInput) string {
	ac := accounting.DefaultAccounting("$", 2)
	return ac.FormatMoneyFloat64(item.UnitCost)
}

func InvoiceItemTotalHelper(item *LineItemInput) string {
	total := item.Quantity * item.UnitCost

	ac := accounting.DefaultAccounting("$", 2)
	return ac.FormatMoneyFloat64(total)
}
