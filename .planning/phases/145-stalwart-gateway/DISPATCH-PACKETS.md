# Phase 145 — executable gated packet contracts

> **Assembly rule:** Common envelope + exactly one packet. No packet is currently eligible.
> **Runtime for code:** `gpt-5.6-terra` / `xhigh`, fallback none.
> **Authority:** `GATE-PLAN.md`; a table row here never self-satisfies a gate.

> **Upstream lifecycle update:** 144-02D docs-only boundary `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition `e563a3d3be8898371e5c8a9861c58f11b72167bc` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; sole Mail next action is three exact 02F closure reviews. Phase 145 remains blocked/not accepted; no packet is eligible and G145-5B/G145-5C remain open.

## Common execution envelope

- Root supplies exact clean SHA, selected licensed engine, immutable image digest and the
  version-specific official API/config evidence required by the packet. Missing or stale
  evidence => `BLOCKED`, no guessing from model memory.
- One behavior/allowed file set, tests first with bounded fake HTTP/JMAP server only where
  allowed; no commit/push/PR. Do not revert concurrent work.
- Never read/change protected testframework/test-rebase/pro-qa-ai zones. Existing Resend
  sender/settings/status/broadcast implementation is a separate compatibility boundary.
- No client-controlled base/capability URL, redirect, engine account id or credential.
  Registry and mailbox methods share fixed `/jmap` but use separate typed auth/method
  contexts; neither credential nor method may cross, and a batch may not mix contexts.
  `/api` permits only individually allowlisted helper/auth path+method pairs. Bound response
  bytes/count/depth/time. Redact
  auth/session/provider bodies/addresses.
- No host test/build/lint/validator/generator/process. Direct Docker/Compose not accepted
  evidence. Runtime only through accepted task-owned launcher with limits/timeout/owner
  report/finally cleanup.
- Code response in Russian: files/symbols/oracles/gates/verification/model provenance/SHA.
  depth/fan-out 0, retries 2, 60m (90 for deployment/runbook/runtime).

## 145-01A — runtime config

- Gates: G145-2; accepted Phase 144 port. The disabled-only shape may be frozen before
  G145-5B, but every enabled path also requires G145-5B. Allowed only
  `backend/internal/mail/mailbox_config.go`,
  `backend/tests/mail/mailbox_config_test.go`.
- Behavior/oracles: disabled default needs no secret/domain and reports not_configured;
  exclusive custom-folder writer defaults false and cannot be enabled by an HTTP request;
  one fixed internal absolute http(s) origin contains no path/query/fragment. Enabled mode
  requires a separately scoped management API-key reference plus the reviewed mailbox auth
  mode/principal/secret references from G145-5B; the legacy single administrator
  user/secret pair is rejected. Until G145-5B, every enabled mode is rejected while disabled
  remains usable. External requires enabled + domain; validate byte/rate/retention bounds;
  errors/string dumps redact values; no environment read inside pure parser (map/lookup
  injected).

## 145-01B — redacting bounded transport client

- Gates: G145-1/2/4/5/5B; dependency 01A. Allowed only
  `backend/internal/mail/stalwart/client.go`, `errors.go`,
  `backend/tests/mail/stalwart_client_test.go`.
- Behavior/oracles: injected `http.Client`, fixed origin and same-origin resolved endpoints;
  management Registry client permits only fixed `/jmap` plus exact `x:*` method allowlist
  and injects only the scoped management API key; mailbox client handles well-known/session/
  upload/download policy, permits only standard mail methods and injects only the accepted
  mailbox credential. Mixed Registry/mail batches, wrong-context credentials and provisioning
  through `/api` fail before I/O.
  Redirects are disabled/validated; request/response/time/count bounds and context
  cancellation are mandatory; map 4xx/5xx/timeout stably; logs/errors omit authorization,
  URL query, provider body/ids. No real network in unit.

## 145-02 — JMAP session/capabilities

- Gates: G145-5/5B; dependency 01B. Allowed only
  `backend/internal/mail/stalwart/jmap.go`, `config.go`,
  `backend/tests/mail/stalwart_jmap_session_test.go`.
- Behavior/oracles: freeze management compatibility as a typed Registry context on fixed
  `/jmap` with `urn:stalwart:jmap` and exact selected-version `x:Domain/*` / `x:Account/*`
  methods. Mail
  discovery reads exact selected-version session JSON from `/.well-known/jmap`; require
  core/mail/submission/blob and optional RFC 9661 sieve explicitly; derive mail URLs only
  through same-origin/session policy. Bound accounts/capabilities; missing/unknown required
  capability fails closed; management evidence is never inferred from mail capabilities;
  Registry and mail method calls cannot share one JMAP batch, and `/api` is not provisioning;
  no credential/session URL in error. The accepted management profile uses API-key
  `Replace` with exactly `domain-create|delete|get|list|update` and
  `individual-create|delete|get|list|update`; any broader principal/key is an external gate
  failure rather than an adapter convenience.

## 145-03A — account/identity mapping

- Gates: G145-5B; dependencies 02 + Phase 144 port. Allowed only
  `backend/internal/mail/stalwart/account.go`, `mapper.go`,
  `backend/tests/mail/stalwart_account_test.go`.
- Behavior/oracles: use the management Registry context on shared `/jmap` and exact
  `x:Account/*` object to
  idempotently ensure the stable Hub key, reserved local part, selected verified domain and
  normalized primary/alias address set. Accept only `@type: "User"`; same key/spec repeats,
  changed spec conflicts and group results never satisfy a user account. Then use the
  separately authenticated mailbox context's standard `Identity/get|set` to reconcile exactly
  one stable server id per desired address. Missing identities are created only after Get;
  a lost create response performs Get-before-retry. Duplicate email identities, identities
  outside the desired address set or immutable-email mismatch are indeterminate/degraded,
  never selected by list position and never replaced by fabricated ids. Conflict/partial
  provisioning is safe; no password/token is persisted/returned; cross-account query is
  impossible and neither credential crosses typed contexts.

## 145-03C — domain desired-state mapping

- Gates: G145-5 exact selected-version domain management API and G145-5C verification
  authority/projection. Dependencies: 02 + Phase 144 port. Allowed only
  `backend/internal/mail/stalwart/domain.go`,
  `backend/tests/mail/stalwart_domain_test.go`.
- Behavior/oracles: use only the management Registry context on shared `/jmap` and exact
  `x:Domain/*` object.
  `EnsureDomain` idempotently applies stable key + captured generation/
  hash + desired present/enabled with manual DNS management; `GetDomain` is read-only
  reconciliation. Map only opaque id, absent/DNS-required/verified/error/indeterminate, the
  exact bounded public DNS projection accepted by G145-5C and DKIM status. Domain existence,
  `isEnabled`, non-empty `dnsZoneFile` or a successful management write never produces
  `verified`; only the accepted verification authority may do so. Before that authority is
  available the only present steady state is `dns_required`, account provisioning stays
  ineligible and no automatic DNS call occurs. Lost create/delete response is observable;
  changed generation conflicts; no private DKIM material, credential/provider body or
  cross-org data.

## 145-03B — folder mapping

- Dependency: 02. Allowed only `stalwart/folder.go`,
  `backend/tests/mail/stalwart_folder_test.go`.
- Behavior/oracles: selected-version mailbox get/set mapping, roles, parent ids, counts and
  state; account id in every request; deterministic pagination; unsupported hierarchy /
  partial failure typed; configure/enforce at most 100 non-deleted folders atomically and
  map over-limit safely; an over-cap get is rejected rather than partially projected;
  no Hub color/order sent as engine authority unless port says so. Expose the injected
  exclusive-custom-folder-writer capability only from accepted trusted runtime config;
  system-role bootstrap folders stay separately typed. False/unknown capability supports
  reads but cannot authorize exact-match adoption or any automatic recovery write.

## 145-04 — message/blob/list mapping

- Dependencies: 02/03B. Allowed only `stalwart/message.go`, `blob.go`,
  `backend/tests/mail/stalwart_message_test.go`.
- Behavior/oracles: exact selected-version properties; bounded ids/page/body/blob; body/
  MIME exists only in returned request-lifetime values; received BCC absent; folder/flags/
  state mapped; account scoping; malformed/missing property safe; no raw response error.
  Map stable submission/draft provenance only when selected-version engine evidence ties to
  an accepted Hub submission/draft identifier. Address/From/alias equality and mutable Sent
  membership never produce confirmed-owner origin; ambiguous provenance maps
  `inbound_or_unknown` so the domain fails toward event admission.
  No engine-search method/query exists; public search is the Hub Meilisearch projection.

## 145-05A — hydrated incremental changes and invalid cursor

- Dependency: 04. Allowed only `stalwart/changes.go`,
  `backend/tests/mail/stalwart_changes_test.go`.
- Behavior/oracles: an adapter-private composite cursor carries the exact selected-version
  Mailbox and Email states. One call returns at most 100 unique account-scoped created/
  updated hydrated snapshots plus destroyed ids and the exact old/new composite cursor;
  `hasMore` never advances past an unreturned change. Expired/cannot-calculate state maps
  typed invalid-cursor; empty-success reset is forbidden. Mixed-account/missing hydration,
  duplicate/conflicting ids, oversized result and provider-body error fail safely. Replay
  of the same page is stable and exposes no cursor/provider text in logs/errors/events.

## 145-05B — base-anchored resumable projection enumeration

- Dependencies: 03B/04/05A. Allowed only `stalwart/enumeration.go`,
  `backend/tests/mail/stalwart_enumeration_test.go`.
- Behavior: capture the selected-version Mailbox/Email object states before enumeration
  (standard empty-id get/state or the reviewed version-specific equivalent) and encode that
  composite base with the unfiltered Email query resume anchor in an opaque adapter-private
  checkpoint. Return at most 100 normalized folder/message snapshots per call. Mailboxes
  are fetched as one bounded set under the contract account cap. Email query uses the same
  unfiltered filter and reviewed immutable comparator set on every request, the prior last
  ID as `anchor` with `anchorOffset=1`, and never a drifting positional offset. JMAP's
  server-dependent tie order is accepted only because RFC 8620 requires it stable between
  calls. Query-state drift is recorded but is not by itself restart-required: result-list
  inserts/deletes after the composite base are converged by mandatory 05A Changes catch-up.
  Hydrated object values may be newer than the base and apply idempotently.
- Oracles: first page/base capture; multi-page resume; exact final base cursor accepted by
  05A; stable duplicate page; current hydration followed by base catch-up; create/delete/
  flag mutation and query-state drift during enumeration. Query drift alone completes only
  after Changes catch-up; anchor-not-found, unsupported/unstable sort-anchor semantics or
  cannotCalculateChanges for the base returns typed restart-required with no partial-final
  signal. An unchanged object present at the base and still present at finalization is never
  skipped. Folder/message/count/address/body bounds and account scope hold. Checkpoint/base/
  query states and provider objects never enter API/event/log/error values. No empty cursor,
  destructive clear or domain event emission exists in this adapter packet.

## 145-06A — draft save/reconciliation

- Dependencies: 04/05A + Phase 144 draft port. Allowed only
  `stalwart/draft.go`, `backend/tests/mail/stalwart_draft_test.go`.
- Behavior/oracles: create/update selected-version JMAP calls with expected opaque state and
  stable mutation id; version conflict distinct; timeout indeterminate; `ReconcileDraft`
  reads/compares current state/content fingerprint and returns applied/absent/indeterminate
  without invoking set/save; body never logged/persisted.

## 145-06B — submission/reconciliation

- Dependencies: 06A + Phase 144 submission port. Allowed only
  `stalwart/submission.go`, `backend/tests/mail/stalwart_submission_test.go`.
- Behavior/oracles: stable engine key bound to selected-version identity; one submit call;
  accepted ids/state mapping; timeout indeterminate; reconciliation queries definitive
  sent/absence/rejection where API supports it and never submits; ambiguous remains
  indeterminate; safe SMTP/provider errors and no duplicate-visible-mail fake scenario.

## 145-07A — message mutations

- Dependencies: 03B/04. Allowed only
  `backend/internal/mail/stalwart/mutation.go` and
  `backend/tests/mail/stalwart_mutation_test.go`.
- Behavior/oracles: account-scoped move/public flags/reserved keyword add-remove/delete;
  reserved markers are typed separately and never accept label name/UUID/public raw input;
  partial per-id errors; hard-delete policy explicit; returned opaque state is reconciled; another account's ids produce no
  call/mutation. Do not implement or upload Sieve in this packet.

## 145-07B — typed Sieve mapping

- Dependencies: 02 plus accepted Phase 144 rule compiler. Allowed only
  `backend/internal/mail/stalwart/sieve.go` and
  `backend/tests/mail/stalwart_sieve_test.go`.
- Behavior/oracles: use mailbox JMAP only with capability
  `urn:ietf:params:jmap:sieve`; accept only the compiler-produced typed representation and
  one fixed Hub script name. Render deterministically, bound bytes/rules, upload the script
  as a bounded blob, then use standard RFC 9661 `SieveScript/get`, `SieveScript/set` and
  `onSuccessActivateScript`; encode generation/hash marker. Native move/read compile
  directly; label action compiles only the resolved opaque reserved keyword. `GetSieve`
  returns only active/marker state. `UpsertSieve` distinguishes applied from indeterminate
  so the worker can Get-before-retry; expose no public raw-script input/provider body.
  Calls to `x:SieveSystemScript` or `x:SieveUserScript` are test-oracle failures. Do not
  implement message mutations here.

> **Current Rule acceptance:** docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `f199546513b758ca7b95328b5ccd7680009b732e`; lifecycle transition/review target `6735ae746c7acde657359e8a709cd897c4dce61c`, not the acceptance commit; Contract/Spec Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. 144-02D is docs-only accepted; 144-02F's three exact closure reviews, a later separate Rule plan and accepted Phase-145 engine-port/carrier remain required; no Rule TDD/source/schema/runtime/product/release authority.
> 145-07B additionally requires accepted engine-port carrier for typed IR/script bytes;
> renderer owns escaping, engine folder ID is not mailbox name, and attachment stays MIME-blocked.

## 145-07C — spam/ham classifier feedback

- Gates: selected-version official message-to-classifier-feedback API and native
  idempotency/reconciliation frozen under G145-5. The documented
  `x:SpamTrainingSample` object by itself does not satisfy this gate.
  Dependencies: 02/04. Allowed only `backend/internal/mail/stalwart/spam.go`,
  `backend/tests/mail/stalwart_spam_test.go`.
- Behavior/oracles: account-scoped unique bounded message ids and explicit `spam|ham` map
  plus stable receipt/message command ids map to documented feedback; per-message accepted/rejected/indeterminate outcomes remain
  distinct from mailbox move/flag. Timeout never becomes accepted, unsupported capability
  is typed rejected, cross-account ids cause zero call, and sample/provider/address data
  never enter logs/errors. G145-5 must evidence how the target mailbox message/blob and
  explicit spam/ham label enter the selected API and whether it provides native
  idempotency/reconciliation. Without both, this packet remains unimplemented; if a future
  accepted API can return indeterminate without reconciliation, that call is terminal
  unknown and the adapter/domain never calls feedback again for the command id.

## 145-08 — adapter composition

- Dependencies: 03A/B/C, 04, 05A/B, 06A/B and 07A/B/C. Allowed only `stalwart/adapter.go`,
  `backend/tests/mail/stalwart_adapter_contract_test.go`.
- Behavior/oracles: compile `MailEngine` assertion; every method, including hydrated
  domain/account, `Changes`, anchored `EnumerateProjection`, `GetSieve` and
  `ClassifyMessages`, delegates exactly once to its accepted
  mapper/client; no hidden fallback; account id required; capabilities cached only with
  bounded/invalidation policy; redacted health; constructor no network.

## 145-09A — mailbox adapter/worker composition and root call

- Dependencies: 08 plus accepted Phase 144 constructor surfaces. Allowed only
  `backend/cmd/server/wiring_mailbox.go` and `backend/cmd/server/main.go`.
- Behavior: `main.go` adds exactly one `wireMailbox(d)` call after `wireAuditMailAI(d)` and
  before module consumers. `wiring_mailbox.go` parses the accepted mailbox config,
  constructs exactly one Stalwart adapter only when enabled, and injects the generic
  `MailEngine` into the already accepted mailbox services/handlers. It then composes the
  accepted Phase 144 domain, account, combined sync/recovery and rule repository/service/workers, starts
  exactly one of each with `d.rootCtx`, and immediately registers bounded reverse-order `Stop`
  callbacks with `d.addCleanup`. Start or
  constructor failure is redacted, follows the frozen startup policy and never reports
  mailbox readiness. Disabled mode creates no adapter, worker, timer/goroutine or probe and
  preserves the explicit not-configured engine supplied by Phase 144.
- Oracles: exact pre/post root-call inventory; one engine instance and four workers;
  provisioning precedes projection eligibility; every start gets the shared root context
  and cleanup drains in LIFO order; start failure has no orphan scheduler; disabled zero
  adapter/worker/probe; no secret/cursor in DI/log; legacy
  `wireAuditMailAI`, Resend readiness/routes and `/admin/mail/settings` registration byte-
  semantically unchanged. This is composition, so no fake isolated Go unit is invented;
  handler/API/runtime proof remains in 09B and the accepted launcher gate.

## 145-09B — mailbox health service/handler mapping

- Dependencies: 09A plus accepted Phase 144 packets 06B/06C. Allowed only
  `backend/internal/mail/mailbox_admin_service.go`,
  `backend/internal/mail/mailbox_admin_handler.go`,
  `backend/tests/mail/mailbox_admin_health_test.go`.
- Behavior: map only the generic `MailEngine` health result wired by 09A. Disabled returns
  the frozen redacted `not_configured` resource without calling the engine; healthy enabled
  returns 200; timeout/auth/capability/unavailable results return redacted `degraded` and
  the contract-selected 503. The existing admin guard runs before any health probe or
  detail. Do not import `internal/mail/stalwart`, expose an endpoint URL/account id/provider
  body, or change mailbox policy/domain behavior.
- Oracles: non-admin and unauthenticated calls cause zero probe; each generic health class
  maps to the exact OpenAPI status/body; output contains no credential/URL/engine id;
  legacy Resend settings/readiness/broadcast service and routes remain unchanged.

## 145-10 — optional bounded deployment profile

- Gates: G145-1/3/4/5/5A/5B/5C and conditional G145-5D; dependencies 09A/09B. Allowed only
  `docker-compose.local.yml`, founder-approved production compose if explicitly in scope,
  `deployments/mail/**`, and names-only env example/docs named in dispatch.
- Behavior/oracles: exact immutable digest; optional `mail` profile; fixed service-network
  `/jmap` endpoint reachable only by Hub for separately authorized Registry and mailbox
  contexts; individually allowlisted `/api` helper path+method exposure only; separate secret
  references; explicit opt-in
  SMTP delivery ports and no user-facing
  IMAP/JMAP/ManageSieve/direct-admin mailbox mutation; named volume; image-specific health; non-root /
  capabilities where supported; CPU/RAM/PID/restart caps; no secret/value/DKIM key; compose
  static/runtime validation only through launcher. `STALWART_RECOVERY_ADMIN`, recovery mode
  and the HTTP configuration listener are absent from steady state; bootstrap/break-glass
  is operator-owned and never a Hub credential. The trusted exclusivity setting may be
  true only in that reviewed exposure shape. Do not invent volume path/command.

## 145-11A — engine backup/restore runbook

- Gates: G145-4/5; dependency 10. Allowed only
  `docs/runbooks/mail-engine-backup-restore.md`.
- Behavior/oracles: exact selected-version quiesce/capture/encryption/retention/isolated
  restore/JMAP verification/cleanup; current Hub PostgreSQL/object-storage backup is
  explicitly insufficient for the engine volume; exact named volume, config checksum and
  image digest are prerequisites. Commands contain no secret or broad destructive target;
  every unexecuted operation is labelled `NOT_RUN`.

## 145-11B — domain deliverability runbook

- Gates: G145-4/5 plus founder domain/operator decision. Allowed only
  `docs/runbooks/mail-domain-deliverability.md`.
- Behavior/oracles: manual MX/PTR/SPF/DKIM/DMARC/TLS, ownership, staged enablement,
  monitoring and rollback; reserved example domains/addresses only; no private key or
  secret; no claim that external delivery works. Every DNS/provider operation is labelled
  manual and `NOT_RUN` until operator evidence exists.

## 145-12 — accepted runtime gate

- Gates: G145-1 through G145-9. Allowed mutations only task-owned launcher report paths
  explicitly handed off by test owner; no product source change.
- Behavior/oracles: execute all 12 cases in `GATE-PLAN.md` with exact SHA/digest/config
  checksum, task resource names, CPU/RAM/PID limits, timeout, idle/load metrics, restart,
  lost response reconciliation, invalid cursor repair, crash/resume in enumeration and catch-
  up phases, command-lease begin/end interlock and expired-owner recovery, two-replica sync-
  lease exclusion, worker cancellation/drain, backup restore and
  unconditional cleanup. External SMTP is a separate opt-in report. Any missing cleanup/
  evidence => FAIL, not partial PASS. Direct compose is forbidden.
