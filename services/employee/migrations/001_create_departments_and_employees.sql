-- =============================================================================
-- Migration: 001_create_departments_and_employees.sql
-- Database: employee_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Departments Table
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    manager_id UUID,
    parent_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_departments_code ON departments(code);

-- 2. Employees Table
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(50),
    job_title VARCHAR(150) NOT NULL,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    manager_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    location VARCHAR(100) DEFAULT 'Headquarters',
    joined_at DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_employees_manager ON employees(manager_id);
CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(status);

-- 3. Seed Initial Departments
INSERT INTO departments (id, name, code)
VALUES 
    ('d0000000-0000-0000-0000-000000000001', 'IT Operations & Infrastructure', 'IT_OPS'),
    ('d0000000-0000-0000-0000-000000000002', 'Software Engineering', 'ENG'),
    ('d0000000-0000-0000-0000-000000000003', 'Human Resources', 'HR'),
    ('d0000000-0000-0000-0000-000000000004', 'Finance & Accounting', 'FIN')
ON CONFLICT (code) DO NOTHING;

-- 4. Seed Initial Employees
INSERT INTO employees (id, user_id, first_name, last_name, email, phone, job_title, department_id, status, location)
VALUES
    (
        'e0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'System',
        'Administrator',
        'admin@eomp.local',
        '+84 901 000 001',
        'Principal Infrastructure Architect',
        'd0000000-0000-0000-0000-000000000001',
        'ACTIVE',
        'Headquarters (Building A)'
    ),
    (
        'e0000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000002',
        'David',
        'Tran',
        'manager@eomp.local',
        '+84 901 000 002',
        'IT Operations Manager',
        'd0000000-0000-0000-0000-000000000001',
        'ACTIVE',
        'Headquarters (Building A)'
    ),
    (
        'e0000000-0000-0000-0000-000000000003',
        'a0000000-0000-0000-0000-000000000003',
        'Alex',
        'Nguyen',
        'agent@eomp.local',
        '+84 901 000 003',
        'Senior IT Support Specialist',
        'd0000000-0000-0000-0000-000000000001',
        'ACTIVE',
        'Headquarters (Building A)'
    ),
    (
        'e0000000-0000-0000-0000-000000000004',
        NULL,
        'Emily',
        'Davis',
        'emily.davis@eomp.local',
        '+84 901 000 004',
        'Senior Frontend Engineer',
        'd0000000-0000-0000-0000-000000000002',
        'ACTIVE',
        'Remote (Da Nang)'
    )
ON CONFLICT (email) DO NOTHING;
