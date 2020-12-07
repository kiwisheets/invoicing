package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/emvi/hide"
	"github.com/google/uuid"
	"github.com/kiwisheets/auth"
	"github.com/kiwisheets/invoicing/graphql/generated"
	"github.com/kiwisheets/invoicing/model"
	"github.com/kiwisheets/util"
)

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

func (r *mutationResolver) CreateInvoice(ctx context.Context, invoice model.CreateInvoiceInput) (*model.Invoice, error) {
	lineItems := make([]model.LineItem, 0)

	for _, l := range invoice.Items {
		lineItems = append(lineItems, model.LineItem{
			Description: l.Description,
			Name:        l.Name,
			Quantity:    l.Quantity,
			TaxRate:     l.TaxRate,
			UnitCost:    l.UnitCost,
		})
	}

	newInvoice := &model.Invoice{
		Client:    invoice.ClientID,
		CreatedBy: auth.For(ctx).UserID,
		LineItems: lineItems,
		Number:    1,
	}

	r.DB.Create(&newInvoice)

	return newInvoice, nil
}

func (r *mutationResolver) CreateInvoicePdf(ctx context.Context, id hide.ID) (string, error) {
	var invoice model.Invoice
	r.DB.Where(id).Where("company_id = ?", auth.For(ctx).CompanyID).Find(&invoice)

	notifyID := uuid.New()

	msg, err := json.Marshal(model.InvoiceRenderMQ{
		Invoice: invoice,
		NotifyConfig: model.Notify{
			Users: []int64{
				int64(auth.For(ctx).UserID),
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

// Invoice returns generated.InvoiceResolver implementation.
func (r *Resolver) Invoice() generated.InvoiceResolver { return &invoiceResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type invoiceResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
