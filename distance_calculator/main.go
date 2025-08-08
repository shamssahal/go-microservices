package main

import (
	"fmt"
	"log"

	"github.com/shamssahal/toll-calculator/aggregator/client"
)

const (
	kafkaTopic             = "obudata"
	httpAggregatorEndpoint = "http://127.0.0.1:3000"
	grpcAggregatorEndpoint = "127.0.0.1:3001"
)

func main() {
	var (
		err           error
		svc           CalculatorServicer
		kafkaConsumer *KafkaConsumer
	)
	svc = NewCalculatorService()
	svc = NewLogMiddleware(svc)
	// httpClient := client.NewHTTPClient(aggregatorEndpoint)
	grpcClient, err := client.NewGRPCClient(grpcAggregatorEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	kafkaConsumer, err = NewKafkaConsumer(kafkaTopic, svc, grpcClient)
	if err != nil {
		log.Fatal(err)
	}
	kafkaConsumer.Start()
	fmt.Println("Distance Calcultor service")
}
