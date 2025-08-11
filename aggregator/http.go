package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shamssahal/toll-calculator/types"
)

type HTTPmetricHandler struct {
	reqCounter prometheus.Counter
	reqLatency prometheus.Histogram
}

func newHTTPMetricHandler(reqName string) *HTTPmetricHandler {
	reqCounter := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "http_request_counter",
		Name:      reqName,
	})
	reqLatency := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "http_request_latency",
		Name:      reqName,
		Buckets:   []float64{0.1, 0.5, 1},
	})
	return &HTTPmetricHandler{
		reqCounter: reqCounter,
		reqLatency: reqLatency,
	}
}

func (h *HTTPmetricHandler) instrument(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			h.reqLatency.Observe(time.Since(start).Seconds())
		}(time.Now())
		h.reqCounter.Inc()
		next(w, r)
	}
}

func handleAggregate(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var distance types.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := svc.AggregateDistance(context.Background(), distance); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{})
	}
}

func handleGetInvoice(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		obuID, err := strconv.Atoi(id)
		if err != nil {
			writeJSON(w, http.StatusBadRequest,
				map[string]string{"error": "missing or incorrect 'obuid' query parameter"})
			return
		}
		invoice, err := svc.CalculateInvoice(context.Background(), obuID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError,
				map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, invoice)
	}
}
