# Phase 143 code-runtime routing record

## 143-03B-T immutable source acceptance and conditional 143-04C gate (2026-08-20)

Calendar 143-03B-T is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `68cbf7cb227799304ce679937b519b66cb6abf9b` (parent `4715f939f24e5b5c2d3f80a38d204babe68bdd34`, tree `1537ba1d2d0fa804545fd8667ffe1dce9e8eaf1e`, subject `feat(calendar): add attendee tombstones`). The exact2 test-first blobs recorded in `IMPLEMENTATION-143-03B-T.md` are schema test `8bc4d44573e5a86e2c78df8c844fe3f1f2fae4cc` then `calendarattendee.go` `3c38799f3e56089e2e264b000ee0951a144b0257`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file saying 143-03B-T is only source-authorized, not accepted, dirty, uncommitted or unaccepted, or that 143-04C is blocked pending 03B-T source acceptance, is `CONSUMED/HISTORICAL` as to that dependency. Status: `COMBINED_EXACT19_SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. Only after one atomic exact19 docs commit and independent Contract/Security/Lifecycle immutable exact-SHA PASS for that commit may root dispatch the frozen 143-04C exact3, test first and in order: `backend/tests/calendar/event_service_update_test.go`, `backend/internal/calendar/event_repository.go`, new `backend/internal/calendar/event_service_update.go`. Until then exact3 remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and dispatch is not effective.

Mail 144-04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS` at `4715f939f24e5b5c2d3f80a38d204babe68bdd34`; 144-04G still requires separate docs authority because its repository port is absent. No adapter, Ent generation, migration, raw iTIP, service/API/UI/product/phase/release or runtime authority follows. Launcher `NOT_READY`; generation/migration/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; existing 143-04A/B/G receipts and other blockers remain unchanged.

## 144-04F immutable source acceptance (2026-08-20)

Mail 144-04F is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `4715f939f24e5b5c2d3f80a38d204babe68bdd34` (parent `7ac8714e99db2935ce290ad67836e1309cc05c78`, tree `6461894a90c4340bf96118ed1a242516c02b48c7`, subject `feat(mail): add sync page policy`). Its exact4 ordered test-first blobs are recorded in `IMPLEMENTATION-144-04F.md`: events test `50a41cefb539d9b840a43da9e1ff34cce6c8653f`, sync-policy test `b6410709e4fb8166c28a5c502cfba535eef08d06`, events source `a3017b0a6064d8fc1017e37e2d4f675ee8732945`, sync-policy source `acd642fcdb6290af79b7fa8e5aa36a5036576f47`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file asserting that Mail exact4 is dirty, `UNCOMMITTED`, `UNACCEPTED`, future-only, source-blocked, awaiting source review/commit, or root-dispatch-next is `CONSUMED/HISTORICAL`; the immutable source receipt above is current. No automatic Mail next source follows: 144-04G remains `NOT_AUTHORIZED / BLOCKED` pending separate docs authority because its required repository port is absent.

The Calendar half of the combined exact17 route is unchanged: 143-03B-T exact2 remains source-authorized but is not `STATIC_ACCEPTED`; any dirty candidate remains uncommitted/unaccepted. 143-04C exact3 stays `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; 143-04A/B/G receipts remain accepted. Launcher `NOT_READY`; tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`. 143-03E, 144-02E, Ent generation and migration stay blocked. No raw iTIP or service/API/UI/product/phase/release authority follows.

## Combined exact17 conditional acceptance sync (2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

> **Detected:** 2026-08-13
> **Scope:** atomic production/test code only
> **Planning/review:** continues
> **Current status:** ROUTE_AUTHORIZED · DISPATCHER_ENFORCED_PROVENANCE_ACCEPTED · RUNTIME_NOT_RUN

> **143-04B source receipt:** `75dafea889055126f768227a4ae62b28c91b7393` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; source `db34bbbc4b4c8a3ace8f768998060f82deecedbe`, test `79225fb73481ffd34ee1f3b9a72965d255e0b14a`, full receipt in `IMPLEMENTATION-143-04B.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 04B `PRECOMMIT_FAIL`/`NOT_ACCEPTED`/current-next clauses `CONSUMED/HISTORICAL`; no next Calendar code packet selected. Launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`; 03E Ent generation/migration blocked.

## Current route — task-local founder amendment

On 2026-08-13 the founder explicitly instructed this Mail branch to use Terra instead of
Codex for implementation. For Phases 143–148 on `codex/mail-module-20260813`:

- requested model: `gpt-5.6-terra`
- reasoning effort: `xhigh`
- fallback: none
- root `gpt-5.6-sol` remains orchestrator/reviewer and does not write atomic product/test
  code

This is a branch/workstream-specific amendment, not a global rewrite of Hub routing. Every
dispatch records its exact requested model/effort and provenance source. An explicit native
dispatch accepted without rejection records `dispatcher-enforced`; actual telemetry, when
exposed, must match. A mismatch, rejection, omitted model/effort or fallback remains
`BLOCKED`. The instruction does not authorize protected-path work, direct Docker fallback,
migration apply, push, PR, merge, deployment, DNS, license purchase or secrets.
The mandatory first-read `AGENTS.md` now records this exact narrow exception so a
fail-closed dispatcher does not have to choose between contradictory routes.

## Superseded route at detection

The prior founder lock required every atomic code task to run as:

- requested model: `gpt-5.3-codex` (5.3 Codex Spark)
- reasoning effort: `xhigh`
- fallback: none

At detection, Root Sol and Terra were not permitted substitutes.

## Attempts and evidence

1. Native collaboration dispatch explicitly requested `gpt-5.3-codex` / `xhigh` for
   Task 143-01. The dispatcher rejected it before execution:
   `Unknown model gpt-5.3-codex`; available models were only `gpt-5.6-sol` and
   `gpt-5.6-terra`.
2. Local exact CLI route used `codex-cli 0.147.0-alpha.6.6`, explicit
   `-m gpt-5.3-codex` and `model_reasoning_effort="xhigh"` with dispatch artifact
   `DISPATCH-143-01.md`.
3. The sandboxed attempt could not reach the API. Per execution policy it was retried with
   approved network access.
4. The approved-network attempt reached the API and returned HTTP 400:
   `The 'gpt-5.3-codex' model is not supported when using Codex with a ChatGPT account.`

CLI thread ids for provenance: `019ff842-5c0f-7782-a8af-adc755392634` (network-blocked)
and `019ff843-2b21-72e0-b91f-0c66590c1cba` (catalog rejection).

## Historical result

- prior requested model: `UNAVAILABLE` (no inference or fallback)
- provider: OpenAI Responses API reached on the second attempt
- code changes from dispatch: none
- tests/build/lint/validators: NOT_RUN
- Task 143-01 code status at that attempt: `BLOCKED_BY_EXACT_RUNTIME`

## Resolution

The prior blocker could be resolved only when either:

1. the native dispatcher exposes exact `gpt-5.3-codex` with `xhigh`; or
2. the ChatGPT-account Codex catalog supports the exact locked slug; or
3. the founder explicitly amends the model-routing lock with another exact route.

Condition 3 was satisfied by the explicit founder directive above. Code dispatch is now
authorized only as exact `gpt-5.6-terra` / `xhigh`. Several authoring and review packets
subsequently reported exact provenance; their evidence is recorded in the matching
`IMPLEMENTATION-*.md` files. Docker-backed verification remains separately `NOT_READY`
until the canonical launcher is accepted.

## 2026-08-13 per-packet telemetry regression and founder resolution

Several earlier Terra packets reported requested and actual model/effort and were reviewed
under that evidence. Later 143-03C and 143-03F review runtimes did not expose actual
telemetry. A clean repeat 143-03C provenance review, the first 144-01B authoring attempt and
the exact 143-03A hash-pair remediation attempt also explicitly requested the required
Terra route, reported actual `UNAVAILABLE`, used no fallback and stopped before reading or
writing files.

Founder question 29 is now answered. For this branch and Phases 143–148, an explicit native
collaboration dispatch with the exact Terra route and effort that the dispatcher accepts and
runs without rejection is sufficient provenance, recorded as `dispatcher-enforced`. Actual
telemetry, if exposed, must match; a mismatch, rejection, omitted model/effort or fallback
remains `BLOCKED`.

Consequences:

- 143-03A is `STATIC_ACCEPTED` on `31bfff8f58ba83d12b59600135503f98a6c92fbd`,
  143-03C is `STATIC_ACCEPTED` on `444c6f7b7b7b3d9b77bc9b24b32d1e13000ebbbc`,
  and 143-03B is `STATIC_ACCEPTED` at clean `dd346e31`;
- 143-03F is `STATIC_ACCEPTED` on `846cd00ecb243402877e18239c89a62bf07e8d14` after
  `SPEC PASS`/`QUALITY PASS` and final docs/whole-boundary review C/I/M `0/0/0`; the intermediate
  C/I/M `0/0/1` docs-sync finding is superseded/closed;
- 143-03D is `STATIC_ACCEPTED` on `10b4ceead119cb74a16b325b33f9121f60d7b1b5` after final independent C/I/M `0/0/0`; requested Terra/xhigh authoring and Terra/high reviews are dispatcher-enforced, actual telemetry `UNKNOWN`.
- 143-03E is `BLOCKED/NOT_RUN` on the accepted launcher `NOT_READY` and the separate founder-approved migration mechanism/gate; all-A does not establish this technical authority or accept an executed migration. 03F is satisfied.
- 143-04L is `STATIC_ACCEPTED` at `4525c559da22729aa8bc51ff5336c29dad4f27e9` after its
  historical initial C/I/M `0/2/2`, independently accepted immutable-envelope/payload amendment,
  and final Terra/high PASS C/I/M `0/0/0`. Implementation is Terra/xhigh; reviews are Terra/high;
  both use dispatcher-enforced provenance and actual telemetry is `UNKNOWN`. Static acceptance
  does not change the launcher `NOT_READY` or executable `NOT_RUN` status.
- 143-04G and 143-04B are `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact receipts are in their implementation records. Former 04G/04B next-route clauses are `CONSUMED/HISTORICAL`. No next Calendar code packet is selected; root must separately dispatch any later packet. All executable evidence remains `NOT_RUN`; launcher `NOT_READY`; 03E remains `BLOCKED/NOT_RUN` only for Ent generation/migration; unrelated blockers remain.
- 144-01B remains subject to its independent contract/source-review gates, but not to a
  telemetry-only provenance block;
- an exact Terra/xhigh worker may author atomic production/test code only through an accepted
  native dispatch; root only orchestrates, reviews, stages and commits; and
- Docker availability does not make the separately missing launcher ready (`NOT_READY`), and
  generation, migration, tests and all executable runtime evidence remain `NOT_RUN`. Actual
  model/effort telemetry for the current accepted dispatcher records remains `UNKNOWN`.

> **143-04C docs candidate:** clause-scoped precedence consumes only the earlier no-next
> Calendar packet route for a separately reviewed sequential 143-03B-T then 143-04C candidate.
> It is `ROOT_SOURCE_AUTHORITY_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED`, never runtime/source
> acceptance. Launcher `NOT_READY`, generation/migration and all executable evidence `NOT_RUN`
> remain unchanged. 03B-T scope is schema/index metadata only; later 04C alone carries
> inactive/re-invite/replacement service behavior, zero-call capability guarding and opaque
> all-and-only result echo rollback.
