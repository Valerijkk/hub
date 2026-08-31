CREATE TABLE "public"."mail_accounts" (
    "id" uuid NOT NULL,
    CONSTRAINT "mail_accounts_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id"),
    CONSTRAINT "mail_accounts_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id")
);

ALTER TABLE ONLY "public"."calendar_events"
    ADD CONSTRAINT "calendar_events_owner_fkey" FOREIGN KEY ("calendar_owner_id") REFERENCES "public"."users" ("id"),
    ADD CONSTRAINT "calendar_events_organization_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id");
