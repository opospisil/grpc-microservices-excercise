package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	c "github.com/opospisil/grpc-microservices-excercise/aggregator_gokit/http/controllers"
	"github.com/opospisil/grpc-microservices-excercise/aggregator_gokit/service"
)

func NewHttpHandler(endpoints c.Set, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	m := http.NewServeMux()
	m.Handle("/aggregate", httptransport.NewServer(
		endpoints.AggregateEndpoint,
		decodeHTTPAggregateRequest,
		encodeHTTPGenericResponse,
		options...,
	))
	m.Handle("/invoice", httptransport.NewServer(
		endpoints.CalculateInvoiceEndpoint,
		decodeHTTPCalculateInvoiceRequest,
		encodeHTTPGenericResponse,
		options...,
	))
	return m
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}

func decodeHTTPAggregateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req c.AggregateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// decodeHTTPConcatRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded concat request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPCalculateInvoiceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req c.CalculateInvoiceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPCalcualteInvoiceResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp c.CalculateInvoiceResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// decodeHTTPConcatResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded concat response from the HTTP response body. If the response
// has a non-200 status code, we will interpret that as an error and attempt to
// decode the specific error message from the response body. Primarily useful in
// a client.
func decodeHTTPAggregateResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp c.AggregateResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// encodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func encodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = io.NopCloser(&buf)
	return nil
}

// encodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func NewHTTPClient(instance string, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) (service.AggregatorService, error) {
	// Quickly sanitize the instance string.
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.

	// global client middlewares
	var options []httptransport.ClientOption

	if zipkinTracer != nil {
		// Zipkin HTTP Client Trace can either be instantiated per endpoint with a
		// provided operation name or a global tracing client can be instantiated
		// without an operation name and fed to each Go kit endpoint as ClientOption.
		// In the latter case, the operation name will be the endpoint's http method.
		options = append(options, zipkin.HTTPClientTrace(zipkinTracer))
	}

	// Each individual endpoint is an http/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	var aggEndpoint endpoint.Endpoint
	{
		aggEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/aggregate"),
			encodeHTTPGenericRequest,
			decodeHTTPAggregateResponse,
			append(options, httptransport.ClientBefore(opentracing.ContextToHTTP(otTracer, logger)))...,
		).Endpoint()
		aggEndpoint = opentracing.TraceClient(otTracer, "Aggregate")(aggEndpoint)
		if zipkinTracer != nil {
			aggEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Aggregate")(aggEndpoint)
		}
		aggEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Sum",
			Timeout: 30 * time.Second,
		}))(aggEndpoint)
	}

	// The Concat endpoint is the same thing, with slightly different
	// middlewares to demonstrate how to specialize per-endpoint.
	var calcInvoiceEndpoint endpoint.Endpoint
	{
		calcInvoiceEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/invoice"),
			encodeHTTPGenericRequest,
			decodeHTTPCalcualteInvoiceResponse,
			append(options, httptransport.ClientBefore(opentracing.ContextToHTTP(otTracer, logger)))...,
		).Endpoint()
		calcInvoiceEndpoint = opentracing.TraceClient(otTracer, "Concat")(calcInvoiceEndpoint)
		if zipkinTracer != nil {
			calcInvoiceEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Concat")(calcInvoiceEndpoint)
		}
		calcInvoiceEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Concat",
			Timeout: 10 * time.Second,
		}))(calcInvoiceEndpoint)
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return c.Set{
		AggregateEndpoint:        aggEndpoint,
		CalculateInvoiceEndpoint: calcInvoiceEndpoint,
	}, nil
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}
