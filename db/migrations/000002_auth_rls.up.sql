BEGIN;

ALTER TABLE "user" ENABLE ROW LEVEL SECURITY;

ALTER TABLE "session" ENABLE ROW LEVEL SECURITY;

ALTER TABLE "account" ENABLE ROW LEVEL SECURITY;

ALTER TABLE "verification" ENABLE ROW LEVEL SECURITY;

ALTER TABLE "jwks" ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS user_owner_select ON "user";
CREATE POLICY user_owner_select ON "user" FOR
SELECT USING (
        id = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS user_owner_insert ON "user";
CREATE POLICY user_owner_insert ON "user" FOR INSERT
WITH
    CHECK (
        id = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS user_owner_update ON "user";
CREATE POLICY user_owner_update ON "user"
FOR UPDATE
    USING (
        id = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    )
WITH
    CHECK (
        id = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS user_owner_delete ON "user";
CREATE POLICY user_owner_delete ON "user" FOR DELETE USING (
    id = current_setting('app.user_id', true)::uuid
    OR current_setting('app.is_internal', true) = 'true'
);

DROP POLICY IF EXISTS session_owner_select ON "session";
CREATE POLICY session_owner_select ON "session" FOR
SELECT USING (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS session_owner_insert ON "session";
CREATE POLICY session_owner_insert ON "session" FOR INSERT
WITH
    CHECK (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS session_owner_update ON "session";
CREATE POLICY session_owner_update ON "session"
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

DROP POLICY IF EXISTS session_owner_delete ON "session";
CREATE POLICY session_owner_delete ON "session" FOR DELETE USING (
    "userId" = current_setting('app.user_id', true)::uuid
    OR current_setting('app.is_internal', true) = 'true'
);

DROP POLICY IF EXISTS account_owner_select ON "account";
CREATE POLICY account_owner_select ON "account" FOR
SELECT USING (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS account_owner_insert ON "account";
CREATE POLICY account_owner_insert ON "account" FOR INSERT
WITH
    CHECK (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS account_owner_update ON "account";
CREATE POLICY account_owner_update ON "account"
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

DROP POLICY IF EXISTS account_owner_delete ON "account";
CREATE POLICY account_owner_delete ON "account" FOR DELETE USING (
    "userId" = current_setting('app.user_id', true)::uuid
    OR current_setting('app.is_internal', true) = 'true'
);

DROP POLICY IF EXISTS verification_owner_select ON "verification";
CREATE POLICY verification_owner_select ON "verification" FOR
SELECT USING (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS verification_owner_insert ON "verification";
CREATE POLICY verification_owner_insert ON "verification" FOR INSERT
WITH
    CHECK (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS verification_owner_update ON "verification";
CREATE POLICY verification_owner_update ON "verification"
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

DROP POLICY IF EXISTS verification_owner_delete ON "verification";
CREATE POLICY verification_owner_delete ON "verification" FOR DELETE USING (
    "userId" = current_setting('app.user_id', true)::uuid
    OR current_setting('app.is_internal', true) = 'true'
);

DROP POLICY IF EXISTS jwks_owner_select ON "jwks";
CREATE POLICY jwks_owner_select ON "jwks" FOR
SELECT USING (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS jwks_owner_insert ON "jwks";
CREATE POLICY jwks_owner_insert ON "jwks" FOR INSERT
WITH
    CHECK (
        "userId" = current_setting('app.user_id', true)::uuid
        OR current_setting('app.is_internal', true) = 'true'
    );

DROP POLICY IF EXISTS jwks_owner_update ON "jwks";
CREATE POLICY jwks_owner_update ON "jwks"
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

DROP POLICY IF EXISTS jwks_owner_delete ON "jwks";
CREATE POLICY jwks_owner_delete ON "jwks" FOR DELETE USING (
    "userId" = current_setting('app.user_id', true)::uuid
    OR current_setting('app.is_internal', true) = 'true'
);

COMMIT;
