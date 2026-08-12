-- TASK-119 / W7-A4-018 — restored_at на backup_records
-- Задача: .planning/backlog/TASK-119-2026-07-28-storage-backup-dashboard-residual.md
--
-- Hub в dev использует Ent Schema.Create; этот файл — канон для ручного/Atlas
-- применения на существующих Postgres-инстансах.
--
-- Additive only: nullable column, без backfill.

ALTER TABLE backup_records
    ADD COLUMN IF NOT EXISTS restored_at TIMESTAMPTZ NULL;
