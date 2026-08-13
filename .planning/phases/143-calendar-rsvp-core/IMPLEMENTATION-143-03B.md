# Task 143-03B — static acceptance evidence

> **Status:** STATIC_ACCEPTED · GENERATION_NOT_RUN · RUNTIME_NOT_RUN
> **Implementation commit:** `61c19f9f0dee5a406005aeacd4af056dcebc6a36`
> **Accepted dependency:** 143-03A `c6f5ec8b54b7e35327125dd4445634c5d1764491`

## Scope and result

- Authoring packet: `/root/mail_143_03b_impl`.
- Requested and reported actual route: exact `gpt-5.6-terra` / `xhigh`; no fallback.
- Independent review packet: `/root/mail_143_03b_review`, exact Terra/high.
- Sole owned file: `backend/ent/schema/calendarattendee.go`.

The schema stores only the required normalized email identity, optional linked Hub user,
role/RSVP state, optional occurrence identity, optimistic version and bounded display name.
It preserves the accepted event edge, nullable-user `SET NULL` behavior and `VersionMixin`.
Its two PostgreSQL partial unique indexes separately enforce
`(event, normalized_email) WHERE recurrence_id IS NULL` for the series row and
`(event, normalized_email, recurrence_id) WHERE recurrence_id IS NOT NULL` for a
materialized occurrence; no nullable composite substitute or duplicate raw email column
was invented.

Independent static review returned PASS with no P0, P1 or P2. The accepted source blob is
`3753d056ecdc00a3ca3f690cf8a29fa6b98eb2b4` (file SHA-256
`ED2021E945C53E7F47943AE0919C460CB77388559DCE529F862224DE4E1670E6`).

Ent generation, schema compilation, migrations and executable tests remain `NOT_RUN`; the
accepted launcher is `NOT_READY`. No generated or migration file was changed.
