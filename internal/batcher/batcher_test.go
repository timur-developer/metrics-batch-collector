package batcher

import (
	"context"
	"sync"
	"testing"
	"time"

	"metrics-batch-collector/internal/event"
)

type repositoryStub struct {
	mu      sync.Mutex
	batches [][]event.Event
}

func (r *repositoryStub) InsertBatch(_ context.Context, events []event.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	batch := append([]event.Event(nil), events...)
	r.batches = append(r.batches, batch)
	return nil
}

func (r *repositoryStub) Close() error {
	return nil
}

func (r *repositoryStub) batchCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.batches)
}

func (r *repositoryStub) totalEvents() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	total := 0
	for _, batch := range r.batches {
		total += len(batch)
	}

	return total
}

func TestBatcherFlushesOnBatchSize(t *testing.T) {
	repository := &repositoryStub{}
	b := New(repository, 2, time.Hour)

	if err := b.Accept(context.Background(), testEvent("u1")); err != nil {
		t.Fatalf("accept first event: %v", err)
	}

	if err := b.Accept(context.Background(), testEvent("u2")); err != nil {
		t.Fatalf("accept second event: %v", err)
	}

	waitForCondition(t, time.Second, func() bool {
		return repository.batchCount() == 1 && repository.totalEvents() == 2
	})

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := b.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown batcher: %v", err)
	}
}

func TestBatcherFlushesOnInterval(t *testing.T) {
	repository := &repositoryStub{}
	b := New(repository, 10, 20*time.Millisecond)

	if err := b.Accept(context.Background(), testEvent("u1")); err != nil {
		t.Fatalf("accept event: %v", err)
	}

	waitForCondition(t, time.Second, func() bool {
		return repository.batchCount() == 1 && repository.totalEvents() == 1
	})

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := b.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown batcher: %v", err)
	}
}

func TestBatcherFlushesPendingEventsOnShutdown(t *testing.T) {
	repository := &repositoryStub{}
	b := New(repository, 10, time.Hour)

	if err := b.Accept(context.Background(), testEvent("u1")); err != nil {
		t.Fatalf("accept event: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := b.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown batcher: %v", err)
	}

	if repository.batchCount() != 1 {
		t.Fatalf("expected 1 batch, got %d", repository.batchCount())
	}

	if repository.totalEvents() != 1 {
		t.Fatalf("expected 1 flushed event, got %d", repository.totalEvents())
	}
}

func TestBatcherReturnsErrorWhenFull(t *testing.T) {
	repository := &repositoryStub{}
	b := New(repository, 1, time.Hour)

	if err := b.Accept(context.Background(), testEvent("u1")); err != nil {
		t.Fatalf("accept first event: %v", err)
	}

	if err := b.Accept(context.Background(), testEvent("u2")); err != ErrBatcherFull {
		t.Fatalf("expected ErrBatcherFull, got %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := b.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown batcher: %v", err)
	}
}

func TestBatcherRejectsEventsAfterShutdown(t *testing.T) {
	repository := &repositoryStub{}
	b := New(repository, 1, time.Hour)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := b.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown batcher: %v", err)
	}

	if err := b.Accept(context.Background(), testEvent("u1")); err != ErrBatcherClosed {
		t.Fatalf("expected ErrBatcherClosed, got %v", err)
	}
}

func testEvent(userID string) event.Event {
	return event.Event{
		EventType: "page_view",
		Source:    "landing",
		UserID:    userID,
		Value:     1,
		CreatedAt: time.Unix(0, 0).UTC(),
	}
}

func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("condition was not met within %s", timeout)
}
