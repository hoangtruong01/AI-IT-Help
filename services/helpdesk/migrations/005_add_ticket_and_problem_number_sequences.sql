-- Allocate human-readable numbers safely under concurrency and after deletions.
CREATE SEQUENCE IF NOT EXISTS ticket_number_seq START WITH 1001;
CREATE SEQUENCE IF NOT EXISTS problem_number_seq START WITH 1001;

DO $$
DECLARE
    highest_ticket BIGINT;
    highest_problem BIGINT;
BEGIN
    SELECT COALESCE(MAX(substring(ticket_number FROM 4)::BIGINT), 1000)
    INTO highest_ticket
    FROM tickets
    WHERE ticket_number ~ '^TK-[0-9]+$';

    SELECT COALESCE(MAX(substring(problem_number FROM 5)::BIGINT), 1000)
    INTO highest_problem
    FROM problems
    WHERE problem_number ~ '^PRB-[0-9]+$';

    PERFORM setval('ticket_number_seq', GREATEST(highest_ticket, 1000), TRUE);
    PERFORM setval('problem_number_seq', GREATEST(highest_problem, 1000), TRUE);
END $$;
