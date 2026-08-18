package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/middleware"
	"eomp/packages/shared/pkg/response"
	"eomp/services/notification/internal/model"
	"eomp/services/notification/internal/service"
)

// NotificationHandler handles HTTP endpoints for In-App and Email Notifications
type NotificationHandler struct {
	svc service.NotificationService
}

// NewNotificationHandler constructs a new NotificationHandler
func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// ListNotifications returns paginated alerts for the authenticated user
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	recipientID := middleware.GetUserID(r.Context())
	if recipientID == "" {
		recipientID = r.URL.Query().Get("recipient_id")
	}

	var isRead *bool
	if isReadStr := r.URL.Query().Get("is_read"); isReadStr != "" {
		val := isReadStr == "true"
		isRead = &val
	}

	query := model.NotificationListQuery{
		RecipientID: recipientID,
		IsRead:      isRead,
		Category:    r.URL.Query().Get("category"),
		Page:        page,
		PageSize:    pageSize,
	}

	resp, err := h.svc.ListNotifications(r.Context(), query)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// SendNotification dispatches a new alert
func (h *NotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	var req model.CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	n, err := h.svc.SendNotification(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, n)
}

// MarkAsRead marks a single notification as read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	if err := h.svc.MarkAsRead(r.Context(), id); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

// MarkAllAsRead marks all user alerts as read
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	recipientID := middleware.GetUserID(r.Context())
	if recipientID == "" {
		recipientID = "all"
	}

	if err := h.svc.MarkAllAsRead(r.Context(), recipientID); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "all marked as read"})
}

// GetStats returns summary counts of unread alerts
func (h *NotificationHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	recipientID := middleware.GetUserID(r.Context())
	if recipientID == "" {
		recipientID = "all"
	}

	stats, err := h.svc.GetStats(r.Context(), recipientID)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, stats)
}
