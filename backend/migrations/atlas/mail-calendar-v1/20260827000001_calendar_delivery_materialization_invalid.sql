ALTER TABLE "calendar_invite_deliveries"
  DROP CONSTRAINT "calendar_invite_deliveries_state_shape";

ALTER TABLE "calendar_invite_deliveries"
  ADD CONSTRAINT "calendar_invite_deliveries_state_shape"
    CHECK (
      COALESCE((
        (state = 'pending' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND terminal_at IS NULL AND next_attempt_at IS NOT NULL AND ((attempts = 0 AND last_error_code IS NULL) OR (attempts BETWEEN 1 AND 8 AND last_error_code IN ('mailbox_unavailable', 'engine_unavailable', 'internal_error'))))
        OR (state = 'submitting' AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NULL)
        OR (state = 'unknown' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND terminal_at IS NULL AND last_error_code IN ('mailbox_unavailable', 'engine_unavailable', 'submission_outcome_unknown'))
        OR (state = 'sent' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL)
        OR (state = 'superseded' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL)
        OR (state = 'failed_permanent' AND lease_token IS NULL AND lease_until IS NULL AND next_attempt_at IS NULL AND last_error_code IN ('submission_rejected', 'invalid_recipient', 'materialization_invalid', 'policy_denied', 'retry_exhausted') AND terminal_at IS NOT NULL)
      ), FALSE)
    );

ALTER TABLE "calendar_reply_deliveries"
  DROP CONSTRAINT "calendar_reply_deliveries_state_shape";

ALTER TABLE "calendar_reply_deliveries"
  ADD CONSTRAINT "calendar_reply_deliveries_state_shape"
    CHECK (
      COALESCE((
        (state = 'pending' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND ((attempts = 0 AND last_error_code IS NULL) OR (attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND last_error_code IN ('mailbox_unavailable', 'engine_unavailable', 'internal_error'))))
        OR (state = 'submitting' AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND mail_submission_id IS NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (state = 'unknown' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND last_error_code IN ('mailbox_unavailable', 'engine_unavailable', 'submission_outcome_unknown'))
        OR (state = 'sent' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (state = 'superseded' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (state = 'failed_permanent' AND lease_token IS NULL AND lease_until IS NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IN ('submission_rejected', 'invalid_recipient', 'materialization_invalid', 'policy_denied', 'retry_exhausted') AND (last_error_code <> 'retry_exhausted' OR attempts = 8))
      ), FALSE)
    );
