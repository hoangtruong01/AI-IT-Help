package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config defines Redis client connection settings.
type Config struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Client wraps the go-redis client with resiliency and health checking utilities.
type Client struct {
	rdb    *redis.Client
	logger *slog.Logger
}

// NewClient initializes a connection pool to Redis.
// If connection fails, it returns an error or a nil-safe client for graceful fallback.
func NewClient(cfg Config, logger *slog.Logger) (*Client, error) {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 6379
	}
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 50
	}
	if cfg.MinIdleConns == 0 {
		cfg.MinIdleConns = 10
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 2 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 500 * time.Millisecond
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 500 * time.Millisecond
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		if logger != nil {
			logger.Warn("redis server unreachable during startup (graceful fallback enabled)",
				slog.String("addr", addr),
				slog.Any("error", err),
			)
		}
		return &Client{rdb: rdb, logger: logger}, fmt.Errorf("failed to ping redis at %s: %w", addr, err)
	}

	if logger != nil {
		logger.Info("connected to redis successfully",
			slog.String("addr", addr),
			slog.Int("db", cfg.DB),
			slog.Int("pool_size", cfg.PoolSize),
		)
	}

	return &Client{rdb: rdb, logger: logger}, nil
}

// Raw returns the underlying *redis.Client.
func (c *Client) Raw() *redis.Client {
	if c == nil {
		return nil
	}
	return c.rdb
}

// Ping checks if Redis is currently reachable.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.rdb == nil {
		return fmt.Errorf("redis client is uninitialized")
	}
	return c.rdb.Ping(ctx).Err()
}

// Close gracefully closes the Redis connection pool.
func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}
