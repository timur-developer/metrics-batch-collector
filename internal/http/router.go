package http

import (
	"encoding/json"
	"net/http"
)

func NewMux(eventService EventService) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/events", NewEventHandler(eventService))
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/metrics", metricsHandler)
	mux.HandleFunc("/", notFoundHandler)

	return mux
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}
