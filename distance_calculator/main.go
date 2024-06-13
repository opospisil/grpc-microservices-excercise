package main

import (
	"github.com/sirupsen/logrus"
)

const kafkaTopic = "obu-data"

type DistanceCalculator struct{}

func main() {
	svc := NewCalculatorService()
	svc = NewLogMiddleware(svc)
  cache := NewDataCache()
	consumer, err := NewKafkaConsumer(kafkaTopic, svc, cache)
	if err != nil {
		logrus.Fatal(err)
	}

	consumer.Start()
}
