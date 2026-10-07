-- +goose Up
ALTER TABLE payroll.runs
 ADD COLUMN approved_by UUID REFERENCES iam.users(id),
 ADD COLUMN approved_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE payroll.runs DROP COLUMN approved_by, DROP COLUMN approved_at;
