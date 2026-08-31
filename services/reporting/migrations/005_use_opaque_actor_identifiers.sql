-- Identity values cross a service boundary and are therefore stored as opaque
-- identifiers. Helpdesk exposes them as strings; Reporting must not reject an
-- otherwise valid domain event merely because an upstream identity is not UUID-shaped.
ALTER TABLE raw_incident_records
    ALTER COLUMN assignee_id TYPE VARCHAR(100) USING assignee_id::text;

ALTER TABLE agent_performance
    ALTER COLUMN agent_id TYPE VARCHAR(100) USING agent_id::text;
