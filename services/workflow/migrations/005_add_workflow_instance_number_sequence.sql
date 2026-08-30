-- Allocate workflow numbers safely under concurrent requests and after deletions.
CREATE SEQUENCE IF NOT EXISTS workflow_instance_number_seq START WITH 1001;
CREATE SEQUENCE IF NOT EXISTS change_number_seq START WITH 2001;

DO $$
DECLARE
    highest_number BIGINT;
	 highest_change BIGINT;
BEGIN
    SELECT COALESCE(MAX(substring(instance_number FROM 5)::BIGINT), 1000)
    INTO highest_number
    FROM workflow_instances
    WHERE instance_number ~ '^WFI-[0-9]+$';

    PERFORM setval('workflow_instance_number_seq', GREATEST(highest_number, 1000), TRUE);

	SELECT COALESCE(MAX(substring(change_number FROM 5)::BIGINT), 2000)
	INTO highest_change
	FROM change_requests
	WHERE change_number ~ '^CHG-[0-9]+$';

	PERFORM setval('change_number_seq', GREATEST(highest_change, 2000), TRUE);
END $$;
