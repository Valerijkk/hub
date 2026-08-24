# Hub Mail contract changelog
> **Current narrow deletion-source authority (2026-08-23; highest precedence):** accepted plan `3f7aa2c58407ac3a10ad427cfa86256e0a9d340a` is `DOCS_PLAN_STATIC_ACCEPTED / SOURCE_AUTHORITY_ACCEPTED` only for one future atomic no-test exact2 production packet: `backend/internal/mail/mailbox_deletion.go` and `backend/internal/mail/mailbox_deletion_fake.go`. Plan Contract/high, Security/xhigh and Lifecycle/high reviews are C/I/M `0/0/0`; `provenance_source: dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own commit and immutable C/S/L triple `0/0/0`. Adapter capability proof is not required only for this provider-neutral ABI plus two standalone fakes exception, but remains mandatory for every concrete provider adapter. Persistence/Ent/schema/generation/migration/SQL/worker/storage drain/account finalization/User-GDPR/HTTP/NATS/Docker/runtime stay `BLOCKED` / `NOT_RUN`. Old blanket deletion-source-blocked or plan-pending claims are `CONSUMED/HISTORICAL` only within this exact2 scope; unrelated gates are unchanged.
> **Engine deletion-port adjudication (2026-08-23):** `STATIC_ACCEPTED / DOCS_DESIGN_ONLY`; former deletion-port-next selection is `CONSUMED/HISTORICAL`, never source authority. Freeze separate opaque future ports: `EngineAccountDeletionPort` (`DeleteAccount`, `ReconcileAccountAbsence`) and `MailObjectDeletionPort` (`DeleteObject`, `ReconcileObjectAbsence`); do not widen normal `MailEngine` or combine account delete with drain. Requests durably retain nonzero `DeletionID` plus account provisioning key+engine account ID or object bucket+key; opaque IDs use `MaxEngineOpaqueIDBytes` and bucket/key existing storage bounds. Results are `deleted|authoritatively_absent`, observations `present|authoritatively_absent`; bare `NotFound` preserves intent unless scoped authoritative absence is proven. No provider/raw IDs or causes are public/logged. Private intent/control is not owner projection; short lease transactions keep I/O outside locks, use exact token+generation+scope CAS, reclaim expired lease only, and stale token/zero rows cause no I/O/compensation. Object drain starts after `upload_drain_until` and attachment/reference terminality; account removal is last after exact child/object/receipt/control/activity manifest; hard-delete only after closed evidence. No Stalwart route/OAuth/DNS/OpenAPI assumption. Adapter capability and independent contract/security/lifecycle reviews are mandatory; future source is bounded ABI/fake then separately planned persistence/worker. Ent/source/generated/migration/SQL/tests/Docker/runtime and physical User/GDPR remain `BLOCKED` / `NOT_RUN`.

## 2026-08-24 - Draft two-phase O/D save representation

- Static review restored the `bytes` import and added final fresh DB-clock lease revalidation before the finish transition; generation, build, and runtime remain `NOT_RUN`.

`MailDraft` now distinguishes confirmed O (`from_identity_id`, reply/forward links, and
confirmed attachments) from a reserved pending D graph. A pending save requires the nullable
pending from identity plus exact 32-byte request hash; its optional reply/forward values and
pending storage-backed attachments are desired state, never a replacement for O before a
confirmed completion. Pending identity/message/attachment foreign keys are `RESTRICT`; the
attachment owner check permits confirmed plus pending ownership only for the same draft and
never alongside a message. Engine-backed attachments forbid pending ownership. The nullable
pending identity has partial index `mail_drafts_pending_from_identity_idx` where
`pending_from_identity_id IS NOT NULL`.

The lifecycle remains repository-owned: `Reserve` preserves O and writes D in the short locked
transaction; provider I/O follows commit. `Applied` promotes D. The provider-neutral sealed
`EngineDraftRetained{Message EngineMessage}` observation and Fake authoritative current-store
proof are implemented. Stalwart retained observation is `SOURCE_NOT_IMPLEMENTED` and fail-closed:
the current request does not carry the exact prior `draft-mutation` marker binding, and a generic
syntactically valid Hub keyword is never a substitute. Durable `SaveRetained` repository/service
behavior remains `SOURCE_NOT_IMPLEMENTED` and forbidden. Every non-`Applied` outcome retains O
and D as `unknown`; the result cannot finalize, retry, clean up, or change O/D by itself. A bare
or unsealed old pair remains insufficient. A sealed live reconcile result plus the exact persisted
O pair and binding can be evidence only for a later separately authorized provider-specific proof
packet; it requires neither reverse mapping provider attachment ids to Hub UUIDs nor a
`confirmed_request_hash`.

The content-free repository O/D source is now implemented as a static packet: records carry
confirmed O, nullable D, canonical request hashes, DB-clock `updated_at`, and nullable
`confirmed_mutation_id`; work carries the effective graph, hash, and binding. The schema change
is additive and nullable with no backfill. `MailboxDraftSaveRetained` is pair-only and Finish
continues to reject it. Generated Ent, migration, SQL validation, build, tests, Docker, and
runtime remain `NOT_RUN`. `EngineDraftReconciliation.PreviousMutationID` and Stalwart exact-prior
marker proof remain `SOURCE_NOT_IMPLEMENTED`, so durable retained completion remains forbidden.
Pending attachment links are live state, so O∩D, O\D, and D\O do not permit generic cleanup or
provider work.

Migration is staged and remains `NOT_MIGRATED`: add target columns nullable first, with no
default; a zero-row `mail_drafts` first-release preflight is accepted. Existing rows require an
exact provider `GetDraft` plus one owner-scoped active, `send_allowed` `MailIdentity` mapping for
the canonical `From`. Unprovable, ambiguous, foreign, or deleted-without-proof rows block rather
than defaulting or fabricating data. Only validated rows may then receive target foreign keys,
`NOT NULL`, and checks. Generated Ent remains `NOT_GENERATED`; build, tests, Docker, and runtime
are `NOT_RUN`.

## 2026-08-23 - Mail-to-Calendar inbound admission additive prerequisite

Documentation-only P0 prerequisite; no source implementation, testframework change, migration,
generation, build, lint, format, Docker, or runtime evidence was run. `mail.received.v1` now
requires `authentication_class` and `calendar_part_locator_sha256` alongside required
`calendar_method`; class is `trusted_internal|dmarc_aligned|untrusted`, locator is null or exact
64 lowercase hex, and method/locator are strictly paired. `engine_message_id` remains only a legacy
v1 compatibility field and is ignored by the future consumer. The envelope carries only the class;
no authentication evidence, headers, explanations, raw MIME, provider locator, or new provider ID
is exposed through public APIs/logs.

The future Mail durable mapping is non-null class/default `untrusted` plus nullable exact-32-byte
domain-separated locator hash, with existing rows `untrusted`/null and no reparse/re-emission.
6C0V2 is not mutated: 6C0V3 introduces `hub.mail.sync.apply.result.v3`, version 3, preserves tags
1..9, and appends authentication class tag 10 and optional raw 32-byte locator-hash tag 11. Matching
stored v2 proofs replay only with v2; v1 remains fail-closed. The receipt contract freezes lease
token/generation, state invariants, exact-second database clock, 60-second TTL, eight attempts,
retry/deferral schedules, expiry behavior, and ABA-proof terminalization. Task 10 now depends on
the P0/P1/P2/P3 prerequisite sequence; future source manifests remain separate.

## 2026-08-22 - Task 6B Ent sync prepare immutable source receipt

`STATIC_ACCEPTED / EXACT_SHA_PASS / RUNTIME_NOT_RUN` at effective HEAD `431022ed`, after source
`6e788cc7` and live-thread correction `431022ed`, for
`backend/internal/mail/mailbox_ent_sync_prepare.go` and authored contract
`backend/tests/mail/mailbox_ent_sync_prepare_contract_test.go`. The bounded read-only Ent
transaction uses one repository clock read and account `FOR UPDATE` under the complete
current+persistent fence/current epochs and unexpired lease; provider/NATS/HTTP/random/write are
absent. It supports exact initial/resumed enumeration, incremental/catch-up and valid empty pages;
uses stable length-delimited SHA-256 UUIDv8 identities with live-ID reuse/collision guards; resolves
live-only provenance, ancestors and threads; rejects duplicate labels within a projection while
deduplicating them across the page; and selects all-and-only unused event candidates before pure
classification.

A live soft-deleted message reappearance retains the message UUID, tombstone previous state and
`ReceivedEventID`, preventing duplicate `mail.received.v1`; previous folders/flags are zero and only
still-live labels remain. Already-deleted/unknown destroyed remains `Projected=false` outside
state/provenance work. Independent immutable Terra/high review of `431022ed` passed C/I/M `0/0/0`,
`dispatcher-enforced`, telemetry `UNKNOWN`. Generated Ent client/predicates are absent; launcher
`NOT_READY`; generation, compile, tests, formatting/lint, Atlas/migration, SQL/DB, Docker and runtime
`NOT_RUN`. Task 6C remains next; no wider acceptance or merge authority follows.

## 2026-08-22 - Mail 144-05D Content plain-only source receipt sync

`SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This non-self-activating dirty docs-only exact9 receipt records immutable source `acc593fe4f20669a21a6f5ba70a8a20293111017` (sole parent `98ebedc9dd0b6ad90dbe06e1faf5810376dce392`, tree `7de89d1237bc0c8f640729f607a2dbfd9189f3c4`, subject `feat(mail): add plain content normalization`) with exact2 test `37c29c421d21e7c173142647d894de0d890894de`, then source `b064a2f6f944d5ffb10651ccb068a77634a9267b`. Source Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` reviews PASS C/I/M `0/0/0`; Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. The sole ordered exact9 receipt manifest is STATE, MAIL-SPEC, Phase-144 PLAN, DISPATCH-INDEX, DISPATCH-REMAINING, `IMPLEMENTATION-144-05D.md`, that review, this changelog and QA-HANDOFF; the receipt commit diff must set-equal it. Before atomic receipt plus immutable receipt C/S/L PASS, acc has no static acceptance and former 05D authority/acceptance/source-unimplemented/current-next routes remain current. Only after activation is acc `STATIC_ACCEPTED / EXACT_SHA_PASS`; those routes become `CONSUMED/HISTORICAL`. No next packet follows.

The pure ABI accepts explicit `plain`/`rich` input and always returns canonical plain output with remote images blocked; input/plain bytes are `8_388_608`, runes `2_000_000`, tokenizer tokens `100_000`, depth `64`, and invalid/too_large errors are typed with safe Error/Unwrap. Plain NFC/newline/whitespace canonicalization rejects invalid UTF-8/NUL/C0/C1/bidi. Rich uses existing `x/net/html` tokenizer only to extract bounded text, dropping active/suppressed elements, attrs/comments/declarations/PIs/URLs/alt/media/images while retaining anchor labels and block breaks. `x/net` `v0.51.0` is direct and `x/text` `v0.35.0` present; no sanitizer dependency or `go.mod` edit is allowed.

Adjudicated literal edges: `MailboxContentErrorKind string` is `Invalid="invalid"` or `TooLarge="too_large"`; the exact sentinels have Error text `mailbox content invalid` and `mailbox content exceeds bounds`; pointer `MailboxContentError{Kind}` Error/Unwrap maps known kinds only, while nil/unknown is `mailbox content error`/nil and every failure returns zero Result. Byte length is checked before UTF-8/rune/allocation; raw controls allow only HT/LF/CR and reject all other C0, DEL/C1 and the exact bidi set. Canonical order is NFC, CRLF/CR/LF/U+2028/U+2029, exact horizontal set, line-edge ASCII trim, >=3 LF to two, outer-LF trim, then output caps. Exported `MaxMailboxContentHTMLTokenBytes=65_536` is maximum accepted raw tokenizer-token bytes including delimiters/attributes/text. Pinned `x/net` fails when `rawLen >= maxBuf`, so Rich calls `tokenizer.SetMaxBuf(MaxMailboxContentHTMLTokenBytes + 1)` before first `Next`: raw 65_536 passes and raw 65_537 is `html.ErrBufferExceeded` -> typed `too_large`/zero Result; static source review owns any other non-EOF-to-`invalid` mapping. It counts every non-Error token before processing, enforces literal 100000/100001 and strict depth 64/65 lowercase void/self-close/match/EOF rules, rechecks decoded entity controls, and applies suppression before every block break: suppression root/descendant/self-closing/void emits neither text nor LF, only non-suppressed blocks may break, and `<span>a</span><form><p>x</p></form><span>b</span>` yields `ab`. The external unit owns exact `"<!--" + strings.Repeat("x", MaxMailboxContentHTMLTokenBytes-7) + "-->"` raw-65_536 empty-plain success and `MaxMailboxContentHTMLTokenBytes-6` payload raw-65_537 `too_large`/zero Result. The separate unterminated quoted-attribute `"<a data-x=\"" + strings.Repeat("x", MaxMailboxContentHTMLTokenBytes)` remains a no-panic/no-leak overflow vector plus a small quoted-attribute success; it does not attempt impossible generic non-EOF error injection.

Only accepted 03C-Attachment and accepted 04A are dependencies; 04L is not. This plain-only route creates no rich sanitizer, remote-image token/proxy, 05D-Remote or 05E authority. Schema/OpenAPI/persistence/Ent/API/UI/testframework/launcher/runtime remain blocked or `NOT_RUN`; no next packet follows.

## 2026-08-22 - Calendar 143-04I inbound iTIP REPLY source receipt sync

`SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT10_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This non-authoritative dirty exact10 docs receipt records immutable source `46f4d6024e6c5475cdd943af9979c4ffadc5474a` (sole parent `04dbde1a14e9499462dbc1ea8784583f83f0b40f`, tree `e3ce6097822f360a04717d0dce61597810ca2046`, subject `feat(calendar): add inbound reply service`) and ordered blobs test `e0c4e64d9d704ebde24c87d9569590cde69d9115`, repository `c5fc1df5bf0a690db199b9df922c5463d27ab2a1`, service `81814a334c17df9dbd24da58a6fb95a776924670`; the exact10 manifest is `IMPLEMENTATION-143-04I.md`. Source Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` are C/I/M `0/0/0`; Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Until one atomic manifest-equal exact10 receipt commit and immutable receipt triple PASS, source is not accepted. Only after activation source is `STATIC_ACCEPTED / EXACT_SHA_PASS` and lower 04I authority/source-unimplemented routes are `CONSUMED/HISTORICAL`. Runtime remains `NOT_RUN`; launcher/testframework external handoff, 04F-I, 03E, 04F-Base, Ent/generated source, generation/migration, HTTP/API/OpenAPI/UI and NATS/iTIP transport remain blocked/not-ready; no next Calendar packet is selected or authorized.

## 2026-08-21 - Calendar 143-04I inbound iTIP REPLY acceptance-sync candidate

`ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This non-self-activating dirty exact9 docs-only sync records immutable candidate `fa69484cde0fff77b5947a204a42f8566808ed73` (sole parent `6e54eb8c1764946826115036b90646b412db4792`, tree `d0a0971258dc34b09c8aa51c2ea71c2c231cb64b`, subject `docs(calendar): propose inbound reply authority`) and the final manifest in `REVIEW-143-04I-INBOUND-REPLY-AUTHORITY.md`. Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` are C/I/M `0/0/0`; Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Until one atomic manifest-equal exact9 acceptance commit plus immutable acceptance-commit triple PASS, fa694 remains current `ROOT_SOURCE_AUTHORITY_CANDIDATE` and source is blocked. After activation only ordered exact3 test → `event_repository.go` → new service is source-authorized, still `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` pending its own lifecycle. Only after that atomic manifest-equal exact9 acceptance commit and immutable Contract/Security/Lifecycle PASS do lower 04I candidate/precommit/root-commit/exact-SHA-next routes become `CONSUMED/HISTORICAL`; 04F-I, Ent/generated source, 03E, HTTP/API/OpenAPI/UI, NATS/iTIP transport, testframework and runtime remain blocked/not-ready/not-run; no next packet follows.

## 2026-08-21 - Calendar 143-04I inbound iTIP REPLY docs authority candidate

`ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / ATOMIC_EXACT9_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. This non-self-activating candidate freezes only the separate inbound-reply service boundary in `REVIEW-143-04I-INBOUND-REPLY-AUTHORITY.md`: inbound mail has no Hub mutation actor; `CalendarInboundReplyScope.EventOperation()` always returns `EventMutationInboundReply`; occurrence REPLY requires an already materialized active exact-recurrence attendee or is quarantined with zero writes; the applied envelope has nil actor. Source cannot start until Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high precommit PASS, atomic exact9 commit, immutable exact-SHA triple PASS and separate docs-only acceptance sync. Future source order is external-package test, repository, service only. This consumes only generic no-next and old 04I exact2 clauses, and leaves Ent/generated source, generation/migration, HTTP/API/OpenAPI/UI, NATS/outbound REPLY transport, launcher and runtime blocked/not-ready/not-run. 04F-I needs 143-03F + 143-04F-Base + accepted 143-04I; 143-03D is forbidden and insufficient.

## 2026-08-21 - Mail 144-04A owner-resolution seam immutable source receipt

> **Narrow receipt / precedence:** `STATIC_ACCEPTED / EXACT_SHA_PASS` applies only to immutable source `7531565879e0d7aa72a246407fd0e1d609a82c49` (parent `2932748df0d6efdcff4f51a3b0eae2b4b3dd3d7b`, tree `448a6db855ba1b3a6873253967c789def305cda4`, subject `feat(mail): add owner resolution seam`): ordered blobs repository `4520db07c57e80faf6318a1444eb5e2aaf51355b`, service `d33a5911fd8a9ec5d14bf2ecddaaa321c5c6b50a`, test `2b414cbd1edece218976699d2777a80f51bd057b`. Contract, Security and Lifecycle precommit plus immutable reviews PASS C/I/M `0/0/0`; provenance `dispatcher-enforced`, Terra high/xhigh/high as applicable, no fallback, telemetry `UNKNOWN`.

This receipt consumes only lower claims that abstract `mailbox_repository.go`, the owner port or the owner seam is absent or unimplemented. It does not consume current claims that concrete sync primitives, `mailbox_ent_repository.go`, an Ent adapter, SQL, 144-04H, 144-02E, generation or migration remain absent, unimplemented, blocked or separately gated.

This dirty exact6 documentation packet is `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT6_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`; it cannot self-activate. No HTTP/API/UI/wiring, runtime, product, phase or release authority follows. Launcher remains `NOT_READY`; all executables, generation, migration and runtime remain `NOT_RUN`; the external-framework handoff remains a gate.

## 2026-08-21 - Calendar 143-04H source receipt-sync candidate

> **Current receipt precedence:** Calendar 143-04H source `73940810046bb3186f2746705a83565ab29368c5` is `STATIC_ACCEPTED / EXACT_SHA_PASS` because exact16 receipt `f9f5d436c08737eeb5a3e965137f329f3ae7f627` is ancestral to HEAD. Mail 144-04I-A source `091f2ef286a71f506e1b048f5386fd37cefa06bc` is `STATIC_ACCEPTED / EXACT_SHA_PASS` because exact10 receipt `03674eace728642fc2c9b39d2e8e2c81bf016033` is ancestral to HEAD. The two named conditional receipt blocks below are `CONSUMED/HISTORICAL` only for their packets. Both preserve PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback and telemetry `UNKNOWN`; no next packet, generation, migration, adapter, API/UI, launcher or runtime authority follows.
>
`SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT16_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty
exact16 documentation receipt grants no acceptance and makes no contract or OpenAPI semantic change.
It records source `73940810046bb3186f2746705a83565ab29368c5` (parent
`9a638f75745aa0fb516a1177a33cbd1c657432fe`, tree
`9f994a8805b43a84c00ed50b8e5544739425edaa`, subject
`feat(calendar): add occurrence cancellation service`) with authored-first test blob
`80e19d3b918ae9ddbe3b3ea8803c6062c2f8a713`, repository
`e0da66bf95d4b8193eb5d4034053a353bd154b66`, service
`2b500e5cdda34d7d80d0ccc0ba9c94427167716f`. Contract Terra/high, Security Terra/xhigh and
Lifecycle Terra/high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; author
Terra/xhigh; dispatcher-enforced/no fallback; telemetry `UNKNOWN`. Only an atomic exact16 commit
equal to `IMPLEMENTATION-143-04H.md` plus immutable triple PASS makes it
`STATIC_ACCEPTED / EXACT_SHA_PASS`; lower 04H authority/precommit/source-unaccepted/current-next
clauses then are `CONSUMED/HISTORICAL`. No next packet; 04J/04K, 04F-H, Ent/adapters/handlers/API/
UI/generation/migration/raw iTIP/launcher/runtime blocks remain unchanged.

## 2026-08-21 - Calendar 143-04H authority acceptance-sync candidate

`AUTHORITY_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT15_COMMIT_REQUIRED /
IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. This dirty exact15 sync cannot self-activate and grants
no authority. Candidate `f2966bdd50631feaa210a42a232880b21ade7e3a` preserves the frozen
ABI/schema/OpenAPI semantics; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high
immutable reviews PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. Only
after atomic exact15 sync commit and immutable acceptance-commit PASS does
the candidate become `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS`; only then are lower routes
`CONSUMED/HISTORICAL` and root may dispatch the ordered exact3. No source/runtime acceptance follows.


## 2026-08-21 - Calendar 143-04H occurrence-cancel root authority candidate

`ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED /
ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. The dirty exact15
carrier is non-authoritative. It freezes the full ABI/state/oracle contract in
`REVIEW-143-04H-OCCURRENCE-CANCEL-AUTHORITY.md`, adds a disjoint schema variant requiring canonical
non-null `recurrence_id` for occurrence cancellation while preserving master-cancel compatibility,
and removes the legacy raw iTIP/serializer premise from the current 04H route. Only after triple
precommit PASS, atomic candidate commit, triple immutable exact-SHA PASS and separate accepted
exact15 sync may root dispatch the ordered test-first exact3. No source, adapter, handler, Ent,
migration, runtime or transport authority follows from these dirty bytes.


## 2026-08-21 - 144-04I-A bounded sync-worker source receipt-sync candidate

`SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT10_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty
exact10 documentation candidate grants no acceptance. It proposes authored-first source commit
`091f2ef286a71f506e1b048f5386fd37cefa06bc` (parent
`6b27643fcde3b702fa13c5fcc9b4bfefcb9fc87a`, tree
`b01be92acf01033935b85446454eaee16b84e01b`, subject
`feat(mail): add bounded sync worker`): test blob
`79c2691971287e3a62c830a5d12c78699414bb86`, then source blob
`6c20fadf563d4d146b2fc1cbacdddea485266eb7`.

Contract, Security and Lifecycle precommit PASS C/I/M `0/0/0`; immutable Contract plus
Security/Lifecycle exact-SHA PASS C/I/M `0/0/0`. Author Terra/xhigh; Contract Terra/high; Security/Lifecycle Terra/xhigh; dispatcher-enforced/no fallback; telemetry `UNKNOWN`. Only one atomic exact10 receipt-sync commit
equal to `IMPLEMENTATION-144-04I-A.md`'s manifest and its immutable receipt review may mark source
`STATIC_ACCEPTED / EXACT_SHA_PASS`. The former candidate/source-unaccepted/current-next routes are
then `CONSUMED/HISTORICAL`; no automatic source follows. This worker still amends neither
schema/event/delivery nor API/OpenAPI. 04F remains accepted. Before the manifest-equal 144-04G
exact12 coherence commit and immutable C/S/L triple PASS, 04G has no current static, source or downstream authority; 04H/02E, Ent/adapters/generation/
migration, SQL/two-replica work, API/UI, launcher and runtime remain blocked or `NOT_RUN`.

## 2026-08-21 - 144-04I-A predecessor authority route - CONSUMED/HISTORICAL

`CONSUMED/HISTORICAL`. This former exact9 acceptance route is superseded by the exact10 source
receipt above; it grants no current authority and records no current unimplemented source.

## 2026-08-21 - 144-04I-A source-authority predecessor - CONSUMED/HISTORICAL

`CONSUMED/HISTORICAL`. The non-self-activating exact9 carrier and frozen ABI/state/safety/oracles
remain in `REVIEW-144-04I-A-SYNC-WORKER-READINESS.md` as provenance only. The implemented worker
preserves raw persisted UTC precision, LeaseEpoch full fault CAS, checked true-result arithmetic
and O(1) capped backoff, without a schema/event/delivery or API/OpenAPI route. All downstream
04H/02E, Ent/adapters/generation/migration, SQL/two-replica, API/UI, launcher and runtime blockers
remain; no next packet is selected.

> **143-04K due-policy source receipt-sync (2026-08-21):**
> `SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT15_COMMIT_REQUIRED / RUNTIME_NOT_RUN`.
> This dirty exact15 documentation candidate grants no authority. Only after one atomic receipt-sync
> commit and immutable acceptance is source `572f911227b9d50e21aa73b87c530ddcaf0f8934` (parent
> `26b377eb`, tree `ac96e2f5`, subject `feat(calendar): add call due policy`)
> `STATIC_ACCEPTED / EXACT_SHA_PASS` for authored-first test blob
> `437db3af0263cedb10dcf217cd53fbbcac03cf23` then source blob
> `376f46c1cc79956fab3479efc38750fd88161b16`. Contract/Spec/Oracle, Security/Adversarial and
> Lifecycle/Dependency/Boundary precommit plus immutable reviews PASS C/I/M `0/0/0`; Terra/xhigh,
> dispatcher-enforced/no fallback, telemetry `UNKNOWN`. It consumes accepted 04J `d7926a27`/
> `e74efa827`, 04A descriptor and 04L envelope contracts, but claims no Calendar-subject admission
> from 04G. It freezes a pure explicit-`Now` call/raw-projection/envelope
> decision: active projection is deep-cloned then reconciled twice with fixed 04J boundaries, while
> inactive/deleted nil-projection/zero-envelope input immediately noops. Existing plural nine-field
> due payload remains; 04F-K alone later owns CAS/locks/rollback/outbox append.
> EventID/OccurredAt/CorrelationID are fresh active candidates: non-due outcomes discard them;
> 04F-K may candidate-call, then reserve/reload only a due outcome and rerun 04K byte-identically.
> 04K locally rechecks all envelope/payload/subject bindings and owns no reservation storage.
> No schema semantic change, legacy subject cutover or runtime authority follows.

> **143-04K receipt lifecycle:** the former authority-candidate, precommit and test-first-dispatch
> clauses become `CONSUMED/HISTORICAL` only after this source receipt activates. The dirty sync
> selects no next packet; `IMPLEMENTATION-143-04K.md` records the immutable exact2. 04F-K alone
> later owns CAS/locks/rollback/outbox append; AQ-02 backfill-only, Ent/generation/migration,
> adapters/API/UI/raw iTIP/legacy cutover, launcher and runtime blockers remain unchanged.

> **143-04J source receipt (2026-08-21):** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT14_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. These proposed facts activate only after one
> atomic exact14 docs commit. Then `e74efa827b52f35e3d3fbe8346adf0c0121ba48d` is
> `STATIC_ACCEPTED / EXACT_SHA_PASS` for the frozen exact2, and lower 04J source-unaccepted routes
> are `CONSUMED/HISTORICAL`. No 04K, persistence, adapter, API/UI or runtime authority follows.

> **143-04J acceptance-sync gate (2026-08-21):** `ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT13_COMMIT_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`.
> These proposed receipt facts have no authority before one atomic exact13 docs commit. After it,
> immutable candidate `5841621720cc03682b3d1f1e2af9cba2953797cf` is
> `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS`; Contract/Spec/Oracle, Security and Lifecycle
> immutable reviews PASS C/I/M `0/0/0`, and lower 04J candidate routes become
> `CONSUMED/HISTORICAL`. Only the frozen ordered test-first exact2 may then be dispatched; no
> source/runtime/downstream acceptance follows.

> **143-04C acceptance-sync gate (2026-08-20):** `ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT12_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact12 working tree grants no
> authority. The `7fd061b5...` receipt and scoped route transitions below activate only after one
> atomic commit of exactly these 12 documents. Until then the prior 04C lifecycle remains current;
> activation selects no next Calendar packet and lifts no downstream blocker.

## 2026-08-22 — 144-04G receipt-coherence sync candidate

> **Coherence gate:** `COHERENCE_SYNC_CANDIDATE / ATOMIC_EXACT12_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This non-self-activating dirty exact12 docs-only coherence candidate records immutable source `3031c53d399accc84fe10cd0c5421a85c089ac7c` and prior acceptance receipt `2a2975056c3fb8001d88863a041310271282989d`; their full metadata, blobs and ordered manifests are in `IMPLEMENTATION-144-04G.md`. Source Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` are C/I/M `0/0/0`; Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Only after an atomic manifest-equal exact12 coherence commit and immutable coherence C/S/L triple PASS is 144-04G `STATIC_ACCEPTED / EXACT_SHA_PASS`; former 04G authority/acceptance/source-unimplemented/current-next routes then become `CONSUMED/HISTORICAL`. Historical coherence selects no next Mail packet and grants no runtime/downstream authority. 04F remains accepted; 04H/02E, Ent/SQL/concrete adapter, generation/migration, API/UI, launcher/testframework and runtime remain blocked `NOT_READY`/`NOT_RUN`.

## 2026-08-20 — 144-04G source-authority acceptance-sync candidate (predecessor route)

> **Historical predecessor boundary:** This exact9 route is historical receipt evidence only. Before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, it confers no current static, source or downstream authority.

- Status: `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED /
  SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. This dirty exact9 working tree grants no
  authority. The `6ba3aed3d9a69a76f98978faf4a9e3362c4bc29b` metadata and the three immutable
  exact-SHA PASS results below are proposed receipt facts only; they activate only after one atomic
  commit of this exact9 acceptance sync. Until activation, the lower candidate/precommit/root-commit/
  exact-SHA-next lifecycle remains current and root cannot dispatch source; the separately false
  accepted-repository premise remains `CONSUMED/HISTORICAL`.

- Historical receipt fact only: after that atomic exact9 sync commit, docs authority was recorded as
  `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` only at immutable
  `6ba3aed3d9a69a76f98978faf4a9e3362c4bc29b` (parent
  `2b6837926c91d652ff7bf30447bd482740261bd0`, tree
  `bdd59cfcb31376c95f5128370a75333cad78bc14`, subject
  `docs(mail): propose sync service authority`); exact9 blobs are recorded in the readiness review.
- Contract, Security and Lifecycle immutable exact-SHA reviews PASS C/I/M `0/0/0`;
  dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Only after that atomic
  exact9 sync commit is the former candidate/precommit/root-commit/exact-SHA-next lifecycle
  `CONSUMED/HISTORICAL`.
- Historical record only: the exact9 route identified the frozen ordered exact3:
  `backend/tests/mail/mailbox_sync_service_test.go`, new interface/types-only
  `backend/internal/mail/mailbox_repository.go`, then new
  `backend/internal/mail/mailbox_sync_service.go`. Source remains
  `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own TDD, reviews, atomic commit and
  immutable exact-SHA acceptance.
- 04F remains accepted. 04H/02E, Ent/SQL/concrete adapter, generation, migration, OpenAPI and
  runtime remain blocked or outside authority; launcher `NOT_READY`, executables/runtime
  `NOT_RUN`. This acceptance sync is docs-only and changes no Calendar state.

## 2026-08-20 — 144-04G sync-service authority candidate (predecessor route)

> **Historical predecessor boundary:** This route is historical receipt evidence only. Before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, it confers no current static, source or downstream authority.

- Status is only `ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED /
  IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`; this is not source authority or acceptance.
- Frozen future test-first exact3: NEW sync-service test, NEW interface/types-only
  `mailbox_repository.go`, NEW sync service. The prior statement that 04G extends an accepted
  repository file is `CONSUMED/HISTORICAL` only for that false premise; the abstract port is part
  of exact3 and no concrete adapter/Ent/SQL is authorized.
- `REVIEW-144-04G-SYNC-SERVICE-READINESS.md` freezes the exact Go ABI and deterministic
  claim/prepare/apply/resolve state machine, including complete fences, bounded one-page engine
  work, base enumeration/catch-up/rebuild, 04F event-ID candidates, deep copy, typed non-leaking
  failures and no external I/O/callbacks under repository locks.
- Required lifecycle is triple precommit PASS -> atomic exact9 candidate -> triple immutable
  exact-SHA PASS -> separate acceptance sync. 04F stays static accepted; 04H/02E, generation,
  migration, launcher and runtime remain blocked/`NOT_READY`/`NOT_RUN`; Calendar 04C state and
  all other receipts are unchanged.
- Fix1 closes lease expiry and pagination authority: Prepare/Apply/Resolve use stable DB
  transaction time, never request.Now, and expired/equal lease means LeaseLost/zero effects even
  without takeover; every Changes and initial/resumed Enumerate request carries exact configured
  PageSize. `MailboxSyncPreparedPage` now has exact `Prepared,Input` to represent Prepare loss.

## 143-03B-T immutable source acceptance and former conditional 143-04C gate (CONSUMED/HISTORICAL as to 04C; 2026-08-20)

Calendar 143-03B-T is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `68cbf7cb227799304ce679937b519b66cb6abf9b` (parent `4715f939f24e5b5c2d3f80a38d204babe68bdd34`, tree `1537ba1d2d0fa804545fd8667ffe1dce9e8eaf1e`, subject `feat(calendar): add attendee tombstones`). The exact2 test-first blobs recorded in `IMPLEMENTATION-143-03B-T.md` are schema test `8bc4d44573e5a86e2c78df8c844fe3f1f2fae4cc` then `calendarattendee.go` `3c38799f3e56089e2e264b000ee0951a144b0257`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file saying 143-03B-T is only source-authorized, not accepted, dirty, uncommitted or unaccepted, or that 143-04C is blocked pending 03B-T source acceptance, is `CONSUMED/HISTORICAL` as to that dependency. Status: `COMBINED_EXACT19_SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. Only after one atomic exact19 docs commit and independent Contract/Security/Lifecycle immutable exact-SHA PASS for that commit may root dispatch the frozen 143-04C exact3, test first and in order: `backend/tests/calendar/event_service_update_test.go`, `backend/internal/calendar/event_repository.go`, new `backend/internal/calendar/event_service_update.go`. Until then exact3 remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and dispatch is not effective.

Mail 144-04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS` at `4715f939f24e5b5c2d3f80a38d204babe68bdd34`; before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, 144-04G has no current static, source or downstream authority. No adapter, Ent generation, migration, raw iTIP, service/API/UI/product/phase/release or runtime authority follows. Launcher `NOT_READY`; generation/migration/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; existing 143-04A/B/G receipts and other blockers remain unchanged.

## 144-04F immutable source acceptance (2026-08-20)

Mail 144-04F is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `4715f939f24e5b5c2d3f80a38d204babe68bdd34` (parent `7ac8714e99db2935ce290ad67836e1309cc05c78`, tree `6461894a90c4340bf96118ed1a242516c02b48c7`, subject `feat(mail): add sync page policy`). Its exact4 ordered test-first blobs are recorded in `IMPLEMENTATION-144-04F.md`: events test `50a41cefb539d9b840a43da9e1ff34cce6c8653f`, sync-policy test `b6410709e4fb8166c28a5c502cfba535eef08d06`, events source `a3017b0a6064d8fc1017e37e2d4f675ee8732945`, sync-policy source `acd642fcdb6290af79b7fa8e5aa36a5036576f47`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file asserting that Mail exact4 is dirty, `UNCOMMITTED`, `UNACCEPTED`, future-only, source-blocked, awaiting source review/commit, or root-dispatch-next is `CONSUMED/HISTORICAL`; the immutable source receipt above is current. No automatic Mail next source follows: before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, 144-04G has no current static, source or downstream authority.

The Calendar half of the combined exact17 route is unchanged: 143-03B-T exact2 remains source-authorized but is not `STATIC_ACCEPTED`; any dirty candidate remains uncommitted/unaccepted. 143-04C exact3 stays `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; 143-04A/B/G receipts remain accepted. Launcher `NOT_READY`; tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`. 143-03E, 144-02E, Ent generation and migration stay blocked. No raw iTIP or service/API/UI/product/phase/release authority follows.

## Combined exact17 conditional acceptance sync (CONSUMED/HISTORICAL; 2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

> **2026-08-20 144-02C source receipt:** `STATIC_ACCEPTED / EXACT_SHA_PASS` applies only to `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`, subject `feat(mail): add delivery schemas`) and its exact 13 blobs recorded in `IMPLEMENTATION-144-02C.md`. Candidate/source-unaccepted/lifecycle-next/implied-next 02C entries are `CONSUMED/HISTORICAL`. No next Mail code packet is selected or authorized; root must select it separately. Provenance Terra/medium, dispatcher-enforced/no fallback, telemetry `UNKNOWN`; launcher `NOT_READY`, executable evidence `NOT_RUN`.

> **2026-08-15 144-02B projection source-authority acceptance:** combined `45fd3369` + `4f3713590fee455fae1c94f2464d7de548e73e76` is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS`; Contract/high, Security/xhigh and Lifecycle/high corrected-child exact reviews PASS C/I/M `0/0/0`. Root may dispatch only this frozen test-first exact13: new `backend/ent/schema/mailfolder.go`, `backend/ent/schema/mailthread.go`, `backend/ent/schema/mailmessage.go`, `backend/ent/schema/mailmessagefolder.go`, `backend/ent/schema/maillabel.go`, `backend/ent/schema/mailmessagelabel.go`, `backend/tests/mail/mail_projection_schema_test.go`; edits `backend/ent/schema/mailaccount.go`, `backend/ent/schema/calendarevent.go`, `backend/ent/schema/mailmutationreceipt.go`, `backend/tests/mail/mail_account_identity_schema_test.go`, `backend/tests/mail/mail_mutation_receipt_schema_test.go`, `backend/tests/calendar/calendar_event_schema_test.go`.
>
> Accepted PERSISTENCE 02B only. 02F-R permits solely `MailMutationReceipt.folders` inverse for `MailFolder.control_receipt_id -> receipt RESTRICT`; actor/message receipt edges scalar-only, all other receipt edges/checks unchanged. `MailAccount` gains folders/threads/messages/labels; `CalendarEvent.mail_messages` inverse pairs with Message FK `SET NULL`; rule targets are service-only locked with no DB FK/normalized invention. Compatibility tests mandatory. Source implementation is unaccepted and still requires test-first authoring, independent review, atomic commit, exact-SHA review and status sync. Candidate/precommit/correction/failure-next and lower 02B eligibility/no-dispatch clauses are `CONSUMED/HISTORICAL`; immutable `45fd3369` failure remains history. former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; 02F-D/02E/Ent generation/migration/runtime/services/API/UI/release BLOCKED; launcher `NOT_READY`, executables `NOT_RUN`. **02C source acceptance: preceding 02C route clauses are CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS; exact13 blobs IMPLEMENTATION-144-02C.md; no next Mail packet selected; named blockers remain.**

> **144-02D static source receipt (2026-08-15):** `2242a4e7846944498eb5ed701ab8f47c986bd804` (parent `e9a1764b001e28074e4074a07ca8ac982e5345ba`, tree `8129303399e627ca5cedac3e70648955c74cfed9`, subject `feat(mail): add user configuration schemas`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for exact blobs account `55b9ccfb4e19a190fc7c083a61350ce7ba8c7e5b`, identity `6b8614c651ee70a70bdb46c00456362c8edad796`, rule `a826916965ab4875a9c61da83fa0a68a8d5d3d4d`, signature `fba258559fe6a8557b045fa10757330be29b5c08`, settings `7643892d0af8934892d23713e4d35b07b0bc455c`, contact `b2216313dfdf492315a377b3567fdc4f71d05a0`, user `b595855cbe9cd39f522ef8263eb72ea4db4f9b55`, A3 compatibility test `74599da63c180942e092a909ee53eb0ca6b7e51f`, and authored-first 02D oracle `b94c686c35e3a5cca2d7838424f7e255012437d8`. Tests first; Terra/xhigh author and Terra/high precommit+exact-SHA PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Static source only: launcher `NOT_READY`; executables/generation/migration/runtime `NOT_RUN`; no service/API/UI/product/phase/release authority. Only former 02D-next/unimplemented/source-blocked clauses are `CONSUMED/HISTORICAL`; prior docs/readiness receipts remain authoritative. Former 02B-waits-for-02F-R, 02F-R-root-dispatch and predecessor `02B requires separate docs authority` clauses are `CONSUMED/HISTORICAL`; current authority is 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 only, implementation unaccepted, and former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; other 02F incl 02F-D unauthorized. **02C source acceptance: preceding 02C route clauses are CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS; exact13 blobs IMPLEMENTATION-144-02C.md; no next Mail packet selected; named blockers remain.**

> **143-04B Calendar create contract freeze (2026-08-15):** docs-only pre-source amendment freezes the 04A transactional `CalendarEventCreateAuthorization` / `ResolveCalendarEventCreateAuthorization` ABI, receipt ordering, resolver-only internal identity authority, guarded ScheduledCall attachment and canonical EXDATE grammar. The present 04B source candidate is `PRECOMMIT_FAIL` / `NOT_ACCEPTED`: it needs an 04A ABI amendment, a corrected 04B source packet and fresh independent review. This is neither runtime, Ent, migration nor schema authority. Hub JSON event create has a canonical request hash but no inbound-iTIP `payload_hash`; existing public link errors remain frozen and only generic `ErrEventConflict` is specified where no public code was already frozen.

> **143-04B review amendment C1/I3/M1 (2026-08-15):** the proposed 04A ABI is now fully
> field-listed in `MAIL-SPEC.md` and expressly remains unaccepted until a separate 04A source
> amendment/review packet. Its resolver binds organizer and optional ScheduledCall owner
> exactly to `EventMutationScope.ActorID`; Hub JSON create has no implicit in-organization
> impersonation/delegation. Actor A -> organizer B is a zero-write `ErrEventConflict` oracle;
> actor A -> call-owner B instead returns content-free `ErrEventNotFound` / 404 with zero
> writes. Accessible actor-owned inactive/linkage/context/mapping/guest failures return
> `ErrEventConflict`. Request and resolved ScheduledCall IDs must be exactly equal when
> nonnil and both nil otherwise; substitution/drop/insertion rejects before writes. The ABI
> text adds the exact method to the existing `EventTransactionBackend` rather than redeclaring
> or replacing it. RRULE moves to the dedicated section 3.11 grammar with
> normalization/order/valid-invalid vectors, display-name/timezone boundary vectors, and
> OpenAPI EXDATE `maxLength: 20` is an explicit closed grammar tightening (timed exactly 20;
> all-day exactly 10), not a descriptions-only assertion. No code acceptance is claimed.

> **144-02A3 static source receipt (2026-08-15):** `e9a1764b001e28074e4074a07ca8ac982e5345ba` (parent `47f8987b269e1625022071e0de95a76da0d2f00a`, tree `c0b59b2ba1168076a4cc787cacb99fe5081cd301`, subject `feat(mail): add account and identity schemas`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for exactly seven paths/blobs: account `30e68ff52171e643e1361f3d6acfd2310289d8e1`, domain `16f5d20600d1359ba737417428d594ac73468360`, identity `1b06e04c9235e5c891a601261a2e400068155065`, organization `df85d2e2f61bcc8b62235f0557f774fb06eca1e3`, user `5594ca1c68b8f3c726ab90a77dcc26b7e7713e0c`, authored-first A3 test `0939618d620c7f9b503df50e4aff04fbf0b3a85d`, and A2 compatibility test `70b0c56f20043aeb8b4b719ca4ae6f3405cb5221`. Terra/xhigh author and Terra/high precommit plus exact-SHA review PASS C/I/M `0/0/0`; provenance is dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Launcher is `NOT_READY`; RED/GREEN/tests/build/gofmt/lint/validators/Docker/Ent generation/migration/runtime are `NOT_RUN`. This grants no service/API/UI/product/phase/release authority. A3-unimplemented/A3-next, 02B-waits-for-02F-R, 02F-R-root-dispatch and predecessor `02B requires separate docs authority` clauses are `CONSUMED/HISTORICAL`; current authority is 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 only, implementation unaccepted, and former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; other 02F incl 02F-D unauthorized. A1/A2 facts and all existing blockers remain unchanged. **02C source acceptance: preceding 02C route clauses are CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS; exact13 blobs IMPLEMENTATION-144-02C.md; no next Mail packet selected; named blockers remain.**

> **144-02A1 DNS static acceptance (2026-08-15):** `5892a90b7582ff4ac35db218428400ac9608c4e3` (parent `29a5c624c0bacc2de3cc4b5a93e3e9176320b5bd`, tree `c78d82ef8ea55a4656d33b82530b2fbb72582201`) accepts only `backend/internal/mail/domain_dns.go` blob `e00586637f209f85c52626c1ec8d96116af934b3` and authored-first `backend/tests/mail/domain_dns_test.go` blob `8a43140964d12cca7cdc09e34675a074ee164d7b` as `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. Terra/high static precommit and immutable exact-SHA review PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN` unless exposed. RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; launcher `NOT_READY`. It is narrow static source acceptance, not Ent/schema generation/migration/service/API/UI/product/phase/release authority. In A1/A2/A3, all three are `STATIC_ACCEPTED`; A2-unimplemented/A3-next clauses are `CONSUMED/HISTORICAL`, and A3 is accepted at `e9a1764b001e28074e4074a07ca8ac982e5345ba`. Founder root-only source sequencing before final integrated Docker does not authorize Ent generation/migrations/runtime. Existing blockers remain unchanged.

> **144-03C-Attachment static acceptance (2026-08-15):** deterministic attachment policy source receipt is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7b2ea5721684e43559d6f6b4b5fc462d8daafb8b` (parent `4e7c609713983a8f6db6fbfeb7474abd73451e8c`, tree `4c3fe4d27e170c5d3f42b116c39887036ce936c6`; source blob `3f67ce1b9f2f056b70bfd3070f1c34f98b5c71c1`, test blob `bf7eda125d5aba1d0cc2594746d14e7bcfd06aa7`). Tests were authored first. It is pure deterministic per-attachment policy with count<=50, NFC output/NFKC structural+extension defense, closed MIME/type deny tables, inline CID, SHA/scan gates, validate-versus-send/read clean-only behavior, typed content-free errors and deep-copy results. Terra/xhigh authoring; Contract Terra/high, Security Terra/xhigh and ABI/Lifecycle Terra/high precommit+exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime remain `NOT_RUN`; launcher `NOT_READY`. Static only: no scanner/storage/network/project/tenant auth/schema/migration/service/API/UI/product/phase/release authority. 144-01B/02/Rule/Sieve/Phase145 blockers and Calendar 143-03E launcher+migration blocker unchanged. Its former no-further-source/root-only-selection routing clause is `CONSUMED/HISTORICAL`. Founder directs root to continue remaining Mail+Calendar atomic source packets before final integrated Docker; each exact packet requires its own test-first/static review/commit. No Ent generation/migration/runtime authority follows.

> **144-02A/B/C/D correction acceptance (2026-08-15):** canonical receipt `576f19231e37f635d5b7187a1ada121ed14ff8dc` (parent `0ef3df9eedfb3683fe70a7f8fc502ca76278facb`, tree `b28904b96c084dd88c4cec37c8e1daa62fde4c5d`) accepts the exact 14 content documents as docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. All candidate/precommit/root-commit/exact-review/status-sync-next clauses for this correction are `CONSUMED/HISTORICAL`. 02F normative acceptance remains unchanged except its dependency trace. No source/schema/generation/migration/service/test/runtime authority; launcher `NOT_READY`, executables `NOT_RUN`. The separate docs-only engine deletion-port contract/adjudication and durable User-deletion-intent/job/route contract remain `BLOCKED` / `NOT_DISPATCHABLE`.

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` remain accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; unrelated gates remain current. No source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; launcher `NOT_READY`, executables `NOT_RUN`.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier acceptance (2026-08-15):** Non-marker typed-IR/Sieve carrier contract `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, sole blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Its base/correction lifecycle is `CONSUMED/HISTORICAL`. No implementation-plan/source/test/runtime authority; root-only separate plan acceptance plus G145-5B/G145-5C/launcher readiness keep Rule Task 1/145-07B blocked. Launcher `NOT_READY`, executables `NOT_RUN`.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept/review/commit instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. The Rule-plan acceptance receipt is the documentation sync; it is not a future Rule Task 4. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

> **144-01B static acceptance (2026-08-15):** immutable `0c064363b5ccb657a7c444802e94075c4ca743e5` (parent `48a7273d100c07788def88e2f1b8d69c7173580f`, tree `9c568a0c58c6bf7b8c3171cfd934760b839db623`, subject `feat(mail): add deterministic engine fake`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for fake `12bcd921668e406f07a98f7f3b0618ad8d44e81d` and authored-first test `b3654103b3a6d96bd887a1e6277aeb0ed6449b3f`. Terra/xhigh author; Spec Terra/high, Security Terra/xhigh, ABI Terra/high precommit and exact-SHA Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 01B Docker-blocked/do-not-start/dependency clauses are `CONSUMED/HISTORICAL`. Launcher `NOT_READY`, executables `NOT_RUN`; no 02E/Ent/generation/migration/service/API/UI/product/phase/release authority.

> **144-02A2 static acceptance (2026-08-15):** `65c91b4824e8a1948fc97c149305aa7f009de061` (parent `5df6b47678364c438ade579252209c8fd238c43c`, tree `08583554939f597443566093f30b6ec7aa9373db`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for domain `4de2ec0882d94db595aaabf8d4117566118ee472`, policy `a70e05122ffa126e9754d029ca00493faae2aeef`, Organization inverse `ef6e2e65430a1397b0f563f1f361ecd89660a39c`, authored-first test `65de04719a60a38d7fc456553e092bb106456957`; Terra/xhigh author, Terra/high precommit+exact-SHA PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Static only: launcher `NOT_READY`, executables `NOT_RUN`, no generation/migration/runtime/service/API/UI/product/phase/release authority. A2-unimplemented/A3-next, 02B-waits-for-02F-R, 02F-R-root-dispatch and predecessor `02B requires separate docs authority` clauses are `CONSUMED/HISTORICAL`; current authority is 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 only, implementation unaccepted, and former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; other 02F incl 02F-D unauthorized. **02C source acceptance: preceding 02C route clauses are CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS; exact13 blobs IMPLEMENTATION-144-02C.md; no next Mail packet selected; named blockers remain.**

> **2026-08-15 — 144-02F-R:** `ce542cb5de839a1fb5200295f710b0ae45bcdbe6` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; parent `1792cdc381eeb4db0ab1ede5b512ba1222eb90b2`, tree `2cadf80c9ae0af99b5c8027c1d78c3ec55af7071`, exact eight blobs in `IMPLEMENTATION-144-02F-R.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Authority-only/source-unaccepted and 02B-separate-authority/02C-predecessor route `CONSUMED/HISTORICAL`; 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, implementation unaccepted, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL, 02F-D blocked; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`. **02C source acceptance: preceding 02C route clauses are CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS; exact13 blobs IMPLEMENTATION-144-02C.md; no next Mail packet selected; named blockers remain.**

## 144-02F pre-acceptance record — 2026-08-14 — CONSUMED/HISTORICAL

> **2026-08-15 — 143-04B static source receipt:** `75dafea889055126f768227a4ae62b28c91b7393` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` (parent `a877848ef4d1ce80a30a4389f97ef0f863615159`, tree `6b86570991e41e0c401934867a49a17849847cb3`); source `db34bbbc4b4c8a3ace8f768998060f82deecedbe`, test `79225fb73481ffd34ee1f3b9a72965d255e0b14a`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Every earlier 04B `PRECOMMIT_FAIL`/`NOT_ACCEPTED`/current-next clause is `CONSUMED/HISTORICAL`. Launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`; 03E remains blocked for Ent generation/migration and no next Calendar code packet is selected.

## 144-02F-R source-authority acceptance — 2026-08-15

`DISPATCH-144-02F-RECEIPT.md` is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` at
`a877848ef4d1ce80a30a4389f97ef0f863615159` (parent
`25cfb36720f9a35c96268a62f7ee81fd91f83503`, tree
`917fed11d1316678008229193aef6186ebaa0abb`). It preserves the accepted 02F normative contract
and authorizes only the exact eight-path, test-first static receipt/target/member source/test
packet for root dispatch, including the
required `NO ACTION` organization and nullable account relations, scalar-only actor/message IDs,
target/member cascades, and only the listed account/organization inverses. It does not amend
PERSISTENCE or OpenAPI and excludes rule deployment.

Independent Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh, and
Lifecycle/Dependency/Boundary Terra/high exact-SHA reviews PASS C/I/M `0/0/0`;
dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Candidate/precommit/commit/review-next
clauses are `CONSUMED/HISTORICAL`. This is not source implementation acceptance: the dispatched
packet still requires its own test-first authoring, review, atomic commit and exact-SHA acceptance.
This 02B-eligibility/02C-predecessor route is `CONSUMED/HISTORICAL`: 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, implementation is unaccepted, and former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; 02F-D is `BLOCKED` / `NOT_AUTHORIZED`, separate/nonblocking. **02C source acceptance:** preceding 02C route clauses are `CONSUMED/HISTORICAL`; `e565523f589fbd7536f6f8e4df7effd39469c1af` is `STATIC_ACCEPTED / EXACT_SHA_PASS`; exact13 blobs are in `IMPLEMENTATION-144-02C.md`; no next Mail packet is selected; named blockers remain.
Generation, migration, runtime, services, API,
UI, release, and launcher status (`NOT_READY`; executable evidence `NOT_RUN`) are unchanged.

Correction commit `fcb5b84fed4a417e736ddb4053256e494a0c25f6` (parent
`55e36b7a1a3516d07b2baf390a5742a24ec8fbe5`, tree `5a0637f2306ff42295adea5560793e2a8f4a294a`,
subject `docs(mail): correct receipt and rule invariants`) changed the exact same 28 docs and
the worktree was clean at review. It is `COMMITTED_UNACCEPTED_AT_FCB5B84F` /
`EXACT_SHA_TRANSITION_REQUIRED` / `IMPLEMENTATION_BLOCKED`, without a version bump. Contract/Spec/Oracle
Terra/high PASS C/I/M `0/0/0`; Security/Adversarial Terra/xhigh PASS `0/0/0`; Lifecycle/Dependency/Boundary
Terra/high FAIL `0/1/0` solely because committed precommit status/instructions became stale after commit,
while Git object/boundary/blockers/exclusions PASS. All routes dispatcher-enforced, fallback none,
telemetry `UNKNOWN`. The prior precommit/root-commit-next sequence is `CONSUMED/HISTORICAL`.
This entire pre-acceptance record is `CONSUMED/HISTORICAL`; it does not route action. The accepted
docs/readiness boundary is correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus transition
`2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260`, `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, after all three
Terra routes PASS C/I/M `0/0/0` under dispatcher-enforced/no-fallback provenance with telemetry
`UNKNOWN` unless exposed. The former Rule-plan acceptance instruction is `CONSUMED/HISTORICAL`; Rule
TDD remains `NOT_AUTHORIZED/BLOCKED` under the current receipt's stated prerequisites. Source/schema/
migration/runtime/release authority remains blocked.

## 144-02D docs-only STATIC_ACCEPTED — 2026-08-14

The exact nine-document content boundary `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb`
(parent `6c424ed2d6703ceccbb269c37d9abc03e75a0410`, tree
`da3fa6e19316bcf2c7c318a63ec09e1b175f68df`) plus same-nine-document lifecycle transition
`e563a3d3be8898371e5c8a9861c58f11b72167bc` (parent `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb`,
tree `b268efdf950609c6d3fa66efe188005d9a39cf40`) is narrowly docs-only `STATIC_ACCEPTED` /
`EXACT_SHA_PASS`. Post-transition Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh,
and Lifecycle/Dependency Terra/high exact reviews PASS C/I/M `0/0/0`; dispatcher-enforced,
fallback none, telemetry `UNKNOWN`. The earlier Lifecycle FAIL C/I/M `0/1/0` and contrary
pre-acceptance next-action text below are `CONSUMED/HISTORICAL`.

This is not source, Ent schema, generation, migration, service, runtime, phase, product or release
acceptance. Launcher remains `NOT_READY`; executable evidence `NOT_RUN`. 02F receipt/deployment
facts from `63e419897aa0683e1d3dfcd442aca78eea1b0727` remain unaccepted. The former direct-closure
instruction and its former Rule-plan acceptance step are `CONSUMED/HISTORICAL`. Rule TDD/source
remains `NOT_AUTHORIZED/BLOCKED` under the current receipt's stated prerequisites.

## 144-02D postcommit lifecycle transition candidate — 2026-08-14 — CONSUMED/HISTORICAL

At that historical point, corrective commit `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` (parent
`6c424ed2d6703ceccbb269c37d9abc03e75a0410`, tree
`da3fa6e19316bcf2c7c318a63ec09e1b175f68df`, exact nine documents) exists and is `COMMITTED_UNACCEPTED` /
`EXACT_SHA_CLOSURE_IN_PROGRESS`, not accepted. Contract/Spec/Oracle Terra/high and
Security/Adversarial Terra/xhigh postcommit exact-SHA reviews PASS C/I/M `0/0/0`;
Lifecycle/Dependency Terra/high exact-SHA review FAIL C/I/M `0/1/0` solely because active text
still said precommit/commit-next, while content/object boundary passed. Dispatcher-enforced,
fallback none, telemetry `UNKNOWN`. This transition record consumes that stale instruction and
does not claim its own commit SHA. Next: fresh independent Lifecycle exact-SHA re-review of `7d`
plus this transition record; fresh Contract/Security trace-only rechecks are permitted. After all
exact reviews PASS, a separate docs-only 02D `STATIC_ACCEPTED` sync is required; only then 02F
closure reviews. All implementation/runtime blockers remain.

## 144-02D RecentContact corrective candidate — 2026-08-14 — CONSUMED/HISTORICAL

Narrow docs-only corrective candidate, no version bump. It supersedes only the RecentContact
portion of initial candidate `63e419897aa0683e1d3dfcd442aca78eea1b0727`: the hard cap applies
only to null-`user_id` external rows, linked rows are exempt, and `user_id ON DELETE SET NULL`
is explicit. It freezes external account-only locking, user-first then sorted complete owned+linked
account-union locking for user lifecycle (so cascades acquire no omitted account lock), user-first
link/unlink, per-account raw-UUID LRU tie-breaking, safe exact-content display-name
validation, link/name privacy and public max-30 projection. The initial candidate and transition
`6c424ed2d6703ceccbb269c37d9abc03e75a0410` are `CONSUMED/HISTORICAL`; neither is accepted.
Receipt/deployment facts from `63e` remain frozen for later 02F closure, which waits on accepted
corrected 02D. Its precommit status/instruction is consumed by the transition record above.
At that historical point, commit `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` remained unaccepted. Launcher `NOT_READY`;
no source, schema, generation, migration, service or runtime authority.

## 144-02D/02F normative correction candidate — 2026-08-14 — CONSUMED/HISTORICAL

Historical initial narrow docs-only candidate, no version bump: corrects identity/signature derived-response versus
Upsert command semantics; freezes named 02D checks, receipt attempts/exhaustion and the exhaustive
operation-target-status-safe-code predicate; adds the internal-only deployment exhaustion terminal
with a bounded ten-slot attempted-tuple budget; and retains flat rule conjunction wording. Its
status `ROOT_CORRECTION_CANDIDATE` / `INDEPENDENT_CLOSURE_PENDING` / `IMPLEMENTATION_BLOCKED` /
`RUNTIME_NOT_RUN` is `CONSUMED/HISTORICAL` and superseded above. Author provenance: `gpt-5.6-terra`/medium,
dispatcher-enforced, fallback none, telemetry `UNKNOWN`. Launcher `NOT_READY`; no
schema/generation/migration/product/release authorization.

> **Current 144-03B authority (2026-08-14):** earlier Folder lifecycle-gate wording below is
> `CONSUMED/HISTORICAL`. Source/oracle commit `f8cb7643986fc591cae16764575337b72ef3a692`
> (parent `32e3dcb5b4f7ae65a88b79781646d25726ae16c6`, tree
> `14e7f6ce9a0f4b7912b955515d122c411159c579`) is narrowly `STATIC_ACCEPTED`, static-only,
> with source blob `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b` and oracle blob
> `efe0db0071cb22ce4b026c635157f0fb08885d64`, after three `EXACT_SHA_PASS` reviews C/I/M
> `0/0/0`. Folder docs are also static-only accepted at `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
> The former `f1995465`/`6735ae74` Rule transition is `CONSUMED/HISTORICAL` and superseded by the
> current `669ef761` sole-plan receipt. It grants no Rule implementation authority. The broad 02D/02F
> candidate/transition and that pre-acceptance sequence are historical. Corrected `7d29d08c` plus
> transition `e563a3d3` plus docs-only acceptance sync `55e36b7a` is narrowly `STATIC_ACCEPTED` /
> `EXACT_SHA_PASS`; the former direct 02F-closure and transition route are `CONSUMED/HISTORICAL`.
> The former Rule-plan author/adjudication instruction is `CONSUMED/HISTORICAL`. Rule TDD/production
> remains `NOT_AUTHORIZED/BLOCKED` under the current receipt's stated prerequisites. Runner `NOT_READY`; all
> executable evidence `NOT_RUN`.

## 144-02A/B/C/D root correction candidate with 02F dependency/status trace — 2026-08-15

> **Current scope precedence:** combined 144-02A/B/C/D correction plus 02F dependency/status trace; exact 14-document route.

Accepted docs/readiness correction freezes the current 02B/02C field tables and carries the correction-scoped 02A/02C/02D lifecycle facts stated in the current banner above. 02F receipt/deployment normative acceptance is unchanged; its broad 02D-unchanged/dependency/writer-replay clause is dependency-traced. Status is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc` after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`; candidate/precommit/root-commit/exact-review/status-sync-next clauses are `CONSUMED/HISTORICAL`. No source/schema/generation/migration/service/test/runtime authority, launcher `NOT_READY`, executables `NOT_RUN`.

## 144-02A representation acceptance receipt — 2026-08-14

The frozen A1/A2/A3 persistence representation at reviewed snapshot
`b631866b2ba372101fd80eea30951eaa267c1931` is narrowly `STATIC_CONTRACT_READY` /
docs-only accepted. Contract/Spec `gpt-5.6-terra/high`, Security/Adversarial
`gpt-5.6-terra/xhigh` and Lifecycle/Dependency `gpt-5.6-terra/high` independently PASSed
C/I/M `0/0/0`; each route is dispatcher-enforced, fallback none, telemetry `UNKNOWN` unless
exposed. This receipt is not an implementation commit and makes no version bump. It accepts
only frozen persistence representation, not Ent/product source or TDD, generated output,
migrations, launcher execution, tests/validators/build/runtime, 02B/C/D/F, services, product,
module, phase or release. Founder migration authority, supported task-owned launcher,
sequential explicit source dispatch and source-level review remain required. G145-5C remains
open. That broad candidate action and pre-acceptance sequence are historical. 144-02D is now
docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; the former direct 02F-closure and transition route are
`CONSUMED/HISTORICAL`; 02F `fcb5b84f` + `2f4381a2` is already docs/readiness-only `STATIC_ACCEPTED` /
`EXACT_SHA_PASS`. The former Rule-plan author/adjudication instruction is `CONSUMED/HISTORICAL`; Rule
TDD/source remains `NOT_AUTHORIZED/BLOCKED` under the current receipt's stated prerequisites, not code.

## 0.1.8 rule compiler candidate — 2026-08-14 — CONSUMED/HISTORICAL

`144-03B-Rule-Compiler` is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
`f199546513b758ca7b95328b5ccd7680009b732e`, not product, module, phase or release acceptance. It freezes
provider-neutral typed IR, never raw Sieve or hash v1. The keyword is `hubk-` plus canonical
lower-case unpadded RFC-4648 base32 of exactly 32 random bytes; validation decode-to-32 then
identical re-encode, not regex alone. 144-02D is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`;
the former direct 02F-closure and transition route are `CONSUMED/HISTORICAL`; 02F `fcb5b84f` +
`2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. The former separate Rule-plan
prerequisite is superseded by the current `669ef761` sole-plan receipt; Rule TDD/production remains
`NOT_AUTHORIZED/BLOCKED` under its stated carrier, G145-5B, G145-5C and launcher-readiness prerequisites;
no Rule source, schema, runtime, product or release authority follows from this docs-only acceptance.

## 144-03B Folder lifecycle sync candidate — 2026-08-14

> **CONSUMED / HISTORICAL PRE-SOURCE:** this preserved entry is superseded by the current 144-03B authority banner above.

Contract `a565af2632bc5f76a7d2fe26ebf3065fb3cefba9` (blob
`0f9ed0adb0cec6ffa83619cb4c7c4cb4d3f3263c`) is narrowly `STATIC_ACCEPTED` after three
independent exact-SHA reviews PASS, C/I/M `0/0/0`. Plan
`a98c7235521852dee1ebc0d09558c7bcbefc4534` (blob
`c79658d61e496489be697bc971d0797ed30fb61f`) is narrowly `EXACT_SHA_PASS` / `STATIC_ACCEPTED`
after exact, security and boundary reviews PASS, C/I/M `0/0/0`. This records the pure bounded
Folder ABI/error/normalization/oracle contract only; it adds no source, test, schema, route
handler, migration, engine or runtime acceptance. The docs-only lifecycle sync itself awaits
three exact-SHA reviews; after them, Task 1 only may create the external RED oracle. Production
source remains forbidden until Task 1 oracle reviews PASS. Runner is `NOT_READY`; executable
evidence `NOT_RUN`; provenance dispatcher-enforced Terra/medium author, Terra/high + Terra/xhigh
+ Terra/high reviews, no fallback, telemetry `UNKNOWN`.

> **Current review status (2026-08-14): five-document 0.1.8 entry amendment PASS at `5e314fb080d420696642058ae47397e251c23207`; Thread contract remains accepted at `cbc12577a6444df9c660e43fedfd0c8133c6e461`; implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS` (source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`), static only; broader Phase 142/cross-contract review remains pending.**

## Candidate lifecycle sync — 2026-08-14

The historical preimplementation entry below is preserved. The exact 19-path boundary is now
narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`; no schemas, routes, migrations,
services, generation or executable evidence are added. Launcher remains `NOT_READY`, runtime
`NOT_RUN`; the accepted contract and plan supersede the prior freeze/adjudication next step.
> The root amendment closes the recorded engine-memory, public-response and Phase 144
> persistence representation gaps. Its independent PASS authorizes only 144-01A entry, not
> product code beyond that packet, generated Ent output, migration or runtime claim. Executable evidence remains
> `NOT_RUN`; the accepted launcher remains `NOT_READY`.

## 0.1.8 preimplementation amendment — 2026-08-14

`144-03A-Thread` callable/encoding/oracle contract is statically accepted at
`cbc12577a6444df9c660e43fedfd0c8133c6e461` (parent `4071ad8120850a8bde2d45d56fdedb5779270047`,
tree `2ee25fa08d343201ee72f80070c3fb58346f223c`, subject `docs(mail): freeze deterministic thread identity`)
with `EXACT_SHA_PASS`: independent Spec/API/Persistence, Security and Lifecycle/Evidence reviews
PASS, C/I/M `0/0/0`. This is docs contract acceptance, not implementation acceptance. It is
separate from the `5e314fb0` five-document 144-01A entry-only amendment and preserves Address
`STATIC_ACCEPTED` `41564d00`. It freezes strict
byte-preserved case-sensitive Message-ID literals/equality and same-account live resolver
provenance, V1 versioned raw key framing and literal SHA oracles, seed-only 1..401 participant
snapshots, 1 MiB typed JSON bounds, and closed public email-only participants. Any x/text
change requires separately accepted version/rekey/dual-read planning.
That packet is now narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`. Its former
Folder freeze/adjudication next step is historical. No 0.1.8 version bump,
schema/generation/migration or runtime acceptance exists.

## 0.1.8 — 2026-08-13

Bounded engine/API and deterministic mailbox persistence amendment. The exact five-document
entry amendment independently PASSed at `5e314fb080d420696642058ae47397e251c23207`, which
authorizes only 144-01A entry; it is not implementation or runtime acceptance and does not
close broader Phase 142/cross-contract review.

- Statically accepted the root-frozen 143-04L `EnvelopeV1` subject amendment: the new builder
  has a 255 UTF-8-byte Hub product/safety bound distinct from the NATS transport/control-line
  ceiling, with the corresponding schema extension and static contract oracles. Independent
  Terra/high review PASS is C/I/M `0/0/0`; 143-04L is `STATIC_ACCEPTED` at
  `4525c559da22729aa8bc51ff5336c29dad4f27e9`, and status-sync commit
  `1c0299bd720ad6f8521314d83e478dedc666eccd` records it. The launcher remains `NOT_READY`,
  runtime remains `NOT_RUN`, and this records no product acceptance.
- Closed the post-freeze machine-contract review gaps: all public HTTP objects are closed;
  summary/detail composition uses internal field bases plus closed wrappers; every non-admin
  operation declares one default-deny `x-hub-authz` policy; mail received/sent addresses and
  calendar invite attendees use their canonical 200-item ceilings, and channel-due events
  carry a structured subject/payload UUID equality invariant. Added standard
  cursor/recovery/idempotent-replay examples, bounded the still-blocked legacy broadcast
  compatibility shape, and made mailbox policy's maximum 300 recipients reachable from the
  three independently capped 100-item draft collections.
  The narrow five-document entry review is accepted at `5e314fb080d420696642058ae47397e251c23207`
  only for 144-01A entry. Broader cross-contract rereview and executable Docker validation remain pending.
- Split list/change/enumeration from hydrated message content. Projection pages now have
  independent count, per-item and aggregate retained-byte ceilings with an exact
  representation-independent variable-byte measure; authorized message/blob reads and draft
  save/reconciliation carry caller budgets and reject whole values on overflow with one safe
  error.
- Closed every request/page/projection/body/blob/draft/submission DTO field. Projection
  headers and attachment metadata are fixed rather than map-shaped, draft attachments carry
  stable keys and bounded bytes, and submission is reference-only with a final-MIME ceiling.
- Added a stable Hub-derived engine draft key for lost initial-create reconciliation and a
  required one-draft `draft_client_id` carried by draft API/storage rows; attachment reuse is
  scalar one-draft ownership rather than an ambiguous many-to-many edge.
- Froze MailboxPolicy creation defaults at 25,000,000 decimal bytes for both message and
  attachment, separately from the engine adapter hard caps of 52,428,800 and 26,214,400
  bytes. Administrator configuration remains bounded by the respective cap and enforces
  attachment <= message.
- Completed OpenAPI list pagination, nested collection/string bounds, 8 MiB cursor-page
  budgets, compact calendar summaries, 366-day ranges and non-truncating 16 MiB legacy
  calendar/call envelopes. Creation quotas now have replay-stable terminal errors.
- Froze exact Phase 144-02A/B/C/D/F storage names, tuple/nullability checks, state matrices,
  normalized-name indexes, delete behavior, quotas and transactional same-owner guards in
  `PERSISTENCE-CONTRACT-0.1.8.md`.
- Required every `MailMessage` projection to persist the engine version supplied with its
  opaque engine id; optional auxiliary state is no longer allowed to stand in for version.
- Split confirmed submission identity into opaque engine submission/result-message ids and
  an immutable Hub message UUID snapshot. Public `sent` now requires a bounded result read
  and atomic projection upsert. Retained-pair recovery retries only exact account/message
  hydration and finalization, derives confirmed submission origin in the domain, and can
  never requeue or resubmit.
- Enforced the 50-attachment maximum both on the public draft value and on the locked durable
  draft-to-attachment association, so sequential saves cannot manufacture an over-limit row.
- Made settings pointers the sole persisted identity/signature default authority, made
  attachment scan state single-source, bounded recent external contacts at 1,000, and added
  immutable attempted tuples for account/domain/rule engine work.
- Froze the generic mutation receipt operation/status/code grammar, bulk target/member
  matrices and canonical rule-set hash v1. No raw key, body, MIME, provider error, secret or
  Sieve source enters PostgreSQL.
- Historical 0.1.7 CVE-v1 semantics remain included; its earlier independent PASS does not
  review these 0.1.8 changes.

## 0.1.7 — 2026-08-13

Canonical inbound iTIP identity amendment. Product runtime remains `NOT_RUN`; 143-01 and
143-02A source acceptance is unchanged. Root and independent Terra/high static review PASS
on exact semantic SHA `b3fb6f48` with no remaining P0/P1/P2, which authorized 143-02B at
the time of this contract amendment. Current implementation status belongs to `STATE.md`
and the phase evidence rather than this historical entry. Review of
initial SHA `7a6ffc1c` found three P1 ambiguities; escape, UTF-8-byte-limit and status-sync
remediation closed them. The 0.1.6 review on `5a351d5b` is historical `STALE`.

- Separated inbound duplicate identity from the accepted outbound serializer. Inbound
  occurrence REQUEST is valid without expanding or weakening the 143-02A outbound subset.
- Froze CVE-v1 as an explicitly tagged, length-delimited, big-endian canonical value stream
  domain-separated by `hub.calendar.itip.payload`, NUL and version byte 1, streamed into
  SHA-256 without retaining another unbounded payload.
- Closed every included field, enum code, sort rule, RRULE normalization, physical-format
  exclusion and unsupported property/parameter rejection. No JSON, map order, locale,
  clock, random value, raw MIME or wire folding participates.
- Closed RFC TEXT and RFC 6868 decoding to explicit allowlists, rejecting terminal/unknown
  escapes and decoded controls except the contracted LF in text/CN. All decoded field
  maxima are post-normalization UTF-8 byte counts, including multibyte boundaries.
- Fixed bounded parser maxima/defaults: 1 MiB total, 8 KiB unfolded line, 4,096 properties,
  200 attendees, 1,000 EXDATE values and exactly one `VCALENDAR` -> `VEVENT` component path.
- Paired every persisted canonical event/exception/object-state/REPLY-receipt hash with its
  version and forbade cross-version equality. The sender-bound REPLY receipt hash now has a
  separate domain prefix and includes that version.
- Chose the unattended fail-closed default for all-day `VALUE=DATE`: canonical timezone
  marker `UTC`, date/exclusive-end preserved, no inferred midnight. Founder question 27
  records the possible trusted-context alternative as a future hash-version amendment.
- Added TC-CALSYNC-039..041 and synchronized Phase 143/147/148 and QA status. OpenAPI only
  receives the contract-version bump; no HTTP request/response shape changes.
- No precomputed digest from ad-hoc host research is accepted as an oracle. The literal
  canonical byte-vector oracle remains executable `NOT_RUN` until the approved launcher is
  available; static review can verify the grammar and test construction only.

## 0.1.6 — 2026-08-13

Mailbox control-plane, safe-content and mutation-recovery amendment. Product code/runtime
remain `NOT_IMPLEMENTED` / `NOT_RUN`; 0.1.5 is historical `STALE`. Root and independent
Terra/high static reviews PASS on exact semantic SHA `5a351d5b` with no P0/P1/P2.

- Added an explicit bounded provisioning service/worker: deterministic immutable primary
  address/key/cutoff reservation, idempotent `EnsureAccount`, exact-lease activation,
  read-only safe `/mail/account` state and no request-path engine creation.
- Made Hub rules canonical desired state with a generation/hash deployment row and separate
  mutation-leased worker. One fixed Hub Sieve script is reconciled through Get-before-retry;
  rule HTTP responses expose honest pending/applying/ready/unknown/error state and no longer
  return 423 for synchronous engine work they do not perform.
- Added explicit spam/ham classifier feedback to the engine port. Bulk requests now preserve
  typed message/thread targets, include label actions, report classifier outcome separately
  from move/flag state, and use one stable mutation receipt with normalized expansion/outcome
  children. Replay never re-expands a changed thread or retries indeterminate training.
- Added generic hashed receipts for folder/label/signature/rule/presign/admin-domain/alias
  creates and message bulk intents. Full replay metadata lasts 30 days and a compact
  key/request tombstone remains for the owning scope lifetime; raw keys/response/content do
  not persist. Generic replay-unavailable/expired conflicts are distinct from specialized
  draft/send/legacy-call receipts; presign replay refreshes only the same pending object URL.
  Generic intents use v4/v7 UUID keys and a raw-keyed canonical request HMAC so small
  payloads cannot be enumerated from database fingerprints.
- Reused the receipt schema in an explicit non-wire `internal` mode for versioned folder
  update/delete recovery. Its injected UUID key is created only after lease/optimistic checks,
  is never accepted from HTTP, and may be removed after the terminal recovery window.
- Froze server-derived plain content, reviewed-sanitizer-or-plain-only behavior, received
  remote-image token rewriting and owner/message-bound proxy checks on every redirect/
  connect. The proxy returns bounded image bytes with fixed private/no-referrer/nosniff/CSP
  headers and never redirects or returns the origin URL.
- Constrained the public remote-image path token to unpadded base64url 32..1024 so traversal,
  escaping, padding and malformed bytes fail before decode, authorization or network work.
- Made label, user-settings, signature, rule-delete, mailbox-policy and domain wire writes
  match their versioned models through `If-Match`; active-rule label deletion now has a
  machine-readable no-mutation conflict.
- Replaced generic new-mail 409s with route-specific header codes for version/name/in-use,
  policy prerequisites, domain invariants and alias state. Global domain/address collisions
  disclose only `*_UNAVAILABLE`, never another tenant's owner/resource.
- Replaced the remaining generic calendar/legacy writer 409s with exact event-version,
  event-not-mutable, occurrence-not-mutable and ScheduledCall-not-mutable codes. Stale v2
  writes expose only the current quoted event ETag and same-origin canonical event Location;
  every conflict retains NFR-13 and mutates nothing.
- Required `If-Match` for draft delete and send as well as save. Same-key send replay is
  resolved first; new stale/not-editable/competing transitions have exact no-mutation codes.
- Added machine-readable attachment confirm/expiry/non-clean conflicts and constrained clean
  download redirects to fixed HTTPS storage with required `no-store`/`no-referrer`.
- Closed new mailbox request objects against unknown properties and made rule conditions/
  actions exact typed unions, preventing client scope ids or ambiguous targets on the wire.
- Made bulk move/label/thread-pin/no-argument actions mutually exclusive and forbade every
  irrelevant folder/label target field instead of silently ignoring it.
- Closed public submission/rule/domain/bulk error codes to explicit allowlists so provider
  text cannot cross persistence, API, audit or client analytics boundaries.
- Made submission message/error fields required-nullable with state-conditional invariants;
  pending cannot masquerade as sent and failed/unknown cannot carry a message id.
- Split enriched response addresses from draft recipient input so client-supplied `user_id`
  is rejected and cannot become identity or authorization input.
- Removed the duplicate message-label replace wire; receipt-backed bulk add/remove now owns
  both single and multi-target label provenance/recovery.
- Added an atomic route-metadata redaction packet because current generic access/audit paths
  would otherwise expose a remote-image token; ingress must independently redact token paths
  and mail-search queries.
- Made folders versioned desired intents with owner-visible pending/unknown/delete states and
  bounded receipt recovery inside the existing projection worker, preventing duplicate
  create or overwrite after a lost engine response without adding a fifth worker.
- Made exact-match/absence folder recovery fail closed behind a reviewed Hub-exclusive
  custom-folder-writer capability. Wave 1 exposes only Hub-internal JMAP/admin plus optional
  SMTP delivery; user IMAP/JMAP/ManageSieve/direct-admin folder mutation stays unpublished.
- Froze Meilisearch hydration for bounded sanitized plain body plus normalized attachment
  filenames, with no raw HTML/MIME/BCC/remote URL and a second current-owner authorization
  before any id/snippet response.
- Added missing folder/label management, remote-image consent, rule-deployment, provisioning
  and admin-alias UI packets. Added founder questions 24–26 for address policy, sanitizer
  dependency and fixed-script ownership.
- Bumped the OpenAPI and routed phase/QA contracts to 0.1.6; no event payload changed.
- Recorded the founder's task-local implementation-route amendment: Mail atomic code uses
  exact `gpt-5.6-terra` / `xhigh`, no fallback; Docker verification remains separately
  gated by the canonical launcher. The narrow exception is also recorded in mandatory
  first-read `AGENTS.md`, while all non-Mail routing remains unchanged.

## 0.1.5 — 2026-08-13

ScheduledCall canonical-cutover and mailbox-projection amendment. Product code and runtime remain
`NOT_IMPLEMENTED` / `NOT_RUN`; the 0.1.4 review is historical `STALE` for this revision and
the current 0.1.6 semantic review PASS is recorded separately on `5a351d5b`.

- Made `CalendarEvent` the only business authority for linked meeting content, time,
  recurrence/exceptions, reminder, status and attendees; `ScheduledCall` is a distinct
  execution/legacy projection with a next effective occurrence.
- Froze atomic legacy create/update/delete, calendar-v2 attach/edit, event-before-call lock
  order, occurrence-cancel projection advancement and occurrence-keyed durable call-due
  outbox behavior on `calls.channel.{channelId}.scheduled_due.v1`. Legacy PUT keeps
  serialized last-write-wins because its wire has no `If-Match`; the successful response
  shape stays compatible.
- Closed reader duplication: legacy aggregate shows a linked pair once as `call`, while v2
  availability shows it once as `event` and loads active unlinked calls only before cutover.
  The backfill now includes paused/completed non-deleted rows so legacy resume cannot restore
  an unlinked writer; any unlinked non-deleted row after activation is an invariant failure.
  New meeting conflict UX must use private v2 availability; legacy conflict detail requires
  membership in the exact conflicting channel.
- Expanded ScheduledCall backfill to organization/project/channel coherence, description,
  duration, reminder, recurrence, participant identity, audit timestamps and deterministic
  hashes. Recurring rows block apply until founder-selected UTC-versus-local semantics and
  bounded occurrence-oracle evidence agree; partial migration and guessing are forbidden.
- Made event `scheduled_call_id` an immutable creation-only, server-authorized link in
  OpenAPI. Attach requires exact normalized equality with the locked call, and the safe
  default rejects external guests until a join-credential delivery contract exists. Series
  attendee state alone drives the internal participant projection; occurrence RSVP cannot.
- Froze mandatory legacy-create idempotency as an explicit compatibility decision, strict
  conflict UUID parsing, exact-channel/project membership and canonical server-side
  occurrence delivery to the browser. Added founder questions 18–23 for recurrence,
  pause/resume, nested call creation, overdue recovery, idempotency compatibility and linked
  external guests.
- Closed the outbox/envelope ambiguity: a shared deterministic builder creates the complete
  versioned envelope, new producers store it unchanged through the existing `PublishInTx`
  seam with `event_type == subject`, and the global worker/legacy payload contract remains
  untouched. Envelope `event_id` is the consumer dedupe id; the outbox row
  `Hub-Event-Id` remains a distinct transport diagnostic id. Phase 144 event mapping now
  precedes transactional producers instead of depending on them.
- Closed the L0 mailbox ingestion/recovery gap: one leased bounded worker owns hydrated
  incremental pages, separate command-lease begin/end epoch fencing, atomic projection/
  event/cursor commits and a
  no-clear resumable base-anchored enumeration/catch-up rebuild. First-seen post-cutoff
  inbound mail emits one stable received event; history and only stably mapped Hub
  submission/draft copies do not. From/alias or mutable Sent membership never proves owner
  origin. Phase 145 now owns incremental and base-state/ID-anchored enumeration adapters
  plus root-context worker composition, while OpenAPI/UI expose bootstrapping/degraded/
  rebuilding freshness without false-ready state.
- Closed the command/sync overlap race: sync cannot capture while an account command lease
  is present; command completion advances a second epoch edge, expired-lease reclaim also
  advances it, and a late former owner must reconcile instead of finalizing stale state.
- Froze the 100-folder total account cap with engine-atomic enforcement, a machine-readable
  422 response and fail-closed degraded projection for an already over-cap engine account.

## 0.1.4 — 2026-08-13

Reminder/CRM race-closure amendment. Product code and runtime remain
`NOT_IMPLEMENTED` / `NOT_RUN`; the prior 0.1.3 root review stays historical `STALE`, while
the 0.1.4 root static review passed on exact semantic SHA `46cfa26c`. Independent review is
still pending.

- Defined timed and all-day reminder due derivation, including deterministic midnight
  overlap/gap/skipped-date behavior, and made both scheduler emission and notification
  finalization recheck current sequence, effective start, non-null offset, derived due and
  owner/attendee eligibility under one lock order. A stable user-row guard also serializes
  quiet-hours and absent/existing preference changes.
- Defined concurrent mutation semantics: mutation-first durably suppresses stale work;
  consumer-first preserves only the already committed valid row. Transient reads/writes
  remain pending/retriable. Telegram/Web Push are post-commit-only best effort, with no
  attempt on rollback and no authority to change the canonical outcome.
- Invalid timezone or incomplete/malformed quiet hours now suppresses with a safe bounded
  reason/signal instead of the legacy UTC/delivery fail-open behavior.
- Added the missing atomic UI packet and QA case for manually linking an existing owned
  inbound/outbound message from the ordinary reader, showing only current-user owned deals
  while preserving independent server authorization.
- Tightened RSVP conflict `Location` to a same-origin canonical UUID event path.
- Bumped OpenAPI metadata to 0.1.4; endpoint and event payload shapes are unchanged.

## 0.1.3 — 2026-08-13

Phase 147 reliability/integration addendum. Product code and runtime remain
`NOT_IMPLEMENTED` / `NOT_RUN`.

- Added durable owner/UID/occurrence ordering and per-message processing receipts for
  authenticated incoming REQUEST/CANCEL, including cancel-before-request tombstones and
  missing-master occurrence recovery. CANCEL tombstones persist for mailbox/account life
  under a hard cap; only deferred missing-master state expires.
- Added immutable external-organizer REPLY delivery reconciliation through stable mailbox
  draft/submission keys and froze its conservative public status mapping.
- Froze privacy-preserving availability: all-or-nothing collaboration-scoped authorization
  for self or same-organization shared-project-or-channel users, bounded event/call/
  policy-visible approved-leave intervals and free complements without private detail.
- Added durable reminder intents and `calendar.reminder.due.v1`; transient policy/store
  failures remain pending/retriable, in-app is persistently deduplicated, and post-commit
  WebSocket/Telegram/push remain explicitly best-effort while Resend stays unchanged.
- Added the CRM compose-context and idempotent manual deal/message link API. `Write email`
  creates one ordinary Mail draft and navigates by its opaque UUID without putting the
  recipient in URL/history. The normal mailbox remains the only send path; no body,
  campaign or tracking data enters CRM.
- Prepared but did not activate the L1 feed contract pending founder question 5; L2/L3
  remain deferred.
- Added exact Phase 147 Spark dispatch packets and named QA scenarios 015–025 plus CRM
  cases. No code packet is eligible while the locked runtime and launcher gates are absent.

## 0.1.2 — 2026-08-13

Implementation-readiness addendum. No runtime or product-code claim.

- Added explicit PostgreSQL metadata models for message-folder projection, drafts, user
  settings, signatures, domains, mailbox policy and calendar mutation receipts.
- Added a mail-owned object-storage authorization/lifecycle model; project-scoped file
  metadata cannot authorize mailbox attachments.
- Froze draft save recovery through opaque engine versions and `ReconcileDraft`; message
  bodies, recipients, subject and raw MIME remain engine-only canonical data.
- Froze submission request hashing, stable engine keys, draft freezing, explicit
  `ReconcileSubmission`, `unknown` behavior and owner-scoped submission lookup.
- Removed the raw idempotency key from `MailSubmission` API responses.
- Added `GET /api/mail/submissions/{submissionId}` and explicit 409 idempotency responses
  for draft create/send.
- Added canonical deletion of one recurrence instance with a cancelled exception and a safe
  occurrence-scoped cancellation-intent envelope; raw iTIP/transport is outside this boundary.
- Split HTTP occurrence RSVP from inbound iTIP REPLY. Inbound replies now use a
  bounded durable per-attendee/sequence/canonical-payload-hash ledger and never an HTTP
  receipt; this prevents delayed `A -> B -> A` redelivery from rolling RSVP backward.
- Made message folder ids and engine-authoritative flag snapshots required in message
  summaries.
- Required `calendar.>` admission in Phase 143 and `mail.>` admission in Phase 144 while
  preserving the outbound webhook filter.
- Added exact Spark dispatch indexes and gate plans for Phases 143–146.

Unchanged compatibility boundaries:

- mandatory NFR-13 error body `{error,request_id}`;
- legacy `GET /api/calendar/events` response `{events,from,to}`;
- canonical calendar CRUD/RSVP under `/api/calendar/v2/**`;
- legacy flat Resend `/api/admin/mail/settings` contract;
- deprecated broadcast remains a documented unresolved governance gap, not new scope.

## 0.1.1 — 2026-08-13

Post-review hardening of the initial frozen contract: legacy calendar compatibility,
canonical v2 calendar paths, timed/all-day validation, same-origin RSVP recovery headers,
persisted calendar idempotency receipts, admin RBAC/audit metadata and explicit legacy
Resend/broadcast boundaries.

## 144-02C authority-bridge candidate — 2026-08-20

The accepted 02B source (`947e2a27650ff1c6db60af9ce13bb33b6bf19cb0`) and docs receipt (`b93c77ae771c6509be54fca8817fbca7617d6853`) consume/historicize only the former clause that 02C waits for accepted 02B source. This changelog now records `CONSUMED/HISTORICAL candidate lifecycle`, not source authority: independent Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency/Boundary Terra/high precommit reviews; one exact eight-document authority commit; the same three immutable exact-child-SHA reviews; then a separate docs-only source-authority acceptance sync. Only then may root dispatch the frozen test-first exact13 source manifest. No source dispatch or acceptance, Ent generation, migration, service, API/UI, runtime, product, phase or release authority is granted. 02F-D/02E, engine deletion-port, physical User/GDPR deletion intent, launcher `NOT_READY` and executable/runtime `NOT_RUN` remain unchanged. **02C source acceptance: preceding 02C route clauses are CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS; exact13 blobs IMPLEMENTATION-144-02C.md; no next Mail packet selected; named blockers remain.**

## 144-04F sync-page policy source-authority receipt — 2026-08-20

Candidate `4e608e890ec12aeed8de965f9649ae503260b94e` (parent
`c53a5fe60e1283cab58be5efaec7ffa131145b31`, tree
`ed7e2bdab6a44254a7c677603ad0afc298aafed8`, subject `docs(mail): propose sync policy authority`)
is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` for immutable exact9 tree blobs; Contract,
Security and Lifecycle exact-SHA reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/
telemetry `UNKNOWN`. The candidate/precommit/commit-review/separate-acceptance-next lifecycle is
`CONSUMED/HISTORICAL` for that receipt. Its former immediate exact4 root-dispatch clause is now
`SUSPENDED / CONSUMED/HISTORICAL` for the cutoff correction route.
**ROOT_CORRECTION_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN:** the following remains only the future immutable test-first source manifest:
`backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`,
`backend/internal/mail/mailbox_events.go`, then `backend/internal/mail/mailbox_sync_policy.go`.
The only event-builder change is the three existing received/changed folder-id and changed label-id
caps `50 -> 100` (100 pass/101 reject), with no field, subject or version change; all new policy
types stay in `mailbox_sync_policy.go`. The exact callable/field ABI, error sentinels/unions and
pure state-only classification contract are frozen in `REVIEW-144-04F-SYNC-POLICY-READINESS.md`
and mirrored in `MAIL-SPEC.md`. The now-consumed lifecycle required three independent precommit PASS
reviews and one atomic exact-nine candidate commit. The former permission for immediate root
dispatch is `SUSPENDED / CONSUMED/HISTORICAL`. Existing uncommitted exact4 source/test candidates
remain `UNACCEPTED` and MUST NOT be committed until new Contract/Security/Lifecycle precommit PASS,
one atomic exact9 correction-candidate commit, all three immutable exact-SHA PASS, and a separate
docs-only acceptance sync. A first-seen inbound/unknown projection requires nonnil `ReceivedAt`;
strictly after `EventCutoff` it requires nonzero `ReceivedEventID`, sets that marker and emits
exactly one `mail.received.v1`, while before/equal accepts only a zero reservation, upserts a nil
marker and emits none; historical/equal with a reservation is invalid. Stable confirmed
submission/draft never emits received. This correction is not accepted.
Executables/runtime are
`NOT_RUN`, and the previously committed 144-04F authority acceptance sync is not a new acceptance claim. OpenAPI/persistence and 02E/Ent/generation/
migration/services/API/UI/product/phase/release remain out of scope, as do 02F-D, engine
deletion/User-GDPR. Launcher remains `NOT_READY`; executables/runtime `NOT_RUN`.

## 143-04J call-projection authority candidate — 2026-08-21

Exact13 docs candidate only: `REVIEW-143-04J-CALL-PROJECTION-AUTHORITY.md` freezes the
future pure `ProjectEventCall` ABI and test-first exact2. It binds canonical Hub event, target,
creator and series attendee identities; exposes only an opaque copied snapshot; defines UTC
whole-second inclusive/strict-after recurrence boundary and safe no-I/O behavior. It changes no
event schema, persistence, OpenAPI, outbox, NATS, adapter, migration, API/UI or runtime contract.
The candidate is not source authority until one atomic exact13 docs commit, the three independent
immutable exact-SHA PASS reviews and a separate docs-only acceptance. AQ-02 remains backfill-only;
launcher/executable evidence remains `NOT_READY` / `NOT_RUN`.

## 143-04C calendar update/master-cancel immutable source receipt — 2026-08-20

Calendar 143-04C is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `7fd061b5099b4608e0c33a6d99edbe1a66587d34` (parent `bb84a47e12516c19682ff08783a5844306fc5766`, tree `0fb65fc72705bb145a1f2950c6efbe16607e6a04`, subject `feat(calendar): add update and cancel service`) for its test-first exact3 blobs: `event_service_update_test.go` `4421204141719fa1af586b3d6272aa130f185af2`, `event_repository.go` `737207280cc23adb873ae8d36952a3e62a0a3a4d`, then `event_service_update.go` `094b68dd8369bdc9345d9348016b3bdc377aa7a9`; `IMPLEMENTATION-143-04C.md` is the complete receipt. Contract/Spec/Oracle, Security and Lifecycle precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; author Terra/xhigh, dispatcher-enforced/no fallback, telemetry `UNKNOWN` unless exposed.

All preceding 143-04C candidate, `ROOT_SOURCE_AUTHORITY_CANDIDATE`, source-unaccepted, precommit and current-next clauses are `CONSUMED/HISTORICAL` only for this receipt. It accepts only the three committed source/test blobs and their safe update/cancel notification-intent contract; no next Calendar packet, adapter, Ent generation/migration, API/UI, raw iTIP, transport delivery or runtime authority follows. Launcher remains `NOT_READY`; executables/runtime remain `NOT_RUN`.

## 0.1.0 — 2026-08-13

Initial mail engine, calendar synchronization, OpenAPI, event schema, UI and QA contract.
> **Local 01B clause reconciliation:** older 144-01B Docker-blocked/do-not-start/dependency wording in this historical 02F sync record is `CONSUMED/HISTORICAL` only. It does not alter 02E/Ent/generation/migration/service/API/UI/product/phase/release gates.

## 144-02F acceptance sync — 2026-08-14 — STATIC_ACCEPTED / EXACT_SHA_PASS

Correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact 28-doc transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only accepted after independent Terra Contract/Spec/Oracle high, Security/Adversarial xhigh and Lifecycle/Dependency/Boundary high PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Former 02F unaccepted/transition instructions and Rule-plan acceptance instruction are `CONSUMED/HISTORICAL`. No implementation authority; Rule TDD remains `NOT_AUTHORIZED/BLOCKED` under the current `669ef761` receipt's stated prerequisites. 144-01B Docker-blocked; Phase 142/148 pending; launcher `NOT_READY`, executables `NOT_RUN`.

> **2026-08-15 — 02B precedence:** source `947e2a27650ff1c6db60af9ce13bb33b6bf19cb0` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; `IMPLEMENTATION-144-02B.md` records exact13 blobs. Tests first; Contract/high, Security/xhigh and Lifecycle/high precommit and exact-SHA PASS C/I/M `0/0/0`. All prior 02B `SOURCE_AUTHORITY_ACCEPTED`, source-unaccepted and 02C-waits-02B-source clauses are `CONSUMED/HISTORICAL`. A separately authored/adjudicated narrow 02C docs/source-authority bridge is required; 02C source is not automatically authorized. Other blockers remain unchanged. **02C source acceptance: preceding 02C route clauses are CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS; exact13 blobs IMPLEMENTATION-144-02C.md; no next Mail packet selected; named blockers remain.**
> **Current deletion-port source receipt (2026-08-23; highest precedence):** immutable source `e4f81bf3d8b932f5ff2b9a8069d41bfa1d1a0e26` (parent `bd76d8469782e9bb7926b440b6b45495bc865e84`, tree `93bac3c6a5798cb75b0a6647fa938416c5e82ede`) is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/internal/mail/mailbox_deletion.go` blob `001b03c879aab36031feac07f0d3894f76e23abd` and `backend/internal/mail/mailbox_deletion_fake.go` blob `7d43b4dee20ad762d251e3c14e428f2f0ccda497`. `provenance_source: dispatcher-enforced`; Terra/xhigh implementation; Contract/high, Security/xhigh, and Lifecycle/high immutable reviews are C/I/M `0/0/0`. This accepts only the provider-neutral deletion ABI and two standalone deterministic fakes. Tests, build, gofmt, lint, parser, validator, generation, migration, SQL, Docker, and runtime are `NOT_RUN`. Within this exact2 packet, former `UNIMPLEMENTED`/`UNCOMMITTED`/`UNACCEPTED`/exact2-next/source-review-pending clauses are `CONSUMED/HISTORICAL`. Concrete provider adapter capability, persistence/Ent/schema, migrations, workers/drains/finalization/User-GDPR/HTTP/NATS remain `BLOCKED`; no next source is selected. Testframework merge facts are unchanged.
