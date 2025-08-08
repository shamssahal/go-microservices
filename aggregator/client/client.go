package client

import (
	"context"

	"github.com/shamssahal/toll-calculator/types"
)

type Client interface {
	Aggregate(context.Context, *types.AggregateRequest) error
}
