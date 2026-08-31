package model

import "time"

// ExecutiveOverview represents the high-level KPI dashboard metrics.
type ExecutiveOverview struct {
	AvgMTTRMinutes   float64 `json:"avg_mttr_minutes"`
	AvgMTTDMinutes   float64 `json:"avg_mttd_minutes"`
	SLACompliancePct float64 `json:"sla_compliance_pct"`
	TotalIncidents   int     `json:"total_incidents"`
	TotalResolved    int     `json:"total_resolved"`
	TotalBreached    int     `json:"total_breached"`
	PeriodLabel      string  `json:"period_label"`
}

// IncidentTrend represents daily incident generation vs resolution.
type IncidentTrend struct {
	Date             string  `json:"date"`
	OpenedCount      int     `json:"opened_count"`
	ResolvedCount    int     `json:"resolved_count"`
	SLACompliancePct float64 `json:"sla_compliance_pct"`
}

// CategoryBreakdown represents incident statistics per category.
type CategoryBreakdown struct {
	CategoryName         string  `json:"category_name"`
	CategoryCode         string  `json:"category_code"`
	Icon                 string  `json:"icon"`
	TotalCount           int     `json:"total_count"`
	ResolvedCount        int     `json:"resolved_count"`
	AvgResolutionMinutes float64 `json:"avg_resolution_minutes"`
	SharePct             float64 `json:"share_pct"`
}

// DepartmentSLAMetric represents department-level SLA compliance breakdown.
type DepartmentSLAMetric struct {
	DepartmentName   string  `json:"department_name"`
	DepartmentCode   string  `json:"department_code"`
	TotalTickets     int     `json:"total_tickets"`
	WithinSLACount   int     `json:"within_sla_count"`
	BreachedSLACount int     `json:"breached_sla_count"`
	SLACompliancePct float64 `json:"sla_compliance_pct"`
	AvgMTTRMinutes   float64 `json:"avg_mttr_minutes"`
}

// AgentScorecard represents technician performance rankings and metrics.
type AgentScorecard struct {
	AgentID          string  `json:"agent_id"`
	AgentName        string  `json:"agent_name"`
	AgentAvatar      string  `json:"agent_avatar"`
	JobTitle         string  `json:"job_title"`
	Department       string  `json:"department"`
	TicketsAssigned  int     `json:"tickets_assigned"`
	TicketsResolved  int     `json:"tickets_resolved"`
	AvgMTTRMinutes   float64 `json:"avg_mttr_minutes"`
	SLACompliancePct float64 `json:"sla_compliance_pct"`
}

// DateFilterQuery captures time range filters.
type DateFilterQuery struct {
	Range     string     `json:"range"` // today, 7d, 30d, quarter, custom
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// RawIncidentRecord represents detailed records for export.
type RawIncidentRecord struct {
	ID            string     `json:"id"`
	TicketNumber  string     `json:"ticket_number"`
	Title         string     `json:"title"`
	Category      string     `json:"category"`
	Priority      string     `json:"priority"`
	Status        string     `json:"status"`
	RequesterName string     `json:"requester_name"`
	AssigneeName  string     `json:"assignee_name"`
	Department    string     `json:"department"`
	MTTDMinutes   int        `json:"mttd_minutes"`
	MTTRMinutes   int        `json:"mttr_minutes"`
	SLAStatus     string     `json:"sla_status"`
	CreatedAt     time.Time  `json:"created_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
}

// ExportReportRequest defines request body for exporting reports.
type ExportReportRequest struct {
	Format    string     `json:"format"` // pdf, csv, excel
	Range     string     `json:"range"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Title     string     `json:"title"`
	LimitRows int        `json:"limit_rows,omitempty"`
}

// ExportReportResponse contains generated document content.
type ExportReportResponse struct {
	Filename         string `json:"filename"`
	MimeType         string `json:"mime_type"`
	ContentBase64    string `json:"content_base64"`
	TotalRecords     int    `json:"total_records"`
	GeneratedAt      string `json:"generated_at"`
	GenerationTimeMs int64  `json:"generation_time_ms"`
}
