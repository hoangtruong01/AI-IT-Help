package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"eomp/services/reporting/internal/model"

	_ "github.com/lib/pq"
)

func getReportingIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	required := os.Getenv("INTEGRATION_REQUIRED") != ""
	dsn := os.Getenv("REPORTING_INTEGRATION_DSN")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}
	if dsn == "" {
		if required {
			t.Fatal("REPORTING_INTEGRATION_DSN is required")
		}
		t.Skip("skipping reporting PostgreSQL integration test (REPORTING_INTEGRATION_DSN not set)")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		if required {
			t.Fatalf("open reporting PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot open reporting PostgreSQL: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		if required {
			t.Fatalf("ping reporting PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot ping reporting PostgreSQL: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestReportingIntegration_CustomDateRangeFiltersAndAggregates(t *testing.T) {
	db := getReportingIntegrationDB(t)
	repo := NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	base := time.Date(2090, 1, 1, 12, 0, 0, 0, time.UTC).AddDate(0, 0, int(time.Now().UnixNano()%1500))
	insideA := base
	insideB := base.AddDate(0, 0, 1)
	outside := base.AddDate(0, 0, 3)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	insideCategory := "GATE_D_INSIDE_" + suffix
	outsideCategory := "GATE_D_OUTSIDE_" + suffix
	insideTicket := "TK-GD-IN-" + suffix
	outsideTicket := "TK-GD-OUT-" + suffix

	for _, row := range []struct {
		date                      time.Time
		total, within, breached   int
		mttd, mttr, compliancePct float64
	}{
		{insideA, 10, 8, 2, 5, 10, 80},
		{insideB, 20, 17, 3, 15, 30, 85},
		{outside, 1000, 1, 999, 999, 999, 0.1},
	} {
		_, err := db.ExecContext(ctx, `
			INSERT INTO sla_metrics_daily(metric_date,total_incidents,within_sla_count,breached_sla_count,avg_mttd_minutes,avg_mttr_minutes,sla_compliance_pct)
			VALUES($1,$2,$3,$4,$5,$6,$7)
		`, row.date, row.total, row.within, row.breached, row.mttd, row.mttr, row.compliancePct)
		if err != nil {
			t.Fatalf("insert SLA metric for %s: %v", row.date.Format("2006-01-02"), err)
		}
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO category_metrics(category_name,category_code,total_count,resolved_count,avg_resolution_minutes,share_pct,period,metric_date)
		VALUES ('Inside category',$1,7,6,12,70,'gate-d',$2), ('Outside category',$3,99,1,999,99,'gate-d',$4)
	`, insideCategory, insideA, outsideCategory, outside)
	if err != nil {
		t.Fatalf("insert category metrics: %v", err)
	}
	for _, row := range []struct {
		ticket, category string
		created          time.Time
	}{
		{insideTicket, insideCategory, insideA.Add(2 * time.Hour)},
		{outsideTicket, outsideCategory, outside.Add(2 * time.Hour)},
	} {
		_, err := db.ExecContext(ctx, `
			INSERT INTO raw_incident_records(ticket_number,title,category,priority,status,requester_name,assignee_name,department,mttd_minutes,mttr_minutes,sla_status,created_at,source_event_at)
			VALUES($1,'Gate D reporting integration',$2,'HIGH','RESOLVED','Requester','Agent','IT',5,20,'WITHIN',$3,$3)
		`, row.ticket, row.category, row.created)
		if err != nil {
			t.Fatalf("insert raw incident %s: %v", row.ticket, err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM raw_incident_records WHERE ticket_number IN ($1,$2)`, insideTicket, outsideTicket)
		_, _ = db.Exec(`DELETE FROM category_metrics WHERE category_code IN ($1,$2)`, insideCategory, outsideCategory)
		_, _ = db.Exec(`DELETE FROM sla_metrics_daily WHERE metric_date IN ($1,$2,$3)`, insideA, insideB, outside)
	})

	start := time.Date(insideA.Year(), insideA.Month(), insideA.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(insideB.Year(), insideB.Month(), insideB.Day(), 23, 59, 59, 0, time.UTC)
	filter := model.DateFilterQuery{Range: "custom", StartDate: &start, EndDate: &end}

	overview, err := repo.GetExecutiveOverview(ctx, filter)
	if err != nil {
		t.Fatalf("get executive overview: %v", err)
	}
	if overview.TotalIncidents != 30 || overview.TotalResolved != 30 || overview.TotalBreached != 5 {
		t.Fatalf("unexpected bounded aggregates: %+v", overview)
	}
	if math.Abs(overview.AvgMTTDMinutes-10) > 0.001 || math.Abs(overview.AvgMTTRMinutes-20) > 0.001 {
		t.Fatalf("unexpected bounded averages: %+v", overview)
	}

	trends, err := repo.GetIncidentTrends(ctx, filter)
	if err != nil {
		t.Fatalf("get incident trends: %v", err)
	}
	if len(trends) != 2 || trends[0].OpenedCount != 10 || trends[1].OpenedCount != 20 {
		t.Fatalf("date filter included or excluded incorrect trend rows: %+v", trends)
	}
	categories, err := repo.GetCategoryBreakdowns(ctx, filter)
	if err != nil {
		t.Fatalf("get category breakdowns: %v", err)
	}
	if len(categories) != 1 || categories[0].CategoryCode != insideCategory {
		t.Fatalf("category date filter returned unexpected rows: %+v", categories)
	}
	records, err := repo.GetRawRecords(ctx, filter, 10)
	if err != nil {
		t.Fatalf("get raw incident records: %v", err)
	}
	if len(records) != 1 || records[0].TicketNumber != insideTicket {
		t.Fatalf("raw record date filter returned unexpected rows: %+v", records)
	}
}
