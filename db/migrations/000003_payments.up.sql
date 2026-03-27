BEGIN;

-- Subscription table
CREATE TABLE "subscription" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    "userId" UUID UNIQUE NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    "externalId" VARCHAR(255),
    tier VARCHAR(50) NOT NULL DEFAULT 'free',
    "cancelAtPeriodEnd" BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    "currentPeriodEnd" TIMESTAMPTZ,
    "createdAt" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Events table
CREATE TABLE "events" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    "userId" UUID NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    "data" JSONB NOT NULL,
    type VARCHAR(100) NOT NULL,
    "createdAt" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_subscription_user_id ON "subscription" ("userId");

CREATE INDEX idx_events_type ON "events" ("userId", "type");

COMMIT;