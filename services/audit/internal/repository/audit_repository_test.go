package repository

import (
	"testing"
	"time"

	"eomp/services/audit/internal/model"
)

func TestComputeHMACBindsPayloadAndPredecessor(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	base := model.AuditLog{
		ID:               "18b77b39-31c4-4da7-a100-0045c2fdad9e",
		EventType:        "ROLE_CHANGE",
		ActorEmail:       "operator@example.invalid",
		ActorRole:        "ROLE_ADMIN",
		ServiceName:      "auth",
		IPAddress:        "127.0.0.1",
		Status:           "SUCCESS",
		ResourceType:     "user",
		ResourceID:       "target",
		NewValues:        map[string]any{"role": "ROLE_MANAGER"},
		PreviousChecksum: chainGenesis,
		CreatedAt:        time.Unix(1_700_000_000, 0).UTC(),
	}
	if err := ComputeHMAC(&base, key); err != nil {
		t.Fatalf("compute base HMAC: %v", err)
	}
	baseChecksum := base.ChecksumSHA256

	tampered := base
	tampered.NewValues = map[string]any{"role": "ROLE_ADMIN"}
	if err := ComputeHMAC(&tampered, key); err != nil {
		t.Fatalf("compute tampered HMAC: %v", err)
	}
	if tampered.ChecksumSHA256 == baseChecksum {
		t.Fatal("payload tampering must change the audit HMAC")
	}

	relinked := base
	relinked.PreviousChecksum = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	if err := ComputeHMAC(&relinked, key); err != nil {
		t.Fatalf("compute relinked HMAC: %v", err)
	}
	if relinked.ChecksumSHA256 == baseChecksum {
		t.Fatal("changing the predecessor must change the audit HMAC")
	}
}

func TestComputeHMACRejectsShortKey(t *testing.T) {
	if err := ComputeHMAC(&model.AuditLog{}, []byte("too-short")); err == nil {
		t.Fatal("expected a short HMAC key to be rejected")
	}
}

func TestComputeHMACNormalizesDatabaseTimestampPrecision(t *testing.T) {
	log := model.AuditLog{
		ID:               "18b77b39-31c4-4da7-a100-0045c2fdad9e",
		PreviousChecksum: chainGenesis,
		CreatedAt:        time.Unix(1_700_000_000, 123_456_789),
	}
	if err := ComputeHMAC(&log, []byte("0123456789abcdef0123456789abcdef")); err != nil {
		t.Fatalf("compute HMAC: %v", err)
	}
	if got, want := log.CreatedAt.Nanosecond(), 123_456_000; got != want {
		t.Fatalf("timestamp precision = %d ns, want %d ns", got, want)
	}
}
