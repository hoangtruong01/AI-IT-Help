package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func getDateRange(filter model.DateFilterQuery) (time.Time, time.Time) {
	now := time.Now()
	if filter.StartDate != nil && filter.EndDate != nil {
		return *filter.StartDate, *filter.EndDate
	}
	switch strings.ToLower(filter.Range) {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, now
	case "7d":
		return now.AddDate(0, 0, -7), now
	case "quarter":
		return now.AddDate(0, -3, 0), now
	case "30d", "":
		fallthrough
	default:
		return now.AddDate(0, 0, -30), now
	}
}

func (r *postgresRepository) GetExecutiveOverview(ctx context.Context, filter model.DateFilterQuery) (*model.ExecutiveOverview, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}

	startDate, endDate := getDateRange(filter)
	query := `
		SELECT 
			COALESCE(AVG(avg_mttr_minutes), 0.0) as avg_mttr,
			COALESCE(AVG(avg_mttd_minutes), 0.0) as avg_mttd,
			COALESCE(SUM(total_incidents), 0) as total_incidents,
			COALESCE(SUM(within_sla_count), 0) as total_within_sla,
			COALESCE(SUM(breached_sla_count), 0) as total_breached
		FROM sla_metrics_daily
		WHERE metric_date >= $1 AND metric_date <= $2
	`

	var avgMTTR, avgMTTD float64
	var totalIncidents, withinSLA, breachedSLA int

	err := r.db.QueryRowContext(ctx, query, startDate, endDate).Scan(&avgMTTR, &avgMTTD, &totalIncidents, &withinSLA, &breachedSLA)
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

	startDate, endDate := getDateRange(filter)
	query := `
		SELECT metric_date, total_incidents, within_sla_count, sla_compliance_pct
		FROM sla_metrics_daily
		WHERE metric_date >= $1 AND metric_date <= $2
		ORDER BY metric_date ASC
		LIMIT 30
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
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

func formatPeriodLabel(r string) string {
	now := time.Now()
	switch r {
	case "today":
		return "Today (Real-time)"
	case "7d":
		return "Last 7 Days"
	case "30d":
		return "Last 30 Days (" + now.Format("January 2006") + ")"
	case "quarter":
		return fmt.Sprintf("Current Quarter (Q%d %d)", (int(now.Month())-1)/3+1, now.Year())
	case "custom":
		return "Custom Selected Range"
	default:
		return now.Format("January 2006") + " (Month-to-Date)"
	}
}
