package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/shamssahal/toll-calculator/types"
	"github.com/sirupsen/logrus"
)

func writeJSON(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Context-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
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
		writeJSON(w, http.StatusOK, map[string]types.Invoice{
			"data": *invoice,
		})
	}
}
func makeHTTPTransportLayer(httpListenAddr string, svc Aggregator) {
	fmt.Println("Starting distance aggregator...")
	timeout := time.Second * 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /aggregate", handleAggregate(svc))
	mux.HandleFunc("GET /invoice", handleGetInvoice(svc))
	srv := &http.Server{
		Addr:              httpListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// only to publish fatal errors
	srvErrCh := make(chan error, 1)
	defer close(srvErrCh)

	go func() {
		srvErrCh <- srv.ListenAndServe()

	}()

	select {
	case sig := <-sigCh:
		logrus.Info("Recived interruption signal. Shutting down gracefully, signal:", sig)
	case err := <-srvErrCh:
		logrus.Errorf("Received unrecoverable error. Restarting gracefully %v", err)
	}
	gracefulShutdown(ctx, timeout, srv)
}

func main() {

	var (
		httpListenAddr = ":3000"
		store          = NewMemoryStore()
		svc            = NewInvoiceAggregator(store)
	)
	svc = NewLogMiddleware(svc)
	makeHTTPTransportLayer(httpListenAddr, svc)

}

func gracefulShutdown(ctx context.Context, timeout time.Duration, srv *http.Server) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := srv.Shutdown(ctx); err != nil {
			logrus.Errorf("Failed to gracefully shutdown server %v", err)
		}
	}()
	wg.Wait()
	logrus.Info("Graceful shutdown complete")

}
