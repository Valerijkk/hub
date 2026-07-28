-- TASK-097 / ADR-ENT-SOFT-DELETE-GAPS-2026 — soft-delete columns
-- Задача: .planning/backlog/TASK-097-2026-07-28-ent-soft-delete-schema-gaps.md
--
-- Hub в dev использует Ent Schema.Create; этот файл — канон для ручного/Atlas
-- применения на существующих Postgres-инстансах.
--
-- OWNER APPROVAL: Accepted 2026-07-28 (owner lock 254 A2 — апрув не блокирует).

ALTER TABLE leave_requests ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS leaverequest_deleted_at ON leave_requests (deleted_at);

ALTER TABLE labels ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE labels ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS label_deleted_at ON labels (deleted_at);
CREATE INDEX IF NOT EXISTS label_project_labels_deleted_at ON labels (project_labels, deleted_at);

ALTER TABLE label_groups ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE label_groups ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS labelgroup_deleted_at ON label_groups (deleted_at);
CREATE INDEX IF NOT EXISTS labelgroup_project_label_groups_deleted_at
	ON label_groups (project_label_groups, deleted_at);

ALTER TABLE custom_field_definitions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE custom_field_definitions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS customfielddefinition_deleted_at ON custom_field_definitions (deleted_at);
CREATE INDEX IF NOT EXISTS customfielddefinition_project_custom_fields_deleted_at
	ON custom_field_definitions (project_custom_fields, deleted_at);

ALTER TABLE sprints ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS sprint_deleted_at ON sprints (deleted_at);
CREATE INDEX IF NOT EXISTS sprint_project_sprints_deleted_at ON sprints (project_sprints, deleted_at);

ALTER TABLE recordings ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS recording_deleted_at ON recordings (deleted_at);
CREATE INDEX IF NOT EXISTS recording_project_id_deleted_at ON recordings (project_id, deleted_at);

ALTER TABLE kudos ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS kudo_deleted_at ON kudos (deleted_at);

ALTER TABLE git_hub_integrations ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS githubintegration_deleted_at ON git_hub_integrations (deleted_at);
CREATE INDEX IF NOT EXISTS githubintegration_project_github_integrations_deleted_at
	ON git_hub_integrations (project_github_integrations, deleted_at);

ALTER TABLE git_lab_integrations ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS gitlabintegration_deleted_at ON git_lab_integrations (deleted_at);
CREATE INDEX IF NOT EXISTS gitlabintegration_project_gitlab_integrations_deleted_at
	ON git_lab_integrations (project_gitlab_integrations, deleted_at);

-- JournalLine: без deleted_at — join-filter через journal_entries.deleted_at (ADR §B).
