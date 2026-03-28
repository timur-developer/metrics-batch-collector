package batcher

import (
	"context"
	"errors"
	"sync"
	"time"

	"metrics-batch-collector/internal/event"
	"metrics-batch-collector/internal/storage"
)

var (
	ErrBatcherClosed = errors.New("batcher is shut down")
	ErrBatcherFull   = errors.New("batcher buffer is full")
)

type Batcher struct {
	repository    storage.Repository
	batchSize     int
	flushInterval time.Duration
	input         chan event.Event

	mu     sync.RWMutex
	closed bool

	done chan struct{}
}

func New(repository storage.Repository, batchSize int, flushInterval time.Duration) *Batcher {
	b := &Batcher{
		repository:    repository,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		input:         make(chan event.Event, batchSize),
		done:          make(chan struct{}),
	}

	go b.run()

	return b
}

func (b *Batcher) Accept(_ context.Context, evt event.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return ErrBatcherClosed
	}

	select {
	case b.input <- evt:
		return nil
	default:
		return ErrBatcherFull
	}
}

func (b *Batcher) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}

	b.closed = true
	close(b.input)
	b.mu.Unlock()

	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *Batcher) run() {
	defer close(b.done)

	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	buffer := make([]event.Event, 0, b.batchSize)

	for {
		select {
		case evt, ok := <-b.input:
			if !ok {
				_ = b.flush(context.Background(), buffer)
				return
			}

			buffer = append(buffer, evt)
			if len(buffer) >= b.batchSize {
				buffer = b.flushAndReset(buffer)
			}
		case <-ticker.C:
			if len(buffer) == 0 {
				continue
			}

			buffer = b.flushAndReset(buffer)
		}
	}
}

func (b *Batcher) flushAndReset(buffer []event.Event) []event.Event {
	_ = b.flush(context.Background(), buffer)
	return make([]event.Event, 0, b.batchSize)
}

func (b *Batcher) flush(ctx context.Context, buffer []event.Event) error {
	if len(buffer) == 0 {
		return nil
	}

	batch := append([]event.Event(nil), buffer...)
	return b.repository.InsertBatch(ctx, batch)
}
