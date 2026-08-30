package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"eomp/services/reporting/internal/model"
)

// Repository defines the contract for accessing reporting and telemetry data.
type Repository interface {
	GetExecutiveOverview(ctx context.Context, filter model.DateFilterQuery) (*model.ExecutiveOverview, error)
	GetIncidentTrends(ctx context.Context, filter model.DateFilterQuery) ([]model.IncidentTrend, error)
	GetCategoryBreakdowns(ctx context.Context, filter model.DateFilterQuery) ([]model.CategoryBreakdown, error)
	GetDepartmentSLAMetrics(ctx context.Context, filter model.DateFilterQuery) ([]model.DepartmentSLAMetric, error)
	GetAgentScorecards(ctx context.Context, filter model.DateFilterQuery) ([]model.AgentScorecard, error)
	GetRawRecords(ctx context.Context, limit int) ([]model.RawIncidentRecord, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository creates a new PostgreSQL reporting repository.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) GetExecutiveOverview(ctx context.Context, filter model.DateFilterQuery) (*model.ExecutiveOverview, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}

	query := `
		SELECT 
			COALESCE(AVG(avg_mttr_minutes), 0.0) as avg_mttr,
			COALESCE(AVG(avg_mttd_minutes), 0.0) as avg_mttd,
			COALESCE(SUM(total_incidents), 0) as total_incidents,
			COALESCE(SUM(within_sla_count), 0) as total_within_sla,
			COALESCE(SUM(breached_sla_count), 0) as total_breached
		FROM sla_metrics_daily
	`

	var avgMTTR, avgMTTD float64
	var totalIncidents, withinSLA, breachedSLA int

	err := r.db.QueryRowContext(ctx, query).Scan(&avgMTTR, &avgMTTD, &totalIncidents, &withinSLA, &breachedSLA)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("query executive overview: %w", err)
	}

	slaPct := 100.0
	if totalIncidents > 0 {
		slaPct = (float64(withinSLA) / float64(totalIncidents)) * 100.0
	}

	return &model.ExecutiveOverview{
		AvgMTTRMinutes:     avgMTTR,
		AvgMTTDMinutes:     avgMTTD,
		SLACompliancePct:   slaPct,
		FCRRatePct:         0,
		CSATRating:         0,
		TotalIncidents:     totalIncidents,
		TotalResolved:      withinSLA + breachedSLA,
		TotalBreached:      breachedSLA,
		MTTRImprovementPct: 0,
		PeriodLabel:        formatPeriodLabel(filter.Range),
	}, nil
}

func (r *postgresRepository) GetIncidentTrends(ctx context.Context, filter model.DateFilterQuery) ([]model.IncidentTrend, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}

	query := `
		SELECT metric_date, total_incidents, within_sla_count, sla_compliance_pct
		FROM sla_metrics_daily
		ORDER BY metric_date ASC
		LIMIT 30
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query incident trends: %w", err)
	}
	defer rows.Close()

	var list []model.IncidentTrend
	for rows.Next() {
		var metricDate time.Time
		var total, within int
		var pct float64
		if err := rows.Scan(&metricDate, &total, &within, &pct); err != nil {
			return nil, fmt.Errorf("scan incident trend: %w", err)
		}
		list = append(list, model.IncidentTrend{Date: metricDate.Format("2006-01-02"), OpenedCount: total, ResolvedCount: within, SLACompliancePct: pct})
	}
	return list, rows.Err()
}

func (r *postgresRepository) GetCategoryBreakdowns(ctx context.Context, filter model.DateFilterQuery) ([]model.CategoryBreakdown, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}

	query := `
		SELECT category_name, category_code, icon, total_count, resolved_count, avg_resolution_minutes, share_pct
		FROM category_metrics
		ORDER BY total_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query category breakdowns: %w", err)
	}
	defer rows.Close()

	var list []model.CategoryBreakdown
	for rows.Next() {
		var item model.CategoryBreakdown
		if err := rows.Scan(&item.CategoryName, &item.CategoryCode, &item.Icon, &item.TotalCount, &item.ResolvedCount, &item.AvgResolutionMinutes, &item.SharePct); err != nil {
			return nil, fmt.Errorf("scan category breakdown: %w", err)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *postgresRepository) GetDepartmentSLAMetrics(ctx context.Context, filter model.DateFilterQuery) ([]model.DepartmentSLAMetric, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}

	query := `
		SELECT department_name, department_code, total_tickets, within_sla_count, breached_sla_count, sla_compliance_pct, avg_mttr_minutes
		FROM department_sla_metrics
		ORDER BY total_tickets DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query department SLA metrics: %w", err)
	}
	defer rows.Close()

	var list []model.DepartmentSLAMetric
	for rows.Next() {
		var item model.DepartmentSLAMetric
		if err := rows.Scan(&item.DepartmentName, &item.DepartmentCode, &item.TotalTickets, &item.WithinSLACount, &item.BreachedSLACount, &item.SLACompliancePct, &item.AvgMTTRMinutes); err != nil {
			return nil, fmt.Errorf("scan department SLA metric: %w", err)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *postgresRepository) GetAgentScorecards(ctx context.Context, filter model.DateFilterQuery) ([]model.AgentScorecard, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}

	query := `
		SELECT agent_id, agent_name, agent_avatar, job_title, department, tickets_assigned, tickets_resolved, avg_mttr_minutes, csat_rating, sla_compliance_pct
		FROM agent_performance
		ORDER BY tickets_resolved DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query agent scorecards: %w", err)
	}
	defer rows.Close()

	var list []model.AgentScorecard
	for rows.Next() {
		var item model.AgentScorecard
		if err := rows.Scan(&item.AgentID, &item.AgentName, &item.AgentAvatar, &item.JobTitle, &item.Department, &item.TicketsAssigned, &item.TicketsResolved, &item.AvgMTTRMinutes, &item.CSATRating, &item.SLACompliancePct); err != nil {
			return nil, fmt.Errorf("scan agent scorecard: %w", err)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *postgresRepository) GetRawRecords(ctx context.Context, limit int) ([]model.RawIncidentRecord, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}

	query := `
		SELECT id, ticket_number, title, category, priority, status, requester_name, assignee_name, department, mttd_minutes, mttr_minutes, sla_status, created_at, resolved_at
		FROM raw_incident_records
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query raw incident records: %w", err)
	}
	defer rows.Close()

	var list []model.RawIncidentRecord
	for rows.Next() {
		var item model.RawIncidentRecord
		var resAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.TicketNumber, &item.Title, &item.Category, &item.Priority, &item.Status, &item.RequesterName, &item.AssigneeName, &item.Department, &item.MTTDMinutes, &item.MTTRMinutes, &item.SLAStatus, &item.CreatedAt, &resAt); err != nil {
			return nil, fmt.Errorf("scan raw incident record: %w", err)
		}
		if resAt.Valid {
			item.ResolvedAt = &resAt.Time
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

// Fallback & Synthetic Data Generators
func (r *postgresRepository) getMockOverview(filter model.DateFilterQuery) *model.ExecutiveOverview {
	if filter.Range == "empty" {
		return &model.ExecutiveOverview{
			AvgMTTRMinutes:     0.0,
			AvgMTTDMinutes:     0.0,
			SLACompliancePct:   0.0,
			FCRRatePct:         0.0,
			CSATRating:         0.0,
			TotalIncidents:     0,
			TotalResolved:      0,
			TotalBreached:      0,
			MTTRImprovementPct: 0.0,
			PeriodLabel:        "No Data Period",
		}
	}

	return &model.ExecutiveOverview{
		AvgMTTRMinutes:     31.8,
		AvgMTTDMinutes:     7.2,
		SLACompliancePct:   97.4,
		FCRRatePct:         88.6,
		CSATRating:         4.86,
		TotalIncidents:     606,
		TotalResolved:      590,
		TotalBreached:      16,
		MTTRImprovementPct: 18.2,
		PeriodLabel:        formatPeriodLabel(filter.Range),
	}
}

func (r *postgresRepository) getMockTrends() []model.IncidentTrend {
	now := time.Now()
	var list []model.IncidentTrend
	for i := 13; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		list = append(list, model.IncidentTrend{
			Date:             d.Format("2006-01-02"),
			OpenedCount:      40 + (i*3)%15,
			ResolvedCount:    39 + (i*3)%14,
			SLACompliancePct: 96.5 + float64(i%4)*0.8,
		})
	}
	return list
}

func (r *postgresRepository) getMockCategories() []model.CategoryBreakdown {
	return []model.CategoryBreakdown{
		{CategoryName: "Network & Connectivity", CategoryCode: "network", Icon: "i-lucide-wifi", TotalCount: 184, ResolvedCount: 178, AvgResolutionMinutes: 38.5, SharePct: 32.5},
		{CategoryName: "Account & Access (SSO/MFA)", CategoryCode: "security", Icon: "i-lucide-shield-check", TotalCount: 142, ResolvedCount: 140, AvgResolutionMinutes: 22.0, SharePct: 25.1},
		{CategoryName: "Hardware & Peripherals", CategoryCode: "hardware", Icon: "i-lucide-laptop", TotalCount: 115, ResolvedCount: 110, AvgResolutionMinutes: 45.2, SharePct: 20.3},
		{CategoryName: "Software & Applications", CategoryCode: "software", Icon: "i-lucide-app-window", TotalCount: 82, ResolvedCount: 79, AvgResolutionMinutes: 34.0, SharePct: 14.5},
		{CategoryName: "Email & Collaboration", CategoryCode: "collaboration", Icon: "i-lucide-mail", TotalCount: 43, ResolvedCount: 42, AvgResolutionMinutes: 18.4, SharePct: 7.6},
	}
}

func (r *postgresRepository) getMockDepartments() []model.DepartmentSLAMetric {
	return []model.DepartmentSLAMetric{
		{DepartmentName: "Software Engineering", DepartmentCode: "ENG", TotalTickets: 185, WithinSLACount: 181, BreachedSLACount: 4, SLACompliancePct: 97.84, AvgMTTRMinutes: 31.4},
		{DepartmentName: "Sales & Business Development", DepartmentCode: "SALES", TotalTickets: 124, WithinSLACount: 120, BreachedSLACount: 4, SLACompliancePct: 96.77, AvgMTTRMinutes: 28.5},
		{DepartmentName: "Human Resources", DepartmentCode: "HR", TotalTickets: 86, WithinSLACount: 84, BreachedSLACount: 2, SLACompliancePct: 97.67, AvgMTTRMinutes: 25.0},
		{DepartmentName: "Finance & Accounting", DepartmentCode: "FIN", TotalTickets: 92, WithinSLACount: 89, BreachedSLACount: 3, SLACompliancePct: 96.74, AvgMTTRMinutes: 33.2},
		{DepartmentName: "Marketing & Operations", DepartmentCode: "MKT", TotalTickets: 79, WithinSLACount: 78, BreachedSLACount: 1, SLACompliancePct: 98.73, AvgMTTRMinutes: 27.8},
	}
}

func (r *postgresRepository) getMockAgents() []model.AgentScorecard {
	return []model.AgentScorecard{
		{AgentID: "u1", AgentName: "Marcus Vance", AgentAvatar: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150", JobTitle: "Senior IT Ops Specialist", Department: "IT Support L2", TicketsAssigned: 68, TicketsResolved: 66, AvgMTTRMinutes: 28.4, CSATRating: 4.92, SLACompliancePct: 98.5},
		{AgentID: "u2", AgentName: "Sarah Jenkins", AgentAvatar: "https://images.unsplash.com/photo-1580489944761-15a19d654956?w=150", JobTitle: "Cybersecurity & IAM Engineer", Department: "IT Security", TicketsAssigned: 54, TicketsResolved: 53, AvgMTTRMinutes: 31.2, CSATRating: 4.88, SLACompliancePct: 98.1},
		{AgentID: "u3", AgentName: "David Kim", AgentAvatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150", JobTitle: "Systems & Cloud Administrator", Department: "DevOps & Infra", TicketsAssigned: 48, TicketsResolved: 46, AvgMTTRMinutes: 35.8, CSATRating: 4.79, SLACompliancePct: 95.8},
		{AgentID: "u4", AgentName: "Elena Rostova", AgentAvatar: "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=150", JobTitle: "IT Support Specialist L1", Department: "IT Helpdesk L1", TicketsAssigned: 72, TicketsResolved: 70, AvgMTTRMinutes: 24.6, CSATRating: 4.95, SLACompliancePct: 97.2},
		{AgentID: "u5", AgentName: "Alex Chen", AgentAvatar: "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=150", JobTitle: "Network Operations Specialist", Department: "Network Engineering", TicketsAssigned: 42, TicketsResolved: 40, AvgMTTRMinutes: 42.0, CSATRating: 4.70, SLACompliancePct: 95.2},
	}
}

func generateSyntheticRawRecords(count int) []model.RawIncidentRecord {
	if count <= 0 {
		count = 100
	}
	categories := []string{"Network & Connectivity", "Account & Access (SSO/MFA)", "Hardware & Peripherals", "Software & Applications", "Email & Collaboration"}
	priorities := []string{"URGENT", "HIGH", "MEDIUM", "LOW"}
	agents := []string{"Marcus Vance", "Sarah Jenkins", "David Kim", "Elena Rostova", "Alex Chen"}
	departments := []string{"Software Engineering", "Sales & Business Development", "Human Resources", "Finance & Accounting", "Marketing & Operations"}

	now := time.Now()
	records := make([]model.RawIncidentRecord, count)
	for i := 0; i < count; i++ {
		cat := categories[i%len(categories)]
		pri := priorities[i%len(priorities)]
		ag := agents[i%len(agents)]
		dept := departments[i%len(departments)]
		slaStatus := "WITHIN_SLA"
		if i%25 == 0 {
			slaStatus = "BREACHED"
		}
		createdAt := now.Add(-time.Duration(i*10) * time.Minute)
		resolvedAt := createdAt.Add(time.Duration(15+(i%45)) * time.Minute)

		records[i] = model.RawIncidentRecord{
			ID:            fmt.Sprintf("raw-%05d", i+1),
			TicketNumber:  fmt.Sprintf("TK-%05d", 1000+i),
			Title:         fmt.Sprintf("Operations Event Incident #%d - %s", i+1, cat),
			Category:      cat,
			Priority:      pri,
			Status:        "RESOLVED",
			RequesterName: fmt.Sprintf("Employee %d", (i%50)+1),
			AssigneeName:  ag,
			Department:    dept,
			MTTDMinutes:   5 + (i % 10),
			MTTRMinutes:   15 + (i % 45),
			SLAStatus:     slaStatus,
			CreatedAt:     createdAt,
			ResolvedAt:    &resolvedAt,
		}
	}
	return records
}

func formatPeriodLabel(r string) string {
	switch r {
	case "today":
		return "Today (Real-time)"
	case "7d":
		return "Last 7 Days"
	case "30d":
		return "Last 30 Days (August 2026)"
	case "quarter":
		return "Current Quarter (Q3 2026)"
	case "custom":
		return "Custom Selected Range"
	default:
		return "August 2026 (Month-to-Date)"
	}
}
