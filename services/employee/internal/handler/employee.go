package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/middleware"
	"eomp/packages/shared/pkg/response"
	"eomp/services/employee/internal/model"
	"eomp/services/employee/internal/service"
)

// EmployeeHandler handles employee and department HTTP requests
type EmployeeHandler struct {
	svc             service.EmployeeService
	assetServiceURL string
	httpClient      *http.Client
}

type employeeDirectoryEntry struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

func requireActor(w http.ResponseWriter, r *http.Request) (middleware.Actor, bool) {
	actor := middleware.GetActor(r.Context())
	if !actor.IsValid() {
		errors.WriteHTTP(w, errors.Unauthorized("valid user identity and role are required"))
		return middleware.Actor{}, false
	}
	return actor, true
}

// NewEmployeeHandler creates a new EmployeeHandler
func NewEmployeeHandler(svc service.EmployeeService) *EmployeeHandler {
	return NewEmployeeHandlerWithAssetURL(svc, "http://localhost:8083")
}

// NewEmployeeHandlerWithAssetURL creates a new EmployeeHandler with explicit asset service URL
func NewEmployeeHandlerWithAssetURL(svc service.EmployeeService, assetServiceURL string) *EmployeeHandler {
	if assetServiceURL == "" {
		assetServiceURL = "http://localhost:8083"
	}
	return &EmployeeHandler{
		svc:             svc,
		assetServiceURL: assetServiceURL,
		httpClient:      &http.Client{Timeout: 5 * time.Second},
	}
}

// ListEmployees returns paginated employees
func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	query := model.EmployeeListQuery{
		Page:         page,
		PageSize:     pageSize,
		DepartmentID: r.URL.Query().Get("department_id"),
		Status:       r.URL.Query().Get("status"),
		Search:       r.URL.Query().Get("search"),
	}
	if actor.IsEmployee() {
		if actor.DepartmentID == "" {
			errors.WriteHTTP(w, errors.Forbidden("employee department scope is required"))
			return
		}
		query.DepartmentID = actor.DepartmentID
	}

	resp, err := h.svc.ListEmployees(r.Context(), query)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	if actor.IsEmployee() {
		entries := make([]employeeDirectoryEntry, 0, len(resp.Data))
		for _, employee := range resp.Data {
			entries = append(entries, employeeDirectoryEntry{FullName: employee.FullName, Email: employee.Email})
		}
		response.JSON(w, http.StatusOK, map[string]any{
			"data": entries, "total": resp.Total, "page": resp.Page,
			"page_size": resp.PageSize, "total_pages": resp.TotalPages,
		})
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// CreateEmployee creates a new employee profile
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	if !actor.IsAdmin() && !actor.IsManager() {
		errors.WriteHTTP(w, errors.Forbidden("employee creation requires manager or admin role"))
		return
	}
	var req model.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}
	if actor.IsManager() {
		if actor.DepartmentID == "" {
			errors.WriteHTTP(w, errors.Forbidden("manager department scope is required"))
			return
		}
		req.DepartmentID = &actor.DepartmentID
	}

	emp, err := h.svc.CreateEmployee(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, emp)
}

// GetEmployee retrieves employee by ID
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		// Fallback for path extraction if pattern is generic
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	emp, err := h.svc.GetEmployeeForActor(r.Context(), id, actor)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, emp)
}

// UpdateEmployee updates employee fields
func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	if !actor.IsAdmin() && !actor.IsManager() {
		errors.WriteHTTP(w, errors.Forbidden("employee update requires manager or admin role"))
		return
	}
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	var req model.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}
	if actor.IsManager() {
		if actor.DepartmentID == "" {
			errors.WriteHTTP(w, errors.Forbidden("manager department scope is required"))
			return
		}
		existing, err := h.svc.GetEmployee(r.Context(), id)
		if err != nil {
			errors.WriteHTTP(w, err)
			return
		}
		if existing.DepartmentID == nil || *existing.DepartmentID != actor.DepartmentID {
			errors.WriteHTTP(w, errors.NotFound("employee not found"))
			return
		}
		if req.DepartmentID != nil && *req.DepartmentID != actor.DepartmentID {
			errors.WriteHTTP(w, errors.Forbidden("manager cannot move employees outside their department"))
			return
		}
	}

	emp, err := h.svc.UpdateEmployee(r.Context(), id, &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, emp)
}

// DeleteEmployee removes employee profile
func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	if !actor.IsAdmin() {
		errors.WriteHTTP(w, errors.Forbidden("employee deletion requires admin role"))
		return
	}
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	version, err := strconv.Atoi(r.URL.Query().Get("version"))
	if err != nil || version <= 0 {
		errors.WriteHTTP(w, errors.BadRequest("version query parameter is required and must be a positive integer"))
		return
	}

	if err := h.svc.DeleteEmployee(r.Context(), id, version); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "employee deleted successfully"})
}

// ListDepartments returns all active departments
func (h *EmployeeHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActor(w, r); !ok {
		return
	}
	depts, err := h.svc.ListDepartments(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, depts)
}

// CreateDepartment registers a new department
func (h *EmployeeHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	if !actor.IsAdmin() {
		errors.WriteHTTP(w, errors.Forbidden("department creation requires admin role"))
		return
	}
	var req model.CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	dept, err := h.svc.CreateDepartment(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, dept)
}

// GetEmployeeAssetHistory returns the asset assignment history for an employee
func (h *EmployeeHandler) GetEmployeeAssetHistory(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("employee id is required"))
		return
	}
	if actor.IsEmployee() {
		employee, err := h.svc.GetEmployeeForActor(r.Context(), id, actor)
		if err != nil {
			errors.WriteHTTP(w, err)
			return
		}
		if employee.UserID == nil || *employee.UserID != actor.ID {
			errors.WriteHTTP(w, errors.NotFound("employee not found"))
			return
		}
	}

	// Forward query to Asset Service
	url := fmt.Sprintf("%s/api/v1/assets/employee/%s/history", h.assetServiceURL, id)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		response.JSON(w, http.StatusOK, []any{})
		return
	}
	for _, header := range []string{"X-User-ID", "X-User-Email", "X-User-Role", "X-User-Department", "X-User-Name"} {
		if value := r.Header.Get(header); value != "" {
			req.Header.Set(header, value)
		}
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		response.JSON(w, http.StatusOK, []any{})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response.JSON(w, http.StatusOK, []any{})
		return
	}

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		response.JSON(w, http.StatusOK, []any{})
		return
	}

	response.JSON(w, http.StatusOK, result)
}
