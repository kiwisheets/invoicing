package model

import (
	"context"

	"github.com/emvi/hide"
	"github.com/kiwisheets/auth"
	"github.com/maxtroughear/logrusextension"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	gqlClient "github.com/kiwisheets/gql-server/client"
)

func DefaultTaxInclusive(ctx context.Context, db *gorm.DB, gqlClient *gqlClient.Client, companyID hide.ID) bool {
	taxInclusive := true
	company, err := GetCompany(ctx, db, gqlClient, auth.For(ctx).CompanyID)
	if err != nil {
		logrusextension.From(ctx).WithFields(logrus.Fields{
			"companyID": auth.For(ctx).CompanyID,
		}).Warn("failed to retrieve company for DefaultTaxInclusive")
	} else {
		taxInclusive = company.InvoiceTaxInclusive
	}
	return taxInclusive
}

func DefaultTaxRate(ctx context.Context, db *gorm.DB, gqlClient *gqlClient.Client, companyID hide.ID) float64 {
	taxRate := 15.00 // default 15.00% tax
	company, err := GetCompany(ctx, db, gqlClient, auth.For(ctx).CompanyID)
	if err != nil {
		logrusextension.From(ctx).WithFields(logrus.Fields{
			"companyID": auth.For(ctx).CompanyID,
		}).Warn("failed to retrieve company for DefaultTaxRate")
	} else {
		taxRate = company.InvoiceTaxRate
	}
	return taxRate
}
