package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metrics-batch-collector/internal/event"
	appmetrics "metrics-batch-collector/internal/metrics"
)

func TestPostEventsAccepted(t *testing.T) {
	router := NewRouter(eventServiceStub{}, appmetrics.NewRegistry())

	body := `{"event_type":"page_view","source":"landing","user_id":"u123","value":0,"created_at":"2026-03-27T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}

func TestPostEventsValidationError(t *testing.T) {
	router := NewRouter(eventServiceStub{}, appmetrics.NewRegistry())

	body := `{"source":"landing","user_id":"u123","value":1,"created_at":"2026-03-27T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "missing required field: event_type") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestPostEventsRejectsMultipleJSONObjects(t *testing.T) {
	router := NewRouter(eventServiceStub{}, appmetrics.NewRegistry())

	body := `{"event_type":"page_view","source":"landing","user_id":"u123","value":1,"created_at":"2026-03-27T12:00:00Z"}{"event_type":"click","source":"landing","user_id":"u124","value":2,"created_at":"2026-03-27T12:00:01Z"}`
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "invalid request body") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestHealthz(t *testing.T) {
	router := NewRouter(eventServiceStub{}, appmetrics.NewRegistry())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	registry := appmetrics.NewRegistry()
	router := NewRouter(eventServiceStub{}, registry)

	eventRequest := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(
		`{"event_type":"page_view","source":"landing","user_id":"u123","value":1,"created_at":"2026-03-27T12:00:00Z"}`,
	))
	eventRecorder := httptest.NewRecorder()
	router.ServeHTTP(eventRecorder, eventRequest)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "http_requests_total{method=\"POST\",path=\"/events\",status=\"202\"} 1") {
		t.Fatalf("expected POST /events request metric, got body: %s", body)
	}

	if !strings.Contains(body, "events_received_total 1") {
		t.Fatalf("expected events_received_total metric, got body: %s", body)
	}

	if !strings.Contains(body, "# HELP http_request_duration_seconds Duration of HTTP requests in seconds.") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

type eventServiceStub struct{}

func (eventServiceStub) Accept(context.Context, event.Event) error {
	return nil
}
