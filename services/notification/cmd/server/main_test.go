package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"eomp/packages/shared/pkg/eventbus"
	"eomp/services/notification/internal/config"
	"eomp/services/notification/internal/handler"
	"eomp/services/notification/internal/model"
)

func TestHealthHandler(t *testing.T) {
	cfg := config.Load()
	h := handler.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["service"] != "notification" {
		t.Errorf("expected service 'notification', got '%v'", body["service"])
	}
}

func TestNotificationConstants(t *testing.T) {
	if model.CategoryIncident != "INCIDENT" {
		t.Errorf("expected INCIDENT, got %s", model.CategoryIncident)
	}
	if model.PriorityUrgent != "URGENT" {
		t.Errorf("expected URGENT, got %s", model.PriorityUrgent)
	}
	if model.ChannelInApp != "IN_APP" {
		t.Errorf("expected IN_APP, got %s", model.ChannelInApp)
	}
}

func TestEventBusPublishSubscribe(t *testing.T) {
	bus := eventbus.NewMemoryEventBus()

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedType string
	err := bus.Subscribe(eventbus.TopicTicketCreated, func(ctx context.Context, event eventbus.Event) error {
		receivedType = event.Type
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	event := eventbus.Event{
		ID:        "test-evt-1",
		Source:    "eomp.helpdesk",
		Type:      eventbus.TopicTicketCreated,
		Data:      map[string]string{"ticket_number": "TK-1001"},
		Timestamp: time.Now(),
	}

	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	wg.Wait()

	if receivedType != eventbus.TopicTicketCreated {
		t.Errorf("expected event type %s, got %s", eventbus.TopicTicketCreated, receivedType)
	}
}
