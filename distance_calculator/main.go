package main

import (
	"github.com/opospisil/grpc-microservices-excercise/aggregator/client"
	"github.com/sirupsen/logrus"
)

const (
  kafkaTopic = "obu-data"
  aggClientHttpEndpoint = "http://localhost:8080"
  aggClientGrpcEndpoint = "localhost:8081"
)

type DistanceCalculator struct{}

func main() {
	svc := NewCalculatorService()
	svc = NewLogMiddleware(svc)
  cache := NewDataCache()
  //aggHttpClient := client.NewHttpClient(aggClientHttpEndpoint)
  aggGrpcClient, err := client.NewGRPCClient(aggClientGrpcEndpoint)
  if err != nil {
    logrus.Fatalf("Error creating gRPC client: %v", err)
  }

	consumer, err := NewKafkaConsumer(kafkaTopic, svc, cache, aggGrpcClient)
	if err != nil {
		logrus.Fatal(err)
	}

	consumer.Start()
}
