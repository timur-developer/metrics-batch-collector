package http

import (
	"context"

	"metrics-batch-collector/internal/event"
)

type EventService interface {
	Accept(context.Context, event.Event) error
}
