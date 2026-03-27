package http

import (
	"context"
	"sync"
	"time"
)

type Event struct {
	EventType string    `json:"event_type"`
	Source    string    `json:"source"`
	UserID    string    `json:"user_id"`
	Value     float64   `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type EventService interface {
	Accept(context.Context, Event) error
}

type InMemoryEventService struct {
	mu     sync.Mutex
	events []Event
}

func NewInMemoryEventService() *InMemoryEventService {
	return &InMemoryEventService{}
}

func (s *InMemoryEventService) Accept(_ context.Context, event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, event)
	return nil
}
