-- TASK-025 / P1-BE-40..42 — additive unique/O2O constraints
-- Задача: .planning/backlog/TASK-025-ent-schema-constraints.md
--
-- Hub в dev использует Ent Schema.Create; этот файл — канон для ручного/Atlas
-- применения на существующих Postgres-инстансах.
--
-- OWNER APPROVAL REQUIRED before prod apply:
-- 1) Проверить дубликаты (запросы ниже). При наличии — dedupe вручную, затем CREATE UNIQUE INDEX.
-- 2) Не применять DROP старого unique, пока новые partial indexes не созданы и не проверены.
-- 3) Schema.Create на чистой БД создаёт индексы сам; этот SQL — только для уже живых БД.

-- ---------------------------------------------------------------------------
-- Pre-check duplicates (должны вернуть 0 строк перед CREATE UNIQUE INDEX)
-- ---------------------------------------------------------------------------
-- CoffeeEnrollment duplicates:
--   SELECT user_coffee_enrollment, COUNT(*) FROM coffee_enrollments
--   GROUP BY 1 HAVING COUNT(*) > 1;
--
-- NotificationPreference global duplicates (NULL project_id):
--   SELECT user_id, notification_type, channel, COUNT(*)
--   FROM notification_preferences WHERE project_id IS NULL
--   GROUP BY 1,2,3 HAVING COUNT(*) > 1;
--
-- NotificationPreference per-project duplicates:
--   SELECT user_id, notification_type, channel, project_id, COUNT(*)
--   FROM notification_preferences WHERE project_id IS NOT NULL
--   GROUP BY 1,2,3,4 HAVING COUNT(*) > 1;
--
-- Invoice→JournalEntry O2O duplicates:
--   SELECT journal_entry_invoice, COUNT(*) FROM invoices
--   WHERE journal_entry_invoice IS NOT NULL
--   GROUP BY 1 HAVING COUNT(*) > 1;

-- ---------------------------------------------------------------------------
-- P1-BE-40: CoffeeEnrollment — unique user (additive)
-- ---------------------------------------------------------------------------
CREATE UNIQUE INDEX IF NOT EXISTS coffeeenrollment_user_coffee_enrollment
	ON coffee_enrollments (user_coffee_enrollment);

-- ---------------------------------------------------------------------------
-- P1-BE-41: NotificationPreference — partial unique (NULL-safe)
-- Replace legacy UNIQUE (user_id, notification_type, channel, project_id) which
-- allows multiple NULL project_id rows in PostgreSQL.
-- ---------------------------------------------------------------------------
CREATE UNIQUE INDEX IF NOT EXISTS notificationpreference_user_type_channel_global
	ON notification_preferences (user_id, notification_type, channel)
	WHERE project_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS notificationpreference_user_type_channel_project
	ON notification_preferences (user_id, notification_type, channel, project_id)
	WHERE project_id IS NOT NULL;

-- Drop legacy composite unique only AFTER partial indexes exist and app regenerated.
-- Name from Ent migrate schema: notificationpreference_user_id_notification_type_channel_project_id
-- DROP INDEX IF EXISTS notificationpreference_user_id_notification_type_channel_project_id;

-- ---------------------------------------------------------------------------
-- P1-BE-42: JournalEntry ↔ Invoice O2O — unique FK on invoices
-- ---------------------------------------------------------------------------
CREATE UNIQUE INDEX IF NOT EXISTS invoice_journal_entry_invoice
	ON invoices (journal_entry_invoice)
	WHERE journal_entry_invoice IS NOT NULL;
