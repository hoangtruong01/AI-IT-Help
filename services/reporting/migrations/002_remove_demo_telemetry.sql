-- Remove the fixed portfolio/demo telemetry without touching independently-created rows.
DELETE FROM agent_performance
WHERE agent_id IN (
    'a1000000-0000-0000-0000-000000000003'::uuid,
    'a1000000-0000-0000-0000-000000000004'::uuid,
    'a1000000-0000-0000-0000-000000000005'::uuid,
    'a1000000-0000-0000-0000-000000000006'::uuid,
    'a1000000-0000-0000-0000-000000000007'::uuid
);

DELETE FROM category_metrics
WHERE period = '2026-08' AND category_code IN ('network', 'security', 'hardware', 'software', 'collaboration');

DELETE FROM department_sla_metrics
WHERE period = '2026-08' AND department_code IN ('ENG', 'SALES', 'HR', 'FIN', 'MKT');

DELETE FROM raw_incident_records
WHERE ticket_number IN ('TK-1001', 'TK-1002', 'TK-1003', 'TK-1004', 'TK-1005');

DELETE FROM sla_metrics_daily
WHERE (total_incidents, within_sla_count, breached_sla_count, avg_mttd_minutes, avg_mttr_minutes, sla_compliance_pct) IN (
    (42, 41, 1, 8.5, 38.2, 97.62),
    (38, 37, 1, 7.2, 34.0, 97.37),
    (45, 43, 2, 9.0, 41.5, 95.56),
    (50, 48, 2, 8.0, 36.8, 96.00),
    (52, 51, 1, 6.8, 32.4, 98.08),
    (28, 28, 0, 5.5, 26.0, 100.00),
    (24, 24, 0, 5.0, 24.5, 100.00),
    (48, 46, 2, 7.8, 35.0, 95.83),
    (54, 52, 2, 8.2, 37.5, 96.30),
    (60, 58, 2, 7.5, 33.2, 96.67),
    (49, 48, 1, 6.9, 31.0, 97.96),
    (55, 53, 2, 7.1, 32.8, 96.36),
    (41, 40, 1, 6.5, 29.4, 97.56),
    (35, 34, 1, 6.2, 28.0, 97.14)
);
