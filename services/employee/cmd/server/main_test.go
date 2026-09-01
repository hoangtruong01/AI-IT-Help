package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eomp/packages/shared/pkg/middleware"
	"eomp/services/employee/internal/config"
	"eomp/services/employee/internal/handler"
	"eomp/services/employee/internal/model"
	"eomp/services/employee/internal/service"
)

func TestHealthHandler(t *testing.T) {
	cfg := config.Load()
	h := handler.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["service"] != "employee" {
		t.Errorf("expected service 'employee', got '%v'", body["service"])
	}
}

// Mock Repository for Service testing
type mockRepo struct{}

func (m *mockRepo) ListEmployees(ctx context.Context, query model.EmployeeListQuery) (*model.EmployeeListResponse, error) {
	return &model.EmployeeListResponse{
		Data:       []model.Employee{{ID: "emp-1", FirstName: "Alex", LastName: "Nguyen", Email: "alex@eomp.local", JobTitle: "Dev"}},
		Total:      1,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}, nil
}
func (m *mockRepo) FindEmployeeByID(ctx context.Context, id string) (*model.Employee, error) {
	return &model.Employee{ID: id, FirstName: "Alex", LastName: "Nguyen", Email: "alex@eomp.local", JobTitle: "Dev", Version: 1}, nil
}
func (m *mockRepo) FindEmployeeByIDForActor(ctx context.Context, id string, actor middleware.Actor) (*model.Employee, error) {
	return m.FindEmployeeByID(ctx, id)
}
func (m *mockRepo) FindEmployeeByEmail(ctx context.Context, email string) (*model.Employee, error) {
	return nil, nil
}
func (m *mockRepo) CreateEmployee(ctx context.Context, emp *model.Employee) error {
	emp.ID = "generated-uuid"
	return nil
}
func (m *mockRepo) UpdateEmployee(ctx context.Context, id string, req *model.UpdateEmployeeRequest) (*model.Employee, error) {
	return &model.Employee{ID: id, FirstName: "Updated", Version: req.Version + 1}, nil
}
func (m *mockRepo) DeleteEmployee(ctx context.Context, id string, expectedVersion int) error {
	return nil
}
func (m *mockRepo) ListDepartments(ctx context.Context) ([]model.Department, error) {
	return []model.Department{{ID: "dept-1", Name: "IT", Code: "IT_OPS"}}, nil
}
func (m *mockRepo) FindDepartmentByID(ctx context.Context, id string) (*model.Department, error) {
	return &model.Department{ID: id, Name: "IT", Code: "IT_OPS"}, nil
}
func (m *mockRepo) CreateDepartment(ctx context.Context, dept *model.Department) error {
	dept.ID = "dept-uuid"
	return nil
}

func TestEmployeeServiceValidation(t *testing.T) {
	svc := service.NewEmployeeService(&mockRepo{})
	ctx := context.Background()

	// Missing required fields
	_, err := svc.CreateEmployee(ctx, &model.CreateEmployeeRequest{
		FirstName: "John",
		// missing last name, email, job title
	})
	if err == nil {
		t.Errorf("expected validation error for missing fields")
	}

	// Valid creation
	created, err := svc.CreateEmployee(ctx, &model.CreateEmployeeRequest{
		FirstName: "Alex",
		LastName:  "Nguyen",
		Email:     "alex@eomp.local",
		JobTitle:  "Dev",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created == nil || created.ID != "generated-uuid" {
		t.Errorf("expected employee with ID 'generated-uuid'")
	}

	if _, err := svc.UpdateEmployee(ctx, "emp-1", &model.UpdateEmployeeRequest{}); err == nil {
		t.Error("expected update without version to fail")
	}
	if err := svc.DeleteEmployee(ctx, "emp-1", 0); err == nil {
		t.Error("expected delete without version to fail")
	}
}
