-- TASK-039 / P2-BE-08 — soft-delete columns for TMS + API gateway
-- Задача: .planning/backlog/TASK-039-be-p2-soft-delete-secrets.md
--
-- Hub в dev использует Ent Schema.Create; этот файл — канон для ручного/Atlas
-- применения на существующих Postgres-инстансах.
--
-- OWNER APPROVAL REQUIRED before prod apply.

ALTER TABLE test_suites ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS testsuite_deleted_at ON test_suites (deleted_at);
CREATE INDEX IF NOT EXISTS testsuite_project_test_suites_deleted_at
	ON test_suites (project_test_suites, deleted_at);

ALTER TABLE components ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS component_deleted_at ON components (deleted_at);
CREATE INDEX IF NOT EXISTS component_project_components_deleted_at
	ON components (project_components, deleted_at);

ALTER TABLE environments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS environment_deleted_at ON environments (deleted_at);
CREATE INDEX IF NOT EXISTS environment_project_environments_deleted_at
	ON environments (project_environments, deleted_at);

ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS apikey_deleted_at ON api_keys (deleted_at);

ALTER TABLE webhook_endpoints ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS webhookendpoint_deleted_at ON webhook_endpoints (deleted_at);
