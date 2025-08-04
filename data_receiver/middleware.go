package main

import (
	"time"

	"github.com/shamssahal/toll-calculator/types"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next DataProducer
}

func NewLogMiddleware(next DataProducer) *LogMiddleware {
	return &LogMiddleware{
		next: next,
	}
}

func (l *LogMiddleware) ProduceData(data types.OBUData) error {
	defer func() {
		start := time.Now()
		logrus.WithFields(logrus.Fields{
			"obuID":     data.OBUID,
			"currLat":   data.CurrLat,
			"currLong":  data.CurrLong,
			"prevLat":   data.PrevLat,
			"prevLong":  data.PrevLong,
			"timestamp": start,
			"took":      time.Since(start),
		}).Info("producing to kafka")
	}()
	return l.next.ProduceData(data)
}

func (l *LogMiddleware) Flush(timeout int) {
	l.next.Flush(timeout)
}

func (l *LogMiddleware) Close() {
	l.next.Close()
}
