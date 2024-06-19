package main

import (
	"net/http"
	"time"

	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next AggregatorService
}

type MetricsMiddleware struct {
	aggCounter    prometheus.Counter
	invCounter    prometheus.Counter
	errAggCounter prometheus.Counter
	errInvCounter prometheus.Counter
	reqLatency    prometheus.Histogram
	next          AggregatorService
}

type HttpMetricsMiddleware struct {
	next              AggHttpHandler
	invoiceReqCounter prometheus.Counter
	invoiceErrCounter prometheus.Counter
}

func NewHttpMetricsMiddleware(next AggHttpHandler) AggHttpHandler {
	return &HttpMetricsMiddleware{
		next: next,
		invoiceReqCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "http_invoices_total",
			Namespace: "aggregator",
			Help:      "Total number of invoice requests coming from HTTP",
		}),
		invoiceErrCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "http_invoices_error_total",
			Namespace: "aggregator",
			Help:      "Total number of errors in invoice requests coming from HTTP",
		}),
	}
}

func (hm *HttpMetricsMiddleware) HandleGetInvoice() apiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		hm.invoiceReqCounter.Inc()
		err := hm.next.HandleGetInvoice()(w, r)
		if err != nil {
			logrus.Errorf("Vyjeb handling request: %+v", err)
			hm.invoiceErrCounter.Inc()
		}
		return err
	}
}

func (hm *HttpMetricsMiddleware) HandleAggregate() apiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return hm.next.HandleAggregate()(w, r)
	}
}

func NewLogMiddleware(next AggregatorService) AggregatorService {
	return &LogMiddleware{
		next: next,
	}
}

func NewMetricsMiddleware(next AggregatorService) AggregatorService {
	return &MetricsMiddleware{
		aggCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "aggregates_total",
			Namespace: "aggregator",
			Help:      "Total number of requests",
		}),
		invCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "invoices_total",
			Namespace: "aggregator",
			Help:      "Total number of invoice requests",
		}),
		errAggCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "aggregates_error_total",
			Namespace: "aggregator",
			Help:      "Total number of errors in aggregate requests",
		}),
		errInvCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "invoices_error_total",
			Namespace: "aggregator",
			Help:      "Total number of errors in invoice requests",
		}),
		reqLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:      "request_latency_seconds",
			Namespace: "aggregator",
			Help:      "Request latency in seconds",
		}),
		next: next,
	}
}

func (mm *MetricsMiddleware) AggregateDistance(distance *model.Distance) (err error) {
	start := time.Now()
	defer func() {
		mm.aggCounter.Inc()
		mm.reqLatency.Observe(time.Since(start).Seconds())
		if err != nil {
			mm.errAggCounter.Inc()
		}
	}()
	return mm.next.AggregateDistance(distance)
}

func (mm *MetricsMiddleware) GetInvoice(obuid int64) (invoice *model.Invoice, err error) {
	start := time.Now()
	defer func() {
		mm.invCounter.Inc()
		mm.reqLatency.Observe(time.Since(start).Seconds())
		if err != nil {
			mm.errInvCounter.Inc()
		}
	}()
	return mm.next.GetInvoice(obuid)
}

func (lm *LogMiddleware) AggregateDistance(distance *model.Distance) (err error) {
	defer func(start time.Time) {
		logrus.WithFields(logrus.Fields{
			"took":     time.Since(start),
			"error":    err,
			"distance": distance,
		}).Info("Aggregating distance")
	}(time.Now())
	err = lm.next.AggregateDistance(distance)
	return
}

func (lm *LogMiddleware) GetInvoice(obuid int64) (invoice *model.Invoice, err error) {
	defer func(start time.Time) {
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"took":  time.Since(start),
				"error": err,
				"obuid": obuid,
			}).Error("Generating invoice")
			return
		}
		logrus.WithFields(logrus.Fields{
			"took":  time.Since(start),
			"obuid": obuid,
		}).Info("Generating invoice")
	}(time.Now())
	invoice, err = lm.next.GetInvoice(obuid)
	return
}
