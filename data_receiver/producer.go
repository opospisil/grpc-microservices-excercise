package main

import (
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/sirupsen/logrus"
)

type DataProducer interface {
	Produce(data model.OBUData) error
}

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

func (kp *KafkaProducer) Produce(data model.OBUData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kp.topic,
			Partition: kafka.PartitionAny,
		},
		Value: b,
	}, nil)
}

func NewKafkaProducer(topic string) (DataProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		return nil, err
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logrus.Errorf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					logrus.Infof("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()
	return &KafkaProducer{
		producer: p,
		topic:    topic,
	}, nil
}
