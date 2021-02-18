package model

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emvi/hide"
	"github.com/maxtroughear/logrusextension"
	"github.com/newrelic/go-agent/v3/newrelic"
	"gorm.io/gorm"

	"github.com/kiwisheets/auth"
	gqlClient "github.com/kiwisheets/gql-server/client"
)

// We are storing a majority of the Company object as we will use it often when generating invoices

// It will be cached by 2 events
// - Synced on request that requires the company (only if doesn't exist)
// - Synced by Company Updated Event (at least once message processing)

const CompanyBillingAddressType = "company_billing"
const CompanyShippingAddressType = "company_shipping"

type Company struct {
	ID        hide.ID `gorm:"type: bigserial;primary_key" json:"id"` // int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Code      string `gorm:"unique_index:idx_code"`
	Name      string
	Website   string

	BillingAddress  Address `gorm:"polymorphic:Addressee;polymorphicValue:company_billing"`
	ShippingAddress Address `gorm:"polymorphic:Addressee;polymorphicValue:company_shipping"`

	InvoiceTaxInclusive bool `json:"invoiceTaxInclusive"`
}

func (Company) IsEntity() {}

func GetCompany(ctx context.Context, db *gorm.DB, gqlClient *gqlClient.Client, id hide.ID) (Company, error) {
	tx := newrelic.FromContext(ctx)
	defer tx.StartSegment("GetCompany").End()

	logger := logrusextension.From(ctx)

	logger.WithField("companyId", id).Debugf("getting company")

	var company Company
	if err := db.Where(id).Find(&company).Error; err != nil || company.ID == 0 {
		// company does not exist, retrieve it
		logger.WithField("companyId", id).Debugf("hydrating company")
		company, err = hydrateCompany(ctx, db, gqlClient, id)
		if err != nil {
			// TODO: Log this

			return company, err
		}
	}
	return company, nil
}

func hydrateCompany(ctx context.Context, db *gorm.DB, gqlClient *gqlClient.Client, id hide.ID) (Company, error) {
	tx := newrelic.FromContext(ctx)
	var es *newrelic.ExternalSegment

	exCompany, err := gqlClient.GetCompany(ctx, func(req *http.Request) {
		es = newrelic.StartExternalSegment(tx, req)
		req.Header.Set("user", auth.For(ctx).OriginalHeader)
	})
	if err != nil {
		return Company{}, errors.New("company does not exist")
	}
	company := exCompanyToCompany(exCompany)

	// create in db
	if err := db.FirstOrCreate(&company).Error; err != nil {
		return Company{}, errors.New("failed to save company")
	}

	if es != nil {
		defer es.End()
		es.Procedure = "HydrateCompany"
	}

	return company, nil
}

func exCompanyToCompany(c *gqlClient.GetCompany) Company {
	return Company{
		ID:        c.Company.ID,
		CreatedAt: c.Company.CreatedAt,
		Code:      c.Company.Code,
		Name:      c.Company.Name,
		Website:   c.Company.Website,
		BillingAddress: Address{
			PostalCode: c.Company.BillingAddress.PostalCode,
			Name:       c.Company.BillingAddress.Name,
			Street1:    c.Company.BillingAddress.Street1,
			Street2:    c.Company.BillingAddress.Street2,
			City:       c.Company.BillingAddress.City,
			State:      c.Company.BillingAddress.State,
			Country:    c.Company.BillingAddress.Country,
		},
		ShippingAddress: Address{
			PostalCode: c.Company.ShippingAddress.PostalCode,
			Name:       c.Company.ShippingAddress.Name,
			Street1:    c.Company.ShippingAddress.Street1,
			Street2:    c.Company.ShippingAddress.Street2,
			City:       c.Company.ShippingAddress.City,
			State:      c.Company.ShippingAddress.State,
			Country:    c.Company.ShippingAddress.Country,
		},
		InvoiceTaxInclusive: true,
	}
}
