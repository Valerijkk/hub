# Task 143-03B — static acceptance evidence

> **Status:** STATIC_ACCEPTED · INDEPENDENT_REVIEW_C/I/M_0/0/0 · RUNTIME_NOT_RUN
> **Implementation commit:** `61c19f9f0dee5a406005aeacd4af056dcebc6a36`
> **Accepted dependency:** 143-03A `31bfff8f58ba83d12b59600135503f98a6c92fbd`
> **Independent re-review boundary:** clean `dd346e318b9ef2eab390f7509b06eb412a65d730`

## Scope and result

- Authoring packet: `/root/mail_143_03b_impl`.
- Requested and reported actual route: exact `gpt-5.6-terra` / `xhigh`; no fallback.
- Independent re-review requested exact `gpt-5.6-terra` / `high`, with
  `provenance_source=dispatcher-enforced`; actual telemetry `UNKNOWN`.
- Sole owned file: `backend/ent/schema/calendarattendee.go`.

The schema stores only the required normalized email identity, optional linked Hub user,
role/RSVP state, optional occurrence identity, optimistic version and bounded display name.
It preserves the accepted event edge, nullable-user `SET NULL` behavior and `VersionMixin`.
Its two PostgreSQL partial unique indexes separately enforce
`(event, normalized_email) WHERE recurrence_id IS NULL` for the series row and
`(event, normalized_email, recurrence_id) WHERE recurrence_id IS NOT NULL` for a
materialized occurrence; no nullable composite substitute or duplicate raw email column
was invented.

Independent re-review returned `STATIC_ACCEPTED`, Critical/Important/Minor `0/0/0`, on clean
`dd346e318b9ef2eab390f7509b06eb412a65d730`. The accepted source blob at the original
implementation commit is
`3753d056ecdc00a3ca3f690cf8a29fa6b98eb2b4` (file SHA-256
`ED2021E945C53E7F47943AE0919C460CB77388559DCE529F862224DE4E1670E6`).

The later allowed 143-03F inverse-edge remediation in
`backend/ent/schema/calendarattendee.go` does not invalidate the 143-03B schema verdict or its
original source commit. Packet 143-03D is technically unblocked by accepted 03A–C, with no 03D
work claimed. Packet 143-03F is `STATIC_ACCEPTED` on
`846cd00ecb243402877e18239c89a62bf07e8d14` after `SPEC PASS`/`QUALITY PASS` and final
docs/whole-boundary review C/I/M `0/0/0`; the intermediate C/I/M `0/0/1` docs-sync finding is
superseded/closed. 143-03E remains separately blocked on accepted 03D plus the launcher.

Ent generation, schema compilation, migrations and executable tests remain `NOT_RUN`; the
accepted launcher is `NOT_READY`. No generated or migration file was changed.

This static acceptance does not accept Phase 143, the Calendar product, a migration, generated
Ent output, tests, or runtime behavior.
