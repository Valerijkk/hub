# SDD ledger — plan: docs/superpowers/plans/2026-08-22-mail-calendar-production-integration.md

Ruling: founder scope is full production Mail plus Calendar availability and colleague scheduling, not MVP — old narrower phase packets do not reduce scope — wrong ruling costs rework across API and UI.
Ruling: tests remain deferred until after merge; new production atoms receive one bounded static review and final runtime evidence comes from one Docker smoke — wrong ruling costs pre-merge automated regression coverage.
Ruling: no repeated review of already accepted immutable commits — only new or fixed risk surfaces are reviewed — wrong ruling could miss cross-atom drift, covered by final whole-branch review.

Task schema-control: complete (commits ac9558eb..07810c2b, review clean)
Task frontend-api-calendar: complete (commits 58d40af2..a60174f2, review clean)
Task mail-account-shell: complete (commits 648f04b5..14e301ab, review clean)
Task mail-folders-api: complete (commits a866059c..172f1d40, review clean)
Task mail-folder-transport: complete (commit 03767c3c, review clean)
Task mail-thread-list-api: in progress
Task real-provider-path: in progress
Task calendar-availability: in progress

Ruling: founder authorization in the active conversation permits source-only Phase 145 and 147 packets that are maintainable and fail closed; this does not waive external credential, DNS, image provenance, migration, or runtime evidence gates.
Ruling: mailbox provider authentication targets per-user OAuth Authorization Code; management credentials are provisioning-only and never authenticate mailbox JMAP; password and client-credentials mailbox access remain forbidden.
Ruling: availability reads are privacy-safe advisory snapshots; colleague call creation must later use one canonical event-before-call mutation with a race-safe server-side reservation/recheck and generic conflict response.

Task mail-thread-list-api-review-fix: in progress (3 medium findings: cursor scope, search/starred, bounded hydration)
Task mailbox-disabled-config: in progress
Task calendar-availability-policy: in progress

Ruling: production source packets may be implemented and statically reviewed before the single final Docker run, but no real provider send is executed before that final smoke - wrong ruling costs delayed discovery of provider/runtime incompatibility.
Ruling: the identity route in `d8ce1ffc` is structurally unauthenticated because Fiber executes handlers in registration order; fix it under the existing authenticated `/api/mail` group before accepting server composition - wrong ruling exposes owner-bound Mail identity metadata.

Task composer-reliability-fix: complete (commits `8bf4e876`, `f8d7d823`; fix round 1 addressed 3/3; scoped re-review new C/I/M 0/0/0; runtime NOT_RUN)
Task stalwart-submission-binding-fix: complete (commit `b4c4500e`, spec PASS, quality approved, review C/I/M 0/0/0, runtime NOT_RUN)
Task narrow-engine-ports: complete (commit `77eb0cc7`, spec PASS, quality approved, review C/I/M 0/0/0, runtime NOT_RUN)
Task calendar-meeting-ui: fix round 1/5 implementation in progress (review C/I/M 1/3/0)
Task mailbox-runtime-wiring-auth-order: complete (commits `ff7cbdf3`, `4fd5c690`; auth review 0/0/0; composition fix round 1 addressed 1/1; new C/I/M 0/0/0; runtime NOT_RUN)

Ruling: public availability responses are source-free while internal availability source remains validated - foreign colleagues reveal only bounded busy/free intervals, not event/call/leave reason - wrong ruling costs public diagnostic granularity and is intentionally paid for privacy.
Ruling: selected IANA wall time rejects nonexistent or ambiguous DST instants instead of guessing - wrong ruling may reject a user-selected edge time that could have been disambiguated with an extra UI choice.

Ruling: Stalwart JMAP upload has no wire idempotency key, client marker, digest lookup, or safe lost-BlobId reconstruction; an ambiguous attachment upload remains indeterminate, is never auto-retried or attached/sent, and relies on bounded server temporary-blob TTL/cleanup while Hub staging remains the byte authority - wrong ruling costs occasional explicit re-upload and is intentionally paid to prevent invented provider state or unbounded orphan retries.

Ruling: Stalwart is the canonical store for draft addresses, subject, body, priority and MIME while Ent retains only owner/revision/idempotency/status and attachment staging control; no draft content columns or second content store are introduced - wrong ruling would duplicate sensitive content and create competing draft authorities.

Task draft-submission-contract: architecture freeze in progress; code waits for Calendar fix review.

## 2026-08-24 - Mail rich content result ABI and deferred oracle

- Marker v2 production/docs is accepted at `376ad146` after same-pair guard review; its exact-field test oracle is deferred.
- Rich sanitizer/read source `43a4db5d18ac0049e1900668c7218863319f9339` passed production security/static review except for the user-locked stale `backend/tests/mail/mailbox_content_test.go` ABI oracle. Future update: replace the stale result expectation `{Format, PlainBody, RemoteImagesBlocked}` with the exact `SourceFormat`, `SanitizerVersion`, `PlainBody`, `SanitizedHTML`, `RemoteImagesBlocked` fields, including rich-only `SanitizedHTML` behavior. Status: `DEFERRED / NOT_RUN`.
- Founder no-test lock remains: no tests changed or run.
