package model

import (
	"github.com/emvi/hide"
	orm "github.com/kiwisheets/orm/model"
)

type Invoice struct {
	orm.SoftDelete
	Number    int
	CompanyID hide.ID
	CreatedBy hide.ID
	Client    hide.ID

	LineItems []LineItem `json:"items"`
}

type InvoiceRenderMQ struct {
	Invoice      Invoice
	NotifyConfig Notify
}
