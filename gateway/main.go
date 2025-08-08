package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/shamssahal/toll-calculator/aggregator/client"
	"github.com/shamssahal/toll-calculator/gateway/config"
	"github.com/shamssahal/toll-calculator/gateway/handler"
	"github.com/shamssahal/toll-calculator/gateway/utils"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		httpListenAddr   = flag.String("httpListenAddr", ":8000", "specify port for the API Gateway")
		readTimeout      = 5 * time.Second
		mux              = http.NewServeMux()
		aggregatorClient = &client.HTTPClient{Endpoint: config.AggregatorService}
		invoiceHandler   = handler.NewInvoiceHandler(aggregatorClient)
	)
	flag.Parse()
	srv := &http.Server{
		Addr:        *httpListenAddr,
		Handler:     mux,
		ReadTimeout: readTimeout,
	}

	mux.HandleFunc("GET /invoice", utils.MakeAPIHandler(invoiceHandler.HandleGetInvoice))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	srvErrCh := make(chan error, 1)
	defer close(srvErrCh)
	go func() {
		fmt.Printf("Starting api gateway on port %s\n", *httpListenAddr)
		srvErrCh <- srv.ListenAndServe()
	}()

	select {
	case sig := <-sigCh:
		logrus.Info("Received interruption signal. Shutting down gracefully, signal:", sig)
	case err := <-srvErrCh:
		if err != nil && err != http.ErrServerClosed {
			logrus.Errorf("Received unrecoverable error. Shutting down gracefully %v", err)
		}
	}
	gracefulShutdown(ctx, readTimeout, srv)
}

func gracefulShutdown(ctx context.Context, readTimeout time.Duration, srv *http.Server) {
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
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
