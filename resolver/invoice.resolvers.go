package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/emvi/hide"
	"github.com/google/uuid"
	"github.com/kiwisheets/auth"
	"github.com/kiwisheets/invoicing/graphql/generated"
	"github.com/kiwisheets/invoicing/helper"
	"github.com/kiwisheets/invoicing/model"
	"github.com/kiwisheets/util"
	"github.com/maxtroughear/logrusextension"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *invoiceResolver) Status(ctx context.Context, obj *model.Invoice) (model.InvoiceStatus, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *invoiceResolver) PaymentStatus(ctx context.Context, obj *model.Invoice) (*model.InvoicePaymentStatus, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *invoiceResolver) Number(ctx context.Context, obj *model.Invoice) (string, error) {
	return strconv.FormatInt(obj.Number.Number, 10), nil
}

func (r *invoiceResolver) CreatedBy(ctx context.Context, obj *model.Invoice) (*model.User, error) {
	return &model.User{
		ID: obj.CreatedBy,
	}, nil
}

func (r *invoiceResolver) Client(ctx context.Context, obj *model.Invoice) (*model.Client, error) {
	return &model.Client{
		ID: obj.ClientID,
	}, nil
}

func (r *invoiceResolver) SubTotal(ctx context.Context, obj *model.Invoice) (float64, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *invoiceResolver) TotalTax(ctx context.Context, obj *model.Invoice) (float64, error) {
	tax := decimal.NewFromInt(0)
	for _, item := range obj.LineItems {
		tax = tax.Add(item.Tax())
	}

	taxFloat, _ := tax.RoundBank(2).Float64()
	return taxFloat, nil
}

func (r *invoiceResolver) Total(ctx context.Context, obj *model.Invoice) (float64, error) {
	total := decimal.NewFromInt(0)
	for _, item := range obj.LineItems {
		total = total.Add(item.Total())
	}

	totalFloat, _ := total.Float64()
	return totalFloat, nil
}

func (r *lineItemResolver) UnitCost(ctx context.Context, obj *model.LineItem) (float64, error) {
	f, _ := obj.UnitCost.Float64()
	return f, nil
}

func (r *lineItemResolver) TaxRate(ctx context.Context, obj *model.LineItem) (float64, error) {
	f, _ := obj.TaxRate.Float64()
	return f, nil
}

func (r *lineItemResolver) Tax(ctx context.Context, obj *model.LineItem) (float64, error) {
	f, _ := obj.Tax().Float64()
	return f, nil
}

func (r *lineItemResolver) Total(ctx context.Context, obj *model.LineItem) (float64, error) {
	totalFloat, _ := obj.Total().Float64()
	return totalFloat, nil
}

func (r *mutationResolver) CreateInvoice(ctx context.Context, invoice model.InvoiceInput) (*model.Invoice, error) {
	lineItems := make([]model.LineItem, len(invoice.Items))
	for i, l := range invoice.Items {
		lineItems[i] = model.LineItem{
			Description: l.Description,
			Name:        l.Name,
			Quantity:    l.Quantity,
			UnitCost:    decimal.NewFromFloat(l.UnitCost),
		}

		if l.TaxRate != nil {
			lineItems[i].TaxRate = decimal.NewFromFloat(*l.TaxRate)
		} else {
			lineItems[i].TaxRate = decimal.NewFromFloat(model.DefaultTaxRate(ctx, r.DB, r.GqlClient, auth.For(ctx).CompanyID))
		}

		if l.TaxInclusive != nil {
			lineItems[i].TaxInclusive = *l.TaxInclusive
		} else {
			lineItems[i].TaxInclusive = model.DefaultTaxInclusive(ctx, r.DB, r.GqlClient, auth.For(ctx).CompanyID)
		}
	}

	newInvoice := &model.Invoice{
		ClientID:  invoice.ClientID,
		CreatedBy: auth.For(ctx).UserID,
		CompanyID: auth.For(ctx).CompanyID,
		LineItems: lineItems,
		Number: model.InvoiceNumber{
			CompanyID: auth.For(ctx).CompanyID,
		},
		DateDue: invoice.DateDue,
	}

	log := logrusextension.From(ctx)
	log.Debugln("creating sequence")
	if err := r.DB.Exec("CREATE SEQUENCE IF NOT EXISTS invoice_number_" + strconv.FormatInt(int64(auth.For(ctx).CompanyID), 10) + " INCREMENT 1 START 1 OWNED BY invoices.number").Error; err != nil {
		log.Errorf("create sequence error %v", err)
	}

	if err := r.DB.Debug().Clauses(clause.Returning{
		Columns: []clause.Column{
			clause.PrimaryColumn,
			{
				Table: clause.CurrentTable,
				Name:  "NUMBER",
			},
		},
	}).Create(&newInvoice).Error; err != nil {
		return nil, err
	}

	return newInvoice, nil
}

func (r *mutationResolver) UpdateInvoice(ctx context.Context, id hide.ID, invoice model.InvoiceInput) (*model.Invoice, error) {
	lineItems := make([]model.LineItem, len(invoice.Items))
	for i, l := range invoice.Items {
		lineItems[i] = model.LineItem{
			Description: l.Description,
			Name:        l.Name,
			Quantity:    l.Quantity,
			UnitCost:    decimal.NewFromFloat(l.UnitCost),
		}

		if l.TaxRate != nil {
			lineItems[i].TaxRate = decimal.NewFromFloat(*l.TaxRate)
		} else {
			lineItems[i].TaxRate = decimal.NewFromFloat(model.DefaultTaxRate(ctx, r.DB, r.GqlClient, auth.For(ctx).CompanyID))
		}

		if l.TaxInclusive != nil {
			lineItems[i].TaxInclusive = *l.TaxInclusive
		} else {
			lineItems[i].TaxInclusive = model.DefaultTaxInclusive(ctx, r.DB, r.GqlClient, auth.For(ctx).CompanyID)
		}
	}

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Invoice{
			ID: id,
		}).Updates(&model.Invoice{
			ClientID: invoice.ClientID,
		}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.LineItem{}, "invoice_id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Invoice{
			ID: id,
		}).Association("LineItems").Append(&lineItems); err != nil {
			return nil
		}

		return nil
	})

	return nil, err
}

func (r *mutationResolver) CreateInvoicePdf(ctx context.Context, id hide.ID) (string, error) {
	var invoice model.Invoice
	r.DB.Where(id).Where("company_id = ?", auth.For(ctx).CompanyID).Preload("LineItems").Find(&invoice)

	if invoice.ID == 0 {
		return "", fmt.Errorf("error processing invoice, does the invoice exist?")
	}

	notifyID := uuid.New()

	msg, err := json.Marshal(model.InvoiceRenderMQ{
		Invoice: invoice,
		NotifyConfig: model.Notify{
			Users: []int64{
				int64(auth.For(ctx).UserID), // notify rendering user when invoice is rendered
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("error processing invoice, bad data")
	}
	r.MQ.RenderProducer.Produce(msg)
	return notifyID.String(), nil
}

func (r *mutationResolver) UpdateCompanyTaxInclusive(ctx context.Context, invoiceTaxInclusive bool) (*model.Company, error) {
	companyID := auth.For(ctx).CompanyID

	company, err := model.GetCompany(ctx, r.DB, r.GqlClient, companyID)
	if err != nil {
		return nil, fmt.Errorf("company not found")
	}
	r.DB.Model(&company).Update("invoice_tax_inclusive", invoiceTaxInclusive)

	company.InvoiceTaxInclusive = invoiceTaxInclusive
	return &company, nil
}

func (r *queryResolver) Invoice(ctx context.Context, id hide.ID) (*model.Invoice, error) {
	var invoice model.Invoice

	if err := r.DB.Where(id).Where("company_id = ?", auth.For(ctx).CompanyID).Preload("LineItems").Find(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found")
	}
	return &invoice, nil
}

func (r *queryResolver) Invoices(ctx context.Context, page *int) ([]*model.Invoice, error) {
	limit := 20
	invoices := make([]*model.Invoice, limit)
	if page == nil {
		page = util.Int(0)
	}
	r.DB.Where("company_id = ?", auth.For(ctx).CompanyID).Limit(limit).Offset(limit * *page).Preload("LineItems").Find(&invoices)

	return invoices, nil
}

func (r *queryResolver) PreviewInvoice(ctx context.Context, invoice model.InvoiceInput) (string, error) {
	// load template and exec, return html
	log := logrusextension.From(ctx)

	tx := newrelic.FromContext(ctx)

	var wg sync.WaitGroup
	var company model.Company
	var client model.Client

	wg.Add(2)
	go func() {
		defer wg.Done()
		var err error
		txg := tx.NewGoroutine()
		ctxg := newrelic.NewContext(ctx, txg)

		company, err = model.GetCompany(ctxg, r.DB, r.GqlClient, auth.For(ctx).CompanyID)

		if err != nil {
			log.Warn(err)
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		txg := tx.NewGoroutine()
		ctxg := newrelic.NewContext(ctx, txg)

		client, err = model.GetClient(ctxg, r.DB, r.GqlClient, invoice.ClientID)
		if err != nil {
			log.Warn(err)
		}
	}()
	wg.Wait()

	// get next number from postgres
	var nextNumber int64
	if err := r.DB.Raw("SELECT last_value FROM invoice_number_" + strconv.FormatInt(int64(auth.For(ctx).CompanyID), 10)).Scan(&nextNumber).Error; err != nil {
		nextNumber = 1
	} else {
		nextNumber++ // use next in sequence
	}

	return helper.RenderInvoice(&model.InvoiceTemplateData{
		Number:  int(nextNumber),
		Items:   invoice.Items,
		Client:  client,
		Company: company,
	})
}

// Invoice returns generated.InvoiceResolver implementation.
func (r *Resolver) Invoice() generated.InvoiceResolver { return &invoiceResolver{r} }

// LineItem returns generated.LineItemResolver implementation.
func (r *Resolver) LineItem() generated.LineItemResolver { return &lineItemResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type invoiceResolver struct{ *Resolver }
type lineItemResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
