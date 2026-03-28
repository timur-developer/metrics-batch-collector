package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"metrics-batch-collector/internal/event"
	appmetrics "metrics-batch-collector/internal/metrics"
)

type EventHandler struct {
	service EventService
	metrics *appmetrics.Registry
}

type eventRequest struct {
	EventType *string    `json:"event_type"`
	Source    *string    `json:"source"`
	UserID    *string    `json:"user_id"`
	Value     *float64   `json:"value"`
	CreatedAt *time.Time `json:"created_at"`
}

func NewEventHandler(service EventService, registry *appmetrics.Registry) *EventHandler {
	return &EventHandler{
		service: service,
		metrics: registry,
	}
}

func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	evt, err := decodeEvent(r)
	if err != nil {
		log.Printf("event request validation failed: remote=%s error=%v", r.RemoteAddr, err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Accept(r.Context(), evt); err != nil {
		log.Printf("event request accept failed: remote=%s error=%v", r.RemoteAddr, err)
		writeError(w, http.StatusInternalServerError, "failed to accept event")
		return
	}

	h.metrics.IncEventsReceived()
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func decodeEvent(r *http.Request) (event.Event, error) {
	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	var request eventRequest
	if err := decoder.Decode(&request); err != nil {
		if errors.Is(err, io.EOF) {
			return event.Event{}, errors.New("invalid request body")
		}

		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			return event.Event{}, errors.New("invalid request body")
		case errors.As(err, &typeErr):
			return event.Event{}, fmt.Errorf("invalid field type: %s", typeErr.Field)
		case strings.Contains(err.Error(), "unknown field"):
			return event.Event{}, errors.New("invalid request body")
		default:
			return event.Event{}, errors.New("invalid request body")
		}
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err == nil {
		return event.Event{}, errors.New("invalid request body")
	} else if !errors.Is(err, io.EOF) {
		return event.Event{}, errors.New("invalid request body")
	}

	evt, err := validateEvent(request)
	if err != nil {
		return event.Event{}, err
	}

	return evt, nil
}

func validateEvent(request eventRequest) (event.Event, error) {
	if request.EventType == nil || strings.TrimSpace(*request.EventType) == "" {
		return event.Event{}, errors.New("missing required field: event_type")
	}

	if request.Source == nil || strings.TrimSpace(*request.Source) == "" {
		return event.Event{}, errors.New("missing required field: source")
	}

	if request.UserID == nil || strings.TrimSpace(*request.UserID) == "" {
		return event.Event{}, errors.New("missing required field: user_id")
	}

	if request.Value == nil {
		return event.Event{}, errors.New("missing required field: value")
	}

	if request.CreatedAt == nil || request.CreatedAt.IsZero() {
		return event.Event{}, errors.New("missing required field: created_at")
	}

	return event.Event{
		EventType: strings.TrimSpace(*request.EventType),
		Source:    strings.TrimSpace(*request.Source),
		UserID:    strings.TrimSpace(*request.UserID),
		Value:     *request.Value,
		CreatedAt: *request.CreatedAt,
	}, nil
}
