ALTER TABLE "mail_accounts"
  ADD COLUMN "password_enrollment_generation" bigint NOT NULL DEFAULT 0,
  ADD COLUMN "password_enrolled_engine_account_id" character varying NULL,
  ADD COLUMN "password_enrolled_at" timestamptz NULL,
  ADD COLUMN "password_enrollment_lease_token" uuid NULL,
  ADD COLUMN "password_enrollment_lease_until" timestamptz NULL;

ALTER TABLE "mail_accounts"
  DROP CONSTRAINT "mail_accounts_deleting_blocked_pending_shape";

ALTER TABLE "mail_accounts"
  ADD CONSTRAINT "mail_accounts_password_enrollment_marker_pair"
    CHECK (("password_enrolled_engine_account_id" IS NULL) = ("password_enrolled_at" IS NULL)),
  ADD CONSTRAINT "mail_accounts_password_enrollment_marker_generation"
    CHECK ("password_enrolled_engine_account_id" IS NULL OR "password_enrollment_generation" >= 1),
  ADD CONSTRAINT "mail_accounts_password_enrollment_marker_engine_account"
    CHECK ("password_enrolled_engine_account_id" IS NULL OR ("engine_account_id" IS NOT NULL AND "password_enrolled_engine_account_id" = "engine_account_id")),
  ADD CONSTRAINT "mail_accounts_password_enrollment_lease_pair"
    CHECK (("password_enrollment_lease_token" IS NULL) = ("password_enrollment_lease_until" IS NULL)),
  ADD CONSTRAINT "mail_accounts_password_enrollment_live_lease"
    CHECK ("password_enrollment_lease_token" IS NULL OR ("status" = 'provisioning' AND "desired_enabled" = TRUE AND "engine_account_id" IS NOT NULL AND "deployed_generation" IS NOT NULL)),
  ADD CONSTRAINT "mail_accounts_password_enrollment_generation_non_negative"
    CHECK ("password_enrollment_generation" >= 0),
  ADD CONSTRAINT "mail_accounts_deleting_blocked_pending_shape"
    CHECK ("status" <> 'deleting' OR ("desired_enabled" = FALSE AND "password_enrollment_lease_token" IS NULL AND "password_enrollment_lease_until" IS NULL AND "sync_lease_token" IS NULL AND "sync_lease_until" IS NULL AND "mutation_lease_token" IS NULL AND "mutation_lease_until" IS NULL AND "provision_lease_token" IS NULL AND "provision_lease_until" IS NULL AND "attempted_control_generation" IS NULL AND "attempted_control_hash_version" IS NULL AND "attempted_control_hash" IS NULL AND "provision_error_code" IS NULL AND "next_provision_at" IS NOT NULL));
