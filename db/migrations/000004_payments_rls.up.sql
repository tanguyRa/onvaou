BEGIN;

ALTER TABLE "subscription" ENABLE ROW LEVEL SECURITY;

ALTER TABLE "events" ENABLE ROW LEVEL SECURITY;

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

DROP POLICY IF EXISTS events_owner_select ON "events";
CREATE POLICY events_owner_select ON "events" FOR
SELECT USING (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS events_owner_insert ON "events";
CREATE POLICY events_owner_insert ON "events" FOR INSERT
WITH
    CHECK (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS events_owner_update ON "events";
CREATE POLICY events_owner_update ON "events"
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

DROP POLICY IF EXISTS events_owner_delete ON "events";
CREATE POLICY events_owner_delete ON "events" FOR DELETE USING (
    "userId" = current_setting('app.user_id', true)::uuid
    OR current_setting('app.is_internal', true) = 'true'
);

COMMIT;
