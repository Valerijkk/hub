-- TASK-107 / A2W8 indexes·FK·polymorphic
-- Задача: .planning/backlog/TASK-107-2026-07-28-ent-indexes-composite-fk-polymorphic.md
-- Finding IDs: A2W8-1-004/005, A2W8-2-005/007/008/009, A2W8-3-002/006
--
-- Hub в dev использует Ent Schema.Create; этот файл — канон для ручного/Atlas
-- применения на существующих Postgres-инстансах.
-- Owner (2026-07-28): migrate OK без отдельного ADR-approve.

BEGIN;

-- ---------------------------------------------------------------------------
-- A2W8-1-004: Board — updated_at, deleted_at, indexes
-- ---------------------------------------------------------------------------
ALTER TABLE boards ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE boards ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS board_deleted_at ON boards (deleted_at);
CREATE INDEX IF NOT EXISTS board_deleted_at_project_boards ON boards (deleted_at, project_boards);

-- ---------------------------------------------------------------------------
-- A2W8-1-005: AIUserPreference.default_provider_id → FK AIProvider
-- ---------------------------------------------------------------------------
-- Pre-check orphans:
--   SELECT default_provider_id FROM ai_user_preferences
--   WHERE default_provider_id IS NOT NULL
--     AND default_provider_id NOT IN (SELECT id FROM ai_providers);
ALTER TABLE ai_user_preferences
	DROP CONSTRAINT IF EXISTS ai_user_preferences_ai_providers_default_provider;
ALTER TABLE ai_user_preferences
	ADD CONSTRAINT ai_user_preferences_ai_providers_default_provider
	FOREIGN KEY (default_provider_id) REFERENCES ai_providers(id)
	ON DELETE SET NULL;

-- ---------------------------------------------------------------------------
-- A2W8-2-005: CrmContact email index; CrmDeal owner+deleted_at
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS crmcontact_email ON crm_contacts (email);
CREATE INDEX IF NOT EXISTS crmdeal_deleted_at_user_crm_deals_owned
	ON crm_deals (deleted_at, user_crm_deals_owned);

-- ---------------------------------------------------------------------------
-- A2W8-2-007: GitLab webhook_secret NOT NULL
-- ---------------------------------------------------------------------------
-- Pre-check / backfill empty secrets before NOT NULL:
--   UPDATE git_lab_integrations SET webhook_secret = 'NEEDS-ROTATE'
--   WHERE webhook_secret IS NULL OR webhook_secret = '';
UPDATE git_lab_integrations
	SET webhook_secret = 'NEEDS-ROTATE-' || id::text
	WHERE webhook_secret IS NULL OR btrim(webhook_secret) = '';
ALTER TABLE git_lab_integrations
	ALTER COLUMN webhook_secret SET NOT NULL;

-- ---------------------------------------------------------------------------
-- A2W8-2-008: GitHub/GitLab links — version + unique (integration, external, type)
-- ---------------------------------------------------------------------------
ALTER TABLE git_hub_links ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE git_lab_links ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

DROP INDEX IF EXISTS githublink_external_id_link_type;
CREATE UNIQUE INDEX IF NOT EXISTS githublink_external_id_link_type_git_hub_integration_github_links
	ON git_hub_links (external_id, link_type, git_hub_integration_github_links);

DROP INDEX IF EXISTS gitlablink_external_id_link_type;
CREATE UNIQUE INDEX IF NOT EXISTS gitlablink_external_id_link_type_git_lab_integration_gitlab_links
	ON git_lab_links (external_id, link_type, git_lab_integration_gitlab_links);

-- ---------------------------------------------------------------------------
-- A2W8-2-009: HubBase/HubColumn/HubRecord composite indexes
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS hubbase_deleted_at_user_hub_bases
	ON hub_bases (deleted_at, user_hub_bases);
CREATE UNIQUE INDEX IF NOT EXISTS hubcolumn_name_hub_table_columns
	ON hub_columns (name, hub_table_columns);
CREATE INDEX IF NOT EXISTS hubcolumn_deleted_at_hub_table_columns
	ON hub_columns (deleted_at, hub_table_columns);
CREATE INDEX IF NOT EXISTS hubrecord_deleted_at_hub_table_records
	ON hub_records (deleted_at, hub_table_records);

-- ---------------------------------------------------------------------------
-- A2W8-3-002: Invite version + revoked_at (soft revoke)
-- ---------------------------------------------------------------------------
ALTER TABLE invites ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE invites ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE invites ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS invite_revoked_at ON invites (revoked_at);

-- ---------------------------------------------------------------------------
-- A2W8-3-006: Notification.entity_id String→UUID + version
-- ---------------------------------------------------------------------------
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE notifications ADD COLUMN IF NOT EXISTS entity_id_uuid UUID NULL;
UPDATE notifications
	SET entity_id_uuid = entity_id::uuid
	WHERE entity_id IS NOT NULL
	  AND entity_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
	  AND entity_id_uuid IS NULL;
-- Drop legacy string column + rename (idempotent-ish for re-runs).
DO $$
BEGIN
	IF EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name = 'notifications' AND column_name = 'entity_id'
		  AND data_type IN ('text', 'character varying')
	) THEN
		ALTER TABLE notifications DROP COLUMN entity_id;
		ALTER TABLE notifications RENAME COLUMN entity_id_uuid TO entity_id;
	ELSIF EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name = 'notifications' AND column_name = 'entity_id_uuid'
	) THEN
		ALTER TABLE notifications DROP COLUMN IF EXISTS entity_id_uuid;
	END IF;
END $$;

COMMIT;
