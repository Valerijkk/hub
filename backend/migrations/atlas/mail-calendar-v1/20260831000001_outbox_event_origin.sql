-- Modify "outbox_events" table
ALTER TABLE "outbox_events" ADD COLUMN "origin_kind" character varying NULL, ADD COLUMN "origin_id" uuid NULL, ADD CONSTRAINT "outbox_events_origin_pair" CHECK ((origin_kind IS NULL AND origin_id IS NULL) OR (origin_kind IS NOT NULL AND origin_id IS NOT NULL));
-- Create index "outboxevent_origin_kind_origin_id_event_type" to table: "outbox_events"
CREATE INDEX "outboxevent_origin_kind_origin_id_event_type" ON "outbox_events" ("origin_kind", "origin_id", "event_type");
