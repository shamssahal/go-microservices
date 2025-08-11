package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shamssahal/toll-calculator/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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
		writeJSON(w, http.StatusOK, invoice)
	}
}

func makeHTTPTransportLayer(httpListenAddr string, svc Aggregator) {
	fmt.Printf("Starting distance aggregator HTTP Transport Layer on port %s\n", httpListenAddr)
	timeout := time.Second * 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /aggregate", handleAggregate(svc))
	mux.HandleFunc("GET /invoice", handleGetInvoice(svc))
	mux.Handle("GET /metrics", promhttp.Handler())
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

func makeGRPCTransport(listenAddr string, svc Aggregator) error {
	fmt.Printf("Starting distance aggregator gRPC Transport Layer on port %s\n", listenAddr)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	serverRegistrar := grpc.NewServer([]grpc.ServerOption{}...)
	server := NewGRPCServer(svc)
	types.RegisterAggregatorServer(serverRegistrar, server)
	return serverRegistrar.Serve(ln)
}

// Chain composes middlewares in order: Chain(A,B,C){H} => A(B(C(H))).
// All middlewares take in aggregator and return a aggregator
func Chain(a Aggregator, mws ...func(Aggregator) Aggregator) Aggregator {
	for i := len(mws) - 1; i >= 0; i-- {
		a = mws[i](a)
	}
	return a
}

func main() {

	var (
		httpListenAddr = os.Getenv("AGG_HTTP_PORT")
		grpcListenAddr = os.Getenv("AGG_GRPC_PORT")
		store          = makeStore()
		svc            = NewInvoiceAggregator(store)
	)
	svc = Chain(
		svc,
		func(s Aggregator) Aggregator { return (NewMetricsMiddleware(s)) },
		func(s Aggregator) Aggregator { return (NewLogMiddleware(s)) },
	)
	go func() {
		log.Fatal(makeGRPCTransport(grpcListenAddr, svc))
	}()
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

func makeStore() Storer {
	storeType := os.Getenv("AGG_STORE_TYPE")
	switch storeType {
	case "memory":
		return NewMemoryStore()
	default:
		log.Fatalf("invalid store type given %s", storeType)
		return nil
	}
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
}
