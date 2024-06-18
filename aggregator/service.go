package main

import (
	"time"

	"github.com/opospisil/grpc-microservices-excercise/model"
)

const (
	DDMMYYYYhhmmss = "2006-01-02 15:04:05"
	basePrice      = 0.13
)

type AggregatorService interface {
	AggregateDistance(*model.Distance) error
	GetInvoice(obuid int64) (*model.Invoice, error)
}

type InvoiceAggregator struct {
	distanceRepo DistanceRepository
}

func NewInvoiceAggregator(distanceRepo DistanceRepository) AggregatorService {
	return &InvoiceAggregator{
		distanceRepo: distanceRepo,
	}
}

func (ia *InvoiceAggregator) AggregateDistance(distance *model.Distance) error {
	return ia.distanceRepo.Store(distance)
}

func (ia *InvoiceAggregator) GetInvoice(obuid int64) (*model.Invoice, error) {
	retrievedDist, err := ia.distanceRepo.Get(obuid)
	if err != nil {
		return &model.Invoice{}, err
	}

	return &model.Invoice{
		OBUID:    obuid,
		Amount:   retrievedDist * basePrice,
		DateTime: time.Now().UTC().Format(DDMMYYYYhhmmss),
	}, nil
}
