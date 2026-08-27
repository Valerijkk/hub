ALTER TABLE "calendar_invite_deliveries"
  DROP CONSTRAINT "calendar_invite_deliveries_event_sequence_positive";

ALTER TABLE "calendar_invite_deliveries"
  ADD CONSTRAINT "calendar_invite_deliveries_event_sequence_non_negative"
    CHECK (event_sequence >= 0);
