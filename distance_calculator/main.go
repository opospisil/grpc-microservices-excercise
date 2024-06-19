package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/opospisil/grpc-microservices-excercise/aggregator/client"
	"github.com/sirupsen/logrus"
)

type DistanceCalculator struct{}

func main() {
	if err := godotenv.Load(); err != nil {
		logrus.Fatal("Error loading .env file")
	}

	var (
		kafkaTopic            = os.Getenv("OBU_KAFKA_TOPIC")
		aggClientGrpcEndpoint = os.Getenv("AGG_GRPC_ADDR")
	)

	svc := NewCalculatorService()
	svc = NewLogMiddleware(svc)
	cache := NewDataCache()
	// aggHttpClient := client.NewHttpClient(aggClientHttpEndpoint)
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
