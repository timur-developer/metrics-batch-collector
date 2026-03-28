package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go/v2"

	"metrics-batch-collector/internal/event"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(ctx context.Context, dsn string) (*Repository, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) InsertBatch(ctx context.Context, events []event.Event) error {
	if len(events) == 0 {
		return nil
	}

	query, args := buildInsertQuery(events)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		log.Printf("clickhouse insert batch failed: size=%d error=%v", len(events), err)
		return fmt.Errorf("insert batch: %w", err)
	}

	return nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func buildInsertQuery(events []event.Event) (string, []any) {
	var builder strings.Builder
	args := make([]any, 0, len(events)*5)

	builder.WriteString(`
		INSERT INTO events (
			event_type,
			source,
			user_id,
			value,
			created_at
		) VALUES
	`)

	for index, item := range events {
		if index > 0 {
			builder.WriteString(",")
		}

		builder.WriteString("(?, ?, ?, ?, ?)")
		args = append(args, item.EventType, item.Source, item.UserID, item.Value, item.CreatedAt)
	}

	return builder.String(), args
}
