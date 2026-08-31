-- Give aggregate rows an explicit date so every reporting endpoint can honor
-- the same requested range. Remove CSAT until a real survey source exists.
ALTER TABLE category_metrics ADD COLUMN IF NOT EXISTS metric_date DATE NOT NULL DEFAULT CURRENT_DATE;
ALTER TABLE department_sla_metrics ADD COLUMN IF NOT EXISTS metric_date DATE NOT NULL DEFAULT CURRENT_DATE;
ALTER TABLE agent_performance ADD COLUMN IF NOT EXISTS metric_date DATE NOT NULL DEFAULT CURRENT_DATE;
ALTER TABLE agent_performance DROP COLUMN IF EXISTS csat_rating;

CREATE UNIQUE INDEX IF NOT EXISTS uq_category_metrics_date_code ON category_metrics(metric_date, category_code);
CREATE UNIQUE INDEX IF NOT EXISTS uq_department_metrics_date_code ON department_sla_metrics(metric_date, department_code);
CREATE UNIQUE INDEX IF NOT EXISTS uq_agent_metrics_date_agent ON agent_performance(metric_date, agent_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_raw_incident_ticket ON raw_incident_records(ticket_number);
