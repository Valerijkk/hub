-- TASK-170 / A3W8-4 — sprint partial unique (active) + issues OnDelete SET NULL
-- Задача: .planning/backlog/TASK-170-2026-07-29-a3-8-4-schemas.md
--
-- Hub в dev/test: Ent Schema.Create подхватывает annotations из codegen.
-- Этот файл — канон для уже живых Postgres-инстансов (additive).
--
-- Marker: needs-owner-apply
-- Approval record: Owner lock TASK-170 (2026-07-29) — schema+SQL+тесты в ветке;
-- перед apply на shared/prod: pre-migrate cleanup duplicate active sprints.

-- ---------------------------------------------------------------------------
-- Pre-migrate check (обязательно до CREATE UNIQUE INDEX):
--   SELECT project_sprints, COUNT(*) AS n
--   FROM sprints
--   WHERE status = 'active' AND deleted_at IS NULL
--   GROUP BY project_sprints
--   HAVING COUNT(*) > 1;
-- При n>1 — cleanup/plan (оставить один active, остальные → planning/completed
-- или soft-delete), иначе CREATE UNIQUE INDEX упадёт.
-- ---------------------------------------------------------------------------

-- ---------------------------------------------------------------------------
-- A3W8-4-001: partial unique — один active sprint на project (без soft-deleted)
-- ---------------------------------------------------------------------------
-- Имя индекса = Ent Symbol (migrate/schema.go: sprint_project_sprints).
CREATE UNIQUE INDEX IF NOT EXISTS sprint_project_sprints
	ON sprints (project_sprints)
	WHERE status = 'active' AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- A3W8-4-002: Sprint → issues optional_ref ON DELETE SET NULL
-- Symbol / колонка — из backend/ent/migrate/schema.go (issues_sprints_issues /
-- sprint_issues). При расхождении:
--   SELECT conname FROM pg_constraint
--   WHERE contype = 'f' AND conrelid = 'issues'::regclass;
-- Применять ALTER только если текущий OnDelete ≠ SetNull.
-- ---------------------------------------------------------------------------
ALTER TABLE issues DROP CONSTRAINT IF EXISTS issues_sprints_issues;
ALTER TABLE issues ADD CONSTRAINT issues_sprints_issues
	FOREIGN KEY (sprint_issues) REFERENCES sprints(id) ON DELETE SET NULL;
