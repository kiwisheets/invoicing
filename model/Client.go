package model

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emvi/hide"
	"github.com/kiwisheets/auth"
	gqlClient "github.com/kiwisheets/gql-server/client"
	"github.com/maxtroughear/logrusextension"
	"github.com/newrelic/go-agent/v3/newrelic"
	"gorm.io/gorm"
)

const ClientBillingAddressType = "client_billing"
const ClientShippingAddressType = "client_shipping"

type Client struct {
	ID        hide.ID `gorm:"type: bigserial;primary_key" json:"id"` // int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Name      string

	Website        *string
	VatNumber      *string
	BusinessNumber *string
	Phone          *string

	BillingAddress  Address `gorm:"polymorphic:Addressee;polymorphicValue:client_billing"`
	ShippingAddress Address `gorm:"polymorphic:Addressee;polymorphicValue:client_shipping"`

	Contacts  []Contact
	CompanyID hide.ID `json:"company"`
	Company   Company `json:"-"`

	Invoices []*Invoice `json:"invoices"`
}

func (Client) IsEntity() {}

func GetClient(ctx context.Context, db *gorm.DB, gqlClient *gqlClient.Client, id hide.ID) (Client, error) {
	tx := newrelic.FromContext(ctx)
	defer tx.StartSegment("GetClient").End()

	logger := logrusextension.From(ctx)

	logger.WithField("clientId", id).Debugf("getting client")

	var client Client
	if err := db.Where(id).Find(&client).Error; err != nil || client.ID == 0 {
		// company does not exist, retrieve it
		logger.WithField("clientId", id).Debugf("hydrating client")
		client, err = hydrateClient(ctx, db, gqlClient, id)
		if err != nil {
			// TODO: Log this

			return client, err
		}
	}
	return client, nil
}

func hydrateClient(ctx context.Context, db *gorm.DB, gqlClient *gqlClient.Client, id hide.ID) (Client, error) {
	tx := newrelic.FromContext(ctx)
	var es *newrelic.ExternalSegment

	exClient, err := gqlClient.GetClientByID(ctx, id, func(req *http.Request) {
		es = newrelic.StartExternalSegment(tx, req)
		req.Header.Set("user", auth.For(ctx).OriginalHeader)
	})
	if err != nil {
		return Client{}, errors.New("client does not exist")
	}
	client := exClientToClient(exClient, auth.For(ctx).CompanyID)

	// check company exists
	if _, err := GetCompany(ctx, db, gqlClient, client.CompanyID); err != nil {
		return client, errors.New("failed to verify if client company exists")
	}

	// create in db
	if err := db.FirstOrCreate(&client).Error; err != nil {
		return client, errors.New("failed to save client")
	}

	if es != nil {
		defer es.End()
		es.Procedure = "HydrateClient"
	}

	return client, nil
}

func exClientToClient(c *gqlClient.GetClientByID, companyID hide.ID) Client {
	client := Client{
		ID:             c.Client.ID,
		CreatedAt:      c.Client.CreatedAt,
		Name:           c.Client.Name,
		Website:        c.Client.Website,
		VatNumber:      c.Client.VatNumber,
		BusinessNumber: c.Client.BusinessNumber,
		Phone:          c.Client.Phone,
		BillingAddress: Address{
			PostalCode: c.Client.BillingAddress.PostalCode,
			Name:       c.Client.BillingAddress.Name,
			Street1:    c.Client.BillingAddress.Street1,
			Street2:    c.Client.BillingAddress.Street2,
			City:       c.Client.BillingAddress.City,
			State:      c.Client.BillingAddress.State,
			Country:    c.Client.BillingAddress.Country,
		},
		ShippingAddress: Address{
			PostalCode: c.Client.ShippingAddress.PostalCode,
			Name:       c.Client.ShippingAddress.Name,
			Street1:    c.Client.ShippingAddress.Street1,
			Street2:    c.Client.ShippingAddress.Street2,
			City:       c.Client.ShippingAddress.City,
			State:      c.Client.ShippingAddress.State,
			Country:    c.Client.ShippingAddress.Country,
		},
		Contacts:  make([]Contact, len(c.Client.Contacts)),
		CompanyID: companyID,
	}

	for i, contact := range c.Client.Contacts {
		client.Contacts[i] = Contact{
			ID:               contact.ID,
			Email:            contact.Email,
			Phone:            contact.Phone,
			Mobile:           contact.Mobile,
			PreferredContact: (*PreferredContact)(contact.PreferredContact),
			Firstname:        contact.Firstname,
			Lastname:         contact.Lastname,
		}
	}

	return client
}
