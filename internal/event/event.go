package event

import "time"

type Event struct {
	EventType string    `json:"event_type"`
	Source    string    `json:"source"`
	UserID    string    `json:"user_id"`
	Value     float64   `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}
