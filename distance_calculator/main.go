package main

import (
	"github.com/opospisil/grpc-microservices-excercise/aggregator/client"
	"github.com/sirupsen/logrus"
)

const (
  kafkaTopic = "obu-data"
  aggClientEndpoint = "http://localhost:8080"
)

type DistanceCalculator struct{}

func main() {
	svc := NewCalculatorService()
	svc = NewLogMiddleware(svc)
  cache := NewDataCache()
  aggClient := client.NewClient(aggClientEndpoint)

	consumer, err := NewKafkaConsumer(kafkaTopic, svc, cache, aggClient)
	if err != nil {
		logrus.Fatal(err)
	}

	consumer.Start()
}
