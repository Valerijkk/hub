-- Marker: needs-owner-apply
-- One-shot additive expansion. Any existing same-name object must fail the
-- transaction for owner triage; do not make this migration idempotent.
BEGIN;

ALTER TABLE calls ADD COLUMN provider VARCHAR NOT NULL DEFAULT 'livekit';
ALTER TABLE calls ADD COLUMN presence_mode VARCHAR NOT NULL DEFAULT 'legacy';
ALTER TABLE calls ADD CONSTRAINT calls_provider_allowed CHECK (provider IN ('livekit', 'jitsi'));
ALTER TABLE calls ADD CONSTRAINT calls_presence_mode_allowed CHECK (presence_mode IN ('legacy', 'durable'));

CREATE TABLE call_participant_connections (
	id UUID NOT NULL PRIMARY KEY,
	call_id UUID NOT NULL,
	user_id UUID NOT NULL,
	epoch UUID NOT NULL,
	state VARCHAR(9) NOT NULL DEFAULT 'unknown',
	connected_at TIMESTAMPTZ NULL,
	left_at TIMESTAMPTZ NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	version INTEGER NOT NULL DEFAULT 1,
	CONSTRAINT call_participant_connections_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_participant_connections_call_id_nonzero CHECK (call_id <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_participant_connections_user_id_nonzero CHECK (user_id <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_participant_connections_epoch_nonzero CHECK (epoch <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_participant_connections_state_allowed CHECK (state IN ('unknown', 'connected', 'left')),
	CONSTRAINT call_participant_connections_version_positive CHECK (version > 0),
	CONSTRAINT call_participant_connections_state_timestamps_valid CHECK (
		(state = 'unknown' AND connected_at IS NULL AND left_at IS NULL)
		OR (state = 'connected' AND connected_at IS NOT NULL AND left_at IS NULL)
		OR (state = 'left' AND left_at IS NOT NULL)
	),
	CONSTRAINT call_participant_connections_timestamp_order_valid CHECK (
		connected_at IS NULL OR left_at IS NULL OR left_at >= connected_at
	),
	CONSTRAINT call_participant_connections_call_fk FOREIGN KEY (call_id) REFERENCES calls(id) ON DELETE CASCADE,
	CONSTRAINT call_participant_connections_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE NO ACTION
);

CREATE UNIQUE INDEX call_participant_connections_call_user_epoch_key
	ON call_participant_connections (call_id, user_id, epoch);
CREATE INDEX call_participant_connections_left_retention_idx
	ON call_participant_connections (left_at) WHERE state = 'left';
CREATE INDEX call_participant_connections_user_idx
	ON call_participant_connections (user_id);

CREATE TABLE call_participant_presences (
	id UUID NOT NULL PRIMARY KEY,
	call_id UUID NOT NULL,
	user_id UUID NOT NULL,
	connected_count BIGINT NOT NULL DEFAULT 0,
	presence_sequence BIGINT NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	CONSTRAINT call_participant_presences_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_participant_presences_call_id_nonzero CHECK (call_id <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_participant_presences_user_id_nonzero CHECK (user_id <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_participant_presences_connected_count_nonnegative CHECK (connected_count >= 0),
	CONSTRAINT call_participant_presences_sequence_nonnegative CHECK (presence_sequence >= 0),
	CONSTRAINT call_participant_presences_call_fk FOREIGN KEY (call_id) REFERENCES calls(id) ON DELETE CASCADE,
	CONSTRAINT call_participant_presences_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE NO ACTION
);

CREATE UNIQUE INDEX call_participant_presences_call_user_key
	ON call_participant_presences (call_id, user_id);
CREATE INDEX call_participant_presences_user_idx
	ON call_participant_presences (user_id);

CREATE TABLE call_provider_deliveries (
	id UUID NOT NULL PRIMARY KEY,
	provider VARCHAR(7) NOT NULL,
	delivery_id VARCHAR(255) NOT NULL,
	kind VARCHAR(12) NOT NULL,
	call_id UUID NOT NULL,
	user_id UUID NULL,
	epoch UUID NOT NULL,
	result VARCHAR(9) NOT NULL DEFAULT 'pending',
	received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	terminal_at TIMESTAMPTZ NULL,
	CONSTRAINT call_provider_deliveries_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_provider_deliveries_call_id_nonzero CHECK (call_id <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_provider_deliveries_user_id_nonzero CHECK (
		user_id IS NULL OR user_id <> '00000000-0000-0000-0000-000000000000'::uuid
	),
	CONSTRAINT call_provider_deliveries_epoch_nonzero CHECK (epoch <> '00000000-0000-0000-0000-000000000000'::uuid),
	CONSTRAINT call_provider_deliveries_provider_allowed CHECK (provider IN ('livekit', 'jitsi')),
	CONSTRAINT call_provider_deliveries_delivery_id_nonempty CHECK (delivery_id <> ''),
	CONSTRAINT call_provider_deliveries_kind_allowed CHECK (kind IN ('connected', 'disconnected')),
	CONSTRAINT call_provider_deliveries_result_allowed CHECK (result IN ('pending', 'applied', 'ignored', 'reconcile')),
	CONSTRAINT call_provider_deliveries_result_terminal_user_valid CHECK (
		(result IN ('pending', 'reconcile') AND terminal_at IS NULL AND user_id IS NOT NULL)
		OR (result IN ('applied', 'ignored') AND terminal_at IS NOT NULL)
	)
);

CREATE UNIQUE INDEX call_provider_deliveries_provider_delivery_key
	ON call_provider_deliveries (provider, delivery_id);
CREATE INDEX call_provider_deliveries_result_terminal_idx
	ON call_provider_deliveries (result, terminal_at);
CREATE INDEX call_provider_deliveries_user_idx
	ON call_provider_deliveries (user_id);

COMMIT;
