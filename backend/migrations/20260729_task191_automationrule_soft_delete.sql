-- TASK-191 / EA-001 — AutomationRule soft-delete (deleted_at)
-- Задача: .planning/backlog/TASK-191-2026-07-29-audit-ent-automationrule-softdelete.md
--
-- Hub в dev/test: Ent Schema.Create подхватывает колонку из codegen.
-- Этот файл — канон для уже живых Postgres-инстансов (additive).
--
-- Marker: needs-owner-apply
-- Approval record: schema+тесты в ветке; SQL не применять в shared/prod без
-- явного owner approve в TASK-191 Approval record.
-- BLOCKED waiting owner: migrate apply.

-- ---------------------------------------------------------------------------
-- EA-001: AutomationRule soft-delete column + list indexes
-- ---------------------------------------------------------------------------
ALTER TABLE automation_rules ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS automationrule_deleted_at ON automation_rules (deleted_at);
CREATE INDEX IF NOT EXISTS automationrule_project_id_deleted_at ON automation_rules (project_id, deleted_at);
