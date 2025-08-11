package main

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shamssahal/toll-calculator/types"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next Aggregator
}

type MetricsMiddleware struct {
	reqCounterAgg  prometheus.Counter
	reqCounterCalc prometheus.Counter
	errCounterAgg  prometheus.Counter
	errCounterCalc prometheus.Counter
	reqLatencyAgg  prometheus.Histogram
	reqLatencyCalc prometheus.Histogram

	next Aggregator
}

func NewLogMiddleware(next Aggregator) Aggregator {
	return &LogMiddleware{
		next: next,
	}
}

func NewMetricsMiddleware(next Aggregator) Aggregator {
	reqCounterAgg := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "aggregator",
		Name:      "request_counter",
	})
	reqCounterCalc := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "caclulator",
		Name:      "request_counter",
	})
	errCounterAgg := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "aggregator",
		Name:      "error_counter",
	})
	errCounterCalc := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "caclulator",
		Name:      "error_counter",
	})
	reqLatencyAgg := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "aggregator",
		Name:      "request_latency",
		Buckets:   []float64{0.1, 0.5, 1},
	})
	reqLatencyCalc := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "calculator",
		Name:      "request_latency",
		Buckets:   []float64{0.1, 0.5, 1},
	})
	return &MetricsMiddleware{
		next:           next,
		reqCounterAgg:  reqCounterAgg,
		reqCounterCalc: reqCounterCalc,
		errCounterAgg:  errCounterAgg,
		errCounterCalc: errCounterCalc,
		reqLatencyAgg:  reqLatencyAgg,
		reqLatencyCalc: reqLatencyCalc,
	}
}

func (m *LogMiddleware) AggregateDistance(ctx context.Context, distance types.Distance) (err error) {
	defer func(start time.Time) {
		logrus.WithFields(logrus.Fields{
			"took":      time.Since(start),
			"err":       err,
			"distance":  distance,
			"requestId": distance.RequestID,
		}).Info("Aggregate distance")
	}(time.Now())
	err = m.next.AggregateDistance(ctx, distance)
	return
}

func (m *LogMiddleware) CalculateInvoice(ctx context.Context, obuID int) (inv *types.Invoice, err error) {
	defer func(start time.Time) {
		var (
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
	}(time.Now())
	inv, err = m.next.CalculateInvoice(ctx, obuID)
	return
}

func (m *MetricsMiddleware) AggregateDistance(ctx context.Context, distance types.Distance) (err error) {
	defer func(start time.Time) {
		m.reqLatencyAgg.Observe(time.Since(start).Seconds())
		m.reqCounterAgg.Inc()
		if err != nil {
			m.errCounterAgg.Inc()
		}
	}(time.Now())
	err = m.next.AggregateDistance(ctx, distance)
	return
}

func (m *MetricsMiddleware) CalculateInvoice(ctx context.Context, obuID int) (inv *types.Invoice, err error) {
	defer func(start time.Time) {
		m.reqLatencyCalc.Observe(time.Since(start).Seconds())
		m.reqCounterCalc.Inc()
		if err != nil {
			m.errCounterCalc.Inc()
		}
	}(time.Now())
	inv, err = m.next.CalculateInvoice(ctx, obuID)
	return
}
