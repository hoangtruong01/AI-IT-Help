package service

import "testing"

func TestEventString(t *testing.T) {
	data := map[string]any{"recipient_id": "user-123", "count": 2}
	if got := eventString(data, "recipient_id"); got != "user-123" {
		t.Fatalf("unexpected recipient: %q", got)
	}
	if got := firstEventString(data, "missing", "recipient_id"); got != "user-123" {
		t.Fatalf("unexpected fallback recipient: %q", got)
	}
	if got := eventString("invalid", "recipient_id"); got != "" {
		t.Fatalf("expected invalid payload to return empty, got %q", got)
	}
}
