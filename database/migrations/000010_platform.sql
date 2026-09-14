-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS platform;

-- Fallback for environments where uuidv7() is not yet native
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_proc WHERE proname = 'uuidv7'
    ) THEN
        CREATE OR REPLACE FUNCTION uuidv7() RETURNS uuid AS $f$
        BEGIN
            RETURN gen_random_uuid();
        END;
        $f$ LANGUAGE plpgsql;
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS platform.outbox_events (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    event_type          TEXT NOT NULL,
    aggregate_type      TEXT NOT NULL,
    aggregate_id        UUID,
    payload             JSONB NOT NULL,
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    available_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at        TIMESTAMPTZ,
    attempt_count       INTEGER NOT NULL DEFAULT 0,
    last_error          TEXT,

    CONSTRAINT ck_outbox_attempt_count
        CHECK (attempt_count >= 0)
);

CREATE INDEX IF NOT EXISTS idx_outbox_pending
ON platform.outbox_events (available_at, occurred_at)
WHERE published_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_outbox_aggregate
ON platform.outbox_events (aggregate_type, aggregate_id);

CREATE TABLE IF NOT EXISTS platform.idempotency_keys (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    scope               TEXT NOT NULL,
    idempotency_key     TEXT NOT NULL,
    request_hash        TEXT NOT NULL,
    state               TEXT NOT NULL DEFAULT 'PROCESSING',
    response_status     INTEGER,
    response_body       JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ NOT NULL,

    CONSTRAINT uq_idempotency_scope_key
        UNIQUE (scope, idempotency_key),

    CONSTRAINT ck_idempotency_state
        CHECK (state IN ('PROCESSING','COMPLETED','FAILED'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS platform.idempotency_keys;
DROP TABLE IF EXISTS platform.outbox_events;
DROP SCHEMA IF EXISTS platform CASCADE;
-- +goose StatementEnd
