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
	"github.com/sirupsen/logrus"
)

type HTTPHandlerWithError func(http.ResponseWriter, *http.Request) error
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

func (h *HTTPmetricHandler) instrumentAndLog(next HTTPHandlerWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func(start time.Time) {
			latency := time.Since(start).Seconds()
			h.reqLatency.Observe(latency)
			h.reqCounter.Inc()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"latency": latency,
					"error":   err,
					"request": r.RequestURI,
				}).Error("http request error:")
			} else {
				logrus.WithFields(logrus.Fields{
					"latency": latency,
					"request": r.RequestURI,
				}).Info("received http request:")
			}
		}(time.Now())
		err = next(w, r)
	}
}

func handleAggregate(svc Aggregator) HTTPHandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		var distance types.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return err
		}
		if err := svc.AggregateDistance(context.Background(), distance); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return err
		}

		writeJSON(w, http.StatusOK, map[string]string{})
		return nil
	}
}

func handleGetInvoice(svc Aggregator) HTTPHandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := r.URL.Query().Get("id")
		obuID, err := strconv.Atoi(id)
		if err != nil {
			writeJSON(w, http.StatusBadRequest,
				map[string]string{"error": "missing or incorrect 'obuid' query parameter"})
			return err
		}
		invoice, err := svc.CalculateInvoice(context.Background(), obuID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError,
				map[string]string{"error": err.Error()})
			return err
		}
		writeJSON(w, http.StatusOK, invoice)
		return nil
	}
}
