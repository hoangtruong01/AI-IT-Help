package service

import (
	"context"
	stdErrors "errors"
	"fmt"
	"strings"

	"eomp/packages/shared/pkg/errors"
	"eomp/services/employee/internal/model"
	"eomp/services/employee/internal/repository"
)

// EmployeeService defines employee and department business logic
type EmployeeService interface {
	ListEmployees(ctx context.Context, query model.EmployeeListQuery) (*model.EmployeeListResponse, error)
	GetEmployee(ctx context.Context, id string) (*model.Employee, error)
	CreateEmployee(ctx context.Context, req *model.CreateEmployeeRequest) (*model.Employee, error)
	UpdateEmployee(ctx context.Context, id string, req *model.UpdateEmployeeRequest) (*model.Employee, error)
	DeleteEmployee(ctx context.Context, id string, expectedVersion int) error

	ListDepartments(ctx context.Context) ([]model.Department, error)
	CreateDepartment(ctx context.Context, req *model.CreateDepartmentRequest) (*model.Department, error)
}

type employeeService struct {
	repo repository.Repository
}

// NewEmployeeService creates a new EmployeeService instance
func NewEmployeeService(repo repository.Repository) EmployeeService {
	return &employeeService{repo: repo}
}

func (s *employeeService) ListEmployees(ctx context.Context, query model.EmployeeListQuery) (*model.EmployeeListResponse, error) {
	return s.repo.ListEmployees(ctx, query)
}

func (s *employeeService) GetEmployee(ctx context.Context, id string) (*model.Employee, error) {
	if id == "" {
		return nil, errors.BadRequest("employee id is required")
	}

	emp, err := s.repo.FindEmployeeByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to get employee: %v", err))
	}
	if emp == nil {
		return nil, errors.NotFound("employee not found")
	}

	return emp, nil
}

func (s *employeeService) CreateEmployee(ctx context.Context, req *model.CreateEmployeeRequest) (*model.Employee, error) {
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.JobTitle == "" {
		return nil, errors.BadRequest("first_name, last_name, email, and job_title are required")
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	existing, err := s.repo.FindEmployeeByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to check existing employee: %v", err))
	}
	if existing != nil {
		return nil, errors.Conflict("employee with this email already exists")
	}

	status := req.Status
	if status == "" {
		status = "ACTIVE"
	}
	location := req.Location
	if location == "" {
		location = "Headquarters"
	}

	emp := &model.Employee{
		UserID:       req.UserID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        req.Phone,
		JobTitle:     req.JobTitle,
		DepartmentID: req.DepartmentID,
		ManagerID:    req.ManagerID,
		Status:       status,
		Location:     location,
		JoinedAt:     req.JoinedAt,
	}

	if err := s.repo.CreateEmployee(ctx, emp); err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to create employee: %v", err))
	}

	return s.repo.FindEmployeeByID(ctx, emp.ID)
}

func (s *employeeService) UpdateEmployee(ctx context.Context, id string, req *model.UpdateEmployeeRequest) (*model.Employee, error) {
	if id == "" {
		return nil, errors.BadRequest("employee id is required")
	}
	if req.Version <= 0 {
		return nil, errors.BadRequest("version is required for optimistic concurrency control")
	}

	existing, err := s.repo.FindEmployeeByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to verify employee: %v", err))
	}
	if existing == nil {
		return nil, errors.NotFound("employee not found")
	}
	if existing.Version != req.Version {
		return nil, errors.Conflict("employee was modified by another request; reload and retry")
	}

	updated, err := s.repo.UpdateEmployee(ctx, id, req)
	if err != nil {
		if stdErrors.Is(err, repository.ErrVersionConflict) {
			return nil, errors.Conflict("employee was modified by another request; reload and retry")
		}
		return nil, errors.InternalServerError(fmt.Sprintf("failed to update employee: %v", err))
	}
	if updated == nil {
		return nil, errors.NotFound("employee not found")
	}

	return updated, nil
}

func (s *employeeService) DeleteEmployee(ctx context.Context, id string, expectedVersion int) error {
	if id == "" {
		return errors.BadRequest("employee id is required")
	}
	if expectedVersion <= 0 {
		return errors.BadRequest("version is required for optimistic concurrency control")
	}

	existing, err := s.repo.FindEmployeeByID(ctx, id)
	if err != nil {
		return errors.InternalServerError(fmt.Sprintf("failed to verify employee: %v", err))
	}
	if existing == nil {
		return errors.NotFound("employee not found")
	}
	if existing.Version != expectedVersion {
		return errors.Conflict("employee was modified by another request; reload and retry")
	}

	if err := s.repo.DeleteEmployee(ctx, id, expectedVersion); err != nil {
		if stdErrors.Is(err, repository.ErrVersionConflict) {
			return errors.Conflict("employee was modified by another request; reload and retry")
		}
		return errors.InternalServerError(fmt.Sprintf("failed to delete employee: %v", err))
	}

	return nil
}

func (s *employeeService) ListDepartments(ctx context.Context) ([]model.Department, error) {
	return s.repo.ListDepartments(ctx)
}

func (s *employeeService) CreateDepartment(ctx context.Context, req *model.CreateDepartmentRequest) (*model.Department, error) {
	if req.Name == "" || req.Code == "" {
		return nil, errors.BadRequest("name and code are required")
	}

	dept := &model.Department{
		Name:      req.Name,
		Code:      strings.ToUpper(strings.TrimSpace(req.Code)),
		ManagerID: req.ManagerID,
		ParentID:  req.ParentID,
	}

	if err := s.repo.CreateDepartment(ctx, dept); err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to create department: %v", err))
	}

	return dept, nil
}
