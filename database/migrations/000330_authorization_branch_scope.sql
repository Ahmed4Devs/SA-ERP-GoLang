-- +goose Up
-- +goose StatementBegin
ALTER TABLE organization.branches
    ADD CONSTRAINT uq_branches_id_entity UNIQUE (id, legal_entity_id);

-- Fail on inconsistent legacy assignments so an operator can review their
-- intended scope. Do not silently broaden, move, or discard existing grants.
ALTER TABLE authz.user_role_assignments
    ADD CONSTRAINT ck_assignment_branch_requires_entity CHECK (branch_id IS NULL OR legal_entity_id IS NOT NULL),
    ADD CONSTRAINT fk_assignment_branch_entity FOREIGN KEY (branch_id, legal_entity_id)
        REFERENCES organization.branches (id, legal_entity_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE authz.user_role_assignments
    DROP CONSTRAINT fk_assignment_branch_entity,
    DROP CONSTRAINT ck_assignment_branch_requires_entity;
ALTER TABLE organization.branches DROP CONSTRAINT uq_branches_id_entity;
-- +goose StatementEnd
