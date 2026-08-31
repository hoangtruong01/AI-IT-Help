package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"eomp/packages/shared/pkg/eventbus"
	"eomp/services/reporting/internal/model"
)

// Repository defines the contract for accessing reporting and telemetry data.
type Repository interface {
	GetExecutiveOverview(ctx context.Context, filter model.DateFilterQuery) (*model.ExecutiveOverview, error)
	GetIncidentTrends(ctx context.Context, filter model.DateFilterQuery) ([]model.IncidentTrend, error)
	GetCategoryBreakdowns(ctx context.Context, filter model.DateFilterQuery) ([]model.CategoryBreakdown, error)
	GetDepartmentSLAMetrics(ctx context.Context, filter model.DateFilterQuery) ([]model.DepartmentSLAMetric, error)
	GetAgentScorecards(ctx context.Context, filter model.DateFilterQuery) ([]model.AgentScorecard, error)
	GetRawRecords(ctx context.Context, filter model.DateFilterQuery, limit int) ([]model.RawIncidentRecord, error)
	ProjectTicketEvent(ctx context.Context, event eventbus.Event) error
}

type ticketProjection struct {
	TicketNumber  string     `json:"ticket_number"`
	Title         string     `json:"title"`
	Category      string     `json:"category"`
	Priority      string     `json:"priority"`
	Status        string     `json:"status"`
	RequesterName string     `json:"requester_name"`
	AssigneeID    string     `json:"assignee_id"`
	AssigneeName  string     `json:"assignee_name"`
	DepartmentID  string     `json:"department_id"`
	SLAStatus     string     `json:"sla_status"`
	CreatedAt     time.Time  `json:"created_at"`
	RespondedAt   *time.Time `json:"responded_at"`
	ResolvedAt    *time.Time `json:"resolved_at"`
}

func (r *postgresRepository) ProjectTicketEvent(ctx context.Context, event eventbus.Event) error {
	if event.ID == "" {
		return fmt.Errorf("ticket event id is required")
	}
	var data ticketProjection
	payload, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("marshal ticket event data: %w", err)
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("decode ticket event data: %w", err)
	}
	if data.TicketNumber == "" || data.CreatedAt.IsZero() {
		return fmt.Errorf("ticket event is missing required projection fields")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin ticket projection: %w", err)
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `
		INSERT INTO reporting_processed_events(event_id, event_type) VALUES($1, $2)
		ON CONFLICT(event_id) DO NOTHING
	`, event.ID, event.Type)
	if err != nil {
		return fmt.Errorf("record processed ticket event: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check processed ticket event: %w", err)
	}
	if rows == 0 {
		return tx.Commit()
	}

	mttd, mttr := 0, 0
	if data.RespondedAt != nil {
		mttd = int(data.RespondedAt.Sub(data.CreatedAt).Minutes())
	}
	if data.ResolvedAt != nil {
		mttr = int(data.ResolvedAt.Sub(data.CreatedAt).Minutes())
	}
	eventAt := event.Timestamp
	if eventAt.IsZero() {
		eventAt = time.Now()
	}
	if data.RespondedAt == nil && data.Status != "OPEN" {
		mttd = int(eventAt.Sub(data.CreatedAt).Minutes())
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO raw_incident_records(
			ticket_number, title, category, priority, status, requester_name,
			assignee_id, assignee_name, department, mttd_minutes, mttr_minutes,
			sla_status, created_at, resolved_at, source_event_at
		) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT(ticket_number) DO UPDATE SET
			title=EXCLUDED.title, category=EXCLUDED.category, priority=EXCLUDED.priority,
			status=EXCLUDED.status, requester_name=EXCLUDED.requester_name,
			assignee_id=EXCLUDED.assignee_id, assignee_name=EXCLUDED.assignee_name,
			department=EXCLUDED.department,
			mttd_minutes=CASE WHEN raw_incident_records.mttd_minutes=0 THEN EXCLUDED.mttd_minutes ELSE raw_incident_records.mttd_minutes END,
			mttr_minutes=CASE WHEN EXCLUDED.mttr_minutes>0 THEN EXCLUDED.mttr_minutes ELSE raw_incident_records.mttr_minutes END,
			sla_status=EXCLUDED.sla_status,
			created_at=EXCLUDED.created_at, resolved_at=EXCLUDED.resolved_at,
			source_event_at=EXCLUDED.source_event_at
		WHERE raw_incident_records.source_event_at <= EXCLUDED.source_event_at
	`, data.TicketNumber, data.Title, data.Category, data.Priority, data.Status, data.RequesterName,
		data.AssigneeID, data.AssigneeName, data.DepartmentID, mttd, mttr, data.SLAStatus,
		data.CreatedAt, data.ResolvedAt, eventAt)
	if err != nil {
		return fmt.Errorf("upsert raw ticket projection: %w", err)
	}
	return tx.Commit()
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository creates a new PostgreSQL reporting repository.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func getDateRange(filter model.DateFilterQuery) (time.Time, time.Time, error) {
	now := time.Now()
	if strings.EqualFold(filter.Range, "custom") {
		if filter.StartDate == nil || filter.EndDate == nil {
			return time.Time{}, time.Time{}, fmt.Errorf("custom range requires start_date and end_date")
		}
		if filter.StartDate.After(*filter.EndDate) {
			return time.Time{}, time.Time{}, fmt.Errorf("start_date must not be after end_date")
		}
		return *filter.StartDate, *filter.EndDate, nil
	}
	switch strings.ToLower(filter.Range) {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, now, nil
	case "7d":
		return now.AddDate(0, 0, -7), now, nil
	case "quarter":
		quarterStartMonth := time.Month(((int(now.Month())-1)/3)*3 + 1)
		start := time.Date(now.Year(), quarterStartMonth, 1, 0, 0, 0, 0, now.Location())
		return start, now, nil
	case "30d", "":
		return now.AddDate(0, 0, -30), now, nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unsupported date range %q", filter.Range)
	}
}

func (r *postgresRepository) GetExecutiveOverview(ctx context.Context, filter model.DateFilterQuery) (*model.ExecutiveOverview, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}

	startDate, endDate, err := getDateRange(filter)
	if err != nil {
		return nil, err
	}
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

	err = r.db.QueryRowContext(ctx, query, startDate, endDate).Scan(&avgMTTR, &avgMTTD, &totalIncidents, &withinSLA, &breachedSLA)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("query executive overview: %w", err)
	}

	slaPct := 0.0
	resolvedCount := withinSLA + breachedSLA
	if resolvedCount > 0 {
		slaPct = (float64(withinSLA) / float64(resolvedCount)) * 100.0
	}

	return &model.ExecutiveOverview{
		AvgMTTRMinutes:   avgMTTR,
		AvgMTTDMinutes:   avgMTTD,
		SLACompliancePct: slaPct,
		TotalIncidents:   totalIncidents,
		TotalResolved:    withinSLA + breachedSLA,
		TotalBreached:    breachedSLA,
		PeriodLabel:      formatPeriodLabel(filter.Range),
	}, nil
}

func (r *postgresRepository) GetIncidentTrends(ctx context.Context, filter model.DateFilterQuery) ([]model.IncidentTrend, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}

	startDate, endDate, err := getDateRange(filter)
	if err != nil {
		return nil, err
	}
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

	startDate, endDate, err := getDateRange(filter)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT category_name, category_code, icon, total_count, resolved_count, avg_resolution_minutes, share_pct
		FROM category_metrics
		WHERE metric_date >= $1 AND metric_date <= $2
		ORDER BY total_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
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

	startDate, endDate, err := getDateRange(filter)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT department_name, department_code, total_tickets, within_sla_count, breached_sla_count, sla_compliance_pct, avg_mttr_minutes
		FROM department_sla_metrics
		WHERE metric_date >= $1 AND metric_date <= $2
		ORDER BY total_tickets DESC
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
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

	startDate, endDate, err := getDateRange(filter)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT agent_id, agent_name, agent_avatar, job_title, department, tickets_assigned, tickets_resolved, avg_mttr_minutes, sla_compliance_pct
		FROM agent_performance
		WHERE metric_date >= $1 AND metric_date <= $2
		ORDER BY tickets_resolved DESC
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query agent scorecards: %w", err)
	}
	defer rows.Close()

	var list []model.AgentScorecard
	for rows.Next() {
		var item model.AgentScorecard
		if err := rows.Scan(&item.AgentID, &item.AgentName, &item.AgentAvatar, &item.JobTitle, &item.Department, &item.TicketsAssigned, &item.TicketsResolved, &item.AvgMTTRMinutes, &item.SLACompliancePct); err != nil {
			return nil, fmt.Errorf("scan agent scorecard: %w", err)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *postgresRepository) GetRawRecords(ctx context.Context, filter model.DateFilterQuery, limit int) ([]model.RawIncidentRecord, error) {
	if r.db == nil {
		return nil, fmt.Errorf("reporting database is unavailable")
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}

	startDate, endDate, err := getDateRange(filter)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT id, ticket_number, title, category, priority, status, requester_name, assignee_name, department, mttd_minutes, mttr_minutes, sla_status, created_at, resolved_at
		FROM raw_incident_records
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate, limit)
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
