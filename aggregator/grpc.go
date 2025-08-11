package main

import (
	"context"

	"github.com/shamssahal/toll-calculator/types"
)

type GRPCAggregatorServer struct {
	types.UnimplementedAggregatorServer
	svc Aggregator
}

func (s *GRPCAggregatorServer) Aggregate(ctx context.Context, req *types.AggregateRequest) (*types.None, error) {
	distance := types.Distance{
		OBUID:     int(req.ObuID),
		Value:     float64(req.Value),
		Unix:      int64(req.Unix),
		RequestID: string(req.RequestID),
	}
	err := s.svc.AggregateDistance(ctx, distance)
	if err != nil {
		return nil, err
	}
	return &types.None{}, nil
}

func NewGRPCServer(svc Aggregator) *GRPCAggregatorServer {
	return &GRPCAggregatorServer{
		svc: svc,
	}
}
