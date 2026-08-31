package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Standard EOMP Event Topics
const (
	TopicTicketCreated       = "ticket.created"
	TopicTicketStatusChanged = "ticket.status_changed"
	TopicTicketAssigned      = "ticket.assigned"
	TopicTicketSLAWarning    = "ticket.sla_warning"
	TopicTicketSLABreached   = "ticket.sla_breached"
	TopicApprovalRequested   = "approval.requested"
	TopicApprovalDecided     = "approval.decided"
	TopicWorkflowStarted     = "workflow.started"
	TopicWorkflowCompleted   = "workflow.completed"
	TopicCABVoteSubmitted    = "cab.vote_submitted"
	TopicAssetAssigned       = "asset.assigned"
	TopicAssetReturned       = "asset.returned"
	TopicSecurityAlert       = "security.alert"
)

// Event follows CloudEvents v1.0 specification
type Event struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	Type      string    `json:"type"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

// EventHandler callback function
type EventHandler func(ctx context.Context, event Event) error

// EventBus interface for publishing and subscribing to domain events
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType string, handler EventHandler) error
}

type memoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

// NewMemoryEventBus constructs an in-memory resilient event bus
func NewMemoryEventBus() EventBus {
	return &memoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (b *memoryEventBus) Publish(ctx context.Context, event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	b.mu.RLock()
	handlers, exists := b.handlers[event.Type]
	wildcardHandlers := b.handlers["*"]
	b.mu.RUnlock()

	var allHandlers []EventHandler
	if exists {
		allHandlers = append(allHandlers, handlers...)
	}
	allHandlers = append(allHandlers, wildcardHandlers...)

	for _, h := range allHandlers {
		go func(handler EventHandler, ev Event) {
			_ = handler(context.Background(), ev)
		}(h, event)
	}

	return nil
}

func (b *memoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	return nil
}

// Helper to serialize event data to JSON
func MarshalEventData(data any) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event data: %w", err)
	}
	return string(bytes), nil
}
