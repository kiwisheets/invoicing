package model

import (
	"context"
	"strconv"

	"github.com/emvi/hide"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InvoiceNumber struct {
	Number    int64
	CompanyID hide.ID
}

func (n InvoiceNumber) GormDataType() string {
	return "bigint"
}

func (n InvoiceNumber) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL: "nextval('invoice_number_" + strconv.FormatInt(int64(n.CompanyID), 10) + "')",
	}
}

func (n *InvoiceNumber) Scan(v interface{}) error {
	logrus.Debugf("scanning InvoiceNumber", n.Number)
	n.Number = v.(int64)
	return nil
}
