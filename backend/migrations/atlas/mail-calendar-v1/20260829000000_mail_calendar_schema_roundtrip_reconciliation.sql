ALTER TABLE "calendar_invite_deliveries"
  DROP CONSTRAINT "calendar_invite_deliveries_state_shape";

ALTER TABLE "calendar_invite_deliveries"
  ADD CONSTRAINT "calendar_invite_deliveries_state_shape"
    CHECK (
      COALESCE((
        (((state)::text = 'pending'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND terminal_at IS NULL AND next_attempt_at IS NOT NULL AND (((attempts = 0) AND last_error_code IS NULL) OR ((attempts BETWEEN 1 AND 8) AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('internal_error'::character varying)::text])))))
        OR (((state)::text = 'submitting'::text) AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NULL)
        OR (((state)::text = 'unknown'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND terminal_at IS NULL AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('submission_outcome_unknown'::character varying)::text])))
        OR (((state)::text = 'sent'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL)
        OR (((state)::text = 'superseded'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL)
        OR (((state)::text = 'failed_permanent'::text) AND lease_token IS NULL AND lease_until IS NULL AND next_attempt_at IS NULL AND ((last_error_code)::text = ANY (ARRAY[('submission_rejected'::character varying)::text, ('invalid_recipient'::character varying)::text, ('materialization_invalid'::character varying)::text, ('policy_denied'::character varying)::text, ('retry_exhausted'::character varying)::text])) AND terminal_at IS NOT NULL)
      ), false)
    );

ALTER TABLE "calendar_reply_deliveries"
  DROP CONSTRAINT "calendar_reply_deliveries_state_shape";

ALTER TABLE "calendar_reply_deliveries"
  ADD CONSTRAINT "calendar_reply_deliveries_state_shape"
    CHECK (
      COALESCE((
        (((state)::text = 'pending'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND (((attempts = 0) AND last_error_code IS NULL) OR ((attempts BETWEEN 1 AND 8) AND next_attempt_at IS NOT NULL AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('internal_error'::character varying)::text])))))
        OR (((state)::text = 'submitting'::text) AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND mail_submission_id IS NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (((state)::text = 'unknown'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('submission_outcome_unknown'::character varying)::text])))
        OR (((state)::text = 'sent'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (((state)::text = 'superseded'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (((state)::text = 'failed_permanent'::text) AND lease_token IS NULL AND lease_until IS NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND ((last_error_code)::text = ANY (ARRAY[('submission_rejected'::character varying)::text, ('invalid_recipient'::character varying)::text, ('materialization_invalid'::character varying)::text, ('policy_denied'::character varying)::text, ('retry_exhausted'::character varying)::text])) AND (((last_error_code)::text <> 'retry_exhausted'::text) OR attempts = 8))
      ), false)
    );

ALTER TABLE "calendar_events"
  RENAME CONSTRAINT "calendar_events_scheduled_calls_calendar_event"
  TO "calendar_events_scheduled_calls_calendar_event_link";

ALTER TABLE "mail_attachments"
  RENAME CONSTRAINT "mail_attachments_mail_messages_attachments"
  TO "mail_attachments_mail_messages_attachment_records";
