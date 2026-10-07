-- +goose Up
-- +goose StatementBegin

ALTER TABLE organization.departments
    ADD CONSTRAINT uq_departments_id_entity UNIQUE (id, legal_entity_id);
ALTER TABLE organization.cost_centers
    ADD CONSTRAINT uq_cost_centers_id_entity UNIQUE (id, legal_entity_id);

-- Validate historical rows too. If legacy dimensions cross company boundaries,
-- stop the upgrade for accounting review; never rewrite posted financial data.
ALTER TABLE accounting.journal_lines
    ADD CONSTRAINT fk_journal_line_branch_entity
        FOREIGN KEY (branch_id, legal_entity_id)
        REFERENCES organization.branches (id, legal_entity_id),
    ADD CONSTRAINT fk_journal_line_department_entity
        FOREIGN KEY (department_id, legal_entity_id)
        REFERENCES organization.departments (id, legal_entity_id),
    ADD CONSTRAINT fk_journal_line_cost_center_entity
        FOREIGN KEY (cost_center_id, legal_entity_id)
        REFERENCES organization.cost_centers (id, legal_entity_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounting.journal_lines
    DROP CONSTRAINT fk_journal_line_cost_center_entity,
    DROP CONSTRAINT fk_journal_line_department_entity,
    DROP CONSTRAINT fk_journal_line_branch_entity;
ALTER TABLE organization.cost_centers DROP CONSTRAINT uq_cost_centers_id_entity;
ALTER TABLE organization.departments DROP CONSTRAINT uq_departments_id_entity;
-- +goose StatementEnd
