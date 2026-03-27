package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostEventsAccepted(t *testing.T) {
	mux := NewMux(NewInMemoryEventService())

	body := `{"event_type":"page_view","source":"landing","user_id":"u123","value":0,"created_at":"2026-03-27T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}

func TestPostEventsValidationError(t *testing.T) {
	mux := NewMux(NewInMemoryEventService())

	body := `{"source":"landing","user_id":"u123","value":1,"created_at":"2026-03-27T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "missing required field: event_type") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestPostEventsRejectsMultipleJSONObjects(t *testing.T) {
	mux := NewMux(NewInMemoryEventService())

	body := `{"event_type":"page_view","source":"landing","user_id":"u123","value":1,"created_at":"2026-03-27T12:00:00Z"}{"event_type":"click","source":"landing","user_id":"u124","value":2,"created_at":"2026-03-27T12:00:01Z"}`
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "invalid request body") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestHealthz(t *testing.T) {
	mux := NewMux(NewInMemoryEventService())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	mux := NewMux(NewInMemoryEventService())

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "metrics are not implemented yet") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}
