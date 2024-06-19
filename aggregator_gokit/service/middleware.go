package service

import (
	"context"

	"github.com/opospisil/grpc-microservices-excercise/model"
)

type Middleware func(AggregatorService) AggregatorService

type aggLoggingMiddleware struct {
	next AggregatorService
}

type aggMetricsMiddleware struct {
	next AggregatorService
}

func newAggLogggingMiddleware() Middleware {
	return func(next AggregatorService) AggregatorService {
		return &aggLoggingMiddleware{next}
	}
}

func (mw *aggLoggingMiddleware) AggregateDistance(ctx context.Context, dst *model.Distance) error {
	return mw.next.AggregateDistance(ctx, dst)
}

func (mw *aggLoggingMiddleware) GetInvoice(ctx context.Context, obuid int64) (*model.Invoice, error) {
	return mw.next.GetInvoice(ctx, obuid)
}

func newAggMetricsMiddleware() Middleware {
	return func(next AggregatorService) AggregatorService {
		return &aggMetricsMiddleware{next}
	}
}

func (mw *aggMetricsMiddleware) AggregateDistance(ctx context.Context, dst *model.Distance) error {
	return mw.next.AggregateDistance(ctx, dst)
}

func (mw *aggMetricsMiddleware) GetInvoice(ctx context.Context, obuid int64) (*model.Invoice, error) {
	return mw.next.GetInvoice(ctx, obuid)
}

