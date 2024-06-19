package service

import (
	"context"
	"time"

	r "github.com/opospisil/grpc-microservices-excercise/aggregator_gokit/repository"
	"github.com/opospisil/grpc-microservices-excercise/model"
)

const (
	DDMMYYYYhhmmss = "2006-01-02 15:04:05"
	basePrice      = 0.13
)

type AggregatorService interface {
	AggregateDistance(ctx context.Context, dst *model.Distance) error
	GetInvoice(ctx context.Context, obuid int64) (*model.Invoice, error)
}

type AggregatorServiceImpl struct {
	distanceRepo r.DistanceRepository
}

func NewAggregatorServiceImp(repo r.DistanceRepository) AggregatorService {
	return &AggregatorServiceImpl{
		distanceRepo: repo,
	}
}

func (svc *AggregatorServiceImpl) AggregateDistance(ctx context.Context, dst *model.Distance) error {
	return svc.distanceRepo.Store(dst)
}

func (svc *AggregatorServiceImpl) GetInvoice(ctx context.Context, obuid int64) (*model.Invoice, error) {
	retrievedDist, err := svc.distanceRepo.Get(obuid)
	if err != nil {
		return &model.Invoice{}, err
	}

	return &model.Invoice{
		OBUID:    obuid,
		Amount:   retrievedDist * basePrice,
		DateTime: time.Now().UTC().Format(DDMMYYYYhhmmss),
	}, nil
}

func NewAggregatorService(repo r.DistanceRepository) AggregatorService {
	var svc AggregatorService
	svc = NewAggregatorServiceImp(repo)
	svc = newAggLogggingMiddleware()(svc)
	svc = newAggMetricsMiddleware()(svc)
	return svc
}
