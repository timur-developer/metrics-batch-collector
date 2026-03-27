package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(eventService EventService) http.Handler {
	router := chi.NewRouter()
	router.Method(http.MethodPost, "/events", NewEventHandler(eventService))
	router.Get("/healthz", healthzHandler)
	router.Get("/metrics", metricsHandler)
	router.NotFound(notFoundHandler)

	return router
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}
