BEGIN;

DO $$
DECLARE
    policy_record RECORD;
BEGIN
    FOR policy_record IN
        SELECT policyname
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'user'
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS %I ON "user";', policy_record.policyname);
    END LOOP;
END $$;

ALTER TABLE "user" DISABLE ROW LEVEL SECURITY;

DO $$
DECLARE
    policy_record RECORD;
BEGIN
    FOR policy_record IN
        SELECT policyname
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'session'
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS %I ON "session";', policy_record.policyname);
    END LOOP;
END $$;

ALTER TABLE "session" DISABLE ROW LEVEL SECURITY;

DO $$
DECLARE
    policy_record RECORD;
BEGIN
    FOR policy_record IN
        SELECT policyname
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'account'
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS %I ON "account";', policy_record.policyname);
    END LOOP;
END $$;

ALTER TABLE "account" DISABLE ROW LEVEL SECURITY;

DO $$
DECLARE
    policy_record RECORD;
BEGIN
    FOR policy_record IN
        SELECT policyname
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'verification'
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS %I ON "verification";', policy_record.policyname);
    END LOOP;
END $$;

ALTER TABLE "verification" DISABLE ROW LEVEL SECURITY;

DO $$
DECLARE
    policy_record RECORD;
BEGIN
    FOR policy_record IN
        SELECT policyname
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'jwks'
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS %I ON "jwks";', policy_record.policyname);
    END LOOP;
END $$;

ALTER TABLE "jwks" DISABLE ROW LEVEL SECURITY;

COMMIT;
