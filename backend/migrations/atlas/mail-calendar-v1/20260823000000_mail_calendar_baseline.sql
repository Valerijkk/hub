-- Create "activity_logs" table
CREATE TABLE "activity_logs" ("id" uuid NOT NULL, "action" character varying NOT NULL, "changes" jsonb NULL, "created_at" timestamptz NOT NULL, "issue_activities" uuid NOT NULL, "user_activity_logs" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "activitylog_created_at_issue_activities" to table: "activity_logs"
CREATE INDEX "activitylog_created_at_issue_activities" ON "activity_logs" ("created_at", "issue_activities");
-- Create "ai_providers" table
CREATE TABLE "ai_providers" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "provider_type" character varying NOT NULL, "base_url" character varying NOT NULL, "api_key_enc" text NULL, "models" jsonb NOT NULL, "enabled" boolean NOT NULL DEFAULT true, "priority" bigint NOT NULL DEFAULT 0, "max_rpm" bigint NOT NULL DEFAULT 60, "ru_resident" boolean NOT NULL DEFAULT false, "scope" character varying NOT NULL DEFAULT 'system', "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "user_ai_providers" uuid NULL, PRIMARY KEY ("id"));
-- Create index "aiprovider_enabled_priority" to table: "ai_providers"
CREATE INDEX "aiprovider_enabled_priority" ON "ai_providers" ("enabled", "priority");
-- Create index "aiprovider_scope" to table: "ai_providers"
CREATE INDEX "aiprovider_scope" ON "ai_providers" ("scope");
-- Create "ai_usage_dailies" table
CREATE TABLE "ai_usage_dailies" ("id" uuid NOT NULL, "usage_date" timestamptz NOT NULL, "tokens_in" bigint NOT NULL DEFAULT 0, "tokens_out" bigint NOT NULL DEFAULT 0, "request_count" bigint NOT NULL DEFAULT 0, "user_ai_usage_daily" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "aiusagedaily_usage_date_user_ai_usage_daily" to table: "ai_usage_dailies"
CREATE UNIQUE INDEX "aiusagedaily_usage_date_user_ai_usage_daily" ON "ai_usage_dailies" ("usage_date", "user_ai_usage_daily");
-- Create "ai_user_preferences" table
CREATE TABLE "ai_user_preferences" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "default_model" character varying NULL, "allow_external_ai" boolean NOT NULL DEFAULT false, "updated_at" timestamptz NOT NULL, "default_provider_id" uuid NULL, "user_ai_preference" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "aiuserpreference_user_ai_preference" to table: "ai_user_preferences"
CREATE UNIQUE INDEX "aiuserpreference_user_ai_preference" ON "ai_user_preferences" ("user_ai_preference");
-- Create "api_keys" table
CREATE TABLE "api_keys" ("id" uuid NOT NULL, "name" character varying NOT NULL, "key_hash" character varying NOT NULL, "prefix" character varying NOT NULL, "last_used_at" timestamptz NULL, "expires_at" timestamptz NULL, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "user_api_keys" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "api_keys_key_hash_key" to table: "api_keys"
CREATE UNIQUE INDEX "api_keys_key_hash_key" ON "api_keys" ("key_hash");
-- Create index "apikey_deleted_at" to table: "api_keys"
CREATE INDEX "apikey_deleted_at" ON "api_keys" ("deleted_at");
-- Create index "apikey_expires_at" to table: "api_keys"
CREATE INDEX "apikey_expires_at" ON "api_keys" ("expires_at");
-- Create "audit_logs" table
CREATE TABLE "audit_logs" ("id" uuid NOT NULL, "action" character varying NOT NULL, "entity_type" character varying NOT NULL, "entity_id" character varying NULL, "user_id" uuid NULL, "project_id" uuid NULL, "details" jsonb NULL, "ip_address" character varying NULL, "user_agent" character varying NULL, "created_at" timestamptz NOT NULL, PRIMARY KEY ("id"));
-- Create index "auditlog_action" to table: "audit_logs"
CREATE INDEX "auditlog_action" ON "audit_logs" ("action");
-- Create index "auditlog_created_at" to table: "audit_logs"
CREATE INDEX "auditlog_created_at" ON "audit_logs" ("created_at");
-- Create index "auditlog_entity_type" to table: "audit_logs"
CREATE INDEX "auditlog_entity_type" ON "audit_logs" ("entity_type");
-- Create index "auditlog_project_id" to table: "audit_logs"
CREATE INDEX "auditlog_project_id" ON "audit_logs" ("project_id");
-- Create index "auditlog_user_id" to table: "audit_logs"
CREATE INDEX "auditlog_user_id" ON "audit_logs" ("user_id");
-- Create "automation_rules" table
CREATE TABLE "automation_rules" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "enabled" boolean NOT NULL DEFAULT true, "trigger_type" character varying NOT NULL, "trigger_config" jsonb NOT NULL, "action_type" character varying NOT NULL, "action_config" jsonb NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "project_id" uuid NOT NULL, "created_by_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "automationrule_deleted_at" to table: "automation_rules"
CREATE INDEX "automationrule_deleted_at" ON "automation_rules" ("deleted_at");
-- Create index "automationrule_project_id_deleted_at" to table: "automation_rules"
CREATE INDEX "automationrule_project_id_deleted_at" ON "automation_rules" ("project_id", "deleted_at");
-- Create index "automationrule_project_id_enabled" to table: "automation_rules"
CREATE INDEX "automationrule_project_id_enabled" ON "automation_rules" ("project_id", "enabled");
-- Create index "automationrule_project_id_trigger_type" to table: "automation_rules"
CREATE INDEX "automationrule_project_id_trigger_type" ON "automation_rules" ("project_id", "trigger_type");
-- Create "backup_records" table
CREATE TABLE "backup_records" ("id" uuid NOT NULL, "s3_key" character varying NOT NULL, "file_size" bigint NOT NULL DEFAULT 0, "status" character varying NOT NULL DEFAULT 'in_progress', "error_message" text NULL, "restored_at" timestamptz NULL, "created_at" timestamptz NOT NULL, PRIMARY KEY ("id"));
-- Create index "backuprecord_created_at" to table: "backup_records"
CREATE INDEX "backuprecord_created_at" ON "backup_records" ("created_at");
-- Create index "backuprecord_status" to table: "backup_records"
CREATE INDEX "backuprecord_status" ON "backup_records" ("status");
-- Create "boards" table
CREATE TABLE "boards" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "filters" jsonb NULL, "is_default" boolean NOT NULL DEFAULT false, "swimlane_mode" character varying NOT NULL DEFAULT 'none', "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "project_boards" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "board_deleted_at" to table: "boards"
CREATE INDEX "board_deleted_at" ON "boards" ("deleted_at");
-- Create index "board_deleted_at_project_boards" to table: "boards"
CREATE INDEX "board_deleted_at_project_boards" ON "boards" ("deleted_at", "project_boards");
-- Create "bookmarks" table
CREATE TABLE "bookmarks" ("id" uuid NOT NULL, "message_id" character varying NOT NULL, "created_at" timestamptz NOT NULL, "channel_id" uuid NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "bookmark_channel_id" to table: "bookmarks"
CREATE INDEX "bookmark_channel_id" ON "bookmarks" ("channel_id");
-- Create index "bookmark_user_id_message_id" to table: "bookmarks"
CREATE UNIQUE INDEX "bookmark_user_id_message_id" ON "bookmarks" ("user_id", "message_id");
-- Create "call_participants" table
CREATE TABLE "call_participants" ("id" uuid NOT NULL, "joined_at" timestamptz NOT NULL, "left_at" timestamptz NULL, "call_id" uuid NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "callparticipant_call_id_user_id" to table: "call_participants"
CREATE UNIQUE INDEX "callparticipant_call_id_user_id" ON "call_participants" ("call_id", "user_id");
-- Create "calls" table
CREATE TABLE "calls" ("id" uuid NOT NULL, "room_name" character varying NOT NULL, "invite_code" character varying NULL, "title" character varying NULL, "status" character varying NOT NULL DEFAULT 'active', "started_at" timestamptz NULL, "ended_at" timestamptz NULL, "participant_count" bigint NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL, "channel_id" uuid NULL, "created_by" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "call_channel_id_status" to table: "calls"
CREATE INDEX "call_channel_id_status" ON "calls" ("channel_id", "status");
-- Create index "calls_invite_code_key" to table: "calls"
CREATE UNIQUE INDEX "calls_invite_code_key" ON "calls" ("invite_code");
-- Create index "calls_room_name_key" to table: "calls"
CREATE UNIQUE INDEX "calls_room_name_key" ON "calls" ("room_name");
-- Create "channel_invites" table
CREATE TABLE "channel_invites" ("id" uuid NOT NULL, "token" character varying NOT NULL, "max_uses" bigint NOT NULL DEFAULT 0, "use_count" bigint NOT NULL DEFAULT 0, "expires_at" timestamptz NULL, "is_revoked" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "channel_id" uuid NOT NULL, "creator_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "channel_invites_token_key" to table: "channel_invites"
CREATE UNIQUE INDEX "channel_invites_token_key" ON "channel_invites" ("token");
-- Create index "channelinvite_channel_id" to table: "channel_invites"
CREATE INDEX "channelinvite_channel_id" ON "channel_invites" ("channel_id");
-- Create index "channelinvite_expires_at" to table: "channel_invites"
CREATE INDEX "channelinvite_expires_at" ON "channel_invites" ("expires_at");
-- Create "channel_members" table
CREATE TABLE "channel_members" ("id" uuid NOT NULL, "role" character varying NOT NULL DEFAULT 'member', "joined_at" timestamptz NOT NULL, "is_starred" boolean NOT NULL DEFAULT false, "muted_until" timestamptz NULL, "notify_mode" character varying NOT NULL DEFAULT 'all', "channel_id" uuid NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "channelmember_channel_id_user_id" to table: "channel_members"
CREATE UNIQUE INDEX "channelmember_channel_id_user_id" ON "channel_members" ("channel_id", "user_id");
-- Create "channels" table
CREATE TABLE "channels" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "channel_type" character varying NOT NULL, "name" character varying NULL, "description" character varying NULL, "is_readonly" boolean NOT NULL DEFAULT false, "dm_key" character varying NULL, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "project_id" uuid NULL, "created_by" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "channel_channel_type" to table: "channels"
CREATE INDEX "channel_channel_type" ON "channels" ("channel_type");
-- Create index "channel_deleted_at" to table: "channels"
CREATE INDEX "channel_deleted_at" ON "channels" ("deleted_at");
-- Create index "channel_project_id" to table: "channels"
CREATE INDEX "channel_project_id" ON "channels" ("project_id");
-- Create index "channels_dm_key_key" to table: "channels"
CREATE UNIQUE INDEX "channels_dm_key_key" ON "channels" ("dm_key");
-- Create "coffee_enrollments" table
CREATE TABLE "coffee_enrollments" ("id" uuid NOT NULL, "is_active" boolean NOT NULL DEFAULT true, "enrolled_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "user_coffee_enrollment" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "coffee_enrollments_user_coffee_enrollment_key" to table: "coffee_enrollments"
CREATE UNIQUE INDEX "coffee_enrollments_user_coffee_enrollment_key" ON "coffee_enrollments" ("user_coffee_enrollment");
-- Create index "coffeeenrollment_is_active" to table: "coffee_enrollments"
CREATE INDEX "coffeeenrollment_is_active" ON "coffee_enrollments" ("is_active");
-- Create index "coffeeenrollment_user_coffee_enrollment" to table: "coffee_enrollments"
CREATE UNIQUE INDEX "coffeeenrollment_user_coffee_enrollment" ON "coffee_enrollments" ("user_coffee_enrollment");
-- Create "coffee_pairs" table
CREATE TABLE "coffee_pairs" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "coffee_round_pairs" uuid NOT NULL, "user_coffee_pairs_a" uuid NOT NULL, "user_coffee_pairs_b" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "coffeepair_coffee_round_pairs__e763bcb5485b6d80e2ed8ba9a78b8519" to table: "coffee_pairs"
CREATE UNIQUE INDEX "coffeepair_coffee_round_pairs__e763bcb5485b6d80e2ed8ba9a78b8519" ON "coffee_pairs" ("coffee_round_pairs", "user_coffee_pairs_a", "user_coffee_pairs_b");
-- Create "coffee_rounds" table
CREATE TABLE "coffee_rounds" ("id" uuid NOT NULL, "pair_count" bigint NOT NULL DEFAULT 0, "unpaired_count" bigint NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL, "user_coffee_rounds_triggered" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "coffeeround_created_at" to table: "coffee_rounds"
CREATE INDEX "coffeeround_created_at" ON "coffee_rounds" ("created_at");
-- Create "comments" table
CREATE TABLE "comments" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "body_json" jsonb NOT NULL, "body_text" text NOT NULL, "is_internal" boolean NOT NULL DEFAULT false, "mentions" jsonb NULL, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "issue_comments" uuid NOT NULL, "user_comments" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "comment_created_at_issue_comments" to table: "comments"
CREATE INDEX "comment_created_at_issue_comments" ON "comments" ("created_at", "issue_comments");
-- Create index "comment_deleted_at" to table: "comments"
CREATE INDEX "comment_deleted_at" ON "comments" ("deleted_at");
-- Create "components" table
CREATE TABLE "components" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "description" text NULL, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "project_components" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "component_deleted_at" to table: "components"
CREATE INDEX "component_deleted_at" ON "components" ("deleted_at");
-- Create index "component_deleted_at_project_components" to table: "components"
CREATE INDEX "component_deleted_at_project_components" ON "components" ("deleted_at", "project_components");
-- Create index "component_name_project_components" to table: "components"
CREATE UNIQUE INDEX "component_name_project_components" ON "components" ("name", "project_components");
-- Create "crm_contacts" table
CREATE TABLE "crm_contacts" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "email" character varying NULL, "phone" character varying NULL, "company" character varying NULL, "notes" text NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "user_crm_contacts_created" uuid NOT NULL, "user_crm_contact_link" uuid NULL, PRIMARY KEY ("id"));
-- Create index "crmcontact_company" to table: "crm_contacts"
CREATE INDEX "crmcontact_company" ON "crm_contacts" ("company");
-- Create index "crmcontact_deleted_at" to table: "crm_contacts"
CREATE INDEX "crmcontact_deleted_at" ON "crm_contacts" ("deleted_at");
-- Create index "crmcontact_email" to table: "crm_contacts"
CREATE INDEX "crmcontact_email" ON "crm_contacts" ("email");
-- Create "crm_deals" table
CREATE TABLE "crm_deals" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "title" character varying NOT NULL, "description" text NULL, "stage" character varying NOT NULL DEFAULT 'lead', "amount_cents" bigint NULL, "currency" character varying NOT NULL DEFAULT 'RUB', "rank" character varying NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "crm_contact_deals" uuid NULL, "project_crm_deals" uuid NULL, "user_crm_deals_owned" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "crmdeal_deleted_at" to table: "crm_deals"
CREATE INDEX "crmdeal_deleted_at" ON "crm_deals" ("deleted_at");
-- Create index "crmdeal_deleted_at_user_crm_deals_owned" to table: "crm_deals"
CREATE INDEX "crmdeal_deleted_at_user_crm_deals_owned" ON "crm_deals" ("deleted_at", "user_crm_deals_owned");
-- Create index "crmdeal_stage_rank" to table: "crm_deals"
CREATE INDEX "crmdeal_stage_rank" ON "crm_deals" ("stage", "rank");
-- Create "custom_field_definitions" table
CREATE TABLE "custom_field_definitions" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "field_type" character varying NOT NULL, "options" jsonb NULL, "required" boolean NOT NULL DEFAULT false, "sort_order" bigint NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "project_custom_fields" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "customfielddefinition_deleted_at" to table: "custom_field_definitions"
CREATE INDEX "customfielddefinition_deleted_at" ON "custom_field_definitions" ("deleted_at");
-- Create index "customfielddefinition_deleted_at_project_custom_fields" to table: "custom_field_definitions"
CREATE INDEX "customfielddefinition_deleted_at_project_custom_fields" ON "custom_field_definitions" ("deleted_at", "project_custom_fields");
-- Create index "customfielddefinition_name_project_custom_fields" to table: "custom_field_definitions"
CREATE UNIQUE INDEX "customfielddefinition_name_project_custom_fields" ON "custom_field_definitions" ("name", "project_custom_fields");
-- Create "dashboard_layouts" table
CREATE TABLE "dashboard_layouts" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "dashboard_type" character varying NOT NULL, "project_id" uuid NULL, "items" jsonb NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "dashboardlayout_user_id_dashboard_type_project_id" to table: "dashboard_layouts"
CREATE UNIQUE INDEX "dashboardlayout_user_id_dashboard_type_project_id" ON "dashboard_layouts" ("user_id", "dashboard_type", "project_id");
-- Create "environments" table
CREATE TABLE "environments" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "env_label" character varying NOT NULL DEFAULT 'custom', "config" jsonb NULL, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "project_environments" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "environment_deleted_at" to table: "environments"
CREATE INDEX "environment_deleted_at" ON "environments" ("deleted_at");
-- Create index "environment_deleted_at_project_environments" to table: "environments"
CREATE INDEX "environment_deleted_at_project_environments" ON "environments" ("deleted_at", "project_environments");
-- Create index "environment_env_label" to table: "environments"
CREATE INDEX "environment_env_label" ON "environments" ("env_label");
-- Create index "environment_name_project_environments" to table: "environments"
CREATE UNIQUE INDEX "environment_name_project_environments" ON "environments" ("name", "project_environments");
-- Create "file_metadata" table
CREATE TABLE "file_metadata" ("id" uuid NOT NULL, "filename" character varying NOT NULL, "content_type" character varying NOT NULL, "size" bigint NOT NULL, "s3_key" character varying NOT NULL, "s3_bucket" character varying NOT NULL, "entity_type" character varying NULL, "entity_id" character varying NULL, "status" character varying NOT NULL DEFAULT 'pending', "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "issue_files" uuid NULL, "project_files" uuid NULL, "test_result_files" uuid NULL, "uploader_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "filemetadata_deleted_at" to table: "file_metadata"
CREATE INDEX "filemetadata_deleted_at" ON "file_metadata" ("deleted_at");
-- Create index "filemetadata_entity_type_entity_id" to table: "file_metadata"
CREATE INDEX "filemetadata_entity_type_entity_id" ON "file_metadata" ("entity_type", "entity_id");
-- Create index "filemetadata_s3_key" to table: "file_metadata"
CREATE INDEX "filemetadata_s3_key" ON "file_metadata" ("s3_key");
-- Create index "filemetadata_status" to table: "file_metadata"
CREATE INDEX "filemetadata_status" ON "file_metadata" ("status");
-- Create "git_hub_integrations" table
CREATE TABLE "git_hub_integrations" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "repo_owner" character varying NOT NULL, "repo_name" character varying NOT NULL, "access_token" character varying NOT NULL, "webhook_secret" character varying NOT NULL, "merge_status_id" uuid NULL, "created_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "project_github_integrations" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "githubintegration_deleted_at" to table: "git_hub_integrations"
CREATE INDEX "githubintegration_deleted_at" ON "git_hub_integrations" ("deleted_at");
-- Create index "githubintegration_deleted_at_project_github_integrations" to table: "git_hub_integrations"
CREATE INDEX "githubintegration_deleted_at_project_github_integrations" ON "git_hub_integrations" ("deleted_at", "project_github_integrations");
-- Create index "githubintegration_repo_owner_r_132321746ff39875c65e9d204687c55a" to table: "git_hub_integrations"
CREATE UNIQUE INDEX "githubintegration_repo_owner_r_132321746ff39875c65e9d204687c55a" ON "git_hub_integrations" ("repo_owner", "repo_name", "project_github_integrations");
-- Create "git_hub_links" table
CREATE TABLE "git_hub_links" ("id" uuid NOT NULL, "link_type" character varying NOT NULL, "external_id" character varying NOT NULL, "title" character varying NOT NULL, "url" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'none', "author" character varying NULL, "branch" character varying NULL, "version" bigint NOT NULL DEFAULT 1, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "git_hub_integration_github_links" uuid NOT NULL, "issue_github_links" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "githublink_external_id_link_ty_c35f20f66fa15ac15fd97ff9bb22bec3" to table: "git_hub_links"
CREATE UNIQUE INDEX "githublink_external_id_link_ty_c35f20f66fa15ac15fd97ff9bb22bec3" ON "git_hub_links" ("external_id", "link_type", "git_hub_integration_github_links");
-- Create index "githublink_issue_github_links" to table: "git_hub_links"
CREATE INDEX "githublink_issue_github_links" ON "git_hub_links" ("issue_github_links");
-- Create "git_lab_integrations" table
CREATE TABLE "git_lab_integrations" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "gitlab_url" character varying NOT NULL DEFAULT 'https://gitlab.com', "project_path" character varying NOT NULL, "access_token" character varying NOT NULL, "webhook_secret" character varying NOT NULL, "created_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "project_gitlab_integrations" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "gitlabintegration_deleted_at" to table: "git_lab_integrations"
CREATE INDEX "gitlabintegration_deleted_at" ON "git_lab_integrations" ("deleted_at");
-- Create index "gitlabintegration_deleted_at_project_gitlab_integrations" to table: "git_lab_integrations"
CREATE INDEX "gitlabintegration_deleted_at_project_gitlab_integrations" ON "git_lab_integrations" ("deleted_at", "project_gitlab_integrations");
-- Create index "gitlabintegration_project_path_project_gitlab_integrations" to table: "git_lab_integrations"
CREATE UNIQUE INDEX "gitlabintegration_project_path_project_gitlab_integrations" ON "git_lab_integrations" ("project_path", "project_gitlab_integrations");
-- Create "git_lab_links" table
CREATE TABLE "git_lab_links" ("id" uuid NOT NULL, "link_type" character varying NOT NULL, "external_id" character varying NOT NULL, "title" character varying NOT NULL, "url" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'none', "author" character varying NULL, "source_branch" character varying NULL, "project_path" character varying NULL, "version" bigint NOT NULL DEFAULT 1, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "git_lab_integration_gitlab_links" uuid NOT NULL, "issue_gitlab_links" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "gitlablink_external_id_link_ty_17ce11025e9cdece2923e053c5116b8b" to table: "git_lab_links"
CREATE UNIQUE INDEX "gitlablink_external_id_link_ty_17ce11025e9cdece2923e053c5116b8b" ON "git_lab_links" ("external_id", "link_type", "git_lab_integration_gitlab_links");
-- Create index "gitlablink_issue_gitlab_links" to table: "git_lab_links"
CREATE INDEX "gitlablink_issue_gitlab_links" ON "git_lab_links" ("issue_gitlab_links");
-- Create "hub_bases" table
CREATE TABLE "hub_bases" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "user_hub_bases" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "hubbase_deleted_at" to table: "hub_bases"
CREATE INDEX "hubbase_deleted_at" ON "hub_bases" ("deleted_at");
-- Create index "hubbase_deleted_at_user_hub_bases" to table: "hub_bases"
CREATE INDEX "hubbase_deleted_at_user_hub_bases" ON "hub_bases" ("deleted_at", "user_hub_bases");
-- Create "hub_columns" table
CREATE TABLE "hub_columns" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "col_type" character varying NOT NULL DEFAULT 'text', "options" jsonb NULL, "sort_order" bigint NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "hub_table_columns" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "hubcolumn_deleted_at" to table: "hub_columns"
CREATE INDEX "hubcolumn_deleted_at" ON "hub_columns" ("deleted_at");
-- Create index "hubcolumn_deleted_at_hub_table_columns" to table: "hub_columns"
CREATE INDEX "hubcolumn_deleted_at_hub_table_columns" ON "hub_columns" ("deleted_at", "hub_table_columns");
-- Create index "hubcolumn_name_hub_table_columns" to table: "hub_columns"
CREATE UNIQUE INDEX "hubcolumn_name_hub_table_columns" ON "hub_columns" ("name", "hub_table_columns");
-- Create index "hubcolumn_sort_order" to table: "hub_columns"
CREATE INDEX "hubcolumn_sort_order" ON "hub_columns" ("sort_order");
-- Create "hub_records" table
CREATE TABLE "hub_records" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "cells" jsonb NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "hub_table_records" uuid NOT NULL, "user_hub_records" uuid NULL, PRIMARY KEY ("id"));
-- Create index "hubrecord_deleted_at" to table: "hub_records"
CREATE INDEX "hubrecord_deleted_at" ON "hub_records" ("deleted_at");
-- Create index "hubrecord_deleted_at_hub_table_records" to table: "hub_records"
CREATE INDEX "hubrecord_deleted_at_hub_table_records" ON "hub_records" ("deleted_at", "hub_table_records");
-- Create "hub_tables" table
CREATE TABLE "hub_tables" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "sort_order" bigint NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "hub_base_tables" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "hubtable_deleted_at" to table: "hub_tables"
CREATE INDEX "hubtable_deleted_at" ON "hub_tables" ("deleted_at");
-- Create index "hubtable_sort_order" to table: "hub_tables"
CREATE INDEX "hubtable_sort_order" ON "hub_tables" ("sort_order");
-- Create "invites" table
CREATE TABLE "invites" ("id" uuid NOT NULL, "code" character varying NOT NULL, "max_uses" bigint NOT NULL DEFAULT 0, "use_count" bigint NOT NULL DEFAULT 0, "project_role" character varying NOT NULL DEFAULT 'developer', "expires_at" timestamptz NULL, "version" bigint NOT NULL DEFAULT 1, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "revoked_at" timestamptz NULL, "project_invites" uuid NOT NULL, "user_invites" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "invite_expires_at" to table: "invites"
CREATE INDEX "invite_expires_at" ON "invites" ("expires_at");
-- Create index "invite_revoked_at" to table: "invites"
CREATE INDEX "invite_revoked_at" ON "invites" ("revoked_at");
-- Create index "invites_code_key" to table: "invites"
CREATE UNIQUE INDEX "invites_code_key" ON "invites" ("code");
-- Create "invoices" table
CREATE TABLE "invoices" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "number" character varying NOT NULL, "customer_name" character varying NOT NULL, "description" text NULL, "amount_cents" bigint NOT NULL, "currency" character varying NOT NULL DEFAULT 'RUB', "status" character varying NOT NULL DEFAULT 'draft', "due_date" timestamptz NULL, "issued_at" timestamptz NULL, "paid_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "crm_deal_invoices" uuid NULL, "journal_entry_invoice" uuid NULL, "user_invoices_created" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "invoice_deleted_at" to table: "invoices"
CREATE INDEX "invoice_deleted_at" ON "invoices" ("deleted_at");
-- Create index "invoice_journal_entry_invoice" to table: "invoices"
CREATE UNIQUE INDEX "invoice_journal_entry_invoice" ON "invoices" ("journal_entry_invoice");
-- Create index "invoice_number" to table: "invoices"
CREATE UNIQUE INDEX "invoice_number" ON "invoices" ("number");
-- Create index "invoice_status" to table: "invoices"
CREATE INDEX "invoice_status" ON "invoices" ("status");
-- Create index "invoices_journal_entry_invoice_key" to table: "invoices"
CREATE UNIQUE INDEX "invoices_journal_entry_invoice_key" ON "invoices" ("journal_entry_invoice");
-- Create "issue_labels" table
CREATE TABLE "issue_labels" ("issue_id" uuid NOT NULL, "label_id" uuid NOT NULL, PRIMARY KEY ("issue_id", "label_id"));
-- Create "issue_links" table
CREATE TABLE "issue_links" ("id" uuid NOT NULL, "link_type" character varying NOT NULL, "custom_type" character varying NULL, "created_at" timestamptz NOT NULL, "issue_outgoing_links" uuid NOT NULL, "issue_link_target" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "issuelink_link_type_issue_outgoing_links_issue_link_target" to table: "issue_links"
CREATE UNIQUE INDEX "issuelink_link_type_issue_outgoing_links_issue_link_target" ON "issue_links" ("link_type", "issue_outgoing_links", "issue_link_target");
-- Create "issue_message_links" table
CREATE TABLE "issue_message_links" ("id" uuid NOT NULL, "message_id" character varying NOT NULL, "mirror_to_thread" boolean NOT NULL DEFAULT true, "created_at" timestamptz NOT NULL, "channel_id" uuid NOT NULL, "issue_id" uuid NOT NULL, "created_by_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "issuemessagelink_channel_id_message_id" to table: "issue_message_links"
CREATE INDEX "issuemessagelink_channel_id_message_id" ON "issue_message_links" ("channel_id", "message_id");
-- Create index "issuemessagelink_issue_id_channel_id_message_id" to table: "issue_message_links"
CREATE UNIQUE INDEX "issuemessagelink_issue_id_channel_id_message_id" ON "issue_message_links" ("issue_id", "channel_id", "message_id");
-- Create "issues" table
CREATE TABLE "issues" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "issue_number" bigint NOT NULL, "issue_key" character varying NOT NULL, "title" character varying NOT NULL, "description_json" jsonb NULL, "description_text" text NULL, "issue_type" character varying NOT NULL, "priority" character varying NOT NULL DEFAULT 'medium', "rank" character varying NULL, "story_points" bigint NULL, "estimate_minutes" bigint NULL, "time_spent_minutes" bigint NOT NULL DEFAULT 0, "checklist" jsonb NULL, "custom_fields" jsonb NULL, "due_date" timestamptz NULL, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "issue_assignee" uuid NULL, "issue_reporter" uuid NULL, "issue_children" uuid NULL, "project_issues" uuid NOT NULL, "sprint_issues" uuid NULL, "workflow_status_issues" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "issues_issues_children" FOREIGN KEY ("issue_children") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "issue_created_at" to table: "issues"
CREATE INDEX "issue_created_at" ON "issues" ("created_at");
-- Create index "issue_deleted_at" to table: "issues"
CREATE INDEX "issue_deleted_at" ON "issues" ("deleted_at");
-- Create index "issue_deleted_at_project_issues" to table: "issues"
CREATE INDEX "issue_deleted_at_project_issues" ON "issues" ("deleted_at", "project_issues");
-- Create index "issue_due_date" to table: "issues"
CREATE INDEX "issue_due_date" ON "issues" ("due_date");
-- Create index "issue_issue_key" to table: "issues"
CREATE INDEX "issue_issue_key" ON "issues" ("issue_key");
-- Create index "issue_issue_number" to table: "issues"
CREATE INDEX "issue_issue_number" ON "issues" ("issue_number");
-- Create index "issue_priority" to table: "issues"
CREATE INDEX "issue_priority" ON "issues" ("priority");
-- Create index "issue_rank" to table: "issues"
CREATE INDEX "issue_rank" ON "issues" ("rank");
-- Create index "issues_issue_key_key" to table: "issues"
CREATE UNIQUE INDEX "issues_issue_key_key" ON "issues" ("issue_key");
-- Create "journal_entries" table
CREATE TABLE "journal_entries" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "entry_date" timestamptz NOT NULL, "description" character varying NOT NULL, "reference" character varying NULL, "status" character varying NOT NULL DEFAULT 'draft', "posted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "user_journal_entries" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "journalentry_deleted_at" to table: "journal_entries"
CREATE INDEX "journalentry_deleted_at" ON "journal_entries" ("deleted_at");
-- Create index "journalentry_entry_date" to table: "journal_entries"
CREATE INDEX "journalentry_entry_date" ON "journal_entries" ("entry_date");
-- Create index "journalentry_status" to table: "journal_entries"
CREATE INDEX "journalentry_status" ON "journal_entries" ("status");
-- Create "journal_lines" table
CREATE TABLE "journal_lines" ("id" uuid NOT NULL, "debit_cents" bigint NOT NULL DEFAULT 0, "credit_cents" bigint NOT NULL DEFAULT 0, "memo" character varying NULL, "created_at" timestamptz NOT NULL, "journal_entry_lines" uuid NOT NULL, "ledger_account_journal_lines" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "journalline_debit_xor_credit" CHECK (((debit_cents > 0) = (credit_cents = 0)) AND ((debit_cents + credit_cents) > 0)));
-- Create index "journalline_debit_cents_credit_cents" to table: "journal_lines"
CREATE INDEX "journalline_debit_cents_credit_cents" ON "journal_lines" ("debit_cents", "credit_cents");
-- Create "kudos" table
CREATE TABLE "kudos" ("id" uuid NOT NULL, "message" character varying NOT NULL, "is_public" boolean NOT NULL DEFAULT true, "created_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "user_kudos_sent" uuid NOT NULL, "user_kudos_received" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "kudo_created_at" to table: "kudos"
CREATE INDEX "kudo_created_at" ON "kudos" ("created_at");
-- Create index "kudo_deleted_at" to table: "kudos"
CREATE INDEX "kudo_deleted_at" ON "kudos" ("deleted_at");
-- Create index "kudo_is_public_created_at" to table: "kudos"
CREATE INDEX "kudo_is_public_created_at" ON "kudos" ("is_public", "created_at");
-- Create "label_groups" table
CREATE TABLE "label_groups" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "project_label_groups" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "labelgroup_deleted_at" to table: "label_groups"
CREATE INDEX "labelgroup_deleted_at" ON "label_groups" ("deleted_at");
-- Create index "labelgroup_deleted_at_project_label_groups" to table: "label_groups"
CREATE INDEX "labelgroup_deleted_at_project_label_groups" ON "label_groups" ("deleted_at", "project_label_groups");
-- Create index "labelgroup_name_project_label_groups" to table: "label_groups"
CREATE UNIQUE INDEX "labelgroup_name_project_label_groups" ON "label_groups" ("name", "project_label_groups");
-- Create "labels" table
CREATE TABLE "labels" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "color" character varying NOT NULL DEFAULT '#3B82F6', "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "label_group_labels" uuid NULL, "project_labels" uuid NOT NULL, "test_case_labels" uuid NULL, PRIMARY KEY ("id"));
-- Create index "label_deleted_at" to table: "labels"
CREATE INDEX "label_deleted_at" ON "labels" ("deleted_at");
-- Create index "label_deleted_at_project_labels" to table: "labels"
CREATE INDEX "label_deleted_at_project_labels" ON "labels" ("deleted_at", "project_labels");
-- Create index "label_name_project_labels" to table: "labels"
CREATE UNIQUE INDEX "label_name_project_labels" ON "labels" ("name", "project_labels");
-- Create "leave_requests" table
CREATE TABLE "leave_requests" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "leave_type" character varying NOT NULL DEFAULT 'vacation', "status" character varying NOT NULL DEFAULT 'pending', "start_date" timestamptz NOT NULL, "end_date" timestamptz NOT NULL, "reason" text NULL, "review_note" text NULL, "reviewed_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "user_leave_requests" uuid NOT NULL, "user_leave_reviews" uuid NULL, PRIMARY KEY ("id"));
-- Create index "leaverequest_deleted_at" to table: "leave_requests"
CREATE INDEX "leaverequest_deleted_at" ON "leave_requests" ("deleted_at");
-- Create index "leaverequest_start_date_end_date" to table: "leave_requests"
CREATE INDEX "leaverequest_start_date_end_date" ON "leave_requests" ("start_date", "end_date");
-- Create index "leaverequest_status_start_date" to table: "leave_requests"
CREATE INDEX "leaverequest_status_start_date" ON "leave_requests" ("status", "start_date");
-- Create "ledger_accounts" table
CREATE TABLE "ledger_accounts" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "code" character varying NOT NULL, "name" character varying NOT NULL, "account_type" character varying NOT NULL, "is_active" boolean NOT NULL DEFAULT true, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "ledgeraccount_code" to table: "ledger_accounts"
CREATE UNIQUE INDEX "ledgeraccount_code" ON "ledger_accounts" ("code");
-- Create index "ledgeraccount_deleted_at" to table: "ledger_accounts"
CREATE INDEX "ledgeraccount_deleted_at" ON "ledger_accounts" ("deleted_at");
-- Create "marketplace_listings" table
CREATE TABLE "marketplace_listings" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "title" character varying NOT NULL, "description" character varying NOT NULL, "category" character varying NOT NULL DEFAULT 'other', "status" character varying NOT NULL DEFAULT 'pending', "price_cents" bigint NULL, "currency" character varying NOT NULL DEFAULT 'RUB', "review_note" text NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "reviewed_at" timestamptz NULL, "deleted_at" timestamptz NULL, "user_marketplace_listings" uuid NOT NULL, "user_marketplace_reviews" uuid NULL, PRIMARY KEY ("id"));
-- Create index "marketplacelisting_deleted_at" to table: "marketplace_listings"
CREATE INDEX "marketplacelisting_deleted_at" ON "marketplace_listings" ("deleted_at");
-- Create index "marketplacelisting_status_category_created_at" to table: "marketplace_listings"
CREATE INDEX "marketplacelisting_status_category_created_at" ON "marketplace_listings" ("status", "category", "created_at");
-- Create "news_poll_options" table
CREATE TABLE "news_poll_options" ("id" uuid NOT NULL, "text" character varying NOT NULL, "position" bigint NOT NULL DEFAULT 0, "votes_count" bigint NOT NULL DEFAULT 0, "poll_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "newspolloption_poll_id_position" to table: "news_poll_options"
CREATE INDEX "newspolloption_poll_id_position" ON "news_poll_options" ("poll_id", "position");
-- Create "news_poll_votes" table
CREATE TABLE "news_poll_votes" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "poll_id" uuid NOT NULL, "option_id" uuid NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "newspollvote_option_id" to table: "news_poll_votes"
CREATE INDEX "newspollvote_option_id" ON "news_poll_votes" ("option_id");
-- Create index "newspollvote_poll_id_user_id" to table: "news_poll_votes"
CREATE UNIQUE INDEX "newspollvote_poll_id_user_id" ON "news_poll_votes" ("poll_id", "user_id");
-- Create "news_polls" table
CREATE TABLE "news_polls" ("id" uuid NOT NULL, "question" character varying NOT NULL, "closed" boolean NOT NULL DEFAULT false, "version" bigint NOT NULL DEFAULT 1, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "news_post_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "news_polls_news_post_id_key" to table: "news_polls"
CREATE UNIQUE INDEX "news_polls_news_post_id_key" ON "news_polls" ("news_post_id");
-- Create index "newspoll_news_post_id" to table: "news_polls"
CREATE UNIQUE INDEX "newspoll_news_post_id" ON "news_polls" ("news_post_id");
-- Create "news_posts" table
CREATE TABLE "news_posts" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "title" character varying NOT NULL, "body" character varying NOT NULL, "pinned" boolean NOT NULL DEFAULT false, "published" boolean NOT NULL DEFAULT true, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "user_news_posts" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "newspost_deleted_at" to table: "news_posts"
CREATE INDEX "newspost_deleted_at" ON "news_posts" ("deleted_at");
-- Create index "newspost_published_pinned_created_at" to table: "news_posts"
CREATE INDEX "newspost_published_pinned_created_at" ON "news_posts" ("published", "pinned", "created_at");
-- Create "notification_preferences" table
CREATE TABLE "notification_preferences" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "notification_type" character varying NOT NULL, "channel" character varying NOT NULL, "enabled" boolean NOT NULL DEFAULT true, "project_id" uuid NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "notificationpreference_user_id_11558c9b5f1ac61cdf810945d78fac2a" to table: "notification_preferences"
CREATE UNIQUE INDEX "notificationpreference_user_id_11558c9b5f1ac61cdf810945d78fac2a" ON "notification_preferences" ("user_id", "notification_type", "channel", "project_id") WHERE (project_id IS NOT NULL);
-- Create index "notificationpreference_user_id_notification_type_channel" to table: "notification_preferences"
CREATE UNIQUE INDEX "notificationpreference_user_id_notification_type_channel" ON "notification_preferences" ("user_id", "notification_type", "channel") WHERE (project_id IS NULL);
-- Create "notifications" table
CREATE TABLE "notifications" ("id" uuid NOT NULL, "type" character varying NOT NULL, "title" character varying NOT NULL, "body" character varying NULL, "entity_type" character varying NULL, "entity_id" uuid NULL, "project_id" uuid NULL, "actor_id" uuid NULL, "is_read" boolean NOT NULL DEFAULT false, "is_urgent" boolean NOT NULL DEFAULT false, "version" bigint NOT NULL DEFAULT 1, "created_at" timestamptz NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "notification_user_id_created_at" to table: "notifications"
CREATE INDEX "notification_user_id_created_at" ON "notifications" ("user_id", "created_at");
-- Create index "notification_user_id_is_read" to table: "notifications"
CREATE INDEX "notification_user_id_is_read" ON "notifications" ("user_id", "is_read");
-- Create index "notification_user_id_type_entity_id_created_at" to table: "notifications"
CREATE INDEX "notification_user_id_type_entity_id_created_at" ON "notifications" ("user_id", "type", "entity_id", "created_at");
-- Create "organizations" table
CREATE TABLE "organizations" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "slug" character varying NOT NULL, "invite_token" character varying NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, PRIMARY KEY ("id"));
-- Create index "organization_invite_token" to table: "organizations"
CREATE INDEX "organization_invite_token" ON "organizations" ("invite_token");
-- Create index "organization_slug" to table: "organizations"
CREATE INDEX "organization_slug" ON "organizations" ("slug");
-- Create index "organizations_invite_token_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_invite_token_key" ON "organizations" ("invite_token");
-- Create index "organizations_slug_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_slug_key" ON "organizations" ("slug");
-- Create "outbox_events" table
CREATE TABLE "outbox_events" ("id" uuid NOT NULL, "event_type" character varying NOT NULL, "payload" bytea NOT NULL, "trace_context" jsonb NULL, "status" character varying NOT NULL DEFAULT 'pending', "attempts" bigint NOT NULL DEFAULT 0, "last_error" character varying NULL, "created_at" timestamptz NOT NULL, "published_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "outboxevent_status_created_at" to table: "outbox_events"
CREATE INDEX "outboxevent_status_created_at" ON "outbox_events" ("status", "created_at");
-- Create "pinned_messages" table
CREATE TABLE "pinned_messages" ("id" uuid NOT NULL, "message_id" character varying NOT NULL, "pinned_by" uuid NOT NULL, "pinned_at" timestamptz NOT NULL, "channel_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "pinnedmessage_channel_id_message_id" to table: "pinned_messages"
CREATE UNIQUE INDEX "pinnedmessage_channel_id_message_id" ON "pinned_messages" ("channel_id", "message_id");
-- Create "poll_options" table
CREATE TABLE "poll_options" ("id" uuid NOT NULL, "text" character varying NOT NULL, "position" bigint NOT NULL DEFAULT 0, "poll_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "polloption_poll_id_position" to table: "poll_options"
CREATE INDEX "polloption_poll_id_position" ON "poll_options" ("poll_id", "position");
-- Create "poll_votes" table
CREATE TABLE "poll_votes" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "poll_id" uuid NOT NULL, "option_id" uuid NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "pollvote_poll_id_option_id_user_id" to table: "poll_votes"
CREATE UNIQUE INDEX "pollvote_poll_id_option_id_user_id" ON "poll_votes" ("poll_id", "option_id", "user_id");
-- Create index "pollvote_poll_id_user_id" to table: "poll_votes"
CREATE INDEX "pollvote_poll_id_user_id" ON "poll_votes" ("poll_id", "user_id");
-- Create "polls" table
CREATE TABLE "polls" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "question" character varying NOT NULL, "is_anonymous" boolean NOT NULL DEFAULT false, "is_multi_choice" boolean NOT NULL DEFAULT false, "is_closed" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "channel_id" uuid NOT NULL, "creator_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "poll_channel_id" to table: "polls"
CREATE INDEX "poll_channel_id" ON "polls" ("channel_id");
-- Create index "poll_creator_id" to table: "polls"
CREATE INDEX "poll_creator_id" ON "polls" ("creator_id");
-- Create "project_members" table
CREATE TABLE "project_members" ("id" uuid NOT NULL, "project_role" character varying NOT NULL DEFAULT 'developer', "joined_at" timestamptz NOT NULL, "project_members" uuid NOT NULL, "role_members" uuid NULL, "user_project_members" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "projectmember_user_project_members_project_members" to table: "project_members"
CREATE UNIQUE INDEX "projectmember_user_project_members_project_members" ON "project_members" ("user_project_members", "project_members");
-- Create "project_teams" table
CREATE TABLE "project_teams" ("project_id" uuid NOT NULL, "team_id" uuid NOT NULL, PRIMARY KEY ("project_id", "team_id"));
-- Create "project_templates" table
CREATE TABLE "project_templates" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "description" text NULL, "template_type" character varying NOT NULL DEFAULT 'custom', "config" jsonb NOT NULL, "is_system" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "user_project_templates" uuid NULL, PRIMARY KEY ("id"));
-- Create index "projecttemplate_is_system" to table: "project_templates"
CREATE INDEX "projecttemplate_is_system" ON "project_templates" ("is_system");
-- Create index "projecttemplate_template_type" to table: "project_templates"
CREATE INDEX "projecttemplate_template_type" ON "project_templates" ("template_type");
-- Create "projects" table
CREATE TABLE "projects" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "key" character varying NOT NULL, "description" character varying NULL, "avatar_url" character varying NULL, "status" character varying NOT NULL DEFAULT 'active', "task_counter" bigint NOT NULL DEFAULT 0, "test_case_counter" bigint NOT NULL DEFAULT 0, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "organization_projects" uuid NULL, PRIMARY KEY ("id"));
-- Create index "project_deleted_at" to table: "projects"
CREATE INDEX "project_deleted_at" ON "projects" ("deleted_at");
-- Create index "project_status" to table: "projects"
CREATE INDEX "project_status" ON "projects" ("status");
-- Create index "projects_key_key" to table: "projects"
CREATE UNIQUE INDEX "projects_key_key" ON "projects" ("key");
-- Create "push_subscriptions" table
CREATE TABLE "push_subscriptions" ("id" uuid NOT NULL, "endpoint" character varying NOT NULL, "p256dh_key" character varying NOT NULL, "auth_key" character varying NOT NULL, "user_agent" character varying NULL, "created_at" timestamptz NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "pushsubscription_endpoint" to table: "push_subscriptions"
CREATE UNIQUE INDEX "pushsubscription_endpoint" ON "push_subscriptions" ("endpoint");
-- Create index "pushsubscription_user_id" to table: "push_subscriptions"
CREATE INDEX "pushsubscription_user_id" ON "push_subscriptions" ("user_id");
-- Create "reactions" table
CREATE TABLE "reactions" ("id" uuid NOT NULL, "emoji" character varying NOT NULL, "created_at" timestamptz NOT NULL, "comment_reactions" uuid NOT NULL, "user_reactions" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "reaction_emoji_comment_reactions_user_reactions" to table: "reactions"
CREATE UNIQUE INDEX "reaction_emoji_comment_reactions_user_reactions" ON "reactions" ("emoji", "comment_reactions", "user_reactions");
-- Create "recordings" table
CREATE TABLE "recordings" ("id" uuid NOT NULL, "egress_id" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'recording', "file_path" character varying NULL, "file_size" bigint NOT NULL DEFAULT 0, "duration_seconds" bigint NOT NULL DEFAULT 0, "started_at" timestamptz NOT NULL, "ended_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "call_id" uuid NOT NULL, "project_id" uuid NOT NULL, "started_by" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "recording_call_id" to table: "recordings"
CREATE INDEX "recording_call_id" ON "recordings" ("call_id");
-- Create index "recording_deleted_at" to table: "recordings"
CREATE INDEX "recording_deleted_at" ON "recordings" ("deleted_at");
-- Create index "recording_egress_id" to table: "recordings"
CREATE INDEX "recording_egress_id" ON "recordings" ("egress_id");
-- Create index "recording_project_id_created_at" to table: "recordings"
CREATE INDEX "recording_project_id_created_at" ON "recordings" ("project_id", "created_at");
-- Create index "recording_project_id_deleted_at" to table: "recordings"
CREATE INDEX "recording_project_id_deleted_at" ON "recordings" ("project_id", "deleted_at");
-- Create "roles" table
CREATE TABLE "roles" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "description" character varying NULL, "permissions" bigint NOT NULL DEFAULT 0, "permission_names" jsonb NULL, "is_system" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "project_roles" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "role_name_project_roles" to table: "roles"
CREATE UNIQUE INDEX "role_name_project_roles" ON "roles" ("name", "project_roles");
-- Create "saved_filters" table
CREATE TABLE "saved_filters" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "filter_config" jsonb NOT NULL, "shared" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "project_id" uuid NOT NULL, "owner_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "savedfilter_owner_id_project_id" to table: "saved_filters"
CREATE INDEX "savedfilter_owner_id_project_id" ON "saved_filters" ("owner_id", "project_id");
-- Create index "savedfilter_project_id_shared" to table: "saved_filters"
CREATE INDEX "savedfilter_project_id_shared" ON "saved_filters" ("project_id", "shared");
-- Create "scheduled_calls" table
CREATE TABLE "scheduled_calls" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "title" character varying NOT NULL, "description" character varying NULL, "scheduled_at" timestamptz NOT NULL, "duration_minutes" bigint NULL DEFAULT 60, "reminder_minutes" bigint NULL, "rrule" character varying NULL, "is_active" boolean NOT NULL DEFAULT true, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "channel_id" uuid NOT NULL, "project_id" uuid NOT NULL, "created_by" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "scheduledcall_channel_id_scheduled_at" to table: "scheduled_calls"
CREATE INDEX "scheduledcall_channel_id_scheduled_at" ON "scheduled_calls" ("channel_id", "scheduled_at");
-- Create index "scheduledcall_deleted_at" to table: "scheduled_calls"
CREATE INDEX "scheduledcall_deleted_at" ON "scheduled_calls" ("deleted_at");
-- Create index "scheduledcall_is_active_scheduled_at" to table: "scheduled_calls"
CREATE INDEX "scheduledcall_is_active_scheduled_at" ON "scheduled_calls" ("is_active", "scheduled_at");
-- Create "sessions" table
CREATE TABLE "sessions" ("id" uuid NOT NULL, "refresh_token" character varying NOT NULL, "device" character varying NOT NULL, "ip_address" character varying NULL, "last_active_at" timestamptz NOT NULL, "expires_at" timestamptz NOT NULL, "created_at" timestamptz NOT NULL, "user_sessions" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "session_expires_at" to table: "sessions"
CREATE INDEX "session_expires_at" ON "sessions" ("expires_at");
-- Create index "sessions_refresh_token_key" to table: "sessions"
CREATE UNIQUE INDEX "sessions_refresh_token_key" ON "sessions" ("refresh_token");
-- Create "sprints" table
CREATE TABLE "sprints" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "goal" character varying NULL, "status" character varying NOT NULL DEFAULT 'planning', "start_date" timestamptz NOT NULL, "end_date" timestamptz NOT NULL, "retro_notes" jsonb NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "project_sprints" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "sprint_deleted_at" to table: "sprints"
CREATE INDEX "sprint_deleted_at" ON "sprints" ("deleted_at");
-- Create index "sprint_deleted_at_project_sprints" to table: "sprints"
CREATE INDEX "sprint_deleted_at_project_sprints" ON "sprints" ("deleted_at", "project_sprints");
-- Create index "sprint_project_sprints" to table: "sprints"
CREATE UNIQUE INDEX "sprint_project_sprints" ON "sprints" ("project_sprints") WHERE (((status)::text = 'active'::text) AND (deleted_at IS NULL));
-- Create index "sprint_start_date_project_sprints" to table: "sprints"
CREATE INDEX "sprint_start_date_project_sprints" ON "sprints" ("start_date", "project_sprints");
-- Create index "sprint_status_project_sprints" to table: "sprints"
CREATE INDEX "sprint_status_project_sprints" ON "sprints" ("status", "project_sprints");
-- Create "system_settings" table
CREATE TABLE "system_settings" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "key" character varying NOT NULL, "value" character varying NOT NULL, "description" character varying NULL, "updated_at" timestamptz NOT NULL, PRIMARY KEY ("id"));
-- Create index "system_settings_key_key" to table: "system_settings"
CREATE UNIQUE INDEX "system_settings_key_key" ON "system_settings" ("key");
-- Create "team_members" table
CREATE TABLE "team_members" ("id" uuid NOT NULL, "joined_at" timestamptz NOT NULL, "team_members" uuid NOT NULL, "user_team_members" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "teammember_team_members_user_team_members" to table: "team_members"
CREATE UNIQUE INDEX "teammember_team_members_user_team_members" ON "team_members" ("team_members", "user_team_members");
-- Create "teams" table
CREATE TABLE "teams" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "description" character varying NULL, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, PRIMARY KEY ("id"));
-- Create index "team_deleted_at" to table: "teams"
CREATE INDEX "team_deleted_at" ON "teams" ("deleted_at");
-- Create "test_case_components" table
CREATE TABLE "test_case_components" ("test_case_id" uuid NOT NULL, "component_id" uuid NOT NULL, PRIMARY KEY ("test_case_id", "component_id"));
-- Create "test_case_linked_issues" table
CREATE TABLE "test_case_linked_issues" ("test_case_id" uuid NOT NULL, "issue_id" uuid NOT NULL, PRIMARY KEY ("test_case_id", "issue_id"));
-- Create "test_case_versions" table
CREATE TABLE "test_case_versions" ("id" uuid NOT NULL, "version_number" bigint NOT NULL, "snapshot" jsonb NOT NULL, "change_description" text NULL, "created_at" timestamptz NOT NULL, "test_case_versions" uuid NOT NULL, "user_test_case_versions" uuid NULL, PRIMARY KEY ("id"));
-- Create index "testcaseversion_version_number_test_case_versions" to table: "test_case_versions"
CREATE UNIQUE INDEX "testcaseversion_version_number_test_case_versions" ON "test_case_versions" ("version_number", "test_case_versions");
-- Create "test_cases" table
CREATE TABLE "test_cases" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "case_number" bigint NOT NULL, "case_key" character varying NOT NULL, "title" character varying NOT NULL, "description" text NULL, "precondition" text NULL, "postcondition" text NULL, "case_type" character varying NOT NULL DEFAULT 'classic', "priority" character varying NOT NULL DEFAULT 'medium', "status" character varying NOT NULL DEFAULT 'draft', "steps" jsonb NULL, "full_name" character varying NULL, "class_name" character varying NULL, "method_name" character varying NULL, "external_id" character varying NULL, "automation_id" character varying NULL, "sync_status" character varying NOT NULL DEFAULT 'unlinked', "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "project_test_cases" uuid NOT NULL, "test_suite_test_cases" uuid NULL, "user_created_test_cases" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "test_cases_case_key_key" to table: "test_cases"
CREATE UNIQUE INDEX "test_cases_case_key_key" ON "test_cases" ("case_key");
-- Create index "testcase_automation_id" to table: "test_cases"
CREATE INDEX "testcase_automation_id" ON "test_cases" ("automation_id");
-- Create index "testcase_case_key" to table: "test_cases"
CREATE INDEX "testcase_case_key" ON "test_cases" ("case_key");
-- Create index "testcase_class_name_method_name" to table: "test_cases"
CREATE INDEX "testcase_class_name_method_name" ON "test_cases" ("class_name", "method_name");
-- Create index "testcase_deleted_at" to table: "test_cases"
CREATE INDEX "testcase_deleted_at" ON "test_cases" ("deleted_at");
-- Create index "testcase_deleted_at_project_test_cases" to table: "test_cases"
CREATE INDEX "testcase_deleted_at_project_test_cases" ON "test_cases" ("deleted_at", "project_test_cases");
-- Create index "testcase_external_id" to table: "test_cases"
CREATE INDEX "testcase_external_id" ON "test_cases" ("external_id");
-- Create index "testcase_full_name" to table: "test_cases"
CREATE INDEX "testcase_full_name" ON "test_cases" ("full_name");
-- Create index "testcase_priority" to table: "test_cases"
CREATE INDEX "testcase_priority" ON "test_cases" ("priority");
-- Create index "testcase_status" to table: "test_cases"
CREATE INDEX "testcase_status" ON "test_cases" ("status");
-- Create "test_results" table
CREATE TABLE "test_results" ("id" uuid NOT NULL, "status" character varying NOT NULL DEFAULT 'unknown', "comment" text NULL, "step_results" jsonb NULL, "duration_ms" bigint NULL, "allure_uuid" character varying NULL, "error_message" text NULL, "error_trace" text NULL, "is_flaky" boolean NOT NULL DEFAULT false, "match_strategy" character varying NOT NULL DEFAULT 'unmapped', "started_at" timestamptz NULL, "completed_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "test_case_results" uuid NULL, "test_run_results" uuid NOT NULL, "user_executed_results" uuid NULL, PRIMARY KEY ("id"));
-- Create index "testresult_allure_uuid" to table: "test_results"
CREATE INDEX "testresult_allure_uuid" ON "test_results" ("allure_uuid");
-- Create index "testresult_is_flaky" to table: "test_results"
CREATE INDEX "testresult_is_flaky" ON "test_results" ("is_flaky");
-- Create index "testresult_status" to table: "test_results"
CREATE INDEX "testresult_status" ON "test_results" ("status");
-- Create "test_runs" table
CREATE TABLE "test_runs" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "title" character varying NOT NULL, "description" text NULL, "run_type" character varying NOT NULL DEFAULT 'manual', "status" character varying NOT NULL DEFAULT 'created', "total_count" bigint NOT NULL DEFAULT 0, "passed_count" bigint NOT NULL DEFAULT 0, "failed_count" bigint NOT NULL DEFAULT 0, "skipped_count" bigint NOT NULL DEFAULT 0, "broken_count" bigint NOT NULL DEFAULT 0, "unknown_count" bigint NOT NULL DEFAULT 0, "allure_run_id" character varying NULL, "started_at" timestamptz NULL, "completed_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "environment_test_runs" uuid NULL, "project_test_runs" uuid NOT NULL, "user_created_runs" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "testrun_allure_run_id" to table: "test_runs"
CREATE INDEX "testrun_allure_run_id" ON "test_runs" ("allure_run_id");
-- Create index "testrun_created_at" to table: "test_runs"
CREATE INDEX "testrun_created_at" ON "test_runs" ("created_at");
-- Create index "testrun_run_type" to table: "test_runs"
CREATE INDEX "testrun_run_type" ON "test_runs" ("run_type");
-- Create index "testrun_status" to table: "test_runs"
CREATE INDEX "testrun_status" ON "test_runs" ("status");
-- Create "test_suites" table
CREATE TABLE "test_suites" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "description" text NULL, "sort_order" bigint NOT NULL DEFAULT 0, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "project_test_suites" uuid NOT NULL, "test_suite_children" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "test_suites_test_suites_children" FOREIGN KEY ("test_suite_children") REFERENCES "test_suites" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "testsuite_deleted_at" to table: "test_suites"
CREATE INDEX "testsuite_deleted_at" ON "test_suites" ("deleted_at");
-- Create index "testsuite_deleted_at_project_test_suites" to table: "test_suites"
CREATE INDEX "testsuite_deleted_at_project_test_suites" ON "test_suites" ("deleted_at", "project_test_suites");
-- Create index "testsuite_name" to table: "test_suites"
CREATE INDEX "testsuite_name" ON "test_suites" ("name");
-- Create "time_logs" table
CREATE TABLE "time_logs" ("id" uuid NOT NULL, "minutes" bigint NOT NULL, "description" character varying NULL, "started_at" timestamptz NULL, "ended_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "issue_time_logs" uuid NOT NULL, "user_time_logs" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "timelog_issue_time_logs" to table: "time_logs"
CREATE INDEX "timelog_issue_time_logs" ON "time_logs" ("issue_time_logs");
-- Create "user_blocks" table
CREATE TABLE "user_blocks" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "blocker_id" uuid NOT NULL, "blocked_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "userblock_blocker_id_blocked_id" to table: "user_blocks"
CREATE UNIQUE INDEX "userblock_blocker_id_blocked_id" ON "user_blocks" ("blocker_id", "blocked_id");
-- Create "users" table
CREATE TABLE "users" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "email" character varying NOT NULL, "name" character varying NULL, "display_name" character varying NULL, "avatar_url" character varying NULL, "telegram_id" character varying NULL, "phone" character varying NULL, "position" character varying NULL, "department" character varying NULL, "timezone" character varying NOT NULL DEFAULT 'UTC', "language" character varying NOT NULL DEFAULT 'ru', "birthday" timestamptz NULL, "system_role" character varying NOT NULL DEFAULT 'member', "status" character varying NOT NULL DEFAULT 'active', "pending_since" timestamptz NULL, "onboarding_completed" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "quiet_hours_start" character varying NULL, "quiet_hours_end" character varying NULL, "custom_fields" jsonb NULL, "updated_at" timestamptz NOT NULL, "organization_users" uuid NULL, "scheduled_call_participants" uuid NULL, "test_run_assignees" uuid NULL, PRIMARY KEY ("id"));
-- Create index "user_email" to table: "users"
CREATE INDEX "user_email" ON "users" ("email");
-- Create index "user_status" to table: "users"
CREATE INDEX "user_status" ON "users" ("status");
-- Create index "user_telegram_id" to table: "users"
CREATE INDEX "user_telegram_id" ON "users" ("telegram_id");
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "users" ("email");
-- Create index "users_telegram_id_key" to table: "users"
CREATE UNIQUE INDEX "users_telegram_id_key" ON "users" ("telegram_id");
-- Create "webhook_deliveries" table
CREATE TABLE "webhook_deliveries" ("id" uuid NOT NULL, "event_type" character varying NOT NULL, "payload" jsonb NOT NULL, "status_code" bigint NOT NULL, "response_body" text NULL, "error" text NULL, "delivered_at" timestamptz NOT NULL, "webhook_endpoint_deliveries" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "webhookdelivery_delivered_at" to table: "webhook_deliveries"
CREATE INDEX "webhookdelivery_delivered_at" ON "webhook_deliveries" ("delivered_at");
-- Create index "webhookdelivery_event_type" to table: "webhook_deliveries"
CREATE INDEX "webhookdelivery_event_type" ON "webhook_deliveries" ("event_type");
-- Create "webhook_endpoints" table
CREATE TABLE "webhook_endpoints" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "url" character varying NOT NULL, "secret" character varying NOT NULL, "events" jsonb NOT NULL, "is_active" boolean NOT NULL DEFAULT true, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "project_webhook_endpoints" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "webhookendpoint_deleted_at" to table: "webhook_endpoints"
CREATE INDEX "webhookendpoint_deleted_at" ON "webhook_endpoints" ("deleted_at");
-- Create index "webhookendpoint_is_active" to table: "webhook_endpoints"
CREATE INDEX "webhookendpoint_is_active" ON "webhook_endpoints" ("is_active");
-- Create "workflow_status" table
CREATE TABLE "workflow_status" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "name" character varying NOT NULL, "category" character varying NOT NULL, "color" character varying NULL DEFAULT '#6B7280', "sort_order" bigint NOT NULL DEFAULT 0, "is_default" boolean NOT NULL DEFAULT false, "wip_limit" bigint NOT NULL DEFAULT 0, "deleted_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "project_workflow_statuses" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "workflowstatus_deleted_at" to table: "workflow_status"
CREATE INDEX "workflowstatus_deleted_at" ON "workflow_status" ("deleted_at");
-- Create index "workflowstatus_deleted_at_project_workflow_statuses" to table: "workflow_status"
CREATE INDEX "workflowstatus_deleted_at_project_workflow_statuses" ON "workflow_status" ("deleted_at", "project_workflow_statuses");
-- Create index "workflowstatus_name_project_workflow_statuses" to table: "workflow_status"
CREATE UNIQUE INDEX "workflowstatus_name_project_workflow_statuses" ON "workflow_status" ("name", "project_workflow_statuses");
-- Create index "workflowstatus_sort_order_project_workflow_statuses" to table: "workflow_status"
CREATE INDEX "workflowstatus_sort_order_project_workflow_statuses" ON "workflow_status" ("sort_order", "project_workflow_statuses");
-- Create "workflow_transitions" table
CREATE TABLE "workflow_transitions" ("id" uuid NOT NULL, "version" bigint NOT NULL DEFAULT 1, "created_at" timestamptz NOT NULL, "workflow_status_from_transitions" uuid NOT NULL, "workflow_transition_to_status" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "workflowtransition_workflow_st_270302d69bb8059a49f686a2e26c54b1" to table: "workflow_transitions"
CREATE UNIQUE INDEX "workflowtransition_workflow_st_270302d69bb8059a49f686a2e26c54b1" ON "workflow_transitions" ("workflow_status_from_transitions", "workflow_transition_to_status");
-- Modify "activity_logs" table
ALTER TABLE "activity_logs" ADD CONSTRAINT "activity_logs_issues_activities" FOREIGN KEY ("issue_activities") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "activity_logs_users_activity_logs" FOREIGN KEY ("user_activity_logs") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "ai_providers" table
ALTER TABLE "ai_providers" ADD CONSTRAINT "ai_providers_users_ai_providers" FOREIGN KEY ("user_ai_providers") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "ai_usage_dailies" table
ALTER TABLE "ai_usage_dailies" ADD CONSTRAINT "ai_usage_dailies_users_ai_usage_daily" FOREIGN KEY ("user_ai_usage_daily") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "ai_user_preferences" table
ALTER TABLE "ai_user_preferences" ADD CONSTRAINT "ai_user_preferences_ai_providers_default_provider" FOREIGN KEY ("default_provider_id") REFERENCES "ai_providers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "ai_user_preferences_users_ai_preference" FOREIGN KEY ("user_ai_preference") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "api_keys" table
ALTER TABLE "api_keys" ADD CONSTRAINT "api_keys_users_api_keys" FOREIGN KEY ("user_api_keys") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "automation_rules" table
ALTER TABLE "automation_rules" ADD CONSTRAINT "automation_rules_projects_automation_rules" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "automation_rules_users_automation_rules" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "boards" table
ALTER TABLE "boards" ADD CONSTRAINT "boards_projects_boards" FOREIGN KEY ("project_boards") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "bookmarks" table
ALTER TABLE "bookmarks" ADD CONSTRAINT "bookmarks_channels_bookmarks" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "bookmarks_users_bookmarks" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "call_participants" table
ALTER TABLE "call_participants" ADD CONSTRAINT "call_participants_calls_participants" FOREIGN KEY ("call_id") REFERENCES "calls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "call_participants_users_call_participations" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "calls" table
ALTER TABLE "calls" ADD CONSTRAINT "calls_channels_calls" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "calls_users_created_calls" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "channel_invites" table
ALTER TABLE "channel_invites" ADD CONSTRAINT "channel_invites_channels_invites" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "channel_invites_users_channel_invites" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "channel_members" table
ALTER TABLE "channel_members" ADD CONSTRAINT "channel_members_channels_members" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "channel_members_users_channel_memberships" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "channels" table
ALTER TABLE "channels" ADD CONSTRAINT "channels_projects_channels" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "channels_users_created_channels" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "coffee_enrollments" table
ALTER TABLE "coffee_enrollments" ADD CONSTRAINT "coffee_enrollments_users_coffee_enrollment" FOREIGN KEY ("user_coffee_enrollment") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "coffee_pairs" table
ALTER TABLE "coffee_pairs" ADD CONSTRAINT "coffee_pairs_coffee_rounds_pairs" FOREIGN KEY ("coffee_round_pairs") REFERENCES "coffee_rounds" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "coffee_pairs_users_coffee_pairs_a" FOREIGN KEY ("user_coffee_pairs_a") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "coffee_pairs_users_coffee_pairs_b" FOREIGN KEY ("user_coffee_pairs_b") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "coffee_rounds" table
ALTER TABLE "coffee_rounds" ADD CONSTRAINT "coffee_rounds_users_coffee_rounds_triggered" FOREIGN KEY ("user_coffee_rounds_triggered") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "comments" table
ALTER TABLE "comments" ADD CONSTRAINT "comments_issues_comments" FOREIGN KEY ("issue_comments") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "comments_users_comments" FOREIGN KEY ("user_comments") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "components" table
ALTER TABLE "components" ADD CONSTRAINT "components_projects_components" FOREIGN KEY ("project_components") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "crm_contacts" table
ALTER TABLE "crm_contacts" ADD CONSTRAINT "crm_contacts_users_crm_contact_link" FOREIGN KEY ("user_crm_contact_link") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "crm_contacts_users_crm_contacts_created" FOREIGN KEY ("user_crm_contacts_created") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "crm_deals" table
ALTER TABLE "crm_deals" ADD CONSTRAINT "crm_deals_crm_contacts_deals" FOREIGN KEY ("crm_contact_deals") REFERENCES "crm_contacts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "crm_deals_projects_crm_deals" FOREIGN KEY ("project_crm_deals") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "crm_deals_users_crm_deals_owned" FOREIGN KEY ("user_crm_deals_owned") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "custom_field_definitions" table
ALTER TABLE "custom_field_definitions" ADD CONSTRAINT "custom_field_definitions_projects_custom_fields" FOREIGN KEY ("project_custom_fields") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "dashboard_layouts" table
ALTER TABLE "dashboard_layouts" ADD CONSTRAINT "dashboard_layouts_users_dashboard_layouts" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "environments" table
ALTER TABLE "environments" ADD CONSTRAINT "environments_projects_environments" FOREIGN KEY ("project_environments") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "file_metadata" table
ALTER TABLE "file_metadata" ADD CONSTRAINT "file_metadata_issues_files" FOREIGN KEY ("issue_files") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_metadata_projects_files" FOREIGN KEY ("project_files") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_metadata_test_results_files" FOREIGN KEY ("test_result_files") REFERENCES "test_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "file_metadata_users_files" FOREIGN KEY ("uploader_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "git_hub_integrations" table
ALTER TABLE "git_hub_integrations" ADD CONSTRAINT "git_hub_integrations_projects_github_integrations" FOREIGN KEY ("project_github_integrations") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "git_hub_links" table
ALTER TABLE "git_hub_links" ADD CONSTRAINT "git_hub_links_git_hub_integrations_github_links" FOREIGN KEY ("git_hub_integration_github_links") REFERENCES "git_hub_integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "git_hub_links_issues_github_links" FOREIGN KEY ("issue_github_links") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "git_lab_integrations" table
ALTER TABLE "git_lab_integrations" ADD CONSTRAINT "git_lab_integrations_projects_gitlab_integrations" FOREIGN KEY ("project_gitlab_integrations") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "git_lab_links" table
ALTER TABLE "git_lab_links" ADD CONSTRAINT "git_lab_links_git_lab_integrations_gitlab_links" FOREIGN KEY ("git_lab_integration_gitlab_links") REFERENCES "git_lab_integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "git_lab_links_issues_gitlab_links" FOREIGN KEY ("issue_gitlab_links") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "hub_bases" table
ALTER TABLE "hub_bases" ADD CONSTRAINT "hub_bases_users_hub_bases" FOREIGN KEY ("user_hub_bases") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "hub_columns" table
ALTER TABLE "hub_columns" ADD CONSTRAINT "hub_columns_hub_tables_columns" FOREIGN KEY ("hub_table_columns") REFERENCES "hub_tables" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "hub_records" table
ALTER TABLE "hub_records" ADD CONSTRAINT "hub_records_hub_tables_records" FOREIGN KEY ("hub_table_records") REFERENCES "hub_tables" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "hub_records_users_hub_records" FOREIGN KEY ("user_hub_records") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "hub_tables" table
ALTER TABLE "hub_tables" ADD CONSTRAINT "hub_tables_hub_bases_tables" FOREIGN KEY ("hub_base_tables") REFERENCES "hub_bases" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "invites" table
ALTER TABLE "invites" ADD CONSTRAINT "invites_projects_invites" FOREIGN KEY ("project_invites") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "invites_users_invites" FOREIGN KEY ("user_invites") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "invoices" table
ALTER TABLE "invoices" ADD CONSTRAINT "invoices_crm_deals_invoices" FOREIGN KEY ("crm_deal_invoices") REFERENCES "crm_deals" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "invoices_journal_entries_invoice" FOREIGN KEY ("journal_entry_invoice") REFERENCES "journal_entries" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "invoices_users_invoices_created" FOREIGN KEY ("user_invoices_created") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "issue_labels" table
ALTER TABLE "issue_labels" ADD CONSTRAINT "issue_labels_issue_id" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "issue_labels_label_id" FOREIGN KEY ("label_id") REFERENCES "labels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "issue_links" table
ALTER TABLE "issue_links" ADD CONSTRAINT "issue_links_issues_outgoing_links" FOREIGN KEY ("issue_outgoing_links") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "issue_links_issues_target" FOREIGN KEY ("issue_link_target") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "issue_message_links" table
ALTER TABLE "issue_message_links" ADD CONSTRAINT "issue_message_links_channels_issue_message_links" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "issue_message_links_issues_message_links" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "issue_message_links_users_issue_message_links" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "issues" table
ALTER TABLE "issues" ADD CONSTRAINT "issues_projects_issues" FOREIGN KEY ("project_issues") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "issues_sprints_issues" FOREIGN KEY ("sprint_issues") REFERENCES "sprints" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "issues_users_assignee" FOREIGN KEY ("issue_assignee") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "issues_users_reporter" FOREIGN KEY ("issue_reporter") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "issues_workflow_status_issues" FOREIGN KEY ("workflow_status_issues") REFERENCES "workflow_status" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "journal_entries" table
ALTER TABLE "journal_entries" ADD CONSTRAINT "journal_entries_users_journal_entries" FOREIGN KEY ("user_journal_entries") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "journal_lines" table
ALTER TABLE "journal_lines" ADD CONSTRAINT "journal_lines_journal_entries_lines" FOREIGN KEY ("journal_entry_lines") REFERENCES "journal_entries" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "journal_lines_ledger_accounts_journal_lines" FOREIGN KEY ("ledger_account_journal_lines") REFERENCES "ledger_accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "kudos" table
ALTER TABLE "kudos" ADD CONSTRAINT "kudos_users_kudos_received" FOREIGN KEY ("user_kudos_received") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "kudos_users_kudos_sent" FOREIGN KEY ("user_kudos_sent") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "label_groups" table
ALTER TABLE "label_groups" ADD CONSTRAINT "label_groups_projects_label_groups" FOREIGN KEY ("project_label_groups") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "labels" table
ALTER TABLE "labels" ADD CONSTRAINT "labels_label_groups_labels" FOREIGN KEY ("label_group_labels") REFERENCES "label_groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "labels_projects_labels" FOREIGN KEY ("project_labels") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "labels_test_cases_labels" FOREIGN KEY ("test_case_labels") REFERENCES "test_cases" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "leave_requests" table
ALTER TABLE "leave_requests" ADD CONSTRAINT "leave_requests_users_leave_requests" FOREIGN KEY ("user_leave_requests") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "leave_requests_users_leave_reviews" FOREIGN KEY ("user_leave_reviews") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "marketplace_listings" table
ALTER TABLE "marketplace_listings" ADD CONSTRAINT "marketplace_listings_users_marketplace_listings" FOREIGN KEY ("user_marketplace_listings") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "marketplace_listings_users_marketplace_reviews" FOREIGN KEY ("user_marketplace_reviews") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "news_poll_options" table
ALTER TABLE "news_poll_options" ADD CONSTRAINT "news_poll_options_news_polls_options" FOREIGN KEY ("poll_id") REFERENCES "news_polls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "news_poll_votes" table
ALTER TABLE "news_poll_votes" ADD CONSTRAINT "news_poll_votes_news_poll_options_votes" FOREIGN KEY ("option_id") REFERENCES "news_poll_options" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "news_poll_votes_news_polls_votes" FOREIGN KEY ("poll_id") REFERENCES "news_polls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "news_poll_votes_users_news_poll_votes" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "news_polls" table
ALTER TABLE "news_polls" ADD CONSTRAINT "news_polls_news_posts_poll" FOREIGN KEY ("news_post_id") REFERENCES "news_posts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "news_posts" table
ALTER TABLE "news_posts" ADD CONSTRAINT "news_posts_users_news_posts" FOREIGN KEY ("user_news_posts") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "notification_preferences" table
ALTER TABLE "notification_preferences" ADD CONSTRAINT "notification_preferences_users_notification_preferences" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "notifications" table
ALTER TABLE "notifications" ADD CONSTRAINT "notifications_users_notifications" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "pinned_messages" table
ALTER TABLE "pinned_messages" ADD CONSTRAINT "pinned_messages_channels_pinned_messages" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "poll_options" table
ALTER TABLE "poll_options" ADD CONSTRAINT "poll_options_polls_options" FOREIGN KEY ("poll_id") REFERENCES "polls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "poll_votes" table
ALTER TABLE "poll_votes" ADD CONSTRAINT "poll_votes_poll_options_votes" FOREIGN KEY ("option_id") REFERENCES "poll_options" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "poll_votes_polls_votes" FOREIGN KEY ("poll_id") REFERENCES "polls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "poll_votes_users_poll_votes" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "polls" table
ALTER TABLE "polls" ADD CONSTRAINT "polls_channels_polls" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "polls_users_polls" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "project_members" table
ALTER TABLE "project_members" ADD CONSTRAINT "project_members_projects_members" FOREIGN KEY ("project_members") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "project_members_roles_members" FOREIGN KEY ("role_members") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "project_members_users_project_members" FOREIGN KEY ("user_project_members") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "project_teams" table
ALTER TABLE "project_teams" ADD CONSTRAINT "project_teams_project_id" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "project_teams_team_id" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "project_templates" table
ALTER TABLE "project_templates" ADD CONSTRAINT "project_templates_users_project_templates" FOREIGN KEY ("user_project_templates") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "projects" table
ALTER TABLE "projects" ADD CONSTRAINT "projects_organizations_projects" FOREIGN KEY ("organization_projects") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "push_subscriptions" table
ALTER TABLE "push_subscriptions" ADD CONSTRAINT "push_subscriptions_users_push_subscriptions" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "reactions" table
ALTER TABLE "reactions" ADD CONSTRAINT "reactions_comments_reactions" FOREIGN KEY ("comment_reactions") REFERENCES "comments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "reactions_users_reactions" FOREIGN KEY ("user_reactions") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "recordings" table
ALTER TABLE "recordings" ADD CONSTRAINT "recordings_calls_recordings" FOREIGN KEY ("call_id") REFERENCES "calls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "recordings_projects_recordings" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "recordings_users_recordings" FOREIGN KEY ("started_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "roles" table
ALTER TABLE "roles" ADD CONSTRAINT "roles_projects_roles" FOREIGN KEY ("project_roles") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "saved_filters" table
ALTER TABLE "saved_filters" ADD CONSTRAINT "saved_filters_projects_saved_filters" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "saved_filters_users_saved_filters" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "scheduled_calls" table
ALTER TABLE "scheduled_calls" ADD CONSTRAINT "scheduled_calls_channels_scheduled_calls" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scheduled_calls_projects_scheduled_calls" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scheduled_calls_users_scheduled_calls" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "sessions" table
ALTER TABLE "sessions" ADD CONSTRAINT "sessions_users_sessions" FOREIGN KEY ("user_sessions") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "sprints" table
ALTER TABLE "sprints" ADD CONSTRAINT "sprints_projects_sprints" FOREIGN KEY ("project_sprints") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "team_members" table
ALTER TABLE "team_members" ADD CONSTRAINT "team_members_teams_members" FOREIGN KEY ("team_members") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "team_members_users_team_members" FOREIGN KEY ("user_team_members") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "test_case_components" table
ALTER TABLE "test_case_components" ADD CONSTRAINT "test_case_components_component_id" FOREIGN KEY ("component_id") REFERENCES "components" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "test_case_components_test_case_id" FOREIGN KEY ("test_case_id") REFERENCES "test_cases" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "test_case_linked_issues" table
ALTER TABLE "test_case_linked_issues" ADD CONSTRAINT "test_case_linked_issues_issue_id" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "test_case_linked_issues_test_case_id" FOREIGN KEY ("test_case_id") REFERENCES "test_cases" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "test_case_versions" table
ALTER TABLE "test_case_versions" ADD CONSTRAINT "test_case_versions_test_cases_versions" FOREIGN KEY ("test_case_versions") REFERENCES "test_cases" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "test_case_versions_users_test_case_versions" FOREIGN KEY ("user_test_case_versions") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "test_cases" table
ALTER TABLE "test_cases" ADD CONSTRAINT "test_cases_projects_test_cases" FOREIGN KEY ("project_test_cases") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "test_cases_test_suites_test_cases" FOREIGN KEY ("test_suite_test_cases") REFERENCES "test_suites" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "test_cases_users_created_test_cases" FOREIGN KEY ("user_created_test_cases") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "test_results" table
ALTER TABLE "test_results" ADD CONSTRAINT "test_results_test_cases_results" FOREIGN KEY ("test_case_results") REFERENCES "test_cases" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "test_results_test_runs_results" FOREIGN KEY ("test_run_results") REFERENCES "test_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "test_results_users_executed_results" FOREIGN KEY ("user_executed_results") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "test_runs" table
ALTER TABLE "test_runs" ADD CONSTRAINT "test_runs_environments_test_runs" FOREIGN KEY ("environment_test_runs") REFERENCES "environments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "test_runs_projects_test_runs" FOREIGN KEY ("project_test_runs") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "test_runs_users_created_runs" FOREIGN KEY ("user_created_runs") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "test_suites" table
ALTER TABLE "test_suites" ADD CONSTRAINT "test_suites_projects_test_suites" FOREIGN KEY ("project_test_suites") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "time_logs" table
ALTER TABLE "time_logs" ADD CONSTRAINT "time_logs_issues_time_logs" FOREIGN KEY ("issue_time_logs") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "time_logs_users_time_logs" FOREIGN KEY ("user_time_logs") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "user_blocks" table
ALTER TABLE "user_blocks" ADD CONSTRAINT "user_blocks_users_blocks_created" FOREIGN KEY ("blocker_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "user_blocks_users_blocks_received" FOREIGN KEY ("blocked_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_organizations_users" FOREIGN KEY ("organization_users") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "users_scheduled_calls_participants" FOREIGN KEY ("scheduled_call_participants") REFERENCES "scheduled_calls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "users_test_runs_assignees" FOREIGN KEY ("test_run_assignees") REFERENCES "test_runs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "webhook_deliveries" table
ALTER TABLE "webhook_deliveries" ADD CONSTRAINT "webhook_deliveries_webhook_endpoints_deliveries" FOREIGN KEY ("webhook_endpoint_deliveries") REFERENCES "webhook_endpoints" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "webhook_endpoints" table
ALTER TABLE "webhook_endpoints" ADD CONSTRAINT "webhook_endpoints_projects_webhook_endpoints" FOREIGN KEY ("project_webhook_endpoints") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "workflow_status" table
ALTER TABLE "workflow_status" ADD CONSTRAINT "workflow_status_projects_workflow_statuses" FOREIGN KEY ("project_workflow_statuses") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "workflow_transitions" table
ALTER TABLE "workflow_transitions" ADD CONSTRAINT "workflow_transitions_workflow_status_from_transitions" FOREIGN KEY ("workflow_status_from_transitions") REFERENCES "workflow_status" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "workflow_transitions_workflow_status_to_status" FOREIGN KEY ("workflow_transition_to_status") REFERENCES "workflow_status" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
