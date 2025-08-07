package main

import (
	"fmt"
	"log"

	"github.com/shamssahal/toll-calculator/aggregator/client"
)

const (
	kafkaTopic         = "obudata"
	aggregatorEndpoint = "http://127.0.0.1:3000/aggregate"
)

func main() {
	var (
		err           error
		svc           CalculatorServicer
		kafkaConsumer *KafkaConsumer
	)
	svc = NewCalculatorService()
	svc = NewLogMiddleware(svc)
	kafkaConsumer, err = NewKafkaConsumer(kafkaTopic, svc, client.NewClient(aggregatorEndpoint))
	if err != nil {
		log.Fatal(err)
	}
	kafkaConsumer.Start()
	fmt.Println("Distance Calcultor service")
}
