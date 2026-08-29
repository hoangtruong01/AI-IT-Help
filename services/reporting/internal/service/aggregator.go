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
			metric_date, total_incidents, resolved_incidents, 
			within_sla_count, breached_sla_count, avg_mttr_minutes, avg_mttd_minutes, created_at
		)
		SELECT 
			CURRENT_DATE,
			COUNT(*),
			COUNT(CASE WHEN status IN ('RESOLVED', 'CLOSED') THEN 1 END),
			COUNT(CASE WHEN sla_status = 'WITHIN_SLA' THEN 1 END),
			COUNT(CASE WHEN sla_status = 'BREACHED' THEN 1 END),
			COALESCE(AVG(mttr_minutes), 0.0),
			COALESCE(AVG(mttd_minutes), 0.0),
			CURRENT_TIMESTAMP
		FROM raw_incident_records
		WHERE DATE(created_at) = CURRENT_DATE
		ON CONFLICT (metric_date) DO UPDATE SET
			total_incidents = EXCLUDED.total_incidents,
			resolved_incidents = EXCLUDED.resolved_incidents,
			within_sla_count = EXCLUDED.within_sla_count,
			breached_sla_count = EXCLUDED.breached_sla_count,
			avg_mttr_minutes = EXCLUDED.avg_mttr_minutes,
			avg_mttd_minutes = EXCLUDED.avg_mttd_minutes;
	`

	if _, err := a.db.ExecContext(ctx, dailySQL); err != nil {
		if a.logger != nil {
			a.logger.Warn("SLA rollup failed on daily metrics table", slog.Any("error", err))
		}
	}

	// 2. Rollup department SLA compliance
	deptSQL := `
		INSERT INTO department_sla_metrics (
			department, total_tickets, within_sla, breached_sla, compliance_pct, updated_at
		)
		SELECT 
			department,
			COUNT(*),
			COUNT(CASE WHEN sla_status = 'WITHIN_SLA' THEN 1 END),
			COUNT(CASE WHEN sla_status = 'BREACHED' THEN 1 END),
			CASE 
				WHEN COUNT(*) > 0 THEN (COUNT(CASE WHEN sla_status = 'WITHIN_SLA' THEN 1 END)::FLOAT / COUNT(*)::FLOAT) * 100.0
				ELSE 100.0
			END,
			CURRENT_TIMESTAMP
		FROM raw_incident_records
		GROUP BY department
		ON CONFLICT (department) DO UPDATE SET
			total_tickets = EXCLUDED.total_tickets,
			within_sla = EXCLUDED.within_sla,
			breached_sla = EXCLUDED.breached_sla,
			compliance_pct = EXCLUDED.compliance_pct,
			updated_at = CURRENT_TIMESTAMP;
	`

	if _, err := a.db.ExecContext(ctx, deptSQL); err != nil {
		if a.logger != nil {
			a.logger.Warn("SLA rollup failed on department metrics table", slog.Any("error", err))
		}
	}

	if a.logger != nil {
		a.logger.Debug("SLA daily metrics rollup completed", slog.Duration("elapsed", time.Since(start)))
	}
}
