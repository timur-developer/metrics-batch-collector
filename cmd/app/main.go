package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"metrics-batch-collector/internal/batcher"
	"metrics-batch-collector/internal/config"
	apphttp "metrics-batch-collector/internal/http"
	appmetrics "metrics-batch-collector/internal/metrics"
	"metrics-batch-collector/internal/storage/clickhouse"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()

	repository, err := clickhouse.NewRepository(startupCtx, cfg.ClickHouseDSN)
	if err != nil {
		log.Fatalf("init clickhouse repository: %v", err)
	}
	defer func() {
		if err := repository.Close(); err != nil {
			log.Printf("close clickhouse repository: %v", err)
		}
	}()

	metricsRegistry := appmetrics.NewRegistry()
	eventBatcher := batcher.New(repository, cfg.BatchSize, cfg.FlushInterval, metricsRegistry)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           apphttp.NewRouter(eventBatcher, metricsRegistry),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("connected to clickhouse")
		log.Printf("starting HTTP server on :%s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("http server failed: %v", err)
		}
	case <-ctx.Done():
		stop()
		log.Println("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown server: %v", err)
	}

	if err := eventBatcher.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown batcher: %v", err)
	}

	log.Println("server stopped")
}
