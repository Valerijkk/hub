-- Marker: needs-owner-apply
-- One-shot additive expansion. Any existing same-name object must fail the
-- transaction for owner triage; do not make this migration idempotent.
BEGIN;

ALTER TABLE calls ADD COLUMN provider VARCHAR NOT NULL DEFAULT 'livekit';
ALTER TABLE calls ADD COLUMN presence_mode VARCHAR NOT NULL DEFAULT 'legacy';
ALTER TABLE calls ADD CONSTRAINT calls_provider_allowed CHECK (provider IN ('livekit', 'jitsi'));
ALTER TABLE calls ADD CONSTRAINT calls_presence_mode_allowed CHECK (presence_mode IN ('legacy', 'durable'));

COMMIT;
