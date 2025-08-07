package main

import (
	"context"
	"time"

	"github.com/shamssahal/toll-calculator/types"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next Aggregator
}

func NewLogMiddleware(next Aggregator) Aggregator {
	return &LogMiddleware{
		next: next,
	}
}

func (m *LogMiddleware) AggregateDistance(ctx context.Context, distance types.Distance) (err error) {
	defer func() {
		start := time.Now()
		logrus.WithFields(logrus.Fields{
			"took":     time.Since(start),
			"err":      err,
			"distance": distance,
		}).Info("Aggregate distance")
	}()
	err = m.next.AggregateDistance(ctx, distance)
	return
}

func (m *LogMiddleware) CalculateInvoice(ctx context.Context, obuID int) (inv *types.Invoice, err error) {
	defer func() {
		var (
			start  = time.Now()
			fields = logrus.Fields{
				"took": time.Since(start),
				"err":  err,
			}
		)
		if inv != nil {
			fields["OBUID"] = inv.OBUID
			fields["totalDist"] = inv.TotalDistance
			fields["totalAmount"] = inv.TotalAmount
		}
		logrus.WithFields(fields).Info("Calculated Invoice: ")
	}()
	inv, err = m.next.CalculateInvoice(ctx, obuID)
	return
}
