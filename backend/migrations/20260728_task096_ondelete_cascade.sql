-- TASK-096 / ADR-ENT-ONDELETE-CASCADE-2026 — explicit OnDelete + IssueLink unique
-- Задача: .planning/backlog/TASK-096-2026-07-28-ent-ondelete-cascade-adr.md
--
-- Hub в dev/test: Ent Schema.Create подхватывает annotations из codegen.
-- Этот файл — канон для уже живых Postgres-инстансов (additive).
--
-- Marker: needs-owner-apply (owner 2026-07-28: отдельный approve не требуется;
-- применять вместе с фичей в test/dev).

-- ---------------------------------------------------------------------------
-- A2W8-3-003: IssueLink unique (source, target, link_type); drop parent/child
-- ---------------------------------------------------------------------------
-- Pre-check: SELECT link_type, COUNT(*) FROM issue_links
--   WHERE link_type IN ('parent','child') GROUP BY 1;

DELETE FROM issue_links WHERE link_type IN ('parent', 'child');

DROP INDEX IF EXISTS issuelink_source_id_target_id;
DROP INDEX IF EXISTS issuelink_issue_outgoing_links_issue_link_target;
CREATE UNIQUE INDEX IF NOT EXISTS issuelink_link_type_issue_outgoing_links_issue_link_target
	ON issue_links (link_type, issue_outgoing_links, issue_link_target);

-- Enum CHECK: после ent generate на чистой БД parent/child отсутствуют.
-- На живой PG слагать CONSTRAINT вручную по фактическому имени CHECK.

-- ---------------------------------------------------------------------------
-- OnDelete Cascade / SetNull / NoAction
-- Имена FK — из ent/migrate/schema.go Symbol. При расхождении:
--   SELECT conname FROM pg_constraint WHERE contype = 'f' AND conrelid = '…'::regclass;
-- ---------------------------------------------------------------------------

-- Channel owned children (A2W8-1-003)
-- Пример паттерна (повторить для Symbol из migrate/schema.go):
-- ALTER TABLE channel_members DROP CONSTRAINT channel_members_channels_members;
-- ALTER TABLE channel_members ADD CONSTRAINT channel_members_channels_members
--   FOREIGN KEY (channel_members) REFERENCES channels(id) ON DELETE CASCADE;

-- Полный перечень Symbol после `go generate ./ent` смотреть в
-- backend/ent/migrate/schema.go (OnDelete: schema.Cascade|SetNull|NoAction).
-- Применять ALTER только если текущий OnDelete отличается от схемы.

-- Org (A2W8-4-002): users/projects остаются ON DELETE NO ACTION.
-- Project owned children: ON DELETE CASCADE (PermanentDelete hard-purge).
-- Poll options/votes (A2W8-4-006): ON DELETE CASCADE.
