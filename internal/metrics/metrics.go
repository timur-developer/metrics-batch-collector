package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Registry struct {
	prometheusRegistry          *prometheus.Registry
	httpRequestsTotal           *prometheus.CounterVec
	httpRequestDurationSeconds  *prometheus.HistogramVec
	eventsReceivedTotal         prometheus.Counter
	batchFlushTotal             prometheus.Counter
	batchSize                   prometheus.Gauge
	clickhouseInsertErrorsTotal prometheus.Counter
}

func NewRegistry() *Registry {
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDurationSeconds := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	eventsReceivedTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "events_received_total",
			Help: "Total number of accepted events.",
		},
	)

	batchFlushTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "batch_flush_total",
			Help: "Total number of batch flush operations.",
		},
	)

	batchSize := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "batch_size",
			Help: "Size of the last successfully flushed batch.",
		},
	)

	clickhouseInsertErrorsTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "clickhouse_insert_errors_total",
			Help: "Total number of ClickHouse insert errors.",
		},
	)

	prometheusRegistry := prometheus.NewRegistry()
	prometheusRegistry.MustRegister(
		httpRequestsTotal,
		httpRequestDurationSeconds,
		eventsReceivedTotal,
		batchFlushTotal,
		batchSize,
		clickhouseInsertErrorsTotal,
	)

	return &Registry{
		prometheusRegistry:          prometheusRegistry,
		httpRequestsTotal:           httpRequestsTotal,
		httpRequestDurationSeconds:  httpRequestDurationSeconds,
		eventsReceivedTotal:         eventsReceivedTotal,
		batchFlushTotal:             batchFlushTotal,
		batchSize:                   batchSize,
		clickhouseInsertErrorsTotal: clickhouseInsertErrorsTotal,
	}
}

func (r *Registry) PrometheusRegistry() *prometheus.Registry {
	if r == nil {
		return nil
	}

	return r.prometheusRegistry
}

func (r *Registry) HTTPRequestsTotal() *prometheus.CounterVec {
	if r == nil {
		return nil
	}

	return r.httpRequestsTotal
}

func (r *Registry) HTTPRequestDurationSeconds() *prometheus.HistogramVec {
	if r == nil {
		return nil
	}

	return r.httpRequestDurationSeconds
}

func (r *Registry) EventsReceivedTotal() prometheus.Counter {
	if r == nil {
		return nil
	}

	return r.eventsReceivedTotal
}

func (r *Registry) BatchFlushTotal() prometheus.Counter {
	if r == nil {
		return nil
	}

	return r.batchFlushTotal
}

func (r *Registry) BatchSize() prometheus.Gauge {
	if r == nil {
		return nil
	}

	return r.batchSize
}

func (r *Registry) ClickHouseInsertErrorsTotal() prometheus.Counter {
	if r == nil {
		return nil
	}

	return r.clickhouseInsertErrorsTotal
}

func (r *Registry) IncEventsReceived() {
	if r == nil {
		return
	}

	r.eventsReceivedTotal.Inc()
}

func (r *Registry) ObserveBatchFlush(size int) {
	if r == nil {
		return
	}

	r.batchFlushTotal.Inc()
	r.batchSize.Set(float64(size))
}

func (r *Registry) IncClickHouseInsertErrors() {
	if r == nil {
		return
	}

	r.clickhouseInsertErrorsTotal.Inc()
}

func (r *Registry) ObserveHTTPRequest(method, path string, status int, duration time.Duration) {
	if r == nil {
		return
	}

	statusCode := strconv.Itoa(status)
	r.httpRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
	r.httpRequestDurationSeconds.WithLabelValues(method, path, statusCode).Observe(duration.Seconds())
}
