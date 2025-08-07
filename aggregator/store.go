package main

import (
	"context"
	"fmt"

	"github.com/shamssahal/toll-calculator/types"
)

type Storer interface {
	Insert(context.Context, types.Distance) error
	Get(context.Context, int) (float64, error)
}

type MemoryStore struct {
	data map[int]float64
}

func (m *MemoryStore) Insert(ctx context.Context, d types.Distance) error {
	m.data[d.OBUID] += d.Value
	return nil
}

func (m *MemoryStore) Get(ctx context.Context, obuID int) (float64, error) {
	if dist, ok := m.data[obuID]; !ok {
		return 0.0, fmt.Errorf("could not find data for obuId %d", obuID)
	} else {
		return dist, nil
	}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[int]float64),
	}
}
