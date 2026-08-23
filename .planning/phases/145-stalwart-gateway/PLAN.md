# Phase 145 — Stalwart Gateway and Docker Mail Profile
> **Current narrow deletion-source authority (2026-08-23; highest precedence):** accepted plan `3f7aa2c58407ac3a10ad427cfa86256e0a9d340a` is `DOCS_PLAN_STATIC_ACCEPTED / SOURCE_AUTHORITY_ACCEPTED` only for one future atomic no-test exact2 production packet: `backend/internal/mail/mailbox_deletion.go` and `backend/internal/mail/mailbox_deletion_fake.go`. Plan Contract/high, Security/xhigh and Lifecycle/high reviews are C/I/M `0/0/0`; `provenance_source: dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own commit and immutable C/S/L triple `0/0/0`. Adapter capability proof is not required only for this provider-neutral ABI plus two standalone fakes exception, but remains mandatory for every concrete provider adapter. Persistence/Ent/schema/generation/migration/SQL/worker/storage drain/account finalization/User-GDPR/HTTP/NATS/Docker/runtime stay `BLOCKED` / `NOT_RUN`. Old blanket deletion-source-blocked or plan-pending claims are `CONSUMED/HISTORICAL` only within this exact2 scope; unrelated gates are unchanged.
> **Engine deletion-port adjudication (2026-08-23):** deletion ABI docs contract is frozen and accepted through `3900b331` nil hardening; former deletion-port-next selection is `CONSUMED/HISTORICAL`, never source authority. It defines separate opaque ports `EngineAccountDeletionPort` (`DeleteAccount`, `ReconcileAccountAbsence`) and `MailObjectDeletionPort` (`DeleteObject`, `ReconcileObjectAbsence`); do not widen normal `MailEngine` or combine account delete with drain. Requests durably retain nonzero `DeletionID` plus account provisioning key+engine account ID or object bucket+key; opaque IDs use `MaxEngineOpaqueIDBytes` and bucket/key existing storage bounds. Account results are `deleted|authoritatively_absent`, observations `present|authoritatively_absent`; bare account `NotFound` preserves intent unless scoped authoritative absence is proven. Object finalization requires only a closed `MailObjectDeleteResult` state `deleted|authoritatively_absent` or scoped `ReconcileObjectAbsence` authoritative absence; the object error ABI has no `not_found`. No provider/raw IDs or causes are public/logged. Private intent/control is not owner projection; short lease transactions keep I/O outside locks, use exact token+generation+scope CAS, reclaim expired lease only, and stale token/zero rows cause no I/O/compensation. Object drain starts after `upload_drain_until` and attachment/reference terminality; account removal is last after exact child/object/receipt/control/activity manifest; hard-delete only after closed evidence. No Stalwart route/OAuth/DNS/OpenAPI assumption. Source/adapter/persistence/worker work remains blocked pending adapter capability proof, accepted source plan, and named dependencies. Ent/source/generated/migration/SQL/tests/Docker/runtime and physical User/GDPR remain `BLOCKED` / `NOT_RUN`.

> **Current Mail docs/readiness receipt (2026-08-15):** `576f19231e37f635d5b7187a1ada121ed14ff8dc` is the exact14 docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS` receipt. Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high reviews PASS C/I/M `0/0/0`; `provenance_source: dispatcher-enforced`, no fallback, telemetry `UNKNOWN` unless exposed. It grants no source/schema/runtime/product/phase/release authority. All lower active 02D/02F-only, candidate, precommit, root-commit, exact-review and status-sync-next routes are `CONSUMED/HISTORICAL` as applicable while retaining unrelated historical facts. 02F accepted and Rule/Sieve plan facts remain accepted, without source/schema/runtime authority. 144-01B, Phase 142/148, G145-5B/G145-5C and all other gates remain unchanged; launcher `NOT_READY`, executables `NOT_RUN`. The deletion ABI docs contract is accepted; implementation remains blocked pending adapter capability proof, accepted source plan, and named dependencies. The separate User deletion-intent contract remains a blocker.
> **Carrier-plan acceptance receipt (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, sole blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); Contract Terra/high, Security/Purity Terra/xhigh and Lifecycle Terra/high reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787` and sync `375891bb` remain accepted. Only obsolete plan author/adjudicate/accept/pending fragments are `CONSUMED/HISTORICAL`; all unrelated G145 gates remain current. No source/test/runtime/adapter/schema/service/worker/product/phase/release authority. Rule Task 1 awaits G145-5B/G145-5C/launcher. Carrier six-path RED/source additionally awaits accepted Rule compiler source/ABI; plan gate is satisfied. 145-07B source/runtime blocked; `NOT_READY`/`NOT_RUN` unchanged.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Carrier acceptance receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, all exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No carrier plan/source/test/runtime authority. Root-only separate carrier-plan acceptance, G145-5B/G145-5C and launcher readiness keep Rule Task 1/145-07B `NOT_AUTHORIZED/BLOCKED`; launcher `NOT_READY`, executables `NOT_RUN`.

> **Current upstream lifecycle (2026-08-15):** the non-marker typed-IR/Sieve carrier contract is docs/design-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `1daef78768048d7672c38a65b78ef0f119f3c207`. This Phase remains blocked; no carrier plan/source authority follows. Rule Task 1 is `NOT_AUTHORIZED/BLOCKED` pending a separately `STATIC_ACCEPTED` carrier implementation plan, G145-5B, G145-5C and launcher readiness.

> Task: TASK-238 · Status: BLOCKED_BY_144_CONTRACT_AND_EXTERNAL_GATES · Depends on: Phase 144

> **Upstream lifecycle update:** 144-02D remains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; 02F `fcb5b84f` plus `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former lifecycle is `CONSUMED/HISTORICAL`. Phase 145 remains blocked/not accepted: carrier contract is accepted docs/design-only at `1daef787`, while the separate carrier implementation plan/source gate, G145-5B, G145-5C and launcher/runtime gates remain open.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

## Hard entry gates

0. Phase 144 engine projection/content/blob byte boundary and admin hard ceilings are
   amended and independently accepted.
1. Founder/legal chooses AGPL-3.0 or SELv2 terms.
2. Accepted Docker-only launcher exists for this task.
3. Candidate image tag and digest provenance are refreshed immediately before use.
4. Shared `/jmap` management Registry and mailbox method contexts are reviewed with
   separate credential providers and non-overlapping method allowlists; authentication is
   selected by typed operation, never URL, and mixed batches fail before I/O. A management
   API key is never treated as a mail credential. Shared-account ACL is not a
   full-engine substitute because exact candidate source keeps Identity, EmailSubmission,
   Sieve and Blob member-only. Founder/security selects per-user Authorization Code OAuth
   (root recommendation) or explicitly accepts broad dedicated-account impersonation for
   zero-touch operation; no OAuth client-credentials flow is assumed.
5. If Authorization Code OAuth is selected, one user-reachable HTTPS origin and exact
   default-deny listener/path/proxy boundary are accepted. `STALWART_PUBLIC_URL` controls
   OAuth/OIDC and JMAP discovery together; separate browser and Hub origins are not assumed.
6. A domain-verification authority and bounded DNS projection are reviewed. Selected
   Stalwart `x:Domain` does not expose manual-DNS verification, so domain existence or
   `dnsZoneFile` cannot satisfy Hub `verified`.
7. Account provisioning is reviewed as two-context reconciliation on shared `/jmap`:
   management `x:Account` owns the address set, while separately authenticated mailbox
   `Identity/get|set` supplies stable RFC 8621 ids. This path remains disabled until gate
   4/G145-5B.

No runtime implementation/claim proceeds by silent fallback.

## Ordered tasks after gates

1. Implement two SSRF-safe, redacting typed Stalwart contexts under one trusted internal
   origin and shared fixed `/jmap`: scoped API-key Registry calls and separately authenticated
   mailbox calls. Mail discovery uses `/.well-known/jmap`; `/api` is limited to individually
   allowlisted helper/auth path+method pairs. Reject mixed Registry/mail batches and wrong-
   context credentials before I/O in `backend/internal/mail/stalwart/**`.
2. Implement domain/account/identity/folder/message/draft/submission/flags/fixed-Sieve
   mappings and, only after its separate evidence gate, explicit spam/ham
   classifier-feedback mapping,
   including opaque draft versions, stable submission keys and explicit submission
   reconciliation; prove adapter contract conformance.
   Folder lost-response recovery is enabled only by a reviewed capability proving Hub is
   the sole custom-folder writer; false/unknown or alternate writer remains degraded.
   Domain/account mappings use selected-version management `x:*` objects. Before the
   domain-verification gate, domain mapping returns at most `dns_required` and account
   provisioning remains ineligible. Account mapping obtains identity ids only from mailbox
   `Identity/get|set` after management provisioning; it never fabricates ids. Per-mailbox
   Sieve uses RFC 9661 `SieveScript/*` on mail JMAP, never the global management
   `x:SieveSystemScript` or `x:SieveUserScript` objects. Spam/ham stays omitted until an
   exact message mapping and idempotency/reconciliation contract is evidenced.
3. Implement bounded projection-only incremental Changes plus a resumable base-state/ID-anchored
   projection enumeration whose composite base cursor is accepted by Changes. Query-state
   drift converges through catch-up; missing anchor, unstable enumeration or expired base
   history returns typed restart-required and never resets silently.
4. Compose exactly one Phase 144 domain-control, account-control, combined leased
   projection/recovery and rule-deployment worker
   only for an enabled adapter on the shared root context, with bounded reverse-order
   cleanup; extend admin health without exposing credentials, cursors or checkpoints.
5. Add optional `mail` profile, named volumes, healthcheck, restart and explicit CPU/RAM/
   PID limits; add names-only env examples. Public mail transport is SMTP-only in wave 1;
   JMAP/admin stay Hub-only and user IMAP/JMAP/ManageSieve remain unpublished. A separately
   reviewed HTTPS surface may expose only OAuth paths if G145-5D selects Authorization Code.
6. Treat invalid/expired change cursors as degraded state followed by a bounded,
   owner-scoped projection rebuild; never silently reset and present stale data as fresh.
7. Through accepted launcher prove health, authenticated JMAP, internal delivery,
   instance-domain SMTP, accepted domain-verification authority, enumeration/catch-up plus
   folder lost-response recovery, replica
   lease exclusion, four-worker cancellation, restart persistence, exact cleanup and actual
   resource use.
8. Document dedicated engine-volume backup/restore and DNS/deliverability prerequisites.

External Internet delivery remains a separate operator acceptance even after local
transport passes.

## Rule compiler boundary blocker

The `f199546513b758ca7b95328b5ccd7680009b732e` / `6735ae746c7acde657359e8a709cd897c4dce61c` lifecycle is `CONSUMED/HISTORICAL`. `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4`, sole-path blob `b86c0423a307fd2673a3d322435922aa2a398e8f`, after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. It accepts the plan only, not engine source or Sieve bytes. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` until all independent accepted typed-IR/script carrier implementation prerequisites, G145-5B, G145-5C and launcher readiness exist, without implied order; post-task runtime is `NOT_RUN`. Engine folder ID is not mailbox name;
attachment IR needs MIME gate. No Rule TDD/source/schema/runtime/product/release authority follows.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition and Rule-plan-next text is `CONSUMED/HISTORICAL`; no source/schema/runtime/product/release authority. Carrier contract `1daef787` is docs/design-only accepted; Rule Task 1 still awaits a separately accepted carrier implementation plan, G145-5B, G145-5C and launcher readiness. Task4 is post-future-Task3 source acceptance only. 144-01B remains Docker-blocked, Phase 142/148 and release pending, launcher `NOT_READY`, executables `NOT_RUN`.
