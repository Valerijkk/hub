# TASK-224 — 🐳 E2E Docker baseline prerequisites

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: **CLOSED** · Фаза: **128**

> Historical closed-card provenance only. The recorded legacy baseline grants
> no current or future launcher acceptance, supported command, runtime
> readiness, or fallback. A future versioned launcher API belongs only to
> `testframework/runner/**`; `testframework/docker/**` supplies immutable
> execution assets only. Target launcher and runtime remain `NOT_READY` /
> `NOT_RUN` pending separate implementation and independent acceptance.

## [OUTCOME]

This closed card records a safe Docker Desktop contract for Hub
backend/frontend/auth seed that produced the historical legacy Playwright
smoke/full baseline. That evidence does not change target runtime from `NOT_RUN`
or authorize a new launcher.

## [ROLE]

- Owner: `devops-engineer` — application stack, readiness и seed contract.
- Verifier / handoff: `qa-auto-ver` — независимый smoke/full rerun и проверка
  cleanup/isolation.

## [SCOPE AND CONSTRAINTS]

- Historical scope allowed documented local test-entrypoint and Compose-overlay
  repair plus a non-secret seed harness only when necessary.
- Historical cleanup removed only exact task-owned containers, network, volume,
  and temporary image after evidence capture; shared services remained untouched.
  Production credentials and remote/deploy actions were forbidden.
- Depends on: TASK-202 and the TASK-203 red attempt of 2026-08-07.

## [ACCEPTANCE]

- Historical Hub app/auth prerequisites had readiness checks, deterministic
  non-production accounts, and a redacted environment-name contract.
- Historical `e2e-smoke` 3/3 and `e2e-full` 20 specs / 53 declared cases
  recorded pass/fail/skip/time and artifacts truthfully.
- Historical repetition did not depend on manual UI state; the no-service
  sentinel remained fail-closed.

## [DISCOVERY 2026-08-07]

- At discovery, the future target verdict was `NOT_READY` and execution was not
  run. Target runtime remains `NOT_RUN`.
- The then-current CI workflow and companion E2E backend helper were recorded
  as legacy execution authority: Docker started infrastructure and LiveKit, the
  backend used non-release DEV_MODE, and Playwright started the frontend dev
  server. This historical discovery is not current or future launcher authority.
- Required ready gates were health 2xx, DEV_MODE status, non-production seed
  2xx, and a reachable frontend.
- Full local Compose over current test infrastructure was forbidden without
  coordination because host ports conflicted. A shared destructive-stop helper
  was not used; only the task-owned backend process was ended.
- A historical PASS required a skip sentinel: environment-induced skips were
  RED and intentional AI/health skips used an explicit allowlist.

## [RESULT 2026-08-07]

- An isolated `hub-e2e224-*` Docker stack ran with explicit CPU/RAM limits and
  a deterministic non-production seed.
- Focused admin-users + Telegram slice: **3/3 PASS**.
- Smoke: **3/3 PASS**.
- First full run: **47 pass / 5 expected skip / 1 flaky**; the only retry was a
  Messenger locator.
- After the semantic locator repair, focused Messenger repeat: **3/3 PASS**;
  final full run with retries disabled: **48 PASS / 5 expected skip / 0 fail**.
- Product root causes fixed atomically: safe admin user-list API and stable
  Telegram routes without a configured bot. The Messenger issue was a test
  locator defect.
- Cleanup complete: task-owned containers, network, node-modules volume, and
  temporary frontend image were absent; shared deployment services remained.
- These historical results grant neither current nor future launcher acceptance.
  The target launcher remains `NOT_READY` and target runtime remains `NOT_RUN`.
