package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/opospisil/grpc-microservices-excercise/aggregator/client"
	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/opospisil/grpc-microservices-excercise/proto"
	"github.com/sirupsen/logrus"
)

type DataConsumer interface {
	Start()
}

type KafkaConsumer struct {
	consumer  *kafka.Consumer
	topic     string
	isRunning bool
	svc       CalculatorService
	cache     DataCache
	ac        client.AggClient
}

func (kc *KafkaConsumer) Start() {
	logrus.Info("Starting consumer")
	kc.isRunning = true

	for kc.isRunning {
		msg, err := kc.consumer.ReadMessage(-1)
		if err != nil {
			logrus.WithError(err).Error("Error reading message")
			continue
		}

		var data model.OBUData
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			logrus.Errorf("Error unmarshalling message: %+v", err)
			continue
		}

		kc.cache.Add(&data)
		pairs, err := kc.cache.PopPairs()
		if err != nil {
			logrus.Errorf("Error popping pairs: %+v", err)
			continue
		}

		for _, pair := range pairs {
			dist, err := kc.svc.CalculateDistance(pair)
			if err != nil {
				logrus.Errorf("Error calculating distance: %+v", err)
				continue
			}

			distance := proto.AggregateDistanceRequest{ 
				Value:     dist,
				Timestamp: time.Now().Unix(),
				ObuID:     pair.Current.OBUID,
			}

			if err := kc.ac.AggregateDistance(context.Background(), &distance); err != nil {
				logrus.Errorf("Error aggregating distance: %+v", err)
				continue
			}
		}
	}
}

func NewKafkaConsumer(topic string, svc CalculatorService, cache DataCache, aggClient client.AggClient) (DataConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	c.SubscribeTopics([]string{topic}, nil)

	return &KafkaConsumer{
		consumer: c,
		topic:    topic,
		svc:      svc,
		cache:    cache,
    ac:       aggClient,
  }, nil
}
