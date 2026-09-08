package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"eomp/packages/shared/pkg/eventbus"
	"eomp/services/helpdesk/internal/repository"
)

// OutboxWorker reliably publishes queued outbox events to RabbitMQ event bus.
type OutboxWorker struct {
	repo     repository.Repository
	bus      eventbus.EventBus
	interval time.Duration
	stopCh   chan struct{}
}

// NewOutboxWorker constructs a resilient outbox worker.
func NewOutboxWorker(repo repository.Repository, bus eventbus.EventBus, interval time.Duration) *OutboxWorker {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &OutboxWorker{
		repo:     repo,
		bus:      bus,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start launches the background polling loop.
func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

// Stop signals the worker loop to terminate.
func (w *OutboxWorker) Stop() {
	close(w.stopCh)
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	events, err := w.repo.FetchPendingOutboxEvents(ctx, 50)
	if err != nil {
		slog.Debug("failed to fetch pending outbox events", "error", err)
		return
	}

	for _, ev := range events {
		var data any
		if err := json.Unmarshal([]byte(ev.Payload), &data); err != nil {
			data = ev.Payload
		}

		busEvent := eventbus.Event{
			ID:        ev.ID,
			Source:    ev.Source,
			Type:      ev.EventType,
			Data:      data,
			Timestamp: ev.CreatedAt,
		}

		if err := w.bus.Publish(ctx, busEvent); err != nil {
			_ = w.repo.MarkOutboxEventFailed(ctx, ev.ID, fmt.Sprintf("failed to publish to bus: %v", err))
		} else {
			_ = w.repo.MarkOutboxEventPublished(ctx, ev.ID)
		}
	}
}
