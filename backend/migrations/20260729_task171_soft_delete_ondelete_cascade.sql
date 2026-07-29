-- TASK-171 / A3W8-5 — Team/WorkflowStatus soft-delete + owned_child OnDelete Cascade
-- Задача: .planning/backlog/TASK-171-2026-07-29-a3-8-5-schemas.md
--
-- Hub в dev/test: Ent Schema.Create подхватывает annotations из codegen.
-- Этот файл — канон для уже живых Postgres-инстансов (additive).
--
-- Marker: needs-owner-apply
-- Approval record: schema+тесты в ветке; SQL не применять в shared/prod без
-- явного owner approve в TASK-171 Approval record.

-- ---------------------------------------------------------------------------
-- A3W8-5-001: Team soft-delete columns
-- ---------------------------------------------------------------------------
ALTER TABLE teams ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE teams ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
CREATE INDEX IF NOT EXISTS team_deleted_at ON teams (deleted_at);

-- ---------------------------------------------------------------------------
-- A3W8-5-002: WorkflowStatus soft-delete columns
-- ---------------------------------------------------------------------------
ALTER TABLE workflow_status ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE workflow_status ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
CREATE INDEX IF NOT EXISTS workflowstatus_deleted_at ON workflow_status (deleted_at);
CREATE INDEX IF NOT EXISTS workflowstatus_deleted_at_project_workflow_statuses
	ON workflow_status (deleted_at, project_workflow_statuses);

-- ---------------------------------------------------------------------------
-- A3W8-5-003…007: OnDelete CASCADE на owned To-edges
-- Имена FK Symbol / колонок — из backend/ent/migrate/schema.go после go generate.
-- При расхождении:
--   SELECT conname FROM pg_constraint WHERE contype = 'f' AND conrelid = '…'::regclass;
-- Применять ALTER только если текущий OnDelete отличается от Cascade.
-- ---------------------------------------------------------------------------

-- A3W8-5-003: Team → members
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS team_members_teams_members;
ALTER TABLE team_members ADD CONSTRAINT team_members_teams_members
	FOREIGN KEY (team_members) REFERENCES teams(id) ON DELETE CASCADE;

-- A3W8-5-004: TestCase → versions
ALTER TABLE test_case_versions DROP CONSTRAINT IF EXISTS test_case_versions_test_cases_versions;
ALTER TABLE test_case_versions ADD CONSTRAINT test_case_versions_test_cases_versions
	FOREIGN KEY (test_case_versions) REFERENCES test_cases(id) ON DELETE CASCADE;

-- A3W8-5-005: TestRun → results
ALTER TABLE test_results DROP CONSTRAINT IF EXISTS test_results_test_runs_results;
ALTER TABLE test_results ADD CONSTRAINT test_results_test_runs_results
	FOREIGN KEY (test_run_results) REFERENCES test_runs(id) ON DELETE CASCADE;

-- A3W8-5-006: WebhookEndpoint → deliveries
ALTER TABLE webhook_deliveries DROP CONSTRAINT IF EXISTS webhook_deliveries_webhook_endpoints_deliveries;
ALTER TABLE webhook_deliveries ADD CONSTRAINT webhook_deliveries_webhook_endpoints_deliveries
	FOREIGN KEY (webhook_endpoint_deliveries) REFERENCES webhook_endpoints(id) ON DELETE CASCADE;

-- A3W8-5-007: WorkflowStatus → from_transitions + to_status
ALTER TABLE workflow_transitions DROP CONSTRAINT IF EXISTS workflow_transitions_workflow_status_from_transitions;
ALTER TABLE workflow_transitions ADD CONSTRAINT workflow_transitions_workflow_status_from_transitions
	FOREIGN KEY (workflow_status_from_transitions) REFERENCES workflow_status(id) ON DELETE CASCADE;

ALTER TABLE workflow_transitions DROP CONSTRAINT IF EXISTS workflow_transitions_workflow_status_to_status;
ALTER TABLE workflow_transitions ADD CONSTRAINT workflow_transitions_workflow_status_to_status
	FOREIGN KEY (workflow_transition_to_status) REFERENCES workflow_status(id) ON DELETE CASCADE;
