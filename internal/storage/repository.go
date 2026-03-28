package storage

import (
	"context"

	"metrics-batch-collector/internal/event"
)

type Repository interface {
	InsertBatch(context.Context, []event.Event) error
	Close() error
}
