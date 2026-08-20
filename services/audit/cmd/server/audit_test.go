package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eomp/services/audit/internal/handler"
	"eomp/services/audit/internal/model"
	"eomp/services/audit/internal/repository"
	"eomp/services/audit/internal/service"
)

func setupAuditTestApp() *handler.AuditHandler {
	repo := repository.NewRepository(nil)
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
			"password":    "NewSuperSecurePassword2026!",
			"api_token":   "sk_live_112984918239",
			"status":      "ACTIVE",
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

	// 2. Verify SHA-256 Checksum was computed
	if len(createdLog.ChecksumSHA256) != 64 {
		t.Errorf("expected 64-character SHA-256 checksum, got %s (len %d)", createdLog.ChecksumSHA256, len(createdLog.ChecksumSHA256))
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
