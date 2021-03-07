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
		}).Warn("failed to retrieve company for defaultTaxInclusive")
	} else {
		taxInclusive = company.InvoiceTaxInclusive
	}
	return taxInclusive
}
