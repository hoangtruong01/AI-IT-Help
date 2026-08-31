package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"eomp/services/reporting/internal/model"
	"eomp/services/reporting/internal/repository"
)

// Service defines the business logic methods for BI Reporting.
type Service interface {
	GetOverview(ctx context.Context, filter model.DateFilterQuery) (*model.ExecutiveOverview, error)
	GetTrends(ctx context.Context, filter model.DateFilterQuery) ([]model.IncidentTrend, error)
	GetCategories(ctx context.Context, filter model.DateFilterQuery) ([]model.CategoryBreakdown, error)
	GetDepartmentsSLA(ctx context.Context, filter model.DateFilterQuery) ([]model.DepartmentSLAMetric, error)
	GetAgents(ctx context.Context, filter model.DateFilterQuery) ([]model.AgentScorecard, error)
	ExportReport(ctx context.Context, req model.ExportReportRequest) (*model.ExportReportResponse, error)
}

type reportingService struct {
	repo repository.Repository
}

// NewService instantiates the Reporting Service.
func NewService(repo repository.Repository) Service {
	return &reportingService{repo: repo}
}

func (s *reportingService) GetOverview(ctx context.Context, filter model.DateFilterQuery) (*model.ExecutiveOverview, error) {
	ov, err := s.repo.GetExecutiveOverview(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Test Case 9.2: Guarantee zero NaN values when dataset is empty
	if ov.TotalIncidents == 0 {
		ov.SLACompliancePct = 0.0
		ov.AvgMTTRMinutes = 0.0
		ov.AvgMTTDMinutes = 0.0
	}

	return ov, nil
}

func (s *reportingService) GetTrends(ctx context.Context, filter model.DateFilterQuery) ([]model.IncidentTrend, error) {
	return s.repo.GetIncidentTrends(ctx, filter)
}

func (s *reportingService) GetCategories(ctx context.Context, filter model.DateFilterQuery) ([]model.CategoryBreakdown, error) {
	return s.repo.GetCategoryBreakdowns(ctx, filter)
}

func (s *reportingService) GetDepartmentsSLA(ctx context.Context, filter model.DateFilterQuery) ([]model.DepartmentSLAMetric, error) {
	return s.repo.GetDepartmentSLAMetrics(ctx, filter)
}

func (s *reportingService) GetAgents(ctx context.Context, filter model.DateFilterQuery) ([]model.AgentScorecard, error) {
	return s.repo.GetAgentScorecards(ctx, filter)
}

// ExportReport generates high-speed PDF / Excel / CSV documents (Test Case 9.1: < 3 seconds for 10k records).
func (s *reportingService) ExportReport(ctx context.Context, req model.ExportReportRequest) (*model.ExportReportResponse, error) {
	start := time.Now()

	limit := req.LimitRows
	if limit <= 0 {
		limit = 1000 // Default limit for export
	}
	if limit > 10000 {
		limit = 10000 // Cap at 10k for memory safety
	}

	records, err := s.repo.GetRawRecords(ctx, model.DateFilterQuery{
		Range: req.Range, StartDate: req.StartDate, EndDate: req.EndDate,
	}, limit)
	if err != nil {
		return nil, err
	}

	format := strings.ToLower(req.Format)
	if format == "" {
		format = "pdf"
	}

	var contentBytes []byte
	var mimeType string
	var filename string
	timestampStr := time.Now().Format("20060102-150405")

	switch format {
	case "csv", "excel", "xlsx":
		filename = fmt.Sprintf("eomp-operations-report-%s.csv", timestampStr)
		mimeType = "text/csv"
		contentBytes = generateCSVReport(req.Title, records)

	case "pdf":
		filename = fmt.Sprintf("eomp-executive-bi-report-%s.pdf", timestampStr)
		mimeType = "application/pdf"
		contentBytes = generatePDFDocument(req.Title, records)

	default:
		filename = fmt.Sprintf("eomp-report-%s.txt", timestampStr)
		mimeType = "text/plain"
		contentBytes = generateCSVReport(req.Title, records)
	}

	elapsed := time.Since(start).Milliseconds()

	return &model.ExportReportResponse{
		Filename:         filename,
		MimeType:         mimeType,
		ContentBase64:    base64.StdEncoding.EncodeToString(contentBytes),
		TotalRecords:     len(records),
		GeneratedAt:      time.Now().Format(time.RFC3339),
		GenerationTimeMs: elapsed,
	}, nil
}

func generateCSVReport(title string, records []model.RawIncidentRecord) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Ticket Number", "Title", "Category", "Priority", "Status", "Requester", "Assignee", "Department", "MTTD (min)", "MTTR (min)", "SLA Status", "Created At", "Resolved At"})

	for _, r := range records {
		resAtStr := "N/A"
		if r.ResolvedAt != nil {
			resAtStr = r.ResolvedAt.Format("2006-01-02 15:04:05")
		}
		_ = w.Write([]string{
			safeSpreadsheetCell(r.TicketNumber), safeSpreadsheetCell(r.Title), safeSpreadsheetCell(r.Category),
			safeSpreadsheetCell(r.Priority), safeSpreadsheetCell(r.Status), safeSpreadsheetCell(r.RequesterName),
			safeSpreadsheetCell(r.AssigneeName), safeSpreadsheetCell(r.Department), strconv.Itoa(r.MTTDMinutes),
			strconv.Itoa(r.MTTRMinutes), safeSpreadsheetCell(r.SLAStatus), r.CreatedAt.Format("2006-01-02 15:04:05"), resAtStr,
		})
	}
	w.Flush()
	return buf.Bytes()
}

func safeSpreadsheetCell(value string) string {
	if value != "" && strings.ContainsRune("=+-@", rune(value[0])) {
		return "'" + value
	}
	return value
}

func escapePDFString(s string) string {
	var ascii strings.Builder
	for _, r := range s {
		if r >= 32 && r <= 126 {
			ascii.WriteRune(r)
		} else {
			ascii.WriteByte('?')
		}
	}
	s = ascii.String()
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	return s
}

func generatePDFDocument(title string, records []model.RawIncidentRecord) []byte {
	// Calculate dynamic KPIs from actual records.
	totalRecords := len(records)
	var avgMTTR, avgMTTD, slaCompliance float64
	if totalRecords > 0 {
		var resolved, withinSLA, sumMTTR, sumMTTD int
		for _, r := range records {
			if r.Status == "RESOLVED" || r.Status == "CLOSED" {
				resolved++
				sumMTTR += r.MTTRMinutes
				if r.SLAStatus == "WITHIN_SLA" {
					withinSLA++
				}
			}
			sumMTTD += r.MTTDMinutes
		}
		if resolved > 0 {
			slaCompliance = (float64(withinSLA) / float64(resolved)) * 100.0
			avgMTTR = float64(sumMTTR) / float64(resolved)
		}
		avgMTTD = float64(sumMTTD) / float64(totalRecords)
	}

	var stream bytes.Buffer
	stream.WriteString("BT\n")
	stream.WriteString("/F1 18 Tf\n50 740 Td\n(EOMP Enterprise Operations Platform - BI Executive Summary) Tj\n")
	stream.WriteString("/F2 11 Tf\n0 -25 Td\n(" + escapePDFString(fmt.Sprintf("Report Title: %s", title)) + ") Tj\n")
	stream.WriteString("0 -16 Td\n(" + escapePDFString(fmt.Sprintf("Generated: %s | Total Records: %d", time.Now().Format("2006-01-02 15:04:05"), totalRecords)) + ") Tj\n")
	stream.WriteString("/F1 11 Tf\n0 -28 Td\n(" + escapePDFString(fmt.Sprintf("Key Metrics: Avg MTTR: %.1fm | Avg MTTD: %.1fm | SLA Compliance: %.1f%%", avgMTTR, avgMTTD, slaCompliance)) + ") Tj\n")
	stream.WriteString("/F2 9 Tf\n0 -25 Td\n(Ticket ID    | Category                 | Priority | Status   | Assignee         | MTTR | SLA Status) Tj\n")
	stream.WriteString("0 -12 Td\n(-----------------------------------------------------------------------------------------------------) Tj\n")

	displayCount := totalRecords
	if displayCount > 25 {
		displayCount = 25
	}

	for i := 0; i < displayCount; i++ {
		r := records[i]
		catTrunc := r.Category
		if len(catTrunc) > 22 {
			catTrunc = catTrunc[:22]
		}
		assigneeTrunc := r.AssigneeName
		if len(assigneeTrunc) > 16 {
			assigneeTrunc = assigneeTrunc[:16]
		}
		line := fmt.Sprintf("%-12s | %-24s | %-8s | %-8s | %-16s | %3dm | %-10s",
			r.TicketNumber, catTrunc, r.Priority, r.Status, assigneeTrunc, r.MTTRMinutes, r.SLAStatus)
		stream.WriteString(fmt.Sprintf("0 -14 Td\n(%s) Tj\n", escapePDFString(line)))
	}

	if totalRecords > displayCount {
		stream.WriteString(fmt.Sprintf("0 -18 Td\n(... and %d additional incident records processed in database query) Tj\n", totalRecords-displayCount))
	}

	stream.WriteString("ET\n")

	objects := []string{
		fmt.Sprintf("<< /Title (%s) /Creator (EOMP Reporting Service) >>", escapePDFString(title)),
		"<< /Type /Catalog /Pages 3 0 R >>",
		"<< /Type /Pages /Kids [4 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 3 0 R /MediaBox [0 0 612 792] /Contents 5 0 R /Resources << /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >> /F2 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >> >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", stream.Len(), stream.String()),
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")
	offsets := make([]int, len(objects)+1)
	for i, object := range objects {
		offsets[i+1] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xrefOffset := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for i := 1; i <= len(objects); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 2 0 R /Info 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xrefOffset)
	return buf.Bytes()
}
