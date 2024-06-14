package main

import "github.com/opospisil/grpc-microservices-excercise/model"

type AggregatorService interface {
AggregateDistance(model.Distance) error
}

type InvoiceAggregator struct {
  distanceRepo DistanceRepository
}

func NewInvoiceAggregator(distanceRepo DistanceRepository) AggregatorService {
  return &InvoiceAggregator{
    distanceRepo: distanceRepo,
  }
}

func (ia *InvoiceAggregator) AggregateDistance(distance model.Distance) error {
  return ia.distanceRepo.Store(distance)
}
