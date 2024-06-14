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

func (lm *LogMiddleware) AggregateDistance(distance *model.Distance) (err error) {
	defer func(start time.Time) {
		logrus.WithFields(logrus.Fields{
			"took":     time.Since(start),
			"error":    err,
			"distance": distance,
		}).Info("Aggregating distance")
	}(time.Now())
	err = lm.next.AggregateDistance(distance)
	return
}

func (lm *LogMiddleware) GetInvoice(obuid int) (invoice *model.Invoice, err error) {
	defer func(start time.Time) {
    if err != nil {
      logrus.WithFields(logrus.Fields{
        "took":  time.Since(start),
        "error": err,
        "obuid": obuid,
      }).Error("Generating invoice")
      return
    }
		logrus.WithFields(logrus.Fields{
			"took":  time.Since(start),
			"obuid": obuid,
		}).Info("Generating invoice")
	}(time.Now())
	invoice, err = lm.next.GetInvoice(obuid)
	return
}
