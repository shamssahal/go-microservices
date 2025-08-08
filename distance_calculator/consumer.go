package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/shamssahal/toll-calculator/aggregator/client"
	"github.com/shamssahal/toll-calculator/types"
	"github.com/sirupsen/logrus"
)

// This can also be called kafka transport
type KafkaConsumer struct {
	consumer    *kafka.Consumer
	isRunning   bool
	calcService CalculatorServicer
	aggClient   client.Client
}

func NewKafkaConsumer(topic string, svc CalculatorServicer, aggClient client.Client) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       "localhost:9092",
		"group.id":                "myGroup",
		"auto.offset.reset":       "earliest",
		"session.timeout.ms":      6000,
		"heartbeat.interval.ms":   2000,
		"max.poll.interval.ms":    300000,
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 1000,
	})
	if err != nil {
		return nil, err
	}
	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer:    c,
		calcService: svc,
		aggClient:   aggClient,
	}, nil
}

func (c *KafkaConsumer) Start() {
	logrus.Info("kafka transport started")
	c.isRunning = true
	c.readMessageLoop()
}

func (c *KafkaConsumer) readMessageLoop() {
	for c.isRunning {
		msg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			logrus.Errorf("kafka consumer error %s", err)
			continue
		}
		var data types.OBUData
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			logrus.Errorf("JSON serialization error: %s", err)
			continue
		}
		dist, err := c.calcService.CalculateDistance(data)
		if err != nil {
			logrus.Errorf("calc service error %s", err)
		}
		req := &types.AggregateRequest{
			Value: dist,
			ObuID: int64(data.OBUID),
			Unix:  time.Now().UnixNano(),
		}
		err = c.aggClient.Aggregate(context.Background(), req)
		if err != nil {
			logrus.Errorf("aggregate client failure: %v", err)
			continue
		}

	}
}
