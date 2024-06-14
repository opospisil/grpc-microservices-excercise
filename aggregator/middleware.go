package main

import (
	"time"

	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next AggregatorService
}

func NewLogMiddleware(next AggregatorService) AggregatorService {
	return &LogMiddleware{
		next: next,
	}
}

func (lm *LogMiddleware) AggregateDistance(distance model.Distance) (err error) {
	defer func(start time.Time) {
		logrus.WithFields(logrus.Fields{
			"took":  time.Since(start),
			"error": err,
      "distance": distance,
		}).Info("Aggregated distance")
	}(time.Now())
	err = lm.next.AggregateDistance(distance)
	return
}
