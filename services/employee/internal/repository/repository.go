package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"eomp/services/employee/internal/model"
)

// ErrVersionConflict indicates that an employee changed after the caller read it.
var ErrVersionConflict = errors.New("employee version conflict")

// Repository interface combining Employee and Department data operations
type Repository interface {
	// Employee methods
	ListEmployees(ctx context.Context, query model.EmployeeListQuery) (*model.EmployeeListResponse, error)
	FindEmployeeByID(ctx context.Context, id string) (*model.Employee, error)
	FindEmployeeByEmail(ctx context.Context, email string) (*model.Employee, error)
	CreateEmployee(ctx context.Context, emp *model.Employee) error
	UpdateEmployee(ctx context.Context, id string, req *model.UpdateEmployeeRequest) (*model.Employee, error)
	DeleteEmployee(ctx context.Context, id string, expectedVersion int) error

	// Department methods
	ListDepartments(ctx context.Context) ([]model.Department, error)
	FindDepartmentByID(ctx context.Context, id string) (*model.Department, error)
	CreateDepartment(ctx context.Context, dept *model.Department) error
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository creates a new postgres repository instance
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ListEmployees(ctx context.Context, query model.EmployeeListQuery) (*model.EmployeeListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	whereClauses := []string{"1=1"}
	args := []any{}
	argIndex := 1

	if query.DepartmentID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("e.department_id = $%d", argIndex))
		args = append(args, query.DepartmentID)
		argIndex++
	}

	if query.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("e.status = $%d", argIndex))
		args = append(args, query.Status)
		argIndex++
	}

	if query.Search != "" {
		pattern := "%" + strings.ToLower(query.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(e.first_name) LIKE $%d OR LOWER(e.last_name) LIKE $%d OR LOWER(e.email) LIKE $%d OR LOWER(e.job_title) LIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, pattern)
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// 1. Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employees e WHERE %s", whereSQL)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count employees: %w", err)
	}

	// 2. Query items
	offset := (query.Page - 1) * query.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			e.id, e.user_id, e.first_name, e.last_name, e.email, e.phone, e.job_title,
			e.department_id, d.name AS department_name, d.code AS department_code,
			e.manager_id, CONCAT(m.first_name, ' ', m.last_name) AS manager_name,
			e.status, e.location, TO_CHAR(e.joined_at, 'YYYY-MM-DD') AS joined_at, e.version,
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN departments d ON e.department_id = d.id
		LEFT JOIN employees m ON e.manager_id = m.id
		WHERE %s
		ORDER BY e.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query employees: %w", err)
	}
	defer rows.Close()

	employees := []model.Employee{}
	for rows.Next() {
		var emp model.Employee
		var joinedAt sql.NullString
		var managerName sql.NullString
		err := rows.Scan(
			&emp.ID, &emp.UserID, &emp.FirstName, &emp.LastName, &emp.Email, &emp.Phone, &emp.JobTitle,
			&emp.DepartmentID, &emp.DepartmentName, &emp.DepartmentCode,
			&emp.ManagerID, &managerName,
			&emp.Status, &emp.Location, &joinedAt, &emp.Version,
			&emp.CreatedAt, &emp.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee row: %w", err)
		}
		emp.FullName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
		if joinedAt.Valid {
			emp.JoinedAt = joinedAt.String
		}
		if managerName.Valid {
			name := strings.TrimSpace(managerName.String)
			if name != "" {
				emp.ManagerName = &name
			}
		}
		employees = append(employees, emp)
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))

	return &model.EmployeeListResponse{
		Data:       employees,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *postgresRepository) FindEmployeeByID(ctx context.Context, id string) (*model.Employee, error) {
	query := `
		SELECT 
			e.id, e.user_id, e.first_name, e.last_name, e.email, e.phone, e.job_title,
			e.department_id, d.name AS department_name, d.code AS department_code,
			e.manager_id, CONCAT(m.first_name, ' ', m.last_name) AS manager_name,
			e.status, e.location, TO_CHAR(e.joined_at, 'YYYY-MM-DD') AS joined_at, e.version,
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN departments d ON e.department_id = d.id
		LEFT JOIN employees m ON e.manager_id = m.id
		WHERE e.id = $1
	`
	var emp model.Employee
	var joinedAt sql.NullString
	var managerName sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&emp.ID, &emp.UserID, &emp.FirstName, &emp.LastName, &emp.Email, &emp.Phone, &emp.JobTitle,
		&emp.DepartmentID, &emp.DepartmentName, &emp.DepartmentCode,
		&emp.ManagerID, &managerName,
		&emp.Status, &emp.Location, &joinedAt, &emp.Version,
		&emp.CreatedAt, &emp.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query employee by id: %w", err)
	}
	emp.FullName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
	if joinedAt.Valid {
		emp.JoinedAt = joinedAt.String
	}
	if managerName.Valid {
		name := strings.TrimSpace(managerName.String)
		if name != "" {
			emp.ManagerName = &name
		}
	}
	return &emp, nil
}

func (r *postgresRepository) FindEmployeeByEmail(ctx context.Context, email string) (*model.Employee, error) {
	query := `
		SELECT id, email, first_name, last_name, job_title, status, version
		FROM employees
		WHERE email = $1
	`
	var emp model.Employee
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&emp.ID, &emp.Email, &emp.FirstName, &emp.LastName, &emp.JobTitle, &emp.Status, &emp.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query employee by email: %w", err)
	}
	return &emp, nil
}

func (r *postgresRepository) CreateEmployee(ctx context.Context, emp *model.Employee) error {
	query := `
		INSERT INTO employees (
			user_id, first_name, last_name, email, phone, job_title,
			department_id, manager_id, status, location, joined_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, COALESCE($11::date, CURRENT_DATE), $12, $13
		)
		RETURNING id, version, created_at, updated_at
	`
	now := time.Now()
	var joinedDate any = nil
	if emp.JoinedAt != "" {
		joinedDate = emp.JoinedAt
	}

	err := r.db.QueryRowContext(
		ctx, query,
		emp.UserID, emp.FirstName, emp.LastName, emp.Email, emp.Phone, emp.JobTitle,
		emp.DepartmentID, emp.ManagerID, emp.Status, emp.Location, joinedDate,
		now, now,
	).Scan(&emp.ID, &emp.Version, &emp.CreatedAt, &emp.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert employee: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateEmployee(ctx context.Context, id string, req *model.UpdateEmployeeRequest) (*model.Employee, error) {
	setClauses := []string{"updated_at = CURRENT_TIMESTAMP", "version = version + 1"}
	args := []any{id}
	argIndex := 2

	if req.FirstName != nil {
		setClauses = append(setClauses, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *req.FirstName)
		argIndex++
	}
	if req.LastName != nil {
		setClauses = append(setClauses, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *req.LastName)
		argIndex++
	}
	if req.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, *req.Phone)
		argIndex++
	}
	if req.JobTitle != nil {
		setClauses = append(setClauses, fmt.Sprintf("job_title = $%d", argIndex))
		args = append(args, *req.JobTitle)
		argIndex++
	}
	if req.DepartmentID != nil {
		setClauses = append(setClauses, fmt.Sprintf("department_id = $%d", argIndex))
		args = append(args, *req.DepartmentID)
		argIndex++
	}
	if req.ManagerID != nil {
		setClauses = append(setClauses, fmt.Sprintf("manager_id = $%d", argIndex))
		args = append(args, *req.ManagerID)
		argIndex++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}
	if req.Location != nil {
		setClauses = append(setClauses, fmt.Sprintf("location = $%d", argIndex))
		args = append(args, *req.Location)
		argIndex++
	}

	args = append(args, req.Version)
	query := fmt.Sprintf("UPDATE employees SET %s WHERE id = $1 AND version = $%d", strings.Join(setClauses, ", "), argIndex)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect employee update result: %w", err)
	}
	if rowsAffected == 0 {
		return nil, ErrVersionConflict
	}

	return r.FindEmployeeByID(ctx, id)
}

func (r *postgresRepository) DeleteEmployee(ctx context.Context, id string, expectedVersion int) error {
	query := "DELETE FROM employees WHERE id = $1 AND version = $2"
	result, err := r.db.ExecContext(ctx, query, id, expectedVersion)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to inspect employee delete result: %w", err)
	}
	if rowsAffected == 0 {
		return ErrVersionConflict
	}
	return nil
}

func (r *postgresRepository) ListDepartments(ctx context.Context) ([]model.Department, error) {
	query := `
		SELECT id, name, code, manager_id, parent_id, created_at, updated_at
		FROM departments
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query departments: %w", err)
	}
	defer rows.Close()

	depts := []model.Department{}
	for rows.Next() {
		var d model.Department
		if err := rows.Scan(&d.ID, &d.Name, &d.Code, &d.ManagerID, &d.ParentID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan department row: %w", err)
		}
		depts = append(depts, d)
	}
	return depts, nil
}

func (r *postgresRepository) FindDepartmentByID(ctx context.Context, id string) (*model.Department, error) {
	query := `
		SELECT id, name, code, manager_id, parent_id, created_at, updated_at
		FROM departments
		WHERE id = $1
	`
	var d model.Department
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.Name, &d.Code, &d.ManagerID, &d.ParentID, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query department by id: %w", err)
	}
	return &d, nil
}

func (r *postgresRepository) CreateDepartment(ctx context.Context, dept *model.Department) error {
	query := `
		INSERT INTO departments (name, code, manager_id, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(
		ctx, query,
		dept.Name, dept.Code, dept.ManagerID, dept.ParentID, now, now,
	).Scan(&dept.ID, &dept.CreatedAt, &dept.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert department: %w", err)
	}
	return nil
}
