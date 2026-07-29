-- TASK-169 / A3W8-3 — HubTable→columns/records + JournalEntry→lines OnDelete Cascade
-- Задача: .planning/backlog/TASK-169-2026-07-29-a3-8-3-schemas.md
--
-- Hub в dev/test: Ent Schema.Create подхватывает annotations из codegen.
-- Этот файл — канон для уже живых Postgres-инстансов (additive).
--
-- Marker: needs-owner-apply
-- Approval record: schema+тесты в ветке; SQL не применять в shared/prod без
-- явного owner approve в TASK-169 Approval record.

-- ---------------------------------------------------------------------------
-- A3W8-3-001: HubTable → columns / records
-- Symbol / колонки — из backend/ent/migrate/schema.go после go generate.
-- ---------------------------------------------------------------------------
ALTER TABLE hub_columns DROP CONSTRAINT IF EXISTS hub_columns_hub_tables_columns;
ALTER TABLE hub_columns ADD CONSTRAINT hub_columns_hub_tables_columns
	FOREIGN KEY (hub_table_columns) REFERENCES hub_tables(id) ON DELETE CASCADE;

ALTER TABLE hub_records DROP CONSTRAINT IF EXISTS hub_records_hub_tables_records;
ALTER TABLE hub_records ADD CONSTRAINT hub_records_hub_tables_records
	FOREIGN KEY (hub_table_records) REFERENCES hub_tables(id) ON DELETE CASCADE;

-- ---------------------------------------------------------------------------
-- A3W8-3-002: JournalEntry → lines
-- ---------------------------------------------------------------------------
ALTER TABLE journal_lines DROP CONSTRAINT IF EXISTS journal_lines_journal_entries_lines;
ALTER TABLE journal_lines ADD CONSTRAINT journal_lines_journal_entries_lines
	FOREIGN KEY (journal_entry_lines) REFERENCES journal_entries(id) ON DELETE CASCADE;
