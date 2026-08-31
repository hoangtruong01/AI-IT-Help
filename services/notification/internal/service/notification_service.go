package service

import (
	"context"
	"database/sql"
	stdErrors "errors"
	"fmt"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/eventbus"
	"eomp/services/notification/internal/model"
	"eomp/services/notification/internal/repository"
)

// NotificationService defines business logic for in-app alerts and notifications
type NotificationService interface {
	ListNotifications(ctx context.Context, query model.NotificationListQuery) (*model.NotificationListResponse, error)
	GetNotification(ctx context.Context, id string) (*model.Notification, error)
	SendNotification(ctx context.Context, req *model.CreateNotificationRequest) (*model.Notification, error)
	MarkAsRead(ctx context.Context, id, recipientID, recipientRole string) error
	MarkAllAsRead(ctx context.Context, recipientID, recipientRole string) error
	GetStats(ctx context.Context, recipientID, recipientRole string) (*model.NotificationStats, error)
	HandleDomainEvent(ctx context.Context, event eventbus.Event) error
}

type notificationService struct {
	repo repository.Repository
}

// NewNotificationService constructs a new NotificationService
func NewNotificationService(repo repository.Repository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) ListNotifications(ctx context.Context, query model.NotificationListQuery) (*model.NotificationListResponse, error) {
	return s.repo.ListNotifications(ctx, query)
}

func (s *notificationService) GetNotification(ctx context.Context, id string) (*model.Notification, error) {
	if id == "" {
		return nil, errors.BadRequest("notification id is required")
	}
	n, err := s.repo.FindNotificationByID(ctx, id)
	if err != nil {
		return nil, errors.Internal(ctx, "notification list", err)
	}
	if n == nil {
		return nil, errors.NotFound("notification not found")
	}
	return n, nil
}

func (s *notificationService) SendNotification(ctx context.Context, req *model.CreateNotificationRequest) (*model.Notification, error) {
	if req.Title == "" || req.Message == "" {
		return nil, errors.BadRequest("title and message are required")
	}

	if req.RecipientID == "" {
		req.RecipientID = "all"
	}
	if req.Category == "" {
		req.Category = model.CategorySystem
	}
	if req.Priority == "" {
		req.Priority = model.PriorityMedium
	}
	if req.Channel == "" {
		req.Channel = model.ChannelInApp
	}
	if req.Channel == model.ChannelEmail && req.RecipientEmail == "" {
		return nil, errors.BadRequest("recipient_email is required for email notifications")
	}

	n := &model.Notification{
		RecipientID:    req.RecipientID,
		RecipientEmail: req.RecipientEmail,
		Title:          req.Title,
		Message:        req.Message,
		Category:       req.Category,
		Priority:       req.Priority,
		IsRead:         false,
		Channel:        req.Channel,
		Metadata:       req.Metadata,
	}

	if err := s.repo.CreateNotification(ctx, n); err != nil {
		return nil, errors.Internal(ctx, "notification create", err)
	}

	return s.repo.FindNotificationByID(ctx, n.ID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id, recipientID, recipientRole string) error {
	if id == "" {
		return errors.BadRequest("notification id is required")
	}
	if recipientID == "" {
		return errors.Forbidden("recipient identity is required")
	}
	if err := s.repo.MarkAsRead(ctx, id, recipientID, recipientRole); err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return errors.NotFound("notification not found")
		}
		return errors.Internal(ctx, "notification mark read", err)
	}
	return nil
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, recipientID, recipientRole string) error {
	if recipientID == "" {
		return errors.Forbidden("recipient identity is required")
	}
	return s.repo.MarkAllAsRead(ctx, recipientID, recipientRole)
}

func (s *notificationService) GetStats(ctx context.Context, recipientID, recipientRole string) (*model.NotificationStats, error) {
	if recipientID == "" {
		return nil, errors.Forbidden("recipient identity is required")
	}
	return s.repo.GetStats(ctx, recipientID, recipientRole)
}

func (s *notificationService) HandleDomainEvent(ctx context.Context, event eventbus.Event) error {
	var title, message, category, priority string
	recipientID := "all"
	recipientEmail := ""

	switch event.Type {
	case eventbus.TopicTicketCreated:
		category = model.CategoryIncident
		priority = model.PriorityHigh
		title = "New Incident Ticket Raised"
		message = fmt.Sprintf("A new support incident was logged via EventBus: %v", event.Data)
		recipientID = eventString(event.Data, "reporter_id")
		recipientEmail = eventString(event.Data, "reporter_email")

	case eventbus.TopicTicketSLAWarning:
		category = model.CategorySLA
		priority = model.PriorityUrgent
		title = "SLA Breach Warning"
		message = fmt.Sprintf("Ticket SLA threshold at 80%%: %v", event.Data)
		recipientID = firstEventString(event.Data, "reporter_id", "requester_id")
		if recipientID == "" {
			recipientID = "ROLE_AGENT"
		}
		recipientEmail = firstEventString(event.Data, "reporter_email", "requester_email")

	case eventbus.TopicApprovalRequested:
		category = model.CategoryApproval
		priority = model.PriorityHigh
		title = "Approval Sign-off Requested"
		message = fmt.Sprintf("Workflow instance requires your decision: %v", event.Data)
		recipientID = eventString(event.Data, "approver_id")

	case eventbus.TopicAssetAssigned:
		category = model.CategoryAsset
		priority = model.PriorityMedium
		title = "Hardware Handover Registered"
		message = fmt.Sprintf("Asset assigned: %v", event.Data)
		recipientID = eventString(event.Data, "user_id")

	default:
		category = model.CategorySystem
		priority = model.PriorityLow
		title = fmt.Sprintf("System Event: %s", event.Type)
		message = fmt.Sprintf("Received event from %s: %v", event.Source, event.Data)
	}
	if recipientID == "" {
		return errors.BadRequest("domain event is missing its notification recipient")
	}

	_, err := s.SendNotification(ctx, &model.CreateNotificationRequest{
		RecipientID:    recipientID,
		RecipientEmail: recipientEmail,
		Title:          title,
		Message:        message,
		Category:       category,
		Priority:       priority,
		Channel:        model.ChannelInApp,
	})
	return err
}

func eventString(data any, key string) string {
	values, ok := data.(map[string]any)
	if !ok {
		return ""
	}
	value, _ := values[key].(string)
	return value
}

func firstEventString(data any, keys ...string) string {
	for _, key := range keys {
		if value := eventString(data, key); value != "" {
			return value
		}
	}
	return ""
}
