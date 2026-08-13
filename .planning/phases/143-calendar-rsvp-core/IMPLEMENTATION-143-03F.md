# Task 143-03F — implementation and provenance blocker

> **Status:** ROOT_CONTENT_REVIEW_PASS · INDEPENDENT_MODEL_PROVENANCE_BLOCKED · NOT_STATIC_ACCEPTED
> **Implementation commit:** `fd681bb9375d4ac64117d832dfe09a8712944d45`
> **Implementation dependency:** 143-03B `61c19f9f0dee5a406005aeacd4af056dcebc6a36` · chain later reopened

## Scope and provenance

- Authoring packet: `/root/mail_143_03f_impl`.
- Requested and reported actual authoring route: exact `gpt-5.6-terra` / `xhigh`; no
  fallback.
- Allowed files only:
  - `backend/ent/schema/calendarinboundreplyreceipt.go`
  - narrow inverse edge in `backend/ent/schema/calendarattendee.go`
- Independent review packet: `/root/mail_143_03f_review`, requested exact
  `gpt-5.6-terra` / `high`, fallback none.

## Root content review

The committed schema matches the frozen 143-03F shape:

- required immutable attendee foreign key;
- non-negative immutable event sequence;
- required payload hash version with minimum one;
- required exact 32-byte SHA-256 payload hash;
- closed immutable `applied|observed_no_change` outcome and immutable creation time;
- unique attendee plus sequence plus version plus hash;
- one additive attendee inverse edge with all prior fields, edges and indexes preserved;
- no raw sender, MIME, HTTP key/response, engine id, event or organizer mutation edge;
- no generated, migration or test file change.

Accepted source blobs at the implementation commit are:

| Path | Blob | SHA-256 |
|---|---|---|
| `backend/ent/schema/calendarinboundreplyreceipt.go` | `dbd3e6abe514dbeac1ead61ac44e005747ab93a8` | `CA156D2C48B7442750D961C5D68EDFC34186F4AC3828CA4E8EE37EA34769F03B` |
| `backend/ent/schema/calendarattendee.go` | `003cf085881aa3d20634582bebf07ec4c072344c` | `06E97C8A8810AFCB5F55682B1EF30325BD738BDBBC4CE9DC5EF27F54213C9485` |

Sender-bound hash computation, the maximum-eight distinct-hash cap and retention remain
service-owned behavior and are not claimed by this schema packet.

## Why acceptance is blocked

The independent reviewer runtime did not expose actual model/effort and therefore stopped
before reading any file. Under the founder routing lock, the root source review cannot be
misrepresented as the required independent Terra/high evidence. The implementation commit
is retained, but 143-03F is not statically accepted and cannot unlock generation or later
dependencies.

Ent generation, schema compilation, migrations and executable tests remain `NOT_RUN`; the
accepted launcher is `NOT_READY`.
