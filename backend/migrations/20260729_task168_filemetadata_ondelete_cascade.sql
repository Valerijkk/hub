-- TASK-168 / A3W8-2 — FileMetadata validators + owned_child OnDelete Cascade
-- Задача: .planning/backlog/TASK-168-2026-07-29-a3-8-2-schemas.md
--
-- Hub в dev/test: Ent Schema.Create подхватывает annotations из codegen.
-- Этот файл — канон для уже живых Postgres-инстансов (additive).
--
-- Marker: needs-owner-apply
-- Approval record: schema+тесты в ветке; SQL не применять в shared/prod без
-- явного owner approve в TASK-168 Approval record.
-- ADR-ENT-ONDELETE-CASCADE-2026: Cascade на owned_child OK без отдельного ADR.

-- ---------------------------------------------------------------------------
-- A3W8-2-001: FileMetadata.uploader Required (NOT NULL)
-- Pre-check: SELECT COUNT(*) FROM file_metadata WHERE uploader_id IS NULL;
-- Если >0 — backfill/удалить orphans ДО NOT NULL.
-- ---------------------------------------------------------------------------
ALTER TABLE file_metadata ALTER COLUMN uploader_id SET NOT NULL;

-- FK: Required edge → NoAction (нельзя SET NULL на required).
ALTER TABLE file_metadata DROP CONSTRAINT IF EXISTS file_metadata_users_files;
ALTER TABLE file_metadata ADD CONSTRAINT file_metadata_users_files
	FOREIGN KEY (uploader_id) REFERENCES users(id) ON DELETE NO ACTION;

-- ---------------------------------------------------------------------------
-- A3W8-2-002…005: OnDelete CASCADE на owned To-edges
-- Имена FK Symbol / колонок — из backend/ent/migrate/schema.go после go generate.
-- При расхождении:
--   SELECT conname FROM pg_constraint WHERE contype = 'f' AND conrelid = '…'::regclass;
-- Применять ALTER только если текущий OnDelete отличается от Cascade.
-- ---------------------------------------------------------------------------

-- A3W8-2-002: CoffeeRound → pairs
ALTER TABLE coffee_pairs DROP CONSTRAINT IF EXISTS coffee_pairs_coffee_rounds_pairs;
ALTER TABLE coffee_pairs ADD CONSTRAINT coffee_pairs_coffee_rounds_pairs
	FOREIGN KEY (coffee_round_pairs) REFERENCES coffee_rounds(id) ON DELETE CASCADE;

-- A3W8-2-003: HubBase → tables (HubTable→columns/records — TASK-169)
ALTER TABLE hub_tables DROP CONSTRAINT IF EXISTS hub_tables_hub_bases_tables;
ALTER TABLE hub_tables ADD CONSTRAINT hub_tables_hub_bases_tables
	FOREIGN KEY (hub_base_tables) REFERENCES hub_bases(id) ON DELETE CASCADE;

-- A3W8-2-004: GitHubIntegration → github_links
ALTER TABLE git_hub_links DROP CONSTRAINT IF EXISTS git_hub_links_git_hub_integrations_github_links;
ALTER TABLE git_hub_links ADD CONSTRAINT git_hub_links_git_hub_integrations_github_links
	FOREIGN KEY (git_hub_integration_github_links) REFERENCES git_hub_integrations(id) ON DELETE CASCADE;

-- A3W8-2-004: GitLabIntegration → gitlab_links
ALTER TABLE git_lab_links DROP CONSTRAINT IF EXISTS git_lab_links_git_lab_integrations_gitlab_links;
ALTER TABLE git_lab_links ADD CONSTRAINT git_lab_links_git_lab_integrations_gitlab_links
	FOREIGN KEY (git_lab_integration_gitlab_links) REFERENCES git_lab_integrations(id) ON DELETE CASCADE;

-- A3W8-2-005: Comment → reactions
ALTER TABLE reactions DROP CONSTRAINT IF EXISTS reactions_comments_reactions;
ALTER TABLE reactions ADD CONSTRAINT reactions_comments_reactions
	FOREIGN KEY (comment_reactions) REFERENCES comments(id) ON DELETE CASCADE;

-- A3W8-2-006: access_token NotEmpty — Ent runtime validator; DDL CHECK не требуется
-- (симметрия webhook_secret NotEmpty без CHECK в live PG).
