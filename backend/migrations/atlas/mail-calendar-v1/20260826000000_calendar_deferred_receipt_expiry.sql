DO $$
BEGIN
  LOCK TABLE "calendar_inbound_message_receipts" IN ACCESS EXCLUSIVE MODE;

  IF EXISTS (
    SELECT 1
    FROM "calendar_inbound_message_receipts"
    WHERE "state" = 'deferred_missing_master'
  ) THEN
    RAISE EXCEPTION 'calendar deferred receipt expiry migration requires no deferred_missing_master receipts';
  END IF;

  ALTER TABLE "calendar_inbound_message_receipts"
    ADD COLUMN "deferred_expires_at" timestamptz NULL,
    ADD COLUMN "deferred_object_state_id" uuid NULL;

  ALTER TABLE "calendar_inbound_message_receipts"
    ADD CONSTRAINT "calendar_inbound_message_receipts_deferred_expiry_second"
      CHECK ("deferred_expires_at" IS NULL OR date_trunc('second', "deferred_expires_at") = "deferred_expires_at"),
    ADD CONSTRAINT "calendar_inbound_message_receipts_deferred_pair"
      CHECK (("deferred_object_state_id" IS NULL) = ("deferred_expires_at" IS NULL)),
    ADD CONSTRAINT "calendar_inbound_message_receipts_deferred_retry"
      CHECK ("deferred_object_state_id" IS NULL OR "next_attempt_at" IS NULL OR "next_attempt_at" <= "deferred_expires_at"),
    ADD CONSTRAINT "calendar_inbound_message_receipts_deferred_state"
      CHECK ("state" <> 'deferred_missing_master' OR ("deferred_object_state_id" IS NOT NULL AND "deferred_expires_at" IS NOT NULL)),
    ADD CONSTRAINT "calendar_inbound_message_recei_d33f430eeb810019d73641cf3eee2d47"
      FOREIGN KEY ("deferred_object_state_id")
      REFERENCES "calendar_inbound_object_states" ("id")
      ON UPDATE NO ACTION
      ON DELETE NO ACTION;

  CREATE INDEX "calendar_inbound_message_receipts_deferred_object_state_idx"
    ON "calendar_inbound_message_receipts" ("deferred_object_state_id");
END
$$;
