package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/emvi/hide"
	"github.com/google/uuid"
	"github.com/kiwisheets/auth"
	gqlServerClient "github.com/kiwisheets/gql-server/client"
	"github.com/kiwisheets/invoicing/graphql/generated"
	"github.com/kiwisheets/invoicing/helper"
	"github.com/kiwisheets/invoicing/model"
	"github.com/kiwisheets/util"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

func (r *invoiceResolver) Number(ctx context.Context, obj *model.Invoice) (int64, error) {
	return obj.Number.Number, nil
}

func (r *invoiceResolver) CreatedBy(ctx context.Context, obj *model.Invoice) (*model.User, error) {
	return &model.User{
		ID: obj.CreatedBy,
	}, nil
}

func (r *invoiceResolver) Client(ctx context.Context, obj *model.Invoice) (*model.Client, error) {
	return &model.Client{
		ID: obj.Client,
	}, nil
}

func (r *invoiceResolver) Items(ctx context.Context, obj *model.Invoice) ([]*model.LineItem, error) {
	lineItems := make([]*model.LineItem, 0)
	r.DB.Where("invoice_id = ?", obj.ID).Find(&lineItems)
	return lineItems, nil
}

func (r *mutationResolver) CreateInvoice(ctx context.Context, invoice model.InvoiceInput) (*model.Invoice, error) {
	lineItems := make([]model.LineItem, len(invoice.Items))
	for i, l := range invoice.Items {
		lineItems[i] = model.LineItem{
			Description: l.Description,
			Name:        l.Name,
			Quantity:    l.Quantity,
			TaxRate:     l.TaxRate,
			UnitCost:    l.UnitCost,
		}
	}

	newInvoice := &model.Invoice{
		Client:    invoice.ClientID,
		CreatedBy: auth.For(ctx).UserID,
		CompanyID: auth.For(ctx).CompanyID,
		LineItems: lineItems,
		Number: model.InvoiceNumber{
			CompanyID: auth.For(ctx).CompanyID,
		},
	}

	r.DB.Exec("CREATE SEQUENCE IF NOT EXISTS invoice_number_" + strconv.FormatInt(int64(auth.For(ctx).CompanyID), 10) + " AS BIGINT INCREMENT 1 START 1 OWNED BY invoices.number")

	err := r.DB.Debug().Clauses(clause.Returning{
		Columns: []clause.Column{
			clause.PrimaryColumn,
			{
				Table: clause.CurrentTable,
				Name:  "NUMBER",
			},
		},
	}).Create(&newInvoice).Error
	if err != nil {
		return nil, err
	}

	return newInvoice, nil
}

func (r *mutationResolver) UpdateInvoice(ctx context.Context, invoice model.InvoiceInput) (*model.Invoice, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateInvoicePdf(ctx context.Context, id hide.ID) (string, error) {
	var invoice model.Invoice
	r.DB.Where(id).Where("company_id = ?", auth.For(ctx).CompanyID).Find(&invoice)

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
		return "", fmt.Errorf("error processing invoice, does the invoice exist?")
	}
	r.RenderProducer.Produce(msg)
	return notifyID.String(), nil
}

func (r *queryResolver) Invoice(ctx context.Context, id hide.ID) (*model.Invoice, error) {
	var invoice model.Invoice
	if err := r.DB.Where(id).Where("company_id = ?", auth.For(ctx).CompanyID).Find(&invoice).Error; err != nil {
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
	r.DB.Where("company_id = ?", auth.For(ctx).CompanyID).Limit(limit).Offset(limit * *page).Find(&invoices)

	return invoices, nil
}

func (r *queryResolver) PreviewInvoice(ctx context.Context, invoice model.PreviewInvoiceInput) (string, error) {
	// load template and exec, return html

	var client *gqlServerClient.GetClientByID
	var company *gqlServerClient.GetCompany
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		var err error
		company, err = r.GqlServerClient.GetCompany(ctx, func(req *http.Request) {
			req.Header.Set("user", auth.For(ctx).OriginalHeader)
		})
		if err != nil {
			logrus.Warn(err)
		}
	}()
	go func() {
		defer wg.Done()
		var err error
		client, err = r.GqlServerClient.GetClientByID(ctx, invoice.ClientID, func(req *http.Request) {
			req.Header.Set("user", auth.For(ctx).OriginalHeader)
		})
		if err != nil {
			logrus.Warn(err)
		}
	}()
	wg.Wait()

	return helper.RenderInvoice(&model.InvoiceTemplateData{
		Number:  invoice.Number,
		Items:   invoice.Items,
		Client:  client,
		Company: company,
	})
}

// Invoice returns generated.InvoiceResolver implementation.
func (r *Resolver) Invoice() generated.InvoiceResolver { return &invoiceResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type invoiceResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
