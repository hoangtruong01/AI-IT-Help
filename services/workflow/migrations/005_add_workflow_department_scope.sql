ALTER TABLE workflow_instances
    ADD COLUMN IF NOT EXISTS department_id VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_wf_inst_department
    ON workflow_instances(department_id);
