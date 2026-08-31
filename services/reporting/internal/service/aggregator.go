package service

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"
)

// SLAAggregator defines the background rollup engine for SLA telemetry.
type SLAAggregator struct {
	db       *sql.DB
	logger   *slog.Logger
	interval time.Duration
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewSLAAggregator creates a new background aggregator instance.
func NewSLAAggregator(db *sql.DB, logger *slog.Logger, interval time.Duration) *SLAAggregator {
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	return &SLAAggregator{
		db:       db,
		logger:   logger,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start launches the background ticker worker.
func (a *SLAAggregator) Start() {
	if a.db == nil {
		if a.logger != nil {
			a.logger.Info("SLA aggregator: running in standalone mode (no database handle)")
		}
		return
	}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if a.logger != nil {
			a.logger.Info("starting SLA daily rollup aggregator worker", slog.Duration("interval", a.interval))
		}

		// Initial rollup on boot
		a.RollupOnce(context.Background())

		ticker := time.NewTicker(a.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				a.RollupOnce(ctx)
				cancel()
			case <-a.stopChan:
				if a.logger != nil {
					a.logger.Info("stopping SLA daily rollup aggregator worker")
				}
				return
			}
		}
	}()
}

// Stop signals the background worker to stop gracefully.
func (a *SLAAggregator) Stop() {
	close(a.stopChan)
	a.wg.Wait()
}

// RollupOnce performs one execution of SLA metrics aggregation from raw records.
func (a *SLAAggregator) RollupOnce(ctx context.Context) {
	if a.db == nil {
		return
	}

	start := time.Now()

	// 1. Rollup daily SLA metrics into sla_metrics_daily
	dailySQL := `
		INSERT INTO sla_metrics_daily (
			metric_date, total_incidents, within_sla_count, breached_sla_count,
			avg_mttr_minutes, avg_mttd_minutes, sla_compliance_pct, created_at
		)
		SELECT 
			CURRENT_DATE,
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED') AND sla_status = 'WITHIN_SLA'),
			COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED') AND sla_status = 'BREACHED'),
			COALESCE(AVG(mttr_minutes) FILTER (WHERE status IN ('RESOLVED', 'CLOSED')), 0.0),
			COALESCE(AVG(mttd_minutes), 0.0),
			CASE WHEN COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED')) > 0
				THEN 100.0 * COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED') AND sla_status = 'WITHIN_SLA')
					/ COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED'))
				ELSE 0.0 END,
			CURRENT_TIMESTAMP
		FROM raw_incident_records
		WHERE DATE(created_at) = CURRENT_DATE
		ON CONFLICT (metric_date) DO UPDATE SET
			total_incidents = EXCLUDED.total_incidents,
			within_sla_count = EXCLUDED.within_sla_count,
			breached_sla_count = EXCLUDED.breached_sla_count,
			avg_mttr_minutes = EXCLUDED.avg_mttr_minutes,
			avg_mttd_minutes = EXCLUDED.avg_mttd_minutes,
			sla_compliance_pct = EXCLUDED.sla_compliance_pct;
	`

	if _, err := a.db.ExecContext(ctx, dailySQL); err != nil {
		if a.logger != nil {
			a.logger.Warn("SLA rollup failed on daily metrics table", slog.Any("error", err))
		}
	}

	// 2. Rollup department SLA compliance
	deptSQL := `
		INSERT INTO department_sla_metrics (
			department_name, department_code, metric_date, total_tickets,
			within_sla_count, breached_sla_count, sla_compliance_pct, avg_mttr_minutes, created_at
		)
		SELECT 
			department,
			LOWER(REGEXP_REPLACE(department, '[^a-zA-Z0-9]+', '-', 'g')),
			CURRENT_DATE,
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED') AND sla_status = 'WITHIN_SLA'),
			COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED') AND sla_status = 'BREACHED'),
			CASE WHEN COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED')) > 0
				THEN 100.0 * COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED') AND sla_status = 'WITHIN_SLA')
					/ COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED'))
				ELSE 0.0 END,
			COALESCE(AVG(mttr_minutes) FILTER (WHERE status IN ('RESOLVED', 'CLOSED')), 0.0),
			CURRENT_TIMESTAMP
		FROM raw_incident_records
		WHERE DATE(created_at) = CURRENT_DATE
		GROUP BY department
		ON CONFLICT (metric_date, department_code) DO UPDATE SET
			department_name = EXCLUDED.department_name,
			total_tickets = EXCLUDED.total_tickets,
			within_sla_count = EXCLUDED.within_sla_count,
			breached_sla_count = EXCLUDED.breached_sla_count,
			sla_compliance_pct = EXCLUDED.sla_compliance_pct,
			avg_mttr_minutes = EXCLUDED.avg_mttr_minutes;
	`

	if _, err := a.db.ExecContext(ctx, deptSQL); err != nil {
		if a.logger != nil {
			a.logger.Warn("SLA rollup failed on department metrics table", slog.Any("error", err))
		}
	}

	categorySQL := `
		INSERT INTO category_metrics(
			category_name, category_code, icon, total_count, resolved_count,
			avg_resolution_minutes, share_pct, period, metric_date, created_at
		)
		SELECT category,
			LOWER(REGEXP_REPLACE(category, '[^a-zA-Z0-9]+', '-', 'g')),
			'i-lucide-folder', COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED')),
			COALESCE(AVG(mttr_minutes) FILTER (WHERE status IN ('RESOLVED', 'CLOSED')), 0.0),
			CASE WHEN SUM(COUNT(*)) OVER () > 0 THEN 100.0 * COUNT(*) / SUM(COUNT(*)) OVER () ELSE 0.0 END,
			TO_CHAR(CURRENT_DATE, 'YYYY-MM'), CURRENT_DATE, CURRENT_TIMESTAMP
		FROM raw_incident_records
		WHERE DATE(created_at) = CURRENT_DATE
		GROUP BY category
		ON CONFLICT(metric_date, category_code) DO UPDATE SET
			category_name=EXCLUDED.category_name, total_count=EXCLUDED.total_count,
			resolved_count=EXCLUDED.resolved_count,
			avg_resolution_minutes=EXCLUDED.avg_resolution_minutes,
			share_pct=EXCLUDED.share_pct;
	`
	if _, err := a.db.ExecContext(ctx, categorySQL); err != nil && a.logger != nil {
		a.logger.Warn("SLA rollup failed on category metrics table", slog.Any("error", err))
	}

	agentSQL := `
		INSERT INTO agent_performance(
			agent_id, agent_name, department, tickets_assigned, tickets_resolved,
			avg_mttr_minutes, sla_compliance_pct, period, metric_date, created_at, updated_at
		)
		SELECT assignee_id, MAX(assignee_name), MAX(department), COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED')),
			COALESCE(AVG(mttr_minutes) FILTER (WHERE status IN ('RESOLVED', 'CLOSED')), 0.0),
			CASE WHEN COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED')) > 0
				THEN 100.0 * COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED') AND sla_status = 'WITHIN_SLA')
					/ COUNT(*) FILTER (WHERE status IN ('RESOLVED', 'CLOSED'))
				ELSE 0.0 END,
			TO_CHAR(CURRENT_DATE, 'YYYY-MM'), CURRENT_DATE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM raw_incident_records
		WHERE DATE(created_at) = CURRENT_DATE AND assignee_id IS NOT NULL
		GROUP BY assignee_id
		ON CONFLICT(metric_date, agent_id) DO UPDATE SET
			agent_name=EXCLUDED.agent_name, department=EXCLUDED.department,
			tickets_assigned=EXCLUDED.tickets_assigned, tickets_resolved=EXCLUDED.tickets_resolved,
			avg_mttr_minutes=EXCLUDED.avg_mttr_minutes,
			sla_compliance_pct=EXCLUDED.sla_compliance_pct, updated_at=CURRENT_TIMESTAMP;
	`
	if _, err := a.db.ExecContext(ctx, agentSQL); err != nil && a.logger != nil {
		a.logger.Warn("SLA rollup failed on agent metrics table", slog.Any("error", err))
	}

	if a.logger != nil {
		a.logger.Debug("SLA daily metrics rollup completed", slog.Duration("elapsed", time.Since(start)))
	}
}
