package http

import (
	"net/http"

	appmetrics "metrics-batch-collector/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func metricsHandler(registry *appmetrics.Registry) http.Handler {
	return promhttp.HandlerFor(registry.PrometheusRegistry(), promhttp.HandlerOpts{})
}
