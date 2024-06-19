package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/log"
	s "github.com/opospisil/grpc-microservices-excercise/aggregator_gokit/service"
	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type Set struct {
	AggregateEndpoint        endpoint.Endpoint
	CalculateInvoiceEndpoint endpoint.Endpoint
}

type AggregateRequest struct {
	Value     float64 `json:"value"`
	OBUID     int64   `json:"obuid"`
	Timestamp int64   `json:"timestamp"`
}

type AggregateResponse struct {
	Error error `json:"error,omitempty"`
}

type CalculateInvoiceRequest struct {
	OBUID int64 `json:"obuid"`
}

type CalculateInvoiceResponse struct {
	OBUID    int64   `json:"obuid"`
	Amount   float64 `json:"amount"`
	DateTime string  `json:"dateTime"`
	Distance float64 `json:"distance"`
	Error    error   `json:"error,omitempty"`
}

func (s Set) AggregateDistance(ctx context.Context, dst *model.Distance) error {
	_, err := s.AggregateEndpoint(ctx, AggregateRequest{
		Value:     dst.Value,
		OBUID:     dst.OBUID,
		Timestamp: dst.Timestamp,
	})
	if err != nil {
		return err
	}
	return nil
}

func MakeAggregateEndpoint(svc s.AggregatorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AggregateRequest)
		err := svc.AggregateDistance(ctx, &model.Distance{
			Value:     req.Value,
			OBUID:     req.OBUID,
			Timestamp: req.Timestamp,
		})
		return AggregateResponse{Error: err}, nil
	}
}

func (s Set) GetInvoice(ctx context.Context, obuid int64) (*model.Invoice, error) {
	resp, err := s.CalculateInvoiceEndpoint(ctx, CalculateInvoiceRequest{obuid})
	if err != nil {
		return nil, err
	}

	result := resp.(CalculateInvoiceResponse)
	if result.Error != nil {
		return nil, result.Error
	}
	return &model.Invoice{
		OBUID:    result.OBUID,
		Amount:   result.Amount,
		DateTime: result.DateTime,
	}, nil
}

func MakeCalculateInvoiceEndpoint(svc s.AggregatorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CalculateInvoiceRequest)
		invoice, err := svc.GetInvoice(ctx, req.OBUID)
		if err != nil {
			return CalculateInvoiceResponse{Error: err}, nil
		}
		return CalculateInvoiceResponse{
			OBUID:    invoice.OBUID,
			Amount:   invoice.Amount,
			DateTime: invoice.DateTime,
		}, nil
	}
}

func NewSet(svc s.AggregatorService, logger log.Logger) Set {
	var aggEndpoint endpoint.Endpoint
	{
		aggEndpoint = MakeAggregateEndpoint(svc)
		aggEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(aggEndpoint)
		aggEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(aggEndpoint)
		aggEndpoint = LoggingMiddleware(log.With(logger, "method", "Aggregate"))(aggEndpoint)
	}

	var invoiceEndpoint endpoint.Endpoint
	{
		invoiceEndpoint = MakeCalculateInvoiceEndpoint(svc)
		invoiceEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(invoiceEndpoint)
		invoiceEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(invoiceEndpoint)
		invoiceEndpoint = LoggingMiddleware(log.With(logger, "method", "Invoice"))(invoiceEndpoint)
	}
	return Set{
		AggregateEndpoint:        aggEndpoint,
		CalculateInvoiceEndpoint: invoiceEndpoint,
	}
}

func LoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				logger.Log("transport_error", err, "took", time.Since(begin))
			}(time.Now())
			return next(ctx, request)
		}
	}
}

func InstrumentingMiddleware(duration metrics.Histogram) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				duration.With("success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
			}(time.Now())
			return next(ctx, request)
		}
	}
}
