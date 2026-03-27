BEGIN;

DO $$
DECLARE
    policy_record RECORD;
BEGIN
    FOR policy_record IN
        SELECT policyname
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'subscription'
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS %I ON "subscription";', policy_record.policyname);
    END LOOP;
END $$;

ALTER TABLE "subscription" DISABLE ROW LEVEL SECURITY;

DO $$
DECLARE
    policy_record RECORD;
BEGIN
    FOR policy_record IN
        SELECT policyname
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'system_events'
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS %I ON "system_events";', policy_record.policyname);
    END LOOP;
END $$;

ALTER TABLE "system_events" DISABLE ROW LEVEL SECURITY;

COMMIT;
