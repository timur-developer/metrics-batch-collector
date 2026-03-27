package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type EventHandler struct {
	service EventService
}

type eventRequest struct {
	EventType *string    `json:"event_type"`
	Source    *string    `json:"source"`
	UserID    *string    `json:"user_id"`
	Value     *float64   `json:"value"`
	CreatedAt *time.Time `json:"created_at"`
}

func NewEventHandler(service EventService) *EventHandler {
	return &EventHandler{service: service}
}

func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	event, err := decodeEvent(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Accept(r.Context(), event); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to accept event")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func decodeEvent(r *http.Request) (Event, error) {
	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	var request eventRequest
	if err := decoder.Decode(&request); err != nil {
		if errors.Is(err, io.EOF) {
			return Event{}, errors.New("invalid request body")
		}

		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			return Event{}, errors.New("invalid request body")
		case errors.As(err, &typeErr):
			return Event{}, fmt.Errorf("invalid field type: %s", typeErr.Field)
		case strings.Contains(err.Error(), "unknown field"):
			return Event{}, errors.New("invalid request body")
		default:
			return Event{}, errors.New("invalid request body")
		}
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err == nil {
		return Event{}, errors.New("invalid request body")
	} else if !errors.Is(err, io.EOF) {
		return Event{}, errors.New("invalid request body")
	}

	event, err := validateEvent(request)
	if err != nil {
		return Event{}, err
	}

	return event, nil
}

func validateEvent(request eventRequest) (Event, error) {
	if request.EventType == nil || strings.TrimSpace(*request.EventType) == "" {
		return Event{}, errors.New("missing required field: event_type")
	}

	if request.Source == nil || strings.TrimSpace(*request.Source) == "" {
		return Event{}, errors.New("missing required field: source")
	}

	if request.UserID == nil || strings.TrimSpace(*request.UserID) == "" {
		return Event{}, errors.New("missing required field: user_id")
	}

	if request.Value == nil {
		return Event{}, errors.New("missing required field: value")
	}

	if request.CreatedAt == nil || request.CreatedAt.IsZero() {
		return Event{}, errors.New("missing required field: created_at")
	}

	return Event{
		EventType: strings.TrimSpace(*request.EventType),
		Source:    strings.TrimSpace(*request.Source),
		UserID:    strings.TrimSpace(*request.UserID),
		Value:     *request.Value,
		CreatedAt: *request.CreatedAt,
	}, nil
}
