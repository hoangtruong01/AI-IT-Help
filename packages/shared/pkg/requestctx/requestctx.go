package requestctx

import "context"

type key string

const requestIDKey key = "request_id"

// WithRequestID returns a context carrying the request correlation identifier.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestID returns the correlation identifier stored in ctx, if any.
func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}
