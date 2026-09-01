-- Idempotent development-only operational data for asset_db.
INSERT INTO assets (
    id, asset_tag, name, category, model, serial_number, purchase_cost,
    current_value, status, location, notes
) VALUES (
    '90000000-0000-0000-0000-000000000101', 'DEV-AST-0001',
    'Development Support Laptop', 'LAPTOP', 'Virtual Demo Model',
    'DEV-SERIAL-0001', 1200, 1000, 'IN_STOCK', 'Development Lab',
    'Created only by scripts/dev_seed.*'
)
ON CONFLICT (asset_tag) DO NOTHING;

INSERT INTO configuration_items (
    id, ci_code, name, ci_type, environment, status, asset_id, description
) VALUES (
    '90000000-0000-0000-0000-000000000102', 'DEV-CI-APP',
    'Development EOMP Application', 'APPLICATION', 'DEVELOPMENT',
    'OPERATIONAL', NULL, 'Created only by scripts/dev_seed.*'
)
ON CONFLICT (ci_code) DO NOTHING;
