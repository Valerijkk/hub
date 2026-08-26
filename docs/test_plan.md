# Hub — План тестирования

> Единая база знаний QA. Обновляется при изменении стратегии, CI или coverage floor.  
> Детали TC-*: `docs/TESTING.md` · Матрица: `.planning/COVERAGE.md` · Triage: `.cursor/skills/test-triage/SKILL.md`

> **Founder greenfield cutover 2026-08-07:** сохраняются и перемещаются только
> доказанные genuine isolated Go/Vitest/RTL units в `backend/tests/**` и
> `frontend/tests/**`. Legacy Go component/API/integration, static-contract
> Vitest, TypeScript Playwright и их support/config/entrypoints удаляются по
> exact allowlist и не портируются. Весь non-unit и load слой создаётся с нуля
> как один Python-проект прямо в
> `testframework/`: `{src/hub_testkit,tests/{api,integration,contract,e2e},selftests,cases,load,docker,reports}/**`
> плюс корневой `pyproject.toml`/pytest config. Исторический 82% Go coverage и
> row-level parity не блокируют cutover; новый набор принимается по рискам,
> требованиям и трассируемости. Канон:
> [test architecture rebase](../.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

---

> **Variant A current state (2026-08-20):** all tracked legacy test files in
> `backend/testframework/**` were removed. Any legacy path, hook, count, coverage result, or job name
> below is historical inventory only. Current backend units are exclusively
> `backend/tests/**` and execute through root Compose `backend-unit`. That
> target also has one closed framework-command boundary: with `hubtest && qa`,
> it compiles and runs only isolated CLI contract packages
> `backend/cmd/qa-{secret-init,schema,backend}/`. They are neither genuine
> product units nor legacy/component/API coverage, and do not permit arbitrary
> `cmd/**` tests.

## Пирамида тестирования

```
        ┌─────────────┐
        │     E2E     │  pytest + playwright-python — критические user paths
        ├─────────────┤
        │ Integration │  pytest — DB/broker/storage/application boundaries
        ├─────────────┤
        │    Unit     │  Go testify + Vitest/RTL, service-free
        └─────────────┘
```

| Уровень | Инструмент | Расположение | Gate |
|---------|------------|--------------|------|
| Unit (BE) | Go testing + testify; closed QA CLI contracts | genuine product units: `backend/tests/{domain}/`; QA-only `hubtest && qa`: `backend/cmd/qa-{secret-init,schema,backend}/` | local Docker `backend-unit` (`build`, `vet`, `unit`, `race` PASS) |
| Unit (FE) | Vitest + RTL | `frontend/tests/{domain}/`, support | local Docker `frontend-unit` PASS |
| Framework selftests | pytest | `testframework/selftests/` | local Docker `python-contract` PASS |
| API / Integration / Contract | pytest + httpx | `testframework/tests/{api,integration,contract}/` | `NOT_RUN`; not a minimal merge gate |
| E2E | pytest + playwright-python | `testframework/tests/e2e/` | `smoke` marker; `full` launcher profile |
| Load | Locust | `testframework/load/` | nightly/manual |
| Build | `tsc`, `vite build`, `-tags release` | CI `build` job | 0 errors |
| Lint | Biome | `npm run lint` | **gate** (не continue-on-error) |
| Security | grep + audit docs | `.planning/audits/code-review-final/` | 0 CRITICAL |

Единственный целевой слой:

| Уровень | Инструмент | Расположение |
|---------|------------|--------------|
| Unit (BE) | Go testing + testify | `backend/tests/{domain}/` — только genuine units |
| Unit (FE) | Vitest + RTL | `frontend/tests/{domain}/` — только units; support в `frontend/tests/support/` |
| API / integration / contract | pytest + httpx | `testframework/tests/{api,integration,contract}/` |
| E2E smoke/full | pytest + playwright-python | `testframework/tests/e2e/`; `smoke` is a marker and `full` is a launcher profile |
| Load | Locust | `testframework/load/` |

Текущие **237** `go-component` rows (31 API + 206 integration) не переносятся в
`backend/tests/**` и не переписываются один-в-один: source удаляется по manifest,
а новый Python coverage проектируется заново от требований и рисков.

## Definition of coverage для greenfield Python

Каждый Python test module обязан иметь ссылки на `task`, requirement (`FR/NFR`)
и стабильный `TC-*`. Для применимых требований тест-дизайн явно покрывает:

- happy path и основной пользовательский результат;
- invalid input, authorization/authentication и forbidden transitions;
- min/max/empty/null/Unicode/timezone/concurrency boundaries;
- edge/corner cases, idempotency, retry/timeout и partial failure;
- recovery/rollback, security, privacy, accessibility и performance risk;
- observable side effects и отсутствие запрещённых side effects.

Количество старых тестов и старый statement coverage не являются целевыми KPI.
Zero collection, неожиданные skips и тест без traceability fail closed.

---

## Test Triage Protocol (REGULATION §10)

При падении любого target test/job:

1. **Воспроизвести функционал** по `docs/system_specification.md`, не только assert
2. **Классифицировать ровно один root:** `code | test | infra | flaky | stop`
3. **Действовать по root:** минимальный code/test fix; восстановление infra;
   доказательство flake; либо STOP без произвольной правки
4. **Записать** root, evidence и решение в QA log / STATE.md

**Запрещено:** зелёные тесты без проверки соответствия спецификации.

**Docker-only:** локальные агенты выполняют tests, validators, lint,
build/vet, Vite/backend/frontend, Playwright, pytest и Locust только внутри
task-owned Docker containers с CPU/RAM limits. Host-run не является baseline и
обязан быть повторён в Docker; host оставлен для редактирования и Git.

---

## Текущая инвентаризация (2026-08-07)

Источник: `.planning/audits/test-rebase/{inventory.json,founder-greenfield-cutover-manifest.json}` · HEAD
`b9f690da`; ledger SHA-256 `83b6a311…` (ротация 2026-08-13).

| Метрика | Значение |
|---------|----------|
| Ledger | **1,049** source rows / **4,246** declared; drift/unmapped = 0; historical inventory, not a current runnable-test count |
| Backend Go total | **832** файлов / **3,674** declared: legacy `backend/testframework/` 831/3,672 + server 1/2; taxonomy 595 genuine unit + 237 component (31 API + 206 integration) |
| Backend `internal/*_test.go` | **0** (TF-M08 ✅) |
| Backend hook sources | **43** files / 582 symbols / 892 links; 245 genuine units protected until TASK-234, prune only TASK-219B |
| Hook-dependent protected subset | **42** distinct hook files / **434** distinct referenced hook symbols / **653** consumer links across the 245 protected units; subset-only, so it does not conflict with the global hook inventory above |
| Statement coverage `internal/` | **82.0373%** — historical Docker baseline for context, не migration gate |
| Security-critical packages ≥80% | Historical baseline 6/6; не заменяет risk/requirement coverage нового набора |
| Frontend Vitest | **180** files / **519** literal cases: 156 unit + 24 static contract |
| Frontend support | **3** source rows → `frontend/tests/support/**` |
| Playwright specs | Legacy **20** files / **53** declared; Docker full: **48 pass / 5 expected skip / 0 fail**, затем delete-not-port |
| Playwright support | Legacy **14** config/shared rows → delete-not-port; новый Python support проектируется greenfield |
| Target Python tests | **0** до Phase 129 |
| Lighthouse CI | artifact job (master push, non-blocking) |
| TypeScript | Docker `tsc --noEmit` + Biome PASS; legacy Vitest 510 pass / 9 fail, remediation отменена founder cutover |
| Data races | 0 (CI `-race`) |

### Целевой runner после greenfield scaffold

Host не запускает test/build/lint процессы. Единственный QA API —
`docker compose -f docker-compose.local.yml run --rm <exact-service>` для
`backend-unit`, `frontend-unit`, `python-contract`, `python-api`,
`python-integration`, `python-e2e-smoke`, `python-e2e-full`, `python-load`.
QA services disposable; integration/e2e используют изолированную инфраструктуру.
Результаты запуска в этом документе не заявляются. Legacy runners — только
источник inventory; использовать их как fallback запрещено.

---

## Current minimal merge gate

Hosted GitHub CI/PR automation is manual-only and GitLab pipelines are off.
Canonical evidence is local Docker through exact root Compose services. This
slice requires only `backend-unit`, `frontend-unit` and framework-only
`python-contract`; API/integration/e2e/load remain `NOT_RUN` and are not merge
gates. Production deploy is a separate manual workflow on current `master`
after local Docker evidence, `FA-06/PG-03`, CTO sign-off and
`confirm_production`; stale SHA is rejected before reset/deploy.

**Historical/forbidden legacy inventory:** `frontend/playwright.config.ts` once selected TypeScript specs under `testframework/stages/e2e/domains/{domain}/`; this is not target execution authority.
Module resolution: `frontend/scripts/link-testframework-node-modules.mjs`  
Pipeline manifests: `testframework/pipeline/jobs/*.yaml`  
Infra: `scripts/ci-e2e-backend.sh` (postgres:17, redis:7.4, mongo:8, nats:2.11).

Эта таблица — inventory, не исполняемый gate. Entry-points не переносятся в
новый Python framework. Cutover заменяет их
job-ами `backend-unit`, `frontend-unit`, `python-contract`, `python-api`,
`python-integration`, `python-e2e-smoke/full`, `python-load`.

### e2e-full spec list (CI)

`login`, `critical-gate`, `navigation`, `logout`, `health`, `admin-audit`, `admin-ai`, `admin-metrics`, `ai-hub`, `automation`, `metrics`, `command-palette`, `shortcuts`, `onboarding-profile`, `notifications`, `project-lifecycle`, `messenger`, `tms`, `a11y`, `v4-superapp`.

---

## Критерии приёмки (общие)

1. Все CI jobs green на PR
2. Новый код покрыт unit-тестами (граничные случаи, error paths)
3. Нет регрессии по `docs/security_guidelines.md`
4. API changes → `docs/system_specification.md` + `docs/wiki/API.md`
5. User-facing critical path → E2E spec или расширение существующего
6. Docs sync: `STATE.md` + матрица из rule 55-docs-sync

---

## Матрица модулей (краткая)

Покрытие unit — **2026-06-22 Docker baseline** (`.planning/COVERAGE.md`).

| Модуль | Unit | E2E | Security |
|--------|------|-----|----------|
| auth | ✅ 82.4% (2026-06-22) | login, logout | SEC-06, SEC-07, magic-link timing |
| tracker | ✅ 82.4% (2026-06-22) | project-lifecycle | optimistic lock |
| messenger | ✅ 82.3% (2026-06-22) | messenger.spec | NF-02 WS authz |
| tms | ✅ 82.0% (2026-06-22) | tms.spec | webhook API key |
| calls | ✅ | partial | SEC-02 webhook sha256 |
| ws | ✅ | real-time | SEC-05, NF-02 |
| search | ✅ ~92% (2026-06-22) | command-palette | NF-03 user scope |
| apigateway | ✅ | — | SEC-01 API key order |
| notification | ✅ ~82% (2026-06-22) | notifications | NF-07 WebPush URL |
| v4 (crm, hr, intranet…) | ✅ | v4-superapp | RBAC |

Полная матрица: `.planning/COVERAGE.md`.

---

## E2E — текущие legacy критические пути

| ID | Сценарий | Spec | Статус |
|----|----------|------|--------|
| E2E-01 | Magic Link / dev login | `login.spec.ts` | ✅ |
| E2E-02 | Create project + issue | `project-lifecycle.spec.ts` | ✅ |
| E2E-03 | Messenger send | `messenger.spec.ts` | ✅ |
| E2E-04 | TMS flow | `tms.spec.ts` | ✅ |
| E2E-05 | Voice call | — | ⚠️ partial (LiveKit infra) |
| E2E-06 | Ctrl+K search | `command-palette.spec.ts` | ✅ |
| E2E-07 | a11y axe | `a11y.spec.ts` | ✅ |

Greenfield Python browser lifecycle/artifacts: task-named modules in `testframework/src/hub_testkit/browser/**`; domain UI pages/components: `testframework/src/hub_testkit/domains/<domain>/ui/{pages,components}/**`.

---

## Техники тест-дизайна

- **Классы эквивалентности:** valid/invalid email, UUID, issue types
- **Граничные значения:** rate limits (4-й auth → 429), WIP limits, file 25MB
- **Попарное тестирование:** filter combinations (status × priority × assignee)
- **Таблицы решений:** workflow transitions, RBAC permission matrix

---

## Security-critical тесты (приоритет)

| ID | Модуль | Требуемый тест |
|----|--------|----------------|
| SEC-01 | apigateway | API key middleware order → 401 without key |
| SEC-02 | calls | Webhook reject missing/tampered sha256 |
| SEC-05 | ws | send after disconnect — no panic |
| SEC-06 | auth | concurrent register / first-admin lock |
| SEC-07 | auth | `ConstantTimeCompare` on magic link |
| NF-02 | ws | non-member subscribe → reject |
| NF-03 | search | results scoped to user projects/channels |
| NF-07 | notification | WebPush invalid endpoint → 400 |

Статус fixes: `docs/security_guidelines.md` (12/12 CRITICAL fixed).

---

## Баг-репорты

```
ID: BUG-<module>-<NN>
Title: краткое описание
Steps: 1. ... 2. ... 3. ...
Expected: ...
Actual: ...
Severity: CRITICAL/HIGH/MEDIUM/LOW
File: path:line
```

Трекинг: `.planning/audits/code-review-final/SUMMARY.md`

---

## Связанные документы

| Документ | Содержание |
|----------|------------|
| `docs/TESTING.md` | TC-AUTH-01…, API code matrix, manual checklist |
| `docs/stack_and_tools.md` | Версии, CI jobs |
| `docker-compose.local.yml` | Единственный QA service contract |

---
*Последнее обновление: 2026-08-07 — founder greenfield cutover: genuine
Go/Vitest units only, exact-manifest legacy cleanup в Phase 129 и новый root
Python QA project без parity/dual-run gate.*

## Phase 151 bounded Messenger evidence (2026-08-26)

The following Docker-only gates are accepted for the current bounded slice:

- `docker compose -f docker-compose.local.yml run --rm backend-unit` — overall `PASS`; build, vet, unit, and race receipts `PASS`.
- `docker compose -f docker-compose.local.yml run --rm frontend-unit` — overall `PASS`; TypeScript `PASS`; Vitest `187/187` files and `577/577` tests `PASS`.

Covered regressions include the read-only channel discovery route order, same-tenant/project filtering, fresh recipient projection, denied-kind omission, and whole-request failure on unexpected repository/authority errors; lazy UI fetch, identity A-to-B cache isolation, 401/403 no-retry payload eviction, stale-refetch hiding, public/can-read defensive filtering, server order, neutral states, and no row navigation/membership action; one-proof-per-distinct-channel bookmark authorization fully before Mongo, source-order preservation, cross-channel preview no-leak, user-keyed caches, disabled no-user sentinel masking, in-flight removal fencing, and foreign-channel pin-row omission after fresh list authority, unexpected `FindPin` failure fencing with zero AddPin/RemovePin, pinned-list A/B key isolation, generic/sentinel no-user masking, channel binding, in-flight unpin completion fencing, and loading/error-retry/successful-empty state separation, list/pin capability deny-by-default, click-time store loss, revoke-during-unpin, and current-DTO downgrade fencing, exact row-pin invalidation, no-user/identity-switch fencing, DTO/store capability loss, and revoke-during-row-toggle fencing; root timeline user-key isolation, delayed A/B responses, no-user fetch/room/data masking, exact WS/reconnect/generic invalidation, optimistic-edit/pinned-preview scoping, raw sentinel-data suppression, paged-reply and production-detail thread A/B isolation, no-user raw/render masking, distinct cache schemas, and schema-aware identity cleanup; fresh recipient capability projection for Create/Update channel, new/existing DM, and AcceptInvite responses; fresh WS room/user/audience authority; empty authorized unread scope with zero Mongo; authorization-before-Mongo for reactions and both forward resources; target `CanPost`; zero side effects on denial; ordinary-author own-delete versus privileged delete-any; authorized rejoin restore; target filtering; identity/revoke barriers; revoke during an in-flight forward; and the unbound pure post-create full-payload/deep-copy/server-field/invalid-before-repository contract.

Additional ThreadPanel convergence regressions cover parent-only detail versus canonical paged replies, initial-fetch and resident-refetch `message_sent` races, 101-event sent overflow recovery, edit/delete/reaction canonical reconciliation, a deferred recovery window in which edited/deleted state must not transiently revert, 101-edit mutation-journal overflow with raw/flattened fail-closed masking, split loading, exact root/thread invalidation, principal-bound retry completion, access-loss reset, and identity/revoke-scoped thread issue links.

Backend thread issue-link list regressions prove actor forwarding and exact `FindChannelInProject → IsProjectMember → IsChannelMember → link read` order with bound project/channel/actor IDs. They cover identical 404 concealment for global/foreign/deleted channels, inactive/non-project actors, and channel non-members; generic 500 on each proof infrastructure failure; zero later-proof/link reads after every denial/error; link-store failure; and route-project row filtering. TDD includes the actor-signature RED plus a focused fresh-ProjectMember order/deny/error RED, followed by Docker backend overall/build/qa-build/swagger-check/vet/unit/race PASS and independent security/code/test-adequacy PASS. Create/unlink policy, client capability controls, atomic ProjectMember/ChannelMember+list TOCTOU closure, and live DB execution remain unverified.

Task 128 OpenAPI regressions parse `swagger.json`, the embedded `docs.go` template, and `swagger.yaml` through one order-insensitive semantic contract. Coverage requires legacy-route absence, scoped promotion `POST` parameters/body/`201`, unread response reference, four `integer/int64` maps, and `any_unread:boolean`. The accepted sequence is legacy-path RED → parser/suffix/mount/coverage/order triage → fresh backend build/qa-build/swagger-check/vet/unit/race PASS plus independent security/code/test PASS. This is static checked-in-source evidence only: no generator or live service ran, and source-annotation regeneration remains separate.

Task 98 promotion-source regressions prove the exact scoped `(channel A, message ID)` lookup, zero generic source-ID reads, one happy continuation, indistinguishable missing/foreign/deleted 404 responses, redacted 500 responses for absent capability/store/invariant failures, and zero downstream target/promoter/link/backlink calls before source success. The normal test double delegates to the same canonical live/channel predicate used by the repository mocks; explicit override is limited to injected store and invariant failures. TDD recorded the initial localized RED, one test-double triage failure, an independent false-PASS review finding, remediation, and final Docker backend build/vet/unit/race PASS. Live Mongo execution and post-read deletion TOCTOU remain `NOT_RUN`/`UNBOUND`.

Task 99 promotion-target regressions prove parse/source/target order, exact actor/project binding, concealed unavailable target, forbidden fresh deny, redacted dependency/query/invariant failures, and zero downstream work until `(true,true,nil)`. RBAC decision tests cover default project roles, custom `permission_names` precedence, empty-name legacy-mask fallback, founder/admin bypass only after target proof, missing membership deny, malformed facts, repository error, and stale-cache non-authority. TDD recorded compile RED, one full GREEN, final test-review remediation, and a fresh post-remediation Docker build/vet/unit/race PASS with independent security/code/test PASS. Real Ent/PostgreSQL foreign/archived/deleted/missing/live target execution remains `NOT_RUN`; sequential authority-to-mutation TOCTOU remains `UNBOUND`.

Biome remains a non-gating repository-wide failure with 369 formatting diagnostics. Vite remains a non-gating failure at `vite.config.ts` because `__dirname` is undefined in ESM. The official GSD reviewer was blocked by workspace ACLs before repository access; root manual review and summary-only independent reviews are recorded instead; the final exact-scope root timeline security re-review returned PASS after its raw-data concern received an additional RED/green cycle, and the separate thread identity/cache-schema security review returned PASS after its own two RED/green stages. The final ThreadPanel convergence security, code-lifecycle, and test-adequacy reviews also returned PASS after principal retry fencing, issue-link scoping, fetch-aware sent/mutation journals, transient overlay, and overflow-mask regressions. Membership lifecycle and canonical post-create/recovery remain `UNBOUND`. Discovery is unit-covered only: live Ent/PostgreSQL behavior and a 101-plus-channel pagination path are `NOT_RUN`; the endpoint intentionally exposes only a maximum-100 seed with no cursor/completeness guarantee. Live Mongo/NATS/WS/browser, `python-integration`, cutover, and Jitsi runtime are `NOT_RUN` and must not be inferred from the unit gates.

## Phase 151 promotion parent/reply evidence (2026-08-26)

Fixture-triaged regressions prove that, after fresh route/source and target proofs, promotion uses a channel-scoped live parent lookup and private `channel_id + thread_id(rootID) + deleted_at:nil` reply lookup before GetChannel/promoter/backlink/linker. They require validation of every returned row, forbid generic parent/reply fallback, and verify deterministic `created_at DESC, _id DESC` selection capped at 100. Fresh Docker `backend-unit` overall/build/vet/unit/race and independent code/security/test/final reviews are `PASS`; qa-build/swagger are not evidence for this atom. Live Mongo/Ent-PostgreSQL/NATS/WS are `NOT_RUN`; source-ID defensive invariant, cross-store delete/revoke TOCTOU, source reply outside newest 100, chronology policy, mutation-time target recheck, and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.
