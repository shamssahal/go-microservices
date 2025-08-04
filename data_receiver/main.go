package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shamssahal/toll-calculator/types"
)

const (
	kafkaBroker     = "localhost:9092"
	httpListenAddr  = ":30000"
	maxKafkaTimeout = 10_000
)

var (
	kafkaTopic = "obudata"
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

func NewDataReceiver() (*DataReceiver, error) {
	p, err := NewKafkaProducer()
	if err != nil {
		return nil, err
	}
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

func (dr *DataReceiver) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := dr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())

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

func (dr *DataReceiver) ServerHTTP(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", dr.handleWS)

	srv := &http.Server{
		Addr:              httpListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	log.Printf("data receiver listening on %s", httpListenAddr)
	return srv.ListenAndServe()
}

func main() {
	fmt.Println("Starting data receiver...")
	receiver, err := NewDataReceiver()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := receiver.ServerHTTP(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server err: %v\n", err)
	}

	receiver.prod.Flush(maxKafkaTimeout)
	receiver.prod.Close()

	fmt.Println("Graceful Shutdown Complete")

}
