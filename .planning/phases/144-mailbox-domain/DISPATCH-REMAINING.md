# Phase 144 — executable packet contracts not split into standalone files

> **Assembly rule:** dispatch the Common execution envelope plus exactly one packet or one
> explicitly named variant below. Combining variants is invalid.
> **Runtime:** `gpt-5.3-codex` / `xhigh`, fallback none.
> **Status:** CODE_RUNTIME_BLOCKED; generation/runtime also launcher-blocked.

Standalone complete prompts exist for 144-01A, 144-04D, 144-04E, 144-05A, 144-05B and
144-06A. This file completes every other index row.

## Common execution envelope

- Role Hub `codex-micro`; TDD external-package genuine unit first, then minimal code. One
  behavior, exact allowed files, no commit/push/PR. Root supplies exact clean HEAD.
- You are not alone: no revert; dirty overlap/missing frozen dependency/model mismatch =>
  `BLOCKED` before write. Russian exported comments; Go 1.25; stable safe errors; bounded
  collections; no new dependency/global state/wall clock/network/listener/real DB/broker.
- Never read/change protected `testframework/**`, `backend/testframework/**`, test-rebase
  audits/phases 128–141/TASK-193–234, `pro-qa-ai/**`, `.env.local`, `.git/index.lock`.
- Existing Resend alert files/routes/flat DTO remain unchanged unless a packet explicitly
  lists a narrow composition call. No raw body/MIME/key/credential/provider response in
  PostgreSQL, error, audit or event.
- Host tests/build/lint/validators/generation forbidden; direct Docker/Compose not fallback.
  Accepted launcher absent => `NOT_RUN`. Real DB/engine proofs route to Python handoff, not
  fake component Go tests in `backend/tests/**`.
- Bounds: depth/fan-out 0; two retries/root cause; 45m (60 service, 75 handler). Response in
  Russian with files/symbols/oracles/verification/risks/requested+actual model/source SHA.

## 144-01B — deterministic engine fake

- Dependency: accepted 01A. Allowed only:
  `backend/internal/mail/mailbox_fake.go`,
  `backend/tests/mail/mailbox_fake_test.go`.
- Behavior: mutex-protected in-memory accounts/folders/messages/blobs/drafts/submissions;
  injected ids/clock; scripted typed errors and lost responses; strict account scoping;
  `ReconcileDraft` and `ReconcileSubmission` never mutate/retry.
- Oracles: compile conformance; same stable submit key one visible result; definitive
  absent vs indeterminate; draft expected-version conflict; changes cursor pages; cross-
  account ids zero disclosure/mutation; deterministic snapshots; no secret/body in errors.

## 144-02A — account/identity/domain/policy schemas

- Dependency: contract 0.1.4 and migration gate recorded. Allowed only:
  `backend/ent/schema/mailaccount.go`, `mailidentity.go`, `maildomain.go`,
  `mailboxpolicy.go`, narrow additive inverse edges in `user.go`, `organization.go`.
- Behavior: exact spec §3 fields/enums/bounds; unique org+user, opaque engine account;
  normalized identity/domain uniqueness; partial one-default indexes; one policy/org; no
  password/admin/DKIM private secret. Preserve all existing edges byte-semantically.
- Oracles: schema field/index/edge table; partial predicates; active/provisioning states;
  exact pre/post existing-edge inventory; no generated/migration edit.

## 144-02B — folder/thread/message/membership/label schemas

- Dependency: reviewed 02A. Allowed only new schema files
  `mailfolder.go`, `mailthread.go`, `mailmessage.go`, `mailmessagefolder.go`, `maillabel.go`.
- Behavior: same-account relations; engine folder/message ids unique per account;
  message-folder join unique; flags/state/freshness rebuildable; labels/pin Hub-owned;
  bounded headers/snippet only, never raw body/MIME; custom root/child sibling uniqueness
  with nullable-parent partial indexes.
- Oracles: exact indexes/predicates; no folder cycle claim at schema alone; soft-delete
  fields; BCC projection policy; no cross-account edge accepted by service contract.

## 144-02C — object/attachment/draft/submission schemas

- Dependency: reviewed 02A. Allowed only new schema files
  `mailobject.go`, `mailattachment.go`, `maildraft.go`, `mailsubmission.go`.
- Behavior: mail object owner/account/lease/key/hash/scan; attachment backing metadata;
  draft opaque engine id/version/revision/states/hashes, no content; submission hashed key,
  request hash, stable engine key/states/result. Express XOR/check through selected
  migration only when authority exists; service invariants remain explicit.
- Oracles: raw key/body absent; unique account+hash and account+engine ids; draft submitting
  state; no project/FileMetadata edge; expiry/audit; no generated/migration edits.

## 144-02D — settings/signature/rule/contact schemas

- Dependency: reviewed 02A. Allowed only new schema files
  `mailusersettings.go`, `mailsignature.go`, `mailrule.go`, `recentcontact.go`.
- Behavior: one settings/account; required same-account default identity for active
  account, nullable signature; one default signature; bounded sanitized config; typed rule
  JSON/version but no raw Sieve; recent contact unique normalized email/account.
- Oracles: ownership/default partial indexes; signature/rule bounds; active Hub user rank
  not encoded as persistence; no message body authority or secret.

## 144-02E — Ent generation inspection

- Dependencies: accepted 02A–D, founder migration choice and launcher. Allowed: no manual
  tracked source; task-owned ignored generated output/evidence only.
- Behavior/oracles: repository Ent generator inside bounded launcher; exact SHA/tool;
  inspect all schema types/edges/index metadata and compile; second generation empty diff;
  exact cleanup. No migration invention/stage. Missing gate => BLOCKED/NOT_RUN.

## 144-03A-Address — address policy

- Dependency: 01A. Allowed only `backend/internal/mail/address_policy.go`,
  `backend/tests/mail/address_policy_test.go`.
- Behavior/oracles: parse normalized mailbox/domain/display name with IDNA/email bounds;
  case/domain policy and duplicate canonical set; CRLF/control/header-injection rejection;
  reserved synthetic fixtures; no DNS/network; stable safe errors.

## 144-03A-Thread — thread identity

- Dependency: accepted 03A-Address. Allowed only
  `backend/internal/mail/threading.go`, `backend/tests/mail/threading_test.go`.
- Behavior/oracles: valid nearest References then In-Reply-To; fallback normalized subject
  + canonical participants; bounded localized prefix stripping; Unicode normalization;
  deterministic SHA-256/base64url; misleading subject cannot beat header ancestor; bounds.

## 144-03B-Folder — folder tree policy

- Dependency: 01A. Allowed only `folder_tree.go`, `backend/tests/mail/folder_tree_test.go`.
- Behavior/oracles: depth ≤2, no cycle/self-parent/cross-account parent; system role cannot
  delete/reparent; normalized sibling collision; deterministic ordering; no engine call.

## 144-03B-Rule — typed rule compiler

- Dependency: accepted 03B-Folder. Allowed only
  `rule_compiler.go`, `backend/tests/mail/rule_compiler_test.go`.
- Behavior/oracles: frozen conjunction conditions/actions only; escape all literals;
  folder/label targets resolved inputs; priority/order deterministic; raw Sieve/new action/
  nesting/bounds rejected; injection fixtures cannot add commands.

## 144-03C-Send — rate/recipient/loop policy

- Dependency: 01A. Allowed only `send_policy.go`,
  `backend/tests/mail/send_policy_test.go`.
- Behavior/oracles: message/recipient/rate bounds, external toggle, authorized identity;
  null sender/Auto-Submitted/precedence/loop headers suppress auto-response; injected clock
  and token-bucket state; explicit retry hint; no network/global clock.

## 144-03C-Attachment — attachment policy

- Dependency: accepted 03C-Send. Allowed only `attachment_policy.go`,
  `backend/tests/mail/attachment_policy_test.go`.
- Behavior/oracles: normalized filename/media/inline CID/size/count; clean-only send/read;
  pending/rejected/error fail closed; unsafe SVG/executable/double-extension policy;
  deterministic no I/O; no project authorization.

## 144-04A — account resolver/repository base

- Dependencies: 01B/02E. Allowed only `mailbox_repository.go`,
  `mailbox_ent_repository.go`, `mailbox_service.go`,
  `backend/tests/mail/mailbox_owner_test.go`.
- Behavior: authenticated user -> active User organization -> active MailAccount -> opaque
  engine account. Public commands never carry org/account. Repository interfaces expose
  transaction/lease primitives for later services; Ent adapter implements only resolution.
- Oracles: cross/suspended/deleted/no account zero engine; no client id; transaction lease
  contract; safe non-disclosure. Real DB constraints remain Python integration.

## 144-04B-Folder — folder commands

- Dependencies: 03B-Folder/04A/01B. Allowed only `folder_service.go`,
  `backend/tests/mail/folder_service_test.go`.
- Behavior/oracles: list/create/rename/delete through resolved account + engine; policy
  first; successful engine mutation reconciles projection/state; timeout/partial failure
  degraded and refetched, not guessed; system roles protected; cross-account zero engine.

## 144-04B-Label — label commands

- Dependency: accepted 04A. Allowed only `label_service.go`,
  `backend/tests/mail/label_service_test.go`.
- Behavior/oracles: Hub-transactional CRUD/assignment, same-account messages, normalized
  uniqueness and partial per-message result; no engine call; thread aggregation derived;
  concurrent/version conflict safe.

## 144-04C — message projection/read/bulk

- Dependencies: 03A-Thread/03B-Folder/04A/01B/02E. Allowed only
  `message_service.go`, `message_projection.go`,
  `backend/tests/mail/message_service_test.go`.
- Behavior: owner filter before engine hydrate; opaque cursor; engine authoritative
  folders/flags, Hub labels/pin; bulk per-message outcomes; projection freshness explicit.
- Oracles: no cross-mailbox hit/count/snippet/timing hydrate; flag/move failure reconcile;
  partial batch; BCC owner-sent only; body never repository; duplicate change idempotent.

## 144-05C-Search — search seam

- Dependency: 04A/C. Allowed only `mailbox_search.go`,
  `backend/tests/mail/mailbox_search_test.go`.
- Behavior/oracles: typed search port/query; mandatory org/account filters embedded before
  query and again before hydrate; bounded cursor/filters/snippet; foreign hits discarded
  without leaking total/id; index document has no raw body beyond approved safe projection.

## 144-05C-Contacts — contact ranking

- Dependency: 04A + 03A-Address. Allowed only `mailbox_contacts.go`,
  `backend/tests/mail/mailbox_contacts_test.go`.
- Behavior/oracles: active same-org Hub users first, then owner recent contacts; normalized
  dedupe; bounded query/result; disabled users removed; no address disclosure cross-org.

## 144-05C-Realtime — websocket/notification seam

- Dependencies: 05B + 04C/E. Allowed only `mailbox_realtime.go`,
  `backend/tests/mail/mailbox_realtime_test.go`.
- Behavior/oracles: internal event -> owner/session bounded invalidation/counter metadata;
  no body/recipient list; duplicate idempotent; unknown version ignored/error-routed; no
  direct websocket network in unit.

## 144-05C-Audit — named audit seam

- Dependency: 04A. Allowed only `mailbox_audit.go`,
  `backend/tests/mail/mailbox_audit_test.go`.
- Behavior/oracles: exact OpenAPI admin action enum plus bounded bulk summaries; actor/org/
  target/result/request id only; no recipient/body/key/secret; mutation audit participates
  in supplied transaction; generic HTTP audit is not substituted.

## 144-06B — mailbox admin service

- Dependencies: 05C-Audit/04A/02E. Allowed only `mailbox_admin_service.go`,
  `backend/tests/mail/mailbox_admin_service_test.go`.
- Behavior/oracles: admin principal + current org; policy get/update, domain list/create/
  update/delete and redacted health; normalized uniqueness/default safety; enabled/external
  preconditions; exact audit intent; no secret/DKIM private material; non-admin zero mutate.

## 144-06C — mailbox admin handler/wiring

- Dependencies: 06A/B. Allowed only `mailbox_admin_handler.go`, its focused unit, and the
  narrow call in `backend/cmd/server/wiring_mailbox.go` needed to register it.
- Behavior/oracles: exact OpenAPI mailbox-settings/domain/health routes and x-hub audit;
  admin middleware; NFR-13; legacy `/admin/mail/settings` response/registration and
  `/broadcast` current state untouched; health redacted; no credential input/output.

## 144-07 — docs sync and independent static verification

- Dependencies: all accepted code packets. Allowed only affected current-truth docs named
  by `.cursor/rules/55-docs-sync.mdc` plus a Phase 144 verification artifact; no product/
  test/protected file.
- Behavior/oracles: implementation-to-OpenAPI/FR/TC matrix; exact diff proves legacy
  Resend untouched, bodies/keys absent from persistence, `mail.>` present once and webhook
  list unchanged; record current SHA and all Docker/test/build/generation statuses honestly.
  Must request independent Terra review; cannot mark runtime PASS from static review.
