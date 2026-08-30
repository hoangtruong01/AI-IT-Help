package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"eomp/services/audit/internal/handler"
	"eomp/services/audit/internal/model"
	"eomp/services/audit/internal/repository"
	"eomp/services/audit/internal/service"
)

type memoryAuditRepository struct {
	logs []model.AuditLog
}

var testAuditKey = []byte("0123456789abcdef0123456789abcdef")

func (m *memoryAuditRepository) ListAuditLogs(context.Context, model.AuditFilterQuery) ([]model.AuditLog, int, error) {
	return m.logs, len(m.logs), nil
}
func (m *memoryAuditRepository) GetAuditLogByID(_ context.Context, id string) (*model.AuditLog, error) {
	for _, item := range m.logs {
		if item.ID == id {
			copy := item
			return &copy, nil
		}
	}
	return nil, nil
}
func (m *memoryAuditRepository) CreateAuditLog(_ context.Context, log *model.AuditLog) error {
	log.PreviousChecksum = "0000000000000000000000000000000000000000000000000000000000000000"
	if len(m.logs) > 0 {
		log.PreviousChecksum = m.logs[len(m.logs)-1].ChecksumSHA256
	}
	if err := repository.ComputeHMAC(log, testAuditKey); err != nil {
		return err
	}
	m.logs = append(m.logs, *log)
	return nil
}
func (m *memoryAuditRepository) VerifyIntegrity(context.Context) (*model.IntegrityReport, error) {
	return &model.IntegrityReport{Valid: true, TotalLogs: int64(len(m.logs)), VerifiedLogs: int64(len(m.logs)), Message: "valid"}, nil
}
func (m *memoryAuditRepository) GetStats(context.Context) (*model.AuditStats, error) {
	total := int64(len(m.logs))
	return &model.AuditStats{TotalLogs: total, SuccessCount: total, ImmutableProofsCount: total}, nil
}
func (m *memoryAuditRepository) GetSecurityEvents(context.Context, int) ([]model.SecurityEvent, error) {
	return []model.SecurityEvent{}, nil
}

func setupAuditTestApp() *handler.AuditHandler {
	repo := &memoryAuditRepository{logs: []model.AuditLog{{ID: "fixture", ChecksumSHA256: "test"}}}
	svc := service.NewService(repo)
	return handler.NewAuditHandler(svc)
}

// Test Case 10.3: Audit Log Creation with automatic Data Masking & SHA-256 Checksum.
func TestAudit_TestCase_10_3_DataMaskingAndChecksum(t *testing.T) {
	h := setupAuditTestApp()

	payload := model.CreateAuditLogRequest{
		EventType:    "USER_PASSWORD_RESET",
		ActorID:      "u1",
		ActorName:    "Administrator",
		ActorEmail:   "admin@eomp.local",
		ActorRole:    "ROLE_ADMIN",
		ServiceName:  "auth",
		IPAddress:    "192.168.1.10",
		Status:       "SUCCESS",
		ResourceType: "user_credentials",
		ResourceID:   "u-4412",
		OldValues: map[string]interface{}{
			"password":    "OldPlainTextPassword123!",
			"api_token":   "sk_test_998129039120",
			"credit_card": "4111 2222 3333 4444",
		},
		NewValues: map[string]interface{}{
			"password":  "NewSuperSecurePassword2026!",
			"api_token": "sk_live_112984918239",
			"status":    "ACTIVE",
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/audit/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateAuditLog(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}

	var createdLog model.AuditLog
	if err := json.Unmarshal(rec.Body.Bytes(), &createdLog); err != nil {
		t.Fatalf("failed to decode created audit log: %v", err)
	}

	// 1. Verify Data Masking was applied
	if createdLog.OldValues["password"] != "********" {
		t.Errorf("expected masked old password, got %v", createdLog.OldValues["password"])
	}
	if createdLog.NewValues["password"] != "********" {
		t.Errorf("expected masked new password, got %v", createdLog.NewValues["password"])
	}
	if createdLog.OldValues["api_token"] != "********" {
		t.Errorf("expected masked api_token, got %v", createdLog.OldValues["api_token"])
	}

	// 2. Verify the keyed chain seal and UUID-compatible identifier.
	if len(createdLog.ChecksumSHA256) != 64 {
		t.Errorf("expected 64-character HMAC-SHA256 checksum, got %s (len %d)", createdLog.ChecksumSHA256, len(createdLog.ChecksumSHA256))
	}
	if createdLog.ChecksumAlgorithm != "HMAC-SHA256" {
		t.Errorf("expected HMAC-SHA256 algorithm, got %q", createdLog.ChecksumAlgorithm)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(createdLog.ID) {
		t.Errorf("expected UUID v4 audit id, got %q", createdLog.ID)
	}
}

// Test Listing and Filtering Audit Logs.
func TestAudit_ListAndStats(t *testing.T) {
	h := setupAuditTestApp()

	// List
	reqList := httptest.NewRequest("GET", "/api/v1/audit/logs?page=1&page_size=10", nil)
	recList := httptest.NewRecorder()
	h.ListAuditLogs(recList, reqList)

	if recList.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for list, got %d", recList.Code)
	}

	// Stats
	reqStats := httptest.NewRequest("GET", "/api/v1/audit/stats", nil)
	recStats := httptest.NewRecorder()
	h.GetStats(recStats, reqStats)

	if recStats.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for stats, got %d", recStats.Code)
	}

	var stats model.AuditStats
	if err := json.Unmarshal(recStats.Body.Bytes(), &stats); err != nil {
		t.Fatalf("failed to decode stats: %v", err)
	}
	if stats.TotalLogs == 0 {
		t.Errorf("expected non-zero total audit logs, got 0")
	}

	// Security Events
	reqSec := httptest.NewRequest("GET", "/api/v1/audit/security-events", nil)
	recSec := httptest.NewRecorder()
	h.GetSecurityEvents(recSec, reqSec)

	if recSec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for security events, got %d", recSec.Code)
	}
}
