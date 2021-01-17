package model

import (
	"time"

	"github.com/emvi/hide"
	"github.com/kiwisheets/gql-server/client"
)

type Invoice struct {
	ID        hide.ID `gorm:"type: bigserial;primary_key" json:"id"` // int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Number    InvoiceNumber `gorm:"default:1"`
	CompanyID hide.ID
	CreatedBy hide.ID
	Client    hide.ID

	LineItems []LineItem `json:"items"`
}

type InvoiceRenderMQ struct {
	Invoice      Invoice
	NotifyConfig Notify
}

type InvoiceTemplateData struct {
	Number  int64
	Client  *client.GetClientByID
	Company *client.GetCompany
	Items   []*LineItemInput
}
