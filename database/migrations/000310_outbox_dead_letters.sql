-- +goose Up
ALTER TABLE platform.outbox_events
    ADD COLUMN IF NOT EXISTS dead_lettered_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_outbox_dead_letters
    ON platform.outbox_events (dead_lettered_at, occurred_at)
    WHERE dead_lettered_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS platform.idx_outbox_dead_letters;

ALTER TABLE platform.outbox_events
    DROP COLUMN IF EXISTS dead_lettered_at;
