package main

import (
	"time"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next CalculatorService
}

func NewLogMiddleware(next CalculatorService) *LogMiddleware {
	return &LogMiddleware{
		next: next,
	}
}

func (lm *LogMiddleware) CalculateDistance(pair CoordsPair) (dist float64, err error) {
	defer func(start time.Time) {
		logrus.WithFields(logrus.Fields{
			"took":     time.Since(start),
			"error":    err,
			"distance": dist,
      "obuidA": pair.Current.OBUID,
		}).Info("Calculating distance")
	}(time.Now())
	dist, err = lm.next.CalculateDistance(pair)
	return
}
