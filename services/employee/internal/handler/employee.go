package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eomp/packages/shared/pkg/errors"
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
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	query := model.EmployeeListQuery{
		Page:         page,
		PageSize:     pageSize,
		DepartmentID: r.URL.Query().Get("department_id"),
		Status:       r.URL.Query().Get("status"),
		Search:       r.URL.Query().Get("search"),
	}

	resp, err := h.svc.ListEmployees(r.Context(), query)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// CreateEmployee creates a new employee profile
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req model.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
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
	id := r.PathValue("id")
	if id == "" {
		// Fallback for path extraction if pattern is generic
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	emp, err := h.svc.GetEmployee(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, emp)
}

// UpdateEmployee updates employee fields
func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
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

	emp, err := h.svc.UpdateEmployee(r.Context(), id, &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, emp)
}

// DeleteEmployee removes employee profile
func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	if err := h.svc.DeleteEmployee(r.Context(), id); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "employee deleted successfully"})
}

// ListDepartments returns all active departments
func (h *EmployeeHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	depts, err := h.svc.ListDepartments(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, depts)
}

// CreateDepartment registers a new department
func (h *EmployeeHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
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

	// Forward query to Asset Service
	url := fmt.Sprintf("%s/api/v1/assets/employee/%s/history", h.assetServiceURL, id)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		response.JSON(w, http.StatusOK, []any{})
		return
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

