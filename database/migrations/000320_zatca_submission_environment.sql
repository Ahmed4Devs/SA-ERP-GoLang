-- +goose Up
-- Preserve the selected gateway environment on every submission audit row.
ALTER TABLE zatca.submissions
    ADD COLUMN IF NOT EXISTS environment VARCHAR(50) NOT NULL DEFAULT 'SIMULATION';

ALTER TABLE zatca.submissions
    DROP CONSTRAINT IF EXISTS chk_zatca_submission_environment;

ALTER TABLE zatca.submissions
    ADD CONSTRAINT chk_zatca_submission_environment
    CHECK (environment IN ('SIMULATION', 'SANDBOX', 'PRODUCTION'));

CREATE INDEX IF NOT EXISTS idx_zatca_sub_environment
    ON zatca.submissions(environment, submitted_at DESC);

-- +goose Down
DROP INDEX IF EXISTS zatca.idx_zatca_sub_environment;
ALTER TABLE zatca.submissions
    DROP CONSTRAINT IF EXISTS chk_zatca_submission_environment;
ALTER TABLE zatca.submissions
    DROP COLUMN IF EXISTS environment;
