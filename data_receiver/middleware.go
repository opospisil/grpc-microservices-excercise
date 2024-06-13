package main

import (
	"time"

	"github.com/opospisil/grpc-microservices-excercise/model"
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

func (lm *LogMiddleware) Produce(data model.OBUData) error {
	defer func(start time.Time) {
		logrus.WithFields(logrus.Fields{
			"obuid":    data.OBUID,
			"lat":      data.Lat,
			"lon":      data.Lon,
			"duration": time.Since(start),
		}).Info("Producing data")
	}(time.Now())
	return lm.next.Produce(data)
}
