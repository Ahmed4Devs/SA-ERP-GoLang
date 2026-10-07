-- +goose Up
CREATE TABLE workflow.approval_limits (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('SALES_INVOICE','TREASURY_PAYMENT','PAYROLL_RUN')),
    maximum_base_amount NUMERIC(24,6) NOT NULL CHECK (maximum_base_amount >= 0),
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0),
    updated_by UUID NOT NULL REFERENCES iam.users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (legal_entity_id,user_id,resource_type)
);
CREATE INDEX approval_limits_entity ON workflow.approval_limits(legal_entity_id);

-- +goose Down
DROP TABLE workflow.approval_limits;
