BEGIN;

ALTER TABLE "subscription" ENABLE ROW LEVEL SECURITY;

ALTER TABLE "system_events" ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS subscription_owner_select ON "subscription";
CREATE POLICY subscription_owner_select ON "subscription" FOR
SELECT USING (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS subscription_owner_insert ON "subscription";
CREATE POLICY subscription_owner_insert ON "subscription" FOR INSERT
WITH
    CHECK (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS subscription_owner_update ON "subscription";
CREATE POLICY subscription_owner_update ON "subscription"
FOR UPDATE
    USING (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    )
WITH
    CHECK (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS subscription_owner_delete ON "subscription";
CREATE POLICY subscription_owner_delete ON "subscription" FOR DELETE USING (
    "userId" = current_setting('app.user_id', true)::uuid
    OR current_setting('app.is_internal', true) = 'true'
);

DROP POLICY IF EXISTS system_events_owner_select ON "system_events";
CREATE POLICY system_events_owner_select ON "system_events" FOR
SELECT USING (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS system_events_owner_insert ON "system_events";
CREATE POLICY system_events_owner_insert ON "system_events" FOR INSERT
WITH
    CHECK (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS system_events_owner_update ON "system_events";
CREATE POLICY system_events_owner_update ON "system_events"
FOR UPDATE
    USING (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    )
WITH
    CHECK (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS system_events_owner_delete ON "system_events";
CREATE POLICY system_events_owner_delete ON "system_events" FOR DELETE USING (
    "userId" = current_setting('app.user_id', true)::uuid
    OR current_setting('app.is_internal', true) = 'true'
);

COMMIT;
