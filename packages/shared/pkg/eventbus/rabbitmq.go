package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DefaultExchange    = "eomp.events"
	DefaultDLXExchange = "eomp.dlx"
	DefaultDLQQueue    = "eomp.dead_letter_queue"
)

// RabbitMQEventBus implements EventBus using RabbitMQ AMQP 0-9-1 protocol.
type RabbitMQEventBus struct {
	url          string
	serviceName  string
	exchange     string
	dlxExchange  string
	conn         *amqp.Connection
	channel      *amqp.Channel
	mu           sync.RWMutex
	memoryBus    EventBus
	isConnected  bool
	isClosed     bool
	subsMu       sync.Mutex
	subs         []subscription
	reconnectSig chan struct{}
}

type subscription struct {
	eventType string
	handler   EventHandler
}

// NewRabbitMQEventBus connects to RabbitMQ and prepares standard topic exchanges.
func NewRabbitMQEventBus(url string, serviceName string) (*RabbitMQEventBus, error) {
	if url == "" {
		return nil, fmt.Errorf("RabbitMQ URL is required")
	}
	if serviceName == "" {
		serviceName = "eomp-service"
	}

	bus := newRabbitMQEventBus(url, serviceName)

	if err := bus.connect(); err != nil {
		return nil, err
	}

	go bus.reconnectSupervisor()
	return bus, nil
}

func newRabbitMQEventBus(url string, serviceName string) *RabbitMQEventBus {
	return &RabbitMQEventBus{
		url:          url,
		serviceName:  serviceName,
		exchange:     DefaultExchange,
		dlxExchange:  DefaultDLXExchange,
		memoryBus:    NewMemoryEventBus(),
		reconnectSig: make(chan struct{}, 1),
	}
}

// NewResilientEventBus always returns a RabbitMQ bus with local subscribers and
// an active reconnect supervisor, even when the broker is unavailable at boot.
func NewResilientEventBus(url string, serviceName string) EventBus {
	if serviceName == "" {
		serviceName = "eomp-service"
	}

	bus := newRabbitMQEventBus(url, serviceName)
	err := bus.connect()
	if err != nil {
		slog.Warn("rabbitMQ broker not reachable on boot; background reconnect enabled",
			slog.String("service", serviceName),
			slog.Any("error", err),
		)
	} else {
		slog.Info("connected to RabbitMQ AMQP broker successfully",
			slog.String("service", serviceName),
			slog.String("exchange", DefaultExchange),
		)
	}
	go bus.reconnectSupervisor()
	return bus
}

func (b *RabbitMQEventBus) connect() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	conn, err := amqp.DialConfig(b.url, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
	})
	if err != nil {
		b.isConnected = false
		return fmt.Errorf("amqp dial failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		b.isConnected = false
		return fmt.Errorf("amqp channel open failed: %w", err)
	}

	// Declare DLX exchange & queue for poison messages
	if err := ch.ExchangeDeclare(
		b.dlxExchange,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("failed declaring DLX exchange: %w", err)
	}

	_, _ = ch.QueueDeclare(
		DefaultDLQQueue,
		true,  // durable
		false, // auto-deleted
		false, // exclusive
		false, // no-wait
		nil,
	)
	_ = ch.QueueBind(DefaultDLQQueue, "#", b.dlxExchange, false, nil)

	// Declare main topic exchange
	if err := ch.ExchangeDeclare(
		b.exchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("failed declaring topic exchange %s: %w", b.exchange, err)
	}

	b.conn = conn
	b.channel = ch
	b.isConnected = true

	return nil
}

func (b *RabbitMQEventBus) reconnectSupervisor() {
	for {
		b.mu.RLock()
		closed := b.isClosed
		b.mu.RUnlock()
		if closed {
			return
		}

		b.mu.RLock()
		conn := b.conn
		b.mu.RUnlock()

		if conn == nil {
			time.Sleep(2 * time.Second)
			if err := b.connect(); err == nil {
				slog.Info("rabbitMQ connected after startup, restoring active subscriptions")
				b.restoreSubscriptions()
			}
			continue
		}

		closeErr := make(chan *amqp.Error, 1)
		conn.NotifyClose(closeErr)

		err, ok := <-closeErr
		if !ok || b.isClosed {
			return
		}

		slog.Warn("rabbitMQ connection lost, attempting auto-reconnect...", slog.Any("reason", err))
		b.mu.Lock()
		b.isConnected = false
		b.mu.Unlock()

		// Exponential backoff reconnect loop
		backoff := 1 * time.Second
		for {
			if b.isClosed {
				return
			}
			time.Sleep(backoff)
			if err := b.connect(); err == nil {
				slog.Info("rabbitMQ reconnected successfully, restoring active subscriptions")
				b.restoreSubscriptions()
				break
			}
			backoff *= 2
			if backoff > 15*time.Second {
				backoff = 15 * time.Second
			}
		}
	}
}

func (b *RabbitMQEventBus) restoreSubscriptions() {
	b.subsMu.Lock()
	defer b.subsMu.Unlock()

	for _, sub := range b.subs {
		_ = b.bindAndConsume(sub.eventType, sub.handler)
	}
}

// Publish publishes an event adhering to CloudEvents v1.0 specification.
func (b *RabbitMQEventBus) Publish(ctx context.Context, event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	if event.Source == "" {
		event.Source = b.serviceName
	}

	// Always publish to local in-memory listeners first
	_ = b.memoryBus.Publish(ctx, event)

	b.mu.RLock()
	ch := b.channel
	connected := b.isConnected
	b.mu.RUnlock()

	if !connected || ch == nil {
		return fmt.Errorf("rabbitMQ publish unavailable: broker is disconnected")
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed marshaling event: %w", err)
	}

	// Routing key mapping: replace wildcard if any
	routingKey := event.Type
	if routingKey == "" {
		routingKey = "general.event"
	}

	msg := amqp.Publishing{
		ContentType:  "application/cloudevents+json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    event.Timestamp,
		MessageId:    event.ID,
		Type:         event.Type,
		AppId:        event.Source,
		Body:         body,
	}

	err = ch.PublishWithContext(
		ctx,
		b.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		msg,
	)
	if err != nil {
		return fmt.Errorf("rabbitMQ publish error: %w", err)
	}

	return nil
}

// Subscribe listens to specific event types or wildcards (e.g. "ticket.*", "#")
func (b *RabbitMQEventBus) Subscribe(eventType string, handler EventHandler) error {
	// Register in memory fallback
	_ = b.memoryBus.Subscribe(eventType, handler)

	b.subsMu.Lock()
	b.subs = append(b.subs, subscription{eventType: eventType, handler: handler})
	b.subsMu.Unlock()

	b.mu.RLock()
	connected := b.isConnected
	b.mu.RUnlock()

	if connected {
		return b.bindAndConsume(eventType, handler)
	}

	return nil
}

func (b *RabbitMQEventBus) bindAndConsume(eventType string, handler EventHandler) error {
	b.mu.RLock()
	ch := b.channel
	b.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("amqp channel is not open")
	}

	sanitized := strings.ReplaceAll(eventType, "*", "wildcard")
	sanitized = strings.ReplaceAll(sanitized, "#", "all")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	queueName := fmt.Sprintf("%s.%s.queue", b.serviceName, sanitized)

	// Declare durable queue with DLX args
	args := amqp.Table{
		"x-dead-letter-exchange": b.dlxExchange,
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,
	)
	if err != nil {
		return fmt.Errorf("failed declaring consumer queue %s: %w", queueName, err)
	}

	// Map topic routing key
	routingKey := eventType
	if routingKey == "*" {
		routingKey = "#"
	}

	if err := ch.QueueBind(
		q.Name,
		routingKey,
		b.exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed binding queue %s to %s with key %s: %w", q.Name, b.exchange, routingKey, err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer tag
		false, // auto-ack (manual ack for resilience)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed starting consumer on queue %s: %w", q.Name, err)
	}

	go func() {
		for msg := range msgs {
			var event Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				slog.Error("failed decoding CloudEvent message from AMQP", slog.Any("error", err))
				_ = msg.Nack(false, false) // send to DLX
				continue
			}

			// Invoke domain event handler
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := handler(ctx, event); err != nil {
				slog.Error("event handler failed, rejecting message",
					slog.String("event_type", event.Type),
					slog.String("event_id", event.ID),
					slog.Any("error", err),
				)
				_ = msg.Nack(false, false) // route to dead-letter queue
			} else {
				_ = msg.Ack(false)
			}
			cancel()
		}
	}()

	return nil
}

// Close closes RabbitMQ channels and connections gracefully.
func (b *RabbitMQEventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.isClosed = true
	b.isConnected = false

	if b.channel != nil {
		_ = b.channel.Close()
	}
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
