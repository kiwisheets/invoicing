package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/emvi/hide"
	"github.com/kiwisheets/invoicing/graphql/generated"
	"github.com/kiwisheets/invoicing/model"
)

func (r *entityResolver) FindClientByID(ctx context.Context, id hide.ID) (*model.Client, error) {
	invoices := make([]*model.Invoice, 0)
	r.DB.Where("created_by = ?", id).Find(&invoices)

	return &model.Client{
		ID:       id,
		Invoices: invoices,
	}, nil
}

func (r *entityResolver) FindCompanyByID(ctx context.Context, id hide.ID) (*model.Company, error) {
	panic(fmt.Errorf("not implemented"))
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
