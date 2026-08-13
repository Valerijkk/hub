# Phase 145 — executable gated packet contracts

> **Assembly rule:** Common envelope + exactly one packet. No packet is currently eligible.
> **Runtime for code:** `gpt-5.3-codex` / `xhigh`, fallback none.
> **Authority:** `GATE-PLAN.md`; a table row here never self-satisfies a gate.

## Common execution envelope

- Root supplies exact clean SHA, selected licensed engine, immutable image digest and the
  version-specific official API/config evidence required by the packet. Missing or stale
  evidence => `BLOCKED`, no guessing from model memory.
- One behavior/allowed file set, tests first with bounded fake HTTP/JMAP server only where
  allowed; no commit/push/PR. Do not revert concurrent work.
- Never read/change protected testframework/test-rebase/pro-qa-ai zones. Existing Resend
  sender/settings/status/broadcast implementation is a separate compatibility boundary.
- No client-controlled base/capability URL, redirect, engine account id or credential.
  Bound response bytes/count/depth/time. Redact auth/session/provider bodies/addresses.
- No host test/build/lint/validator/generator/process. Direct Docker/Compose not accepted
  evidence. Runtime only through accepted task-owned launcher with limits/timeout/owner
  report/finally cleanup.
- Code response in Russian: files/symbols/oracles/gates/verification/model provenance/SHA.
  depth/fan-out 0, retries 2, 60m (90 for deployment/runbook/runtime).

## 145-01A — runtime config

- Gates: G145-2; accepted Phase 144 port. Allowed only
  `backend/internal/mail/mailbox_config.go`,
  `backend/tests/mail/mailbox_config_test.go`.
- Behavior/oracles: disabled default needs no secret/domain and reports not_configured;
  enabled requires fixed internal absolute http(s) URL plus admin identity/secret;
  external requires enabled + domain; validate byte/rate/retention bounds; errors/string
  dumps redact values; no environment read inside pure parser (map/lookup injected).

## 145-01B — redacting bounded transport client

- Gates: G145-1/2/4/5; dependency 01A. Allowed only
  `backend/internal/mail/stalwart/client.go`, `errors.go`,
  `backend/tests/mail/stalwart_client_test.go`.
- Behavior/oracles: injected `http.Client`, fixed base and same-origin resolved endpoints;
  redirects disabled/validated; auth generated internally; request/response/time/count
  bounds; context cancellation; stable mapping of 4xx/5xx/timeout; logs/errors omit header,
  URL query, provider body/ids. No real network in unit.

## 145-02 — JMAP session/capabilities

- Gates: G145-5; dependency 01B. Allowed only
  `backend/internal/mail/stalwart/jmap.go`, `config.go`,
  `backend/tests/mail/stalwart_jmap_session_test.go`.
- Behavior/oracles: exact selected-version session JSON; require mail/submission/core and
  optional sieve capabilities explicitly; derive URLs only same-origin/fixed-session
  policy; bound accounts/capabilities; missing/unknown required capability fails closed;
  no credential/session URL in error.

## 145-03A — account/identity mapping

- Gates: prior; dependencies 02 + Phase 144 port. Allowed only
  `backend/internal/mail/stalwart/account.go`, `mapper.go`,
  `backend/tests/mail/stalwart_account_test.go`.
- Behavior/oracles: idempotent EnsureAccount by stable Hub identity; normalized primary/
  alias mapping; no group send until authorized; opaque returned ids; conflict/partial
  provisioning safe; no password/token persisted/returned; cross-account query impossible.

## 145-03B — folder mapping

- Dependency: 02. Allowed only `stalwart/folder.go`,
  `backend/tests/mail/stalwart_folder_test.go`.
- Behavior/oracles: selected-version mailbox get/set mapping, roles, parent ids, counts and
  state; account id in every request; deterministic pagination; unsupported hierarchy /
  partial failure typed; no Hub color/order sent as engine authority unless port says so.

## 145-04 — message/blob/list/search mapping

- Dependencies: 02/03B. Allowed only `stalwart/message.go`, `blob.go`,
  `backend/tests/mail/stalwart_message_test.go`.
- Behavior/oracles: exact selected-version properties; bounded ids/page/body/blob; body/
  MIME exists only in returned request-lifetime values; received BCC absent; folder/flags/
  state mapped; account scoping; malformed/missing property safe; no raw response error.

## 145-05 — changes and invalid cursor

- Dependency: 04. Allowed only `stalwart/changes.go`,
  `backend/tests/mail/stalwart_changes_test.go`.
- Behavior/oracles: opaque since/new state, bounded hasMore loop contract, created/updated/
  destroyed ids; expired/cannotCalculateChanges maps typed invalid-cursor; it never returns
  empty-success cursor; caller can mark degraded + bounded rebuild; duplicate page stable.

## 145-06A — draft save/reconciliation

- Dependencies: 04/05 + Phase 144 draft port. Allowed only
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
- Behavior/oracles: account-scoped move/flags/delete; partial per-id errors; hard-delete
  policy explicit; returned opaque state is reconciled; another account's ids produce no
  call/mutation. Do not implement or upload Sieve in this packet.

## 145-07B — typed Sieve mapping

- Dependencies: 02 plus accepted Phase 144 rule compiler. Allowed only
  `backend/internal/mail/stalwart/sieve.go` and
  `backend/tests/mail/stalwart_sieve_test.go`.
- Behavior/oracles: accept only the compiler-produced typed representation; render the
  selected-version script deterministically; bound name/bytes/rules; upload then activate
  with explicit partial-failure state; expose no public raw-script input. Do not implement
  message mutations in this packet.

## 145-08 — adapter composition

- Dependencies: 03–06 plus 07A/07B. Allowed only `stalwart/adapter.go`,
  `backend/tests/mail/stalwart_adapter_contract_test.go`.
- Behavior/oracles: compile `MailEngine` assertion; every method delegates exactly once to
  accepted mapper/client; no hidden fallback; account id required; capabilities cached only
  with bounded/invalidation policy; redacted health; constructor no network.

## 145-09A — mailbox adapter composition and root call

- Dependencies: 08 plus accepted Phase 144 constructor surfaces. Allowed only
  `backend/cmd/server/wiring_mailbox.go` and `backend/cmd/server/main.go`.
- Behavior: `main.go` adds exactly one `wireMailbox(d)` call after `wireAuditMailAI(d)` and
  before module consumers. `wiring_mailbox.go` parses the accepted mailbox config,
  constructs exactly one Stalwart adapter only when enabled, and injects the generic
  `MailEngine` into the already accepted mailbox services/handlers. Disabled mode creates
  no adapter, performs no probe/network call and preserves the explicit not-configured
  engine supplied by Phase 144. Constructor failure is redacted and follows the frozen
  startup policy; no background probe is invented.
- Oracles: exact pre/post root-call inventory; one engine instance; disabled zero adapter/
  probe; no secret in DI/log; legacy `wireAuditMailAI`, Resend readiness/routes and
  `/admin/mail/settings` registration byte-semantically unchanged. This is composition,
  so no fake isolated Go unit is invented; handler/API/runtime proof remains in 09B and the
  accepted launcher gate.

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

- Gates: G145-1/3/4/5; dependencies 09A/09B. Allowed only
  `docker-compose.local.yml`, founder-approved production compose if explicitly in scope,
  `deployments/mail/**`, and names-only env example/docs named in dispatch.
- Behavior/oracles: exact immutable digest; optional `mail` profile; fixed internal JMAP/
  admin; explicit opt-in SMTP/IMAP ports; named volume; image-specific health; non-root /
  capabilities where supported; CPU/RAM/PID/restart caps; no secret/value/DKIM key; compose
  static/runtime validation only through launcher. Do not invent volume path/command.

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
  lost response reconciliation, invalid cursor repair, backup restore and unconditional
  cleanup. External SMTP is a separate opt-in report. Any missing cleanup/evidence => FAIL,
  not partial PASS. Direct compose is forbidden.
