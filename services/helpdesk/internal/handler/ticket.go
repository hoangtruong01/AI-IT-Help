package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/middleware"
	"eomp/packages/shared/pkg/response"
	"eomp/services/helpdesk/internal/model"
	"eomp/services/helpdesk/internal/service"
)

// TicketHandler handles HTTP endpoints for Helpdesk and Service Catalog
type TicketHandler struct {
	svc service.TicketService
}

// NewTicketHandler constructs a new TicketHandler instance
func NewTicketHandler(svc service.TicketService) *TicketHandler {
	return &TicketHandler{svc: svc}
}

// ListTickets returns paginated tickets with filtering
func (h *TicketHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	query := model.TicketListQuery{
		Page:        page,
		PageSize:    pageSize,
		Status:      r.URL.Query().Get("status"),
		Priority:    r.URL.Query().Get("priority"),
		Category:    r.URL.Query().Get("category"),
		AssigneeID:  r.URL.Query().Get("assignee_id"),
		RequesterID: r.URL.Query().Get("requester_id"),
		Search:      r.URL.Query().Get("search"),
	}
	if middleware.GetUserRole(r.Context()) == "ROLE_EMPLOYEE" {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			errors.WriteHTTP(w, errors.Forbidden("employee identity is required"))
			return
		}
		query.RequesterID = userID
	}

	resp, err := h.svc.ListTickets(r.Context(), query)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// CreateTicket creates a new incident or service request
func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	userID := middleware.GetUserID(r.Context())
	userEmail := middleware.GetUserEmail(r.Context())
	if middleware.GetUserRole(r.Context()) == "ROLE_EMPLOYEE" {
		if userID == "" || userEmail == "" {
			errors.WriteHTTP(w, errors.Forbidden("employee identity is required"))
			return
		}
		req.RequesterID = userID
		req.RequesterEmail = userEmail
		req.RequesterName = userEmail
	} else {
		if req.RequesterID == "" {
			req.RequesterID = userID
		}
		if req.RequesterEmail == "" {
			req.RequesterEmail = userEmail
		}
	}

	ticket, err := h.svc.CreateTicket(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, ticket)
}

// GetTicket retrieves ticket details
func (h *TicketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	ticket, err := h.svc.GetTicket(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	if !employeeCanAccessTicket(r, ticket) {
		errors.WriteHTTP(w, errors.NotFound("ticket not found"))
		return
	}

	response.JSON(w, http.StatusOK, ticket)
}

// UpdateStatus changes the lifecycle state of a ticket
func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	var req model.UpdateTicketStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	actorID := middleware.GetUserID(r.Context())
	if actorID == "" {
		actorID = "system"
	}
	actorName := middleware.GetUserEmail(r.Context())
	if actorName == "" {
		actorName = "Operator"
	}

	ticket, err := h.svc.UpdateStatus(r.Context(), id, &req, actorID, actorName)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, ticket)
}

// AssignTicket assigns ticket to technician
func (h *TicketHandler) AssignTicket(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	var req model.AssignTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	actorID := middleware.GetUserID(r.Context())
	actorName := middleware.GetUserEmail(r.Context())

	ticket, err := h.svc.AssignTicket(r.Context(), id, &req, actorID, actorName)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, ticket)
}

// AddComment appends a comment to a ticket thread
func (h *TicketHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	var req model.AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}
	if err := h.authorizeTicketAccess(r, id); err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	if middleware.GetUserRole(r.Context()) == "ROLE_EMPLOYEE" {
		req.IsInternal = false
	}

	authorID := middleware.GetUserID(r.Context())
	if authorID == "" {
		authorID = "system"
	}
	authorName := middleware.GetUserEmail(r.Context())
	if authorName == "" {
		authorName = "Support Agent"
	}
	authorRole := middleware.GetUserRole(r.Context())
	if authorRole == "" {
		authorRole = "ROLE_AGENT"
	}

	comment, err := h.svc.AddComment(r.Context(), id, &req, authorID, authorName, authorRole)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, comment)
}

// ListComments gets all comments for a ticket
func (h *TicketHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}
	if err := h.authorizeTicketAccess(r, id); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	comments, err := h.svc.ListComments(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	if middleware.GetUserRole(r.Context()) == "ROLE_EMPLOYEE" {
		visible := make([]model.TicketComment, 0, len(comments))
		for _, comment := range comments {
			if !comment.IsInternal {
				visible = append(visible, comment)
			}
		}
		comments = visible
	}

	response.JSON(w, http.StatusOK, comments)
}

// ListTimeline gets audit trail for a ticket
func (h *TicketHandler) ListTimeline(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}
	if err := h.authorizeTicketAccess(r, id); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	timeline, err := h.svc.ListTimeline(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	if middleware.GetUserRole(r.Context()) == "ROLE_EMPLOYEE" {
		for i := range timeline {
			timeline[i].Notes = nil
		}
	}

	response.JSON(w, http.StatusOK, timeline)
}

// ListServiceCategories returns categories
func (h *TicketHandler) ListServiceCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.ListServiceCategories(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, cats)
}

// ListServiceCatalogItems returns service items
func (h *TicketHandler) ListServiceCatalogItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListServiceCatalogItems(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// ListTicketsByAsset returns tickets associated with an asset ID
func (h *TicketHandler) ListTicketsByAsset(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserRole(r.Context()) == "ROLE_EMPLOYEE" {
		errors.WriteHTTP(w, errors.Forbidden("asset-linked ticket lookup requires an operator role"))
		return
	}
	assetID := r.PathValue("assetId")
	if assetID == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 1 {
			assetID = parts[len(parts)-1]
		}
	}

	if assetID == "" {
		errors.WriteHTTP(w, errors.BadRequest("asset id is required"))
		return
	}

	tickets, err := h.svc.GetTicketsByAssetID(r.Context(), assetID)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, tickets)
}

func employeeCanAccessTicket(r *http.Request, ticket *model.Ticket) bool {
	return middleware.GetUserRole(r.Context()) != "ROLE_EMPLOYEE" ||
		(middleware.GetUserID(r.Context()) != "" && ticket.RequesterID == middleware.GetUserID(r.Context()))
}

func (h *TicketHandler) authorizeTicketAccess(r *http.Request, id string) error {
	if middleware.GetUserRole(r.Context()) != "ROLE_EMPLOYEE" {
		return nil
	}
	ticket, err := h.svc.GetTicket(r.Context(), id)
	if err != nil {
		return err
	}
	if !employeeCanAccessTicket(r, ticket) {
		return errors.NotFound("ticket not found")
	}
	return nil
}
