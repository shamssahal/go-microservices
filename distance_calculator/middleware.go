package main

import (
	"time"

	"github.com/shamssahal/toll-calculator/types"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next CalculatorServicer
}

func NewLogMiddleware(next CalculatorServicer) CalculatorServicer {
	return &LogMiddleware{
		next: next,
	}
}

func (m *LogMiddleware) CalculateDistance(data types.OBUData) (dist float64, err error) {
	defer func() {
		start := time.Now()
		logrus.WithFields(logrus.Fields{
			"took":     time.Since(start),
			"err":      err,
			"distance": dist,
		}).Info("calculate distance")
	}()
	dist, err = m.next.CalculateDistance(data)
	return
}
