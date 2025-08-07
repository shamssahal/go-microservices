package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shamssahal/toll-calculator/types"
	"github.com/sirupsen/logrus"
)

const (
	kafkaBroker     = "localhost:9092"
	httpListenAddr  = ":30000"
	maxKafkaTimeout = 10_000
	kafkaTopic      = "obudata"
)

type DataReceiver struct {
	upgrader websocket.Upgrader
	prod     DataProducer
}

func (dr *DataReceiver) wsReceiveLoop(ctx context.Context, conn *websocket.Conn, cancel context.CancelFunc) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var data types.OBUData
			if err := conn.ReadJSON(&data); err != nil {
				if websocket.IsCloseError(
					err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
				) {
					log.Println("Websocket connection closed")
					return
				}
				log.Println("read error : ", err)
				continue
			}
			if err := dr.produceData(data); err != nil {
				log.Printf("kafka producer err: %v\n", err)
			}
		}
	}
}

func (dr *DataReceiver) produceData(data types.OBUData) error {
	return dr.prod.ProduceData(data)
}

func (dr *DataReceiver) cleanup() {
	dr.prod.Flush(maxKafkaTimeout)
	dr.prod.Close()
}

func NewDataReceiver() (*DataReceiver, error) {
	var (
		p   DataProducer
		err error
	)
	p, err = NewKafkaProducer(kafkaTopic)
	if err != nil {
		return nil, err
	}
	p = NewLogMiddleware(p)
	return &DataReceiver{
		prod: p,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}, nil
}

// uses decorator pattern to pass the context to websocket handler
// which returns the http handler for the /ws enpoint
func (dr *DataReceiver) handleWS(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := dr.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		ctx, cancel := context.WithCancel(ctx)
		//contex watcher for canceled context
		// closes the connection the momemt
		// the passed context ctx is cancelled in receive loop
		go func() {
			<-ctx.Done()
			conn.Close()
		}()
		// receive loop to listen to
		go dr.wsReceiveLoop(ctx, conn, cancel)
	}
}

func (dr *DataReceiver) makeHTTPTransportLayer(ctx context.Context) {
	mux := http.NewServeMux()
	timeout := time.Second * 10
	mux.HandleFunc("/ws", dr.handleWS(ctx))

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
		logrus.Info("Received interruption signal. Shutting down gracefully, signal:", sig)
	case err := <-srvErrCh:
		if err != nil && err != http.ErrServerClosed {
			logrus.Errorf("Received unrecoverable error. Shutting down gracefully %v", err)
		}
	}
	
	dr.cleanup()
	gracefulShutdown(ctx, timeout, srv)
}

func main() {
	fmt.Println("Starting data receiver...")
	receiver, err := NewDataReceiver()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	receiver.makeHTTPTransportLayer(ctx)

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
