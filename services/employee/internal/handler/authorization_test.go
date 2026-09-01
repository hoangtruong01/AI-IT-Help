package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"eomp/packages/shared/pkg/middleware"
	"eomp/services/employee/internal/model"
)

type authorizationServiceStub struct {
	listQuery model.EmployeeListQuery
	createReq *model.CreateEmployeeRequest
}

func (s *authorizationServiceStub) ListEmployees(_ context.Context, query model.EmployeeListQuery) (*model.EmployeeListResponse, error) {
	s.listQuery = query
	return &model.EmployeeListResponse{Data: []model.Employee{{
		FullName: "Scoped Employee", Email: "scoped@example.test", Phone: "private", JobTitle: "private",
	}}, Total: 1, Page: 1, PageSize: 20, TotalPages: 1}, nil
}
func (s *authorizationServiceStub) GetEmployee(context.Context, string) (*model.Employee, error) {
	return &model.Employee{}, nil
}
func (s *authorizationServiceStub) GetEmployeeForActor(context.Context, string, middleware.Actor) (*model.Employee, error) {
	return &model.Employee{}, nil
}
func (s *authorizationServiceStub) CreateEmployee(_ context.Context, req *model.CreateEmployeeRequest) (*model.Employee, error) {
	s.createReq = req
	return &model.Employee{DepartmentID: req.DepartmentID}, nil
}
func (s *authorizationServiceStub) UpdateEmployee(context.Context, string, *model.UpdateEmployeeRequest) (*model.Employee, error) {
	return &model.Employee{}, nil
}
func (s *authorizationServiceStub) DeleteEmployee(context.Context, string, int) error { return nil }
func (s *authorizationServiceStub) ListDepartments(context.Context) ([]model.Department, error) {
	return []model.Department{}, nil
}
func (s *authorizationServiceStub) CreateDepartment(context.Context, *model.CreateDepartmentRequest) (*model.Department, error) {
	return &model.Department{}, nil
}

func withActorHeaders(req *http.Request, id, role, department string) {
	req.Header.Set("X-User-ID", id)
	req.Header.Set("X-User-Role", role)
	req.Header.Set("X-User-Department", department)
}

func TestEmployeeDirectoryForcesDepartmentAndRedactsFields(t *testing.T) {
	stub := &authorizationServiceStub{}
	h := NewEmployeeHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees?department_id=other", nil)
	withActorHeaders(req, "user-1", "ROLE_EMPLOYEE", "dept-it")
	rec := httptest.NewRecorder()

	middleware.ExtractGatewayHeaders()(http.HandlerFunc(h.ListEmployees)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if stub.listQuery.DepartmentID != "dept-it" {
		t.Fatalf("department scope=%q", stub.listQuery.DepartmentID)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	encoded := rec.Body.String()
	if !strings.Contains(encoded, "scoped@example.test") || strings.Contains(encoded, "phone") || strings.Contains(encoded, "job_title") {
		t.Fatalf("employee directory response leaked fields: %s", encoded)
	}
}

func TestEmployeeListFailsClosedWithoutActor(t *testing.T) {
	h := NewEmployeeHandler(&authorizationServiceStub{})
	rec := httptest.NewRecorder()
	h.ListEmployees(rec, httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestManagerCreateForcesOwnDepartment(t *testing.T) {
	stub := &authorizationServiceStub{}
	h := NewEmployeeHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees", strings.NewReader(`{
		"first_name":"Scoped","last_name":"Employee","email":"scoped@example.test",
		"job_title":"Engineer","department_id":"other"
	}`))
	withActorHeaders(req, "manager-1", "ROLE_MANAGER", "dept-it")
	rec := httptest.NewRecorder()

	middleware.ExtractGatewayHeaders()(http.HandlerFunc(h.CreateEmployee)).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if stub.createReq == nil || stub.createReq.DepartmentID == nil || *stub.createReq.DepartmentID != "dept-it" {
		t.Fatalf("manager department was not forced: %#v", stub.createReq)
	}
}
