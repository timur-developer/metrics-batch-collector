package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	appmetrics "metrics-batch-collector/internal/metrics"
)

func NewRouter(eventService EventService, registry *appmetrics.Registry) http.Handler {
	router := chi.NewRouter()
	router.Method(http.MethodPost, "/events", instrumentHandler(registry, "/events", NewEventHandler(eventService, registry)))
	router.Method(http.MethodGet, "/healthz", instrumentHandler(registry, "/healthz", http.HandlerFunc(healthzHandler)))
	router.Method(http.MethodGet, "/metrics", instrumentHandler(registry, "/metrics", metricsHandler(registry)))
	router.NotFound(notFoundHandler)

	return router
}

func instrumentHandler(registry *appmetrics.Registry, path string, next http.Handler) http.Handler {
	if registry == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		startedAt := time.Now()

		next.ServeHTTP(recorder, r)

		registry.ObserveHTTPRequest(r.Method, path, recorder.statusCode, time.Since(startedAt))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}
