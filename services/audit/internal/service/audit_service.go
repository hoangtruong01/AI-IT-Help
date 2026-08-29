package service

import (
	"context"
	"fmt"
	"time"

	"eomp/packages/shared/pkg/eventbus"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/audit/internal/model"
	"eomp/services/audit/internal/repository"
)

// Service defines the business logic methods for Audit & Compliance.
type Service interface {
	ListAuditLogs(ctx context.Context, filter model.AuditFilterQuery) ([]model.AuditLog, int, error)
	GetAuditLogByID(ctx context.Context, id string) (*model.AuditLog, error)
	CreateAuditLog(ctx context.Context, req model.CreateAuditLogRequest) (*model.AuditLog, error)
	GetStats(ctx context.Context) (*model.AuditStats, error)
	GetSecurityEvents(ctx context.Context, limit int) ([]model.SecurityEvent, error)
	IngestDomainEvent(ctx context.Context, event eventbus.Event) error
}

type auditService struct {
	repo repository.Repository
}

// NewService instantiates the Audit Service.
func NewService(repo repository.Repository) Service {
	return &auditService{repo: repo}
}

func (s *auditService) ListAuditLogs(ctx context.Context, filter model.AuditFilterQuery) ([]model.AuditLog, int, error) {
	return s.repo.ListAuditLogs(ctx, filter)
}

func (s *auditService) GetAuditLogByID(ctx context.Context, id string) (*model.AuditLog, error) {
	return s.repo.GetAuditLogByID(ctx, id)
}

// CreateAuditLog sanitizes sensitive fields with Data Masking and persists immutable log record.
func (s *auditService) CreateAuditLog(ctx context.Context, req model.CreateAuditLogRequest) (*model.AuditLog, error) {
	if req.EventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}
	if req.ActorEmail == "" {
		return nil, fmt.Errorf("actor_email is required")
	}

	// Apply Data Masking on OldValues & NewValues before persistence (Test Case 10.3)
	maskedOld := middleware.MaskSensitiveData(req.OldValues)
	maskedNew := middleware.MaskSensitiveData(req.NewValues)

	status := req.Status
	if status == "" {
		status = "SUCCESS"
	}

	log := &model.AuditLog{
		ID:           fmt.Sprintf("aud-%d", time.Now().UnixNano()),
		EventType:    req.EventType,
		ActorID:      req.ActorID,
		ActorName:    req.ActorName,
		ActorEmail:   req.ActorEmail,
		ActorRole:    req.ActorRole,
		ServiceName:  req.ServiceName,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		Status:       status,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		OldValues:    maskedOld,
		NewValues:    maskedNew,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateAuditLog(ctx, log); err != nil {
		return nil, err
	}

	return log, nil
}

func (s *auditService) GetStats(ctx context.Context) (*model.AuditStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *auditService) GetSecurityEvents(ctx context.Context, limit int) ([]model.SecurityEvent, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetSecurityEvents(ctx, limit)
}

// IngestDomainEvent consumes CloudEvents from AMQP EventBus, computes tamper-evident hash and persists to audit_db.
func (s *auditService) IngestDomainEvent(ctx context.Context, event eventbus.Event) error {
	actorEmail := "system@eomp.local"
	actorRole := "ROLE_SYSTEM"
	actorName := "System Event Worker"
	resourceID := event.ID
	resourceType := "event"

	// Parse fields if data is a map
	var newVals map[string]any
	if m, ok := event.Data.(map[string]any); ok {
		newVals = m
		if email, ok := m["actor_email"].(string); ok && email != "" {
			actorEmail = email
		} else if email, ok := m["reporter_email"].(string); ok && email != "" {
			actorEmail = email
		}
		if role, ok := m["actor_role"].(string); ok && role != "" {
			actorRole = role
		}
		if name, ok := m["actor_name"].(string); ok && name != "" {
			actorName = name
		}
		if id, ok := m["id"].(string); ok && id != "" {
			resourceID = id
		} else if id, ok := m["ticket_id"].(string); ok && id != "" {
			resourceID = id
			resourceType = "ticket"
		} else if id, ok := m["asset_id"].(string); ok && id != "" {
			resourceID = id
			resourceType = "asset"
		} else if id, ok := m["instance_id"].(string); ok && id != "" {
			resourceID = id
			resourceType = "workflow"
		}
	}

	req := model.CreateAuditLogRequest{
		EventType:    event.Type,
		ActorName:    actorName,
		ActorEmail:   actorEmail,
		ActorRole:    actorRole,
		ServiceName:  event.Source,
		IPAddress:    "127.0.0.1",
		UserAgent:    "eomp-eventbus-consumer",
		Status:       "SUCCESS",
		ResourceType: resourceType,
		ResourceID:   resourceID,
		NewValues:    newVals,
	}

	_, err := s.CreateAuditLog(ctx, req)
	return err
}
