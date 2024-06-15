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
  //aggHttpClient := client.NewHttpClient(aggClientEndpoint)
  aggGrpcClient, err := client.NewGRPCClient("localhost:8081")
  if err != nil {
    logrus.Fatalf("Error creating gRPC client: %v", err)
  }

	consumer, err := NewKafkaConsumer(kafkaTopic, svc, cache, aggGrpcClient)
	if err != nil {
		logrus.Fatal(err)
	}

	consumer.Start()
}
