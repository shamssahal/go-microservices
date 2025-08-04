package main

import (
	"fmt"
	"log"
)

const kafkaTopic = "obudata"

func main() {
	var (
		err           error
		svc           CalculatorServicer
		kafkaConsumer *KafkaConsumer
	)
	svc = NewCalculatorService()
	kafkaConsumer, err = NewKafkaConsumer(kafkaTopic, svc)
	if err != nil {
		log.Fatal(err)
	}
	kafkaConsumer.Start()
	fmt.Println("Distance Calcultor service")
}
