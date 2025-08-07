package main

import (
	"context"
	"fmt"

	"github.com/shamssahal/toll-calculator/types"
)

const basePrice = 3.7

type Aggregator interface {
	AggregateDistance(context.Context, types.Distance) error
	CalculateInvoice(context.Context, int) (*types.Invoice, error)
}

type InvoiceAggregator struct {
	store Storer
}

func (i *InvoiceAggregator) AggregateDistance(ctx context.Context, distance types.Distance) error {
	return i.store.Insert(ctx, distance)
}

func (i *InvoiceAggregator) CalculateInvoice(ctx context.Context, obuID int) (*types.Invoice, error) {
	dist, err := i.store.Get(ctx, obuID)
	if err != nil {
		return nil, fmt.Errorf("could not find data for the obuid %d", obuID)
	}
	inv := &types.Invoice{
		OBUID:         obuID,
		TotalDistance: dist,
		TotalAmount:   basePrice * dist,
	}
	return inv, nil

}
func NewInvoiceAggregator(store Storer) Aggregator {
	return &InvoiceAggregator{
		store: store,
	}
}
