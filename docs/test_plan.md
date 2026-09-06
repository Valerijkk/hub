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

Reaction lifecycle (2026-09-06):25 genuine UI cases связывают prop/live rights
и текущий scope с одним HTTP toggle, canonical failure и picker open-session.
Полный составной frontend-unit:TypeScript+199files/809testsPASS,118.81s,
без skipped/todo/failures или whole-run unhandled errors. Window rejection
spy не является самостоятельным Node gate; сохранены реальные RED receipts.
Browser/API/WS остаются NOT_RUN.
[Контракт](superpowers/plans/2026-09-06-message-reaction-lifecycle.md).

Edit lifecycle (2026-09-06): 24 новых genuine UI units проверяют admission,
single-flight, optimistic key, current failure и stale/ABA settlement,
сброс draft при потере scope. Принято только на unit-уровне в составе
TypeScript + 758 tests PASS; серверный порядок/rollback и browser NOT_RUN.
[Контракт](superpowers/plans/2026-09-06-message-edit-lifecycle.md).

SEND recovery (2026-09-06): два новых single-flight сценария и четыре
REAL-Toaster сценария дополняют nonce/snapshot oracle; callback mock сам по
себе не подтверждает настоящие Retry/Cancel. Принято в составе TypeScript +
758 unit tests PASS; API/storage/browser gate этим не заменяется.
[Контракт](superpowers/plans/2026-09-06-composer-retry-lifecycle.md).

Resource envelope2026-09-06: максимум2/minimum1 worker в обеих Vitest pool
families задаётся только в canonical frontend-unit service. При прежних
2GiB/no-swap/6CPU и неизменных assertions/таймаутах составное дерево
resource+edit+composer прошло TypeScript и197files/758tests за114.95s.
Тесты не исключались, skipped/todo/failed0, unhandled errors отсутствуют.
Это composed-tree receipt, не проверка промежуточных коммитов по отдельности.
Исторические смешанные failures и worker errors остаются в отчётах.
[Resource contract](superpowers/plans/2026-09-06-frontend-unit-worker-budget.md).

Дополнение `calls-feature-codex` (2026-09-05): MSG151-FOL-01/02 и CHN-02
имеют 29 изолированных ThreadPanel follow-control сценариев с реальным QueryClient
и mock HTTP-границами. Приёмка: TypeScript PASS, 190 файлов / 630 тестов PASS в
canonical Docker frontend-unit с 2 GiB/no-swap. Режимы/права, auth/revoke,
дубликаты, changed=false, RU/EN error/retry и post-dispatch refetch race
проверены; видимое прочтение, browser/API/storage/WS/media — отдельные gates.
Результаты и source hashes: [контракт](superpowers/plans/2026-09-05-thread-follow-control.md).

Дополнение2026-09-06: read-through UI покрыт36 новыми genuine unit-кейсами
через реальные ThreadPanel/Query/hooks/stores и mock HTTP/IntersectionObserver.
Курсор продвигается через старые ответы по видимому концу выбранного ответа;
это не доказательство чтения каждого старого сообщения. Oracle проверяет
focus/visibility, full-intersection, fresh gates, transport/refresh separation,
single-flight/pending/watermark и устаревшие scope callbacks. Финальный Docker
frontend-unit: TypeScript PASS,191 файлов/666 тестов PASS,58.20s,2GiB/no-swap.
Геометрия настоящего браузера, backend/API/DB и пользовательские сценарии
по-прежнему требуют отдельных Docker SUT-gates.
[Контракт и triage](superpowers/plans/2026-09-05-thread-read-receipts.md).

Следующий принятый составной gate2026-09-06: TypeScript PASS,
195 файлов/728 тестов PASS,64.12s,2GiB/no-swap. Добавлены35 genuine search
случаев (real Dialog/Query + local session hook),20 delete lifecycle
(real MessageItem/Actions + HTTP/toast boundaries),7 thread capability
propagation (real owner or panel, не mock тестируемой границы).
Отрицательные и ABA/late-completion случаи входят в эти62 tests.
RED684/32 с3 uncaught DELETE errors сохранён; GREEN не ослабляет oracle.
Это unit acceptance составного дерева, не runtime proof промежуточных
коммитов, browser keyboard, API прав, хранилищ или полной готовности.
[Search evidence](superpowers/plans/2026-09-06-search-dialog-lifecycle.md),
[delete](superpowers/plans/2026-09-06-message-delete-lifecycle.md),
[thread wire](superpowers/plans/2026-09-06-thread-capability-wiring.md).

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
| Unit (BE) | Go testing + testify; closed QA CLI contracts | genuine product units: `backend/tests/{domain}/`; QA-only `hubtest && qa`: `backend/cmd/qa-{secret-init,schema,backend}/` | historical Docker baseline `PASS`; hard2g candidate is `STATIC_ACCEPTED / EXECUTABLE_NOT_RUN` |
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

For the hard2g `backend-unit` candidate, the observer is managed separately
and does **not** launch Compose: start it, wait for `READY`, then separately
issue exactly once `docker compose -f docker-compose.local.yml run --rm backend-unit`.

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

The following Docker-only gates are historical accepted baselines for this
bounded slice; they are not the executable canon of the hard2g candidate:

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

Fixture-triaged regressions prove that, after fresh route/source and target proofs, promotion uses a channel-scoped live parent lookup and private `channel_id + thread_id(rootID) + deleted_at:nil` reply lookup before GetChannel/promoter/backlink/linker. They require validation of every returned row, forbid generic parent/reply fallback, and verify deterministic `created_at DESC, _id DESC` selection capped at 100. Fresh Docker `backend-unit` overall/build/vet/unit/race and independent code/security/test/final reviews are `PASS`; qa-build/swagger are not evidence for this atom. Live Mongo/Ent-PostgreSQL/NATS/WS are `NOT_RUN`; cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Phase 151 selected-reply retention evidence (2026-08-26)

The full-page and short-page regressions prove that a selected reply absent from the private newest-100 page is retained exactly once: the short page appends it and the full page uses first 99 rows plus it. They also prove the parent remains separate, already-present reply data is unchanged, and a fresh assembly slice preserves the repository-owned page. Duplicate IDs and over-100 pages yield redacted 500 before GetChannel/effects. The first attempted RED was noncanonical because a test-only newline literal left build `PASS` but vet/unit/race `FAIL`. Corrected RED kept build/vet `PASS`, with unit/race failure only in the new full/short/duplicate/oversize cases. Production GREEN passed; test review initially `BLOCKED` the surviving backing-array mutant until snapshot remediation. Final fresh Docker `backend-unit` overall/build/vet/unit/race and independent code/security plus mutation re-review are `PASS`. Live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`; cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Phase 151 promotion source-ID defensive evidence (2026-08-26)

Initial two-case RED proved empty/nonempty source-ID mismatch could reach a successful issue response (build/vet `PASS`, unit/race `FAIL`). Test review found the explicit empty branch was not mutation-isolated; a direct-service route/source empty-ID regression was added. Controlled removal of only the empty guard kept build/vet `PASS` and made unit/race `FAIL` with only that new test failing; restored guard gave final Docker `backend-unit` overall/build/vet/unit/race `PASS`. Independent security/code and test re-review are `PASS`. The contract accepts `found=true` only for nonnil, live same-channel source with nonempty ID exactly equal to route `messageId`; malformed found data returns fixed/redacted 500 before target/thread/channel/effects, while `found=false` is concealed 404. Live Mongo/Ent-PostgreSQL/NATS/WS are `NOT_RUN`; cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`.

## Phase 151 mutation-adjacent target recheck evidence (2026-08-26)

Regressions prove that `PromoteToIssue` snapshots one checker instance, calls it for an early fresh target proof, and calls that same instance again with identical actor/project after scoped parent/replies/GetChannel/all description display-name reads and immediately before `CreateIssueFromPromotion`. They prove the stable mapping nil/error/false-true → redacted 500, false-false → concealed 404, and true-false → 403; a second rejection has zero promoter/backlink/linker calls. Root/reply ordering and same-instance replacement are covered. Valid TDD RED kept build/vet `PASS` and limited unit/race failure to new source/recheck/success cases. Final fresh Docker `backend-unit` overall/build/vet/unit/race, code/security review, and replacement mutation review are `PASS`; the direct initial test reviewer was ACL-blocked/`UNKNOWN`, while replacement contract review is `PASS`. Live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`; parent/reply/channel delete/revoke TOCTOU, post-recheck source deletion before issue creation, canonical backlink/post-create lifecycle, and post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create) remain `OPEN/UNBOUND`.

## Phase 151 chronological promotion description evidence (2026-08-26)

Four regression oracles cover parent-first chronological rendering, equal-time `ID ASC` tie-break, selected reply placement outside a full newest-100 page, selected reply placement on a short page, and preservation of repository-owned page order. The valid test-only RED left build/vet `PASS`; unit/race failed only those four tests. After the fresh-copy `CreatedAt ASC, ID ASC` implementation, Docker `backend-unit` overall/build/vet/unit/race is `PASS`, each chronology test passes exactly once in unit and race, and both logs contain zero fail-actions. Root mutation review kills no-sort, descending-sort, missing-ID-tie, and in-place-sort variants. Independent post-code review is `UNKNOWN` because the runtime agent-thread limit was exhausted. No live Mongo/Ent-PostgreSQL/NATS/WS evidence is claimed; parent/reply/channel delete/revoke TOCTOU, post-recheck source deletion before issue creation, canonical backlink/post-create lifecycle, and target authorization-plus-create atomicity remain `OPEN/UNBOUND`.

## Phase 151 promotion dependency snapshot evidence (2026-08-26)

The deterministic regression replaces checker, promoter, and linker during the first target proof and requires the in-flight request to keep the original three-object snapshot. The concurrent regression repeatedly writes checker/linker setters while reading the hubtest exports. In valid RED, build/vet passed, unit failed only the mixed-dependency snapshot test, and race failed that test plus the concurrent access test with warnings at the former direct field reads. After the shared `RWMutex` and request-entry snapshot, Docker overall/build/vet/unit/race is `PASS`; both tests pass once in unit and race, with zero fail-actions and zero race warnings. Root code/security/test review is `PASS`; independent post-code review remains `UNKNOWN` because no new agent thread was available. Atomic multi-setter hot reload, live stores, cross-store delete/revoke, target authorization-plus-create, and canonical post-create lifecycle remain outside this evidence.

## Phase 151 mutation-adjacent live-source recheck evidence (2026-08-26)

The second-proof table covers missing, store error redaction, nil found data, and mismatched source identity after successful initial source/target/thread/description work, requiring two exact route-bound calls, one early target proof, and zero promoter/backlink/linker effects. Successful source, target-recheck, root/reply chronology, and request-snapshot oracles require the second source read before final target authority. The first RED revealed a false self-thread fixture branch and was triaged; the corrected RED kept build/vet `PASS` and failed unit/race only in the intended matrix and success/order tests. Final Docker overall/build/vet/unit/race is `PASS`; eight named regression groups pass once in unit/race and fail-actions are zero. Root review is `PASS`; independent post-code review is `UNKNOWN` because no new agent thread was available. Live stores and atomic source/parent/reply/channel/target-to-create lifecycle remain outside this evidence.

## Phase 151 promotion live-parent recheck evidence (2026-08-26)

The test-first packet adds a per-call parent-result double, updates the successful reply call-order oracle to require two identical root lookups, and covers second-proof missing, repository error/redaction, found-nil, empty ID, mismatched ID, foreign channel, deleted root, and nested parent. RED evidence kept build/vet `PASS` while unit/race failed only the updated reply-success group and the new matrix. The final sequential `docker compose -f docker-compose.local.yml run --rm --no-deps backend-unit` receipt records overall/build/vet/unit/race `PASS`; unit and race logs have zero fail-actions and zero race warnings, and all named cases pass in both modes. This remains unit-only evidence; live Mongo/Ent-PostgreSQL/tracker timing and integration are `NOT_RUN`.

## Phase 151 mutation-adjacent source-channel authority evidence (2026-08-27)

The test-first packet requires two exact `ChannelActionPromote` calls for scoped success and two exact legacy membership calls when no resolver is configured. Root/reply success and final target-rejection order prove the second source-channel proof occurs after live source/root rechecks and before the final target proof. The late-failure matrix covers resolver infrastructure error, revoked promotion capability, revoked legacy membership, and legacy membership-store error; it asserts stable public status/message mapping, raw-error redaction, one early target proof, and zero promoter/backlink/linker effects.

The first RED was noncanonical only for the legacy branch because a hubtest constructor intentionally panicked on a `nil` resolver. After fixture-only triage to the ordinary constructor, corrected Docker RED kept build/vet `PASS`, made unit/race fail only in intended authority/order groups, and recorded zero panic/race warnings. Final Docker `backend-unit` overall/build/vet/unit/race receipts are `PASS`; named cases pass in both unit and race, fail-actions and race warnings are zero. Independent review is `UNKNOWN` because the runtime agent-thread limit rejected the required dispatch. Live stores/events and post-proof cross-store races remain outside this unit evidence.

## Phase 151 canonical create-result validation evidence (2026-08-27)

The lookup matrix covers nil message, lookup `Created=true`, empty ID, foreign channel, foreign sender, different nonce, unknown state, result/document state mismatch, and missing result state. The persist matrix covers `Created=false` without a canonical message plus the same identity/state failures. Call-order oracles require malformed lookup to stop after idempotency and malformed persist to stop after mention resolution/persist, with zero follow, reply-count, completion-CAS, or invalidation effects. Positive regressions preserve a legacy lookup with missing document state/result complete and a `Created=false` persist race returning an existing complete canonical message.

Initial Docker RED recorded overall `FAIL`, build/vet `PASS`, unit/race `FAIL` only in those new malformed groups, zero panics, and zero race warnings. The first post-code run exposed `ROOT=test` in two older realtime cases: their recorder omitted the ID assigned by concrete Mongo and their retry canonical omitted the request nonce. Test-only correction produced final Docker `backend-unit` overall/build/vet/unit/race `PASS`; all four named canonical/realtime groups pass in unit and race, with zero fail events, panics, or race warnings. Root review is `PASS`; independent review is `UNKNOWN` because the runtime agent-thread limit rejected dispatch. Live Mongo/NATS/WS and integration are `NOT_RUN`; strict fingerprint v1 is covered below; production binding and live invocation of the frozen typed nonce mapper, recovery/fencing, and outbox ownership remain `UNBOUND`.

## Phase 151 durable client-nonce fingerprint v1 evidence (2026-08-27)

Tests prove persisted lowercase 64-hex metadata/version `1`; equivalent JSON map/numeric forms, mention/group set order and duplicates, nil/empty attachments, and different server times; caller metadata rejection before repository access; strict missing/unknown/malformed/different metadata rejection; and producer/message/poll/JSON/text/thread-projection/forward/voice/attachment/mention/group mutation detection. A persist mismatch stops before follow/reply-count/CAS/invalidation. Positive retries prove current post-create edits do not replace stored original intent and earlier legacy-effect-state compatibility still works only when v1 fingerprint metadata is present.

Valid Docker RED kept build/vet `PASS`, failed unit/race only in the new durable/fingerprint groups, and recorded zero panic/race warnings. Final Docker `backend-unit` overall/build/vet/unit/race is `PASS`; nine named durable/fingerprint/compatibility/realtime groups pass in unit and race, with zero fail, panic, or race-warning events. Root mutation/security review is `PASS`; independent review is `UNKNOWN` because runtime thread allocation was exhausted. Live stores/integration are `NOT_RUN`; the no-replay version-0 reservation policy below, live invocation of the frozen different-payload/deleted-nonce mapper, live writer binding, recovery/fencing, and outbox ownership remain `UNBOUND`.

## Phase 151 typed client-nonce outcome evidence (2026-08-27)

The new table-driven regressions cover both canonical lookup and persist-race paths for a valid-v1 different-payload conflict and a deleted nonce reservation whose fingerprint metadata is absent. A separate oracle proves `Created=true + deleted` is an invalid repository result. Existing strict-v1 cases prove uppercase/malformed/missing/unknown metadata remains invalid; existing different payload is typed conflict; and `Created=true` fingerprint mismatch remains invalid. Exact call sequences and zero follow/reply-count/CAS/invalidation effects are asserted.

The compile-level RED recorded build `PASS` and vet/unit/race `FAIL` only for the two undefined sentinel names, with no panic or race warning. The final sequential Docker `backend-unit` receipt records overall/build/vet/unit/race `PASS`; eight named groups pass once in both unit and race, with zero fail records, panics, or data-race warnings. Root mutation/security review is `PASS`; independent review is `UNKNOWN` because the required Terra/high dispatch hit the agent-thread limit. qa-build/swagger and live Mongo/NATS/WS/integration are not evidence for this atom; live HTTP invocation and writer binding remain `UNBOUND`.

## Phase 151 post-create service error mapper evidence (2026-08-27)

The mapping matrix covers direct/wrapped different-payload conflict, direct/wrapped deleted reservation, invalid and wrapped-invalid canonical results, trusted-producer denial, arbitrary repository detail, and nil. It requires exact AppError 409 status plus the two legacy public messages for typed user conflicts; all other failures require one fixed AppError 500 save-message text, empty request/details/retry fields, and no sensitive raw substring. Nil remains nil.

Docker RED recorded overall `FAIL`, production build `PASS`, and vet/unit/race `FAIL` only on the undefined mapper symbol; panic and data-race warnings were zero. Final sequential `backend-unit` records overall/build/vet/unit/race `PASS`. Three mapping groups and eight typed nonce/fingerprint groups each pass once in unit and race; fail, panic, and data-race counts are zero. Root mutation/security review is `PASS`; independent Terra/high review is `UNKNOWN` because the runtime rejected dispatch at the thread limit. No handler/live writer uses the adapter; qa-build/swagger and live Mongo/NATS/WS/integration are not evidence for this atom.

## Phase 151 legacy client-nonce reservation evidence (2026-08-27)

The new lifecycle regressions cover a fully unversioned existing lookup and `Created=false` persist race, requiring the typed legacy sentinel, zero result, exact pre-effect calls, and no follow/reply-count/CAS/invalidation. A `Created=true` unversioned result must remain invalid. Earlier deleted-with-missing-metadata tests prove deleted precedence, while missing-only, version-missing, unknown, malformed, and uppercase cases prove partial/noncanonical data is not reclassified as legacy. Direct/wrapped mapper cases require the generic different-message 409.

Compile RED records overall `FAIL`, production build `PASS`, vet/unit/race `FAIL` only for the missing legacy sentinel, and zero panic/data-race warnings. Final Docker `backend-unit` overall/build/vet/unit/race is `PASS`; 13 mapping/nonce/fingerprint/legacy groups each pass once in unit and race, with zero fail, panic, or data-race records. Root mutation/security review is `PASS`; independent Terra/high review is `UNKNOWN` because dispatch hit the thread limit. No migration/backfill, live writer, qa-build/swagger, live Mongo/NATS/WS, or integration evidence is claimed.

## Phase 151 Mongo reply-count reconciliation evidence (2026-08-27)

The focused unit contract asserts exact BSON for the live direct-reply count, same-channel live root, absent-or-empty root `thread_id`, and `$max` update. Fail-closed coverage includes nil message, unknown producer, foreign channel/sender, non-reply, self-root, deleted reply, completed state/timestamp, and zero/negative observed count. A compile-time local port assertion proves `MongoRepository` exposes `ReconcileReplyCount`.

Test-only Docker RED: overall `FAIL`, production build `PASS`; vet/unit/race fail only on the absent repository method and plan/update symbols, with no panic or data-race marker. Final sequential Docker `backend-unit`: overall/build/vet/unit/race `PASS`; all three top-level reconciliation groups pass in unit and race, with zero fail/panic/data-race markers. Root review is `PASS`; independent Terra/high is `UNKNOWN` because dispatch was rejected at the thread limit.

No live Mongo/integration/handler/writer/recovery/qa-build/swagger evidence is claimed. The tests freeze create-side retry and concurrent-create monotonicity; historical overcount repair and concurrent deletion remain explicit uncovered lifecycle boundaries.

## Phase 151 automatic thread-follow effect evidence (2026-08-27)

The focused suite freezes author-first unique mentions, root/reply identity, auto mode and UTC cursor, one fresh PG scope call before every Mongo write, candidate-local typed denial/target loss, infrastructure stop after safe partial progress, malformed user/tenant rejection, trusted producer and system-bot proof, canonical state validation before authority, and mandatory context/resolver/writer dependencies. A compile assertion requires `CombinedRepository.EnsureAutoThreadFollows`; existing Mongo follow tests prove automatic insert-only behavior preserves manual intent.

Compile RED recorded overall `FAIL`, production build `PASS`, and vet/unit/race failures only on the missing method/coordinator/plan. Initial GREEN led root mutation review to add explicit untrusted-producer, malformed bot, foreign-tenant, and nil-context coverage. Final sequential Docker `backend-unit` records overall/build/vet/unit/race `PASS`; seven top-level groups pass in both unit and race and critical markers are zero. Root review is `PASS`; independent Terra/high is `UNKNOWN` because thread capacity rejected dispatch.

No live PostgreSQL/Mongo, handler/writer, recovery, NATS/WS, qa-build, or Swagger evidence is claimed. Multi-user partial writes converge by insert-only retry but are not a cross-store transaction.

## Phase 151 recipient unread invalidation evidence (2026-08-27)

The focused tests freeze root and reply identifier payloads, exact strict JSON shape, `unread_changed` audience selection, stable recipient order, sender/nil/duplicate filtering, zero-publish audience failure, trusted-producer and canonical pending-state validation before authority, non-UUID/self-root rejection, mandatory dependencies, nil/cancelled context, and compile compatibility with the pipeline port and `WSChannelAuthority`.

Test-only Docker RED kept production build `PASS` and failed vet/unit/race only on the missing event/payload/plan/effect symbols. An initial GREEN was mutation-reviewed, then exact reply JSON, zero `created_at`, and cancelled-context assertions were added. The final sequential `backend-unit` receipt records overall/build/vet/unit/race `PASS`. All five top-level unread-invalidation groups pass once in unit and once under race; searches found no `FAIL`, `panic:`, or `DATA RACE` marker. Root review is `PASS`; independent Terra/high review is `UNKNOWN` because the runtime rejected dispatch at the agent-thread limit.

No `qa-build`, Swagger, live PostgreSQL/Mongo/NATS/WS/browser, or integration evidence is claimed. The tests prove a standalone adapter only. They do not prove full repository composition, production writer routing, publish acknowledgement, outbox delivery, or recovery-worker fencing.

## Phase 151 eligible mention resolver evidence (2026-08-27)

Six focused top-level groups cover individual/`all` intersection, `online` and unique project-role intersection, role skip outside projects, authority/audience/presence/role infrastructure failures, invalid producer/request/tenant/channel facts plus missing sender audience, constructor and nil/cancelled context behavior, and the empty no-read fast path. Exact assertions freeze `ReceiveEvent` action, audience channel/event, role project ID, deterministic UUID order, sender/malformed/foreign/duplicate omission, and compile compatibility with `WSChannelAuthority`, `PresenceService`, and `EntRepository`.

The test-only RED recorded overall `FAIL`, production build `PASS`, and vet/unit/race failures only on undefined resolver symbols. The first GREEN was mutation-reviewed; the final sequential `backend-unit` run records overall/build/vet/unit/race `PASS`. All six top-level groups pass once in unit and once under race; no `FAIL`, `panic:`, or `DATA RACE` marker is present. Root review is `PASS`; independent Terra/high review is `UNKNOWN` because the runtime rejected dispatch at the agent-thread limit.

No `qa-build`, Swagger, live PostgreSQL/Redis/Mongo/NATS/WS/browser, writer, recovery, or integration evidence is claimed. Candidate-specific audience authorization errors remain an untested live completeness boundary even though unauthorized inclusion is prevented.

## Phase 151 post-create repository composition evidence (2026-08-27)

Four focused top-level groups freeze complete pipeline-interface satisfaction, exact eligible-mention and unread-invalidation delegation, defensive mention slice copying, delegate error identity, mandatory Combined/Ent/Mongo dependencies, ordinary and typed-nil delegate rejection, and fail-closed nil-receiver behavior for the explicit ports. Promoted storage methods are intentionally compile-checked but not called.

RED preserved build and failed vet/unit/race only on missing composition symbols. After the first GREEN, mutation review added the typed-nil cases and production guard. Final sequential Docker evidence is overall/build/vet/unit/race `PASS`; all four groups pass in unit and race with no `FAIL`, panic, or `DATA RACE` marker. Root review is `PASS`; independent review is `UNKNOWN` because retained agents exhaust all subagent slots. No `qa-build`, Swagger, live store, integration, recovery, or writer-cutover evidence is claimed.

## Phase 151 recovery claim/completion-fence plan evidence (2026-08-27)

Four top-level test groups freeze the exact Mongo claim filter/update, exact completion fence filter/update, UTC normalization, creation/lease chronology, trusted scope and canonical pending-state validation, tenant/message/channel/owner/fence mismatch rejection, server-owned request rejection, and durable BSON field manifest. Claim admits only absent/expired leases; completion at lease equality is rejected.

Initial RED preserved build and failed vet/unit/race on missing symbols/fields. First GREEN passed. Mutation RED preserved build/vet and failed unit/race on claim/completion timestamps before creation; a fixed-zone assertion mismatch was classified as a test bug because production correctly normalized UTC. Final minimal-diff Docker evidence is overall/build/vet/unit/race `PASS`; all four top-level groups pass in both unit and race with no critical markers. Root review is `PASS`; independent review is `UNKNOWN`. No live Mongo CAS, recovery runner, producer provenance, integration, Swagger, `qa-build`, or browser evidence is claimed.

## Phase 151 Mongo recovery claim/completion adapter evidence (2026-08-27)

Five top-level groups freeze exact claim/completion plan delegation, successful return-after acquisition, clean claim miss, error identity, cancelled-context short circuit, every returned-document identity/state/owner/fence/lease invariant, completion miss mapping, nil/typed-nil construction, nil receiver behavior, production interface satisfaction, and owned recovery-critical result fields.

Initial RED preserved backend build and failed vet/unit/race on missing adapter symbols. First GREEN passed. Mutation RED preserved build/vet and failed unit/race only because the store result still owned the resolved-mention slice and lease pointer. Final Docker evidence is overall/build/vet/unit/race `PASS`; all five top-level groups pass in unit and race with no `FAIL`, panic, or `DATA RACE` marker. Root review is `PASS`; independent review is `UNKNOWN`. No live Mongo, recovery-loop, provenance, integration, Swagger, `qa-build`, or browser evidence is claimed.

## Phase 151 durable producer provenance evidence (2026-08-27)

The focused unit matrix covers user and system-bot stamping, caller-spoof rejection, all five closed producer reconstructions, invalid target/document/version/producer/bot shapes, canonical-result provenance drift before effects, durable-manifest inclusion, exact claim/completion provenance predicates, post-update drift rejection, and defensive acquisition ownership. Initial missing-symbol RED kept build green and failed vet/unit/race only because the new API did not exist. First GREEN passed the full launcher.

A controlled mutation removed only the server stamping assignment. Build/vet remained `PASS`; unit/race failed on nil durable provenance, killing the mutant. The mutation also caused a later drift-table dereference after the primary failure; test triage classified that as fixture coupling and added a nil-safe valid provenance baseline so drift cases remain independent. Production was restored to its exact pre-mutation SHA. Root review then added a focused CAS RED: build/vet stayed `PASS`, while unit/race failed on missing claim/completion provenance predicates, accepted scope drift, accepted post-update drift, and store-owned pointer aliasing. A third RED kept build/vet `PASS` and failed unit/race on the nil-message panic in the exported claim-plan validator; moving structural validation before provenance access closed it. Final Docker overall/build/vet/unit/race is `PASS`; all seven top-level provenance groups pass in unit and race; fail/panic/data-race marker search is empty.

Evidence remains unit/compile only. No live Mongo persistence/readback, recovery-loop execution, integration reauthorization, lease behavior, Ent/PostgreSQL/Redis/NATS/WS/browser path, `qa-build`, or Swagger run is claimed. Root review is `PASS`; independent review is `UNKNOWN` because retained subagent threads exhausted the runtime limit.

## Phase 151 claimed recovery effect runner evidence (2026-08-27)

Six focused groups freeze root and reply effect order, canonical-sender scope reconstruction, clean claim contention, exact returned owner/fence/lease/message/provenance validation before effects, error identity and stop position for every repository stage, nil/typed-nil dependency handling, cancelled context, zero clocks, and completion at lease expiry.

Initial RED kept production build `PASS` and failed vet/unit/race only on the missing runner interface, clock, and exported function. A first acquisition-validation mutation was compile-invalid because it left the expected plan unused and is not counted. The corrected compilation-preserving no-op kept build/vet `PASS`; unit/race failed in the malformed-acquisition group on foreign owner/fence/lease/message acceptance and forbidden downstream effects. Production was restored to exact SHA-256 `16011D65EE3D5E57FB5B312F77E6FB96CCF9B3093C5FD443310CD024601A2192`.

Final sequential Docker `backend-unit` records overall/build/vet/unit/race `PASS`. All six groups pass in unit and race and the final logs contain no `FAIL`, panic, or `DATA RACE` marker. Root review is `PASS`; independent review is `UNKNOWN` because retained agents exhaust all runtime slots. No recovery-loop wiring, scheduler, live writer/store, lease renewal, integration, Swagger, `qa-build`, NATS/WS, or browser evidence is claimed.

## Phase 151 recovery coordinator adapter evidence (2026-08-27)

Six focused groups freeze end-to-end bounded scan-to-runner composition, exact worker/lease forwarding, claim-miss continuation to the next document, runner-error identity and stop position, nil/typed-nil construction, malformed target and tenant/channel/actor/producer authority drift, exact scan argument/error delegation, nil receiver, and nil context.

Initial RED kept production build `PASS` and failed vet/unit/race only on undefined coordinator adapter API. First GREEN passed. A controlled compilation-preserving mutation replaced only `ReconcilePendingPostCreateEffects` runner invocation with a no-op. Build/vet remained `PASS`; unit/race failed exactly `BindsBoundedScanToRunner`, `ContinuesAfterClaimMiss`, and `PreservesRunnerError`. Production was restored to SHA-256 `01DB81738F777572EC43601E1E997359EBC0A5DBEA52BA5F9189D8BAC991DB38`.

Final sequential Docker `backend-unit` records overall/build/vet/unit/race `PASS`; all six groups pass in unit and race and the final logs contain no `FAIL`, panic, or `DATA RACE` marker. Root review is `PASS`; independent review is `UNKNOWN` because all three retained subagent slots are exhausted. This older receipt does not prove a scanner bound into the coordinator, server, or a production scheduler; live store/effect, lease renewal, integration, Swagger, `qa-build`, NATS/WS, and browser evidence remain unclaimed.

## Phase 151 promotion reply chronology evidence (2026-08-27)

`TestPhase151PromoteToIssueRejectsNonDescendingReplyPagesBeforeEffects` covers a `created_at` inversion and an equal-timestamp `_id` tie-break inversion. Both cases require a concealed internal response, no GetChannel/display/effect work, and an exact stop after source action, scoped source, target authority, and scoped reply resolution. Existing root/reply chronological rendering and selected-reply-retention groups verify accepted pages still render parent first and replies oldest-to-newest.

Initial RED kept build/vet `PASS` and failed unit/race only in the new group. A compilation-preserving no-op chronology mutant also kept build/vet `PASS` and failed unit/race exactly in both inversion scenarios. After exact restoration to SHA-256 `9003A6B8DA409D8732C52DACD5A52E316C41EB6E4D06E0EE6143B6905BF852F4`, final sequential Docker `backend-unit` is overall/build/vet/unit/race `PASS`; all four relevant groups pass in unit and race, and no `FAIL`, panic, or `DATA RACE` marker remains.

Root review is `PASS`; independent review is `UNKNOWN` because retained subagent slots exhaust runtime capacity. Live Mongo, cross-store behavior, integration, Swagger, `qa-build`, NATS/WS, and browser execution are `NOT_RUN`. Recovery-scanner tests are intentionally not started until the authority identity decision is frozen.

## Messenger completion CAS and Jitsi seam evidence (2026-08-27)

Messenger commit `7d248b98` unit-tests the exact post-create completion CAS predicate and minimal completion update across trusted producer scopes. A controlled predicate mutation produced exit `1`; final Docker `backend-unit` passed and independent re-review found no findings. This historical unit receipt predates the accepted recovery principal/reauthorizer/scanner/passive composition/callable once capability. Production installer, scheduler/process owner, live writer/runtime/stores, integration, and direct-writer cutover remain `UNBOUND/NOT_RUN`.

Jitsi backend commit `63a3e8bf` preserves the complete typed `JoinCredential` while retaining legacy token compatibility; final Docker `backend-unit` passed and independent review was `Ready`. Jitsi frontend commit `0751ad2a` tests strict self-hosted origin/provider/canonical-room/JWT handling in a pure contract with no network or logging. Final Docker `frontend-unit` target, TypeScript, and Vitest passed: `188` files and `591` tests, after two documented test-only oracle/typo corrections; independent review was `Ready`.

No browser or live provider test ran. LiveKit remains active; Jitsi public join DTO/route/iframe/deployment/media/cutover/recording/runtime are `NOT_RUN`/absent. Future coverage needs browser-reachable LiveKit URL semantics, call-aware issuer and room identity, expiry/refresh, and guest-policy tests.
Messenger commit `c90e75d5` adds constructor/wiring regressions for exact concrete post-create identity, nil and typed-nil rejection, legacy-constructor compatibility, and zero effect calls during construction. Final Docker `backend-unit` passed build/vet/unit/race (exit `0`); independent review found no findings. No writer or live store path ran; recovery and nonce compatibility remain required before cutover tests can be valid.

## Calls canonical-room/expiry and Messenger pending-scan evidence (2026-08-28)

Focused backend units cover exact Jitsi `Call.RoomName` binding and invalid canonical-room forms (`343d1d0d`), exact equality between internal credential expiry and JWT expiry with legacy token compatibility (`366bb825`), and the exact bounded Messenger recovery-scan filter/order/limit plus fail-closed authority drift and absence of lease predicates (`44b3a95e`).

One combined isolated Docker `backend-unit` run exited `0`; build, vet, unit, and race all passed without Mail overlap. Self-hosted Jitsi deployment/route/browser/media E2E, live provider runtime, live Mongo recovery, production scheduler/principal, and `SendMessage` cutover are `NOT_RUN/UNBOUND`; LiveKit remains active.

## Typed Jitsi join projection evidence (2026-08-28)

Commit `0a332760` is covered for exact provider/room/service-URL/token projection, a single private issuance clock call, token claim agreement, and zero-result failures for invalid server inputs. The first attempted RED contained a test typo and is not counted; the corrected replacement RED failed on the exact undefined method. Final isolated Docker `backend-unit` exited `0` with build, vet, unit, and race passing while Mail was absent.

No provider factory, public route/DTO, frontend, self-hosted deployment, cutover, live provider runtime, or browser E2E was exercised; these remain `UNBOUND/NOT_RUN`, and LiveKit remains active.

## Inert Jitsi host and recovery page-guard evidence (2026-08-28)

Commit `ee0387e5` tests one exact empty host/ref for valid input, fail-closed invalid/null input and rerender, absence of credential values in DOM, and zero fetch/storage/logging/external-API/script/iframe activation. Its valid RED was the missing module; final isolated Docker `frontend-unit` exited `0` with TypeScript and Vitest passing.

Commit `ce46a0b2` tests cursor/page structure, tuple order and advancement, exact `Next`, size bounds, future nonempty kinds, error identity, zero partial pages, and alias isolation. Its valid RED kept build/vet passing and failed unit/race on the old pass-through; final isolated Docker `backend-unit` exited `0` with build, vet, unit, and race passing. Live provider/recovery runtime, deployment, routing, `CallPanel`, cutover, integration, and browser E2E remain `NOT_RUN/UNBOUND`; LiveKit remains active.

## Messenger server-owned recovery authority regression (2026-08-28)

Focused unit/race regressions cover the sole closed `Phase151RecoveryServicePrincipal` value `messenger.post_create_recovery`, exact `Phase151RecoveryTargetAuthority{Principal,TenantID,ChannelID,Kind}` acceptance, and rejection of malformed or principal/tenant/channel/kind-drifting facts as `ErrPostCreatePipelineInvalid`. The contract deliberately proves absence of actor, producer, bot, and integration fields rather than inventing a recovery actor.

The valid RED kept build `PASS` and failed vet/unit/race only on missing symbols. After removal of one stale test-only unused `errors` import exposed by the first GREEN, final sequential Docker `backend-unit --no-deps` recorded target/build/vet/unit/race `PASS`, including the exact new case under unit and race, with no Mail overlap; independent Messenger code/security and test reviews are `Ready`. This receipt predates the accepted scanner/passive composition/callable once capability. No production migration/installer, scheduler/process owner, live writer/runtime/stores, integration, Swagger, or live PostgreSQL/Mongo/Redis/NATS/WS/browser execution is covered; these remain `UNBOUND/NOT_RUN`.

## Calls signed credential-expiry regression (2026-08-28)

Focused tests verify exact equality between `JoinCredential.ExpiresAt` and the signed JWT `exp` for Jitsi, authenticated LiveKit, and both guest LiveKit credentials; the existing five-minute Jitsi and 24-hour LiveKit lifetime assertions remain. The valid RED kept production build `PASS` and failed vet/unit/race on the absent field. After unrelated Messenger test-only triage, final sequential Docker `backend-unit --no-deps` recorded target/build/vet/unit/race `PASS`, with every named expiry case passing in unit and race and no Mail overlap; independent crypto/code and test reviews are `Ready`.

This is executable unit/race evidence, not deployment or browser evidence. Public DTO/route/refresh/client enforcement, self-hosted Jitsi runtime, provider cutover, and browser/media behavior remain `UNBOUND/NOT_RUN`.

## Unverified closed-authority reconstruction and call-aware issuer coverage (2026-08-28)

Focused source tests cover closed-target authority reconstruction into canonical producer scope, including invalid authority, canonical-message, and producer-provenance paths. Separate source tests cover `CallJoinCredentialIssuer` structural validation, compile satisfaction by LiveKit/Jitsi, and one-delegation typed credential projection. Current Docker evidence for these two atoms is `NOT_RUN` because Mail owns Docker; no prior backend-unit receipt is inherited.

Future execution must keep the boundary explicit: accepted Messenger scanner/passive composition/callable once capability does not prove fresh database authorization, production migration/installation, scheduler/process ownership, writer cutover, live stores, or recovery runtime. Calls coverage does not prove call loading, RBAC/membership/invite/guest authorization, service/factory/route/client binding, Jitsi deployment/cutover, or browser/media runtime. LiveKit remains active; all such Jitsi and runtime gates remain `UNBOUND/NOT_RUN`.

## Recovery-authority migration and Calls projector evidence boundary (2026-08-28)

No Docker result is recorded for the recovery authority migration or Calls projector atom: Mail owns Docker and invalidates the earlier overlap, so no receipt is inherited. That historical Messenger static scope was limited to closed-authority reconstruction; later accepted scanner/passive composition/callable once capability supersede the uncomposed status. Production installer, scheduler/process owner, writer cutover, live stores, and runtime remain `UNBOUND/NOT_RUN`.

Calls static scope is limited to unbound `CallJoinCredentialProjector` port/constructor/method behavior: active exact server-loaded `Call`, channel-member or instant-creator policy, server-known name, and exactly one typed issuer call. Ent source, service/`JoinCall`/HTTP, participant/factory/client, Jitsi cutover, and tenant claim have no evidence. LiveKit remains active. No code/test/Docker/Git/ROADMAP/ADR outcome is implied beyond these boundaries.

## Fresh isolated Docker unit/race receipt (2026-08-28)

This entry supersedes the latest NOT_RUN execution status only for the named
source-test atoms below. After a confirmed 60-second empty Docker window, one
docker compose -f docker-compose.local.yml run --rm --no-deps backend-unit
run completed with target PASS, first exit 0, and build/vet/unit/race all PASS;
monitoring found no Mail overlap. The Calls issuer/projector and Messenger
authority-reconstruction, claimed-runner, and scan-proof tests each passed
under both unit and race.

The earlier run was stopped because of Mail overlap and remains invalid history;
no receipt is inherited from it. This is unit/race evidence only. All existing
UNBOUND, runtime, deployment, browser, and cutover statements remain unchanged.

## Calls membership argument-order regression (2026-08-28)

Focused projector tests assert channel-first membership invocation: the persisted call `channelID` is the first UUID and the authenticated caller is the second. They preserve channel denial/error short-circuiting, instant-call creator policy, server-known-name ordering, and exactly one issuer invocation. The behavioral RED kept build/vet `PASS` and failed unit/race with the expected channel UUID receiving the actor UUID and vice versa. The corrected implementation then passed one isolated Docker `backend-unit --no-deps` run with target/build/vet/unit/race `PASS`, first exit `0`, and no Mail overlap; independent review found no issue.

This regression does not execute or prove an Ent/DB source, service/HTTP/factory/client binding, tenant authorization, self-hosted Jitsi runtime, provider cutover, or browser/media behavior. Those gates remain `UNBOUND/NOT_RUN`, and LiveKit remains active.

## Calls projector structural fail-closed regression (2026-08-28)

Focused tests require nil and zero-value projectors to return `JoinCredential{}` plus `ErrCallJoinCredentialInvalid` without panic. The malformed server-loaded call table covers blank, whitespace-only, and padded room names, a nonnil zero channel ID, and an instant call with a zero creator; each path permits only `load_active_call` and forbids membership, name, and issuer operations.

RED-1 recorded target `FAIL`, build/vet `PASS`, and unit/race nil-pointer panics. RED-2 recorded target `FAIL`, build/vet `PASS`, and five malformed-call failures while both receiver cases passed. The final isolated Docker `backend-unit --no-deps` run recorded target/build/vet/unit/race `PASS`, first exit `0`; both named suites passed under unit and race and monitoring found no Mail overlap. Ent/DB/service/HTTP/factory/client, self-hosted Jitsi runtime/cutover, and browser/media gates remain `UNBOUND/NOT_RUN`; LiveKit remains active.

## Messenger closed-authority scan-plan regression (2026-08-28)

Focused tests prove an exact plan for the closed recovery authority, legacy parity even when the legacy scope carries an unrelated actor and a valid forward producer, the 500-row cap, and fail-closed empty/arbitrary principal, zero tenant/channel, empty kind, and nonpositive limits. The new API accepts no actor or producer input.

RED recorded target `FAIL`, build `PASS`, and vet/unit/race failures only on the absent export. The first GREEN attempt was stopped and invalidated on Mail overlap. After Mail completed and a new 60-second empty window, the isolated rerun recorded target/build/vet/unit/race `PASS`, first exit `0`; both named suites passed under unit and race without overlap. Mongo repository/port/runner, cross-store composition of the narrow Ent pager and pure reauthorizer, scheduler, writer, runtime/cutover, and fresh DB/RBAC coverage remain `UNBOUND/NOT_RUN`.

## Calls Ent projection-source regression (2026-08-28)

The compile/fail-closed suite covers interface satisfaction plus nil receiver/client/context/UUID inputs. Independent review exposed a typed-nil context gap; its behavioral RED kept build/vet `PASS` and failed unit/race with Ent-client panics for both source methods. The shared typed-nil detector then closed both paths.

The final fresh isolated Docker `backend-unit --no-deps` run passed target/build/vet/unit/race with first exit `0`; the invalid-input and typed-nil suites passed under unit and race, no race warning or Mail overlap was observed. Positive Ent predicates, name fallback against stored rows, and live PostgreSQL remain integration `NOT_RUN`; production wiring and Jitsi cutover remain unchanged.

## Phase 151 closed recovery authority evidence (2026-08-28)

Valid RED for 9f8d0b66 kept production build PASS; vet, unit, and race failed only because production still used old scope-plus-target pending scan signatures. Typo, undefined-identifier, and unrelated RED failures were invalid.

Implementer GREEN and final root GREEN used exactly docker compose -f docker-compose.local.yml run --rm --no-deps backend-unit, with no --build or pull. Final root receipt 2026-08-28T17:42:08.0869425+03:00 records target/build/vet/unit/race PASS, first exit 0, four stderr logs empty, and named authority/coordinator tests passing under unit and race.

Regression coverage proves valid authority reaches scan/reconcile once unchanged; reauthorization error, foreign principal/tenant/channel/kind, or malformed authority produces zero Mongo scan and zero reconcile. Claimed-runner provenance reconstruction remains green; Mongo retains exact filter/order/cap through the authority scan-plan builder.

Review evidence only: invalid System.Object[] was rejected before the regenerated real diff received Spec Compliance PASS, Task quality Approved, and no Critical/Important breakage.

Historical 2026-08-28 composition evidence is superseded: passive `NewPhase151PendingRecoveryComposition` now composes reauthorizer -> scanner -> process-local gate -> claimed coordinator, and a callable once adapter reaches the existing executor. Runtime evidence remains `UNBOUND/NOT_RUN`: no scheduler/process owner, production installer, writer cutover, service/server wiring, or live PostgreSQL/Mongo/NATS/WS/browser/integration has run; Phase 151 is partial.

## Messenger coordinator interval-gate composition coverage (2026-08-28)

Commit `a5851053` narrows `Phase151PendingRecoveryScanRepository` so it no longer owns interval admission. Focused deterministic unit/race coverage proves that `Phase151ClaimedRecoveryCoordinatorAdapter` requires a plain nonnil `Phase151PendingRecoveryIntervalGate`, validates interval input, and delegates only interval acquisition to that gate; false, error, or invalid admission causes no scan. Scanner paging, reauthorization, and document listing remain separate dependencies, and executor admission-before-scan ordering is unchanged.

Valid RED `worktree-master-backend-unit-run-d0fe21905ddb` had target `FAIL`, first exit `1`, build `PASS`, and expected narrowed-interface compile mismatch in vet/unit/race. Root-read GREEN `worktree-master-backend-unit-run-e9844841ebae` has target/build/vet/unit/race `PASS`, first exit `0`, and four stderr lengths `0`. Independent spec and quality reviews are `PASS`/`APPROVED`, with `C0/I0/M0` after removing an over-broad white-box request.

This is composition/unit/race evidence only. The standalone gate remains per constructed instance; no process singleton, scheduler/process-lifetime wiring, pager/store/runtime/live transport/browser, distributed, or exactly-once behavior is claimed. Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

Docker policy allows concurrent branch-isolated Compose projects, one task-owned container per branch, no --build/pull/image creation/stopping another branch, and a separate shared-layer numeric Jitsi disk-budget gate.

## Calls service credential facade coverage (2026-08-28)

Commit `b9faba3c920371f182a7416e039a6f2708954fbb` is covered at the compile/unit/race boundary for nil-projector constructor rejection; nil receiver and legacy-constructor fail-closed behavior; exact context/call/authenticated-user forwarding and source order; all five credential fields; one issuer call; zero provider calls; and exact error identity plus zero result for source and issuer failures. The additive constructor calls the legacy provider constructor once and leaves legacy constructors unchanged.

Valid RED container `worktree-master-backend-unit-run-e80469df6d84` finished `2026-08-28T18:40:12+03:00` with target `FAIL`, first exit `1`, build `PASS`, and vet/unit/race failure only because the constructor was absent. Final GREEN `worktree-master-backend-unit-run-866fc0b8b37b` finished `2026-08-28T18:49:00+03:00` with target/build/vet/unit/race `PASS`, first exit `0`, and four empty stderr logs. Independent review returned Spec Compliance `PASS` and Task quality `Approved`, with no Critical/Important/Minor findings; root independently read exact raw Docker receipts after the reviewer could not inspect them.

No server/runtime constructor, `JoinCall`, handlers/routes/DTO, participant lifecycle, provider factory/client selection, tenant authorization, self-hosted Jitsi deployment, provider cutover, integration, or browser/media path was exercised. These remain `UNBOUND/NOT_RUN`; LiveKit remains active and Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

## Phase 151 PG recovery-target reauthorizer evidence (2026-08-28)

Focused unit/race coverage for `326fb23c` and hardening `a57f5da5` verifies singleton-tenant resolution, exact target facts, live-channel requirement, exact `public|private -> channel` and `dm|group -> dm`, active tenant project for project channels, tenant creator for projectless channels, and closed-principal-only output with no actor/sender/bot/integration/Mongo authority. `67abe088` is whitespace-only fixture alignment.

Final accepted Docker container `worktree-master-backend-unit-run-4bb4f9e2e610` completed `2026-08-28T19:48:05+03:00` with target/build/vet/unit/race `PASS`, exit `0`, four empty stderr logs. Initial RED `worktree-master-backend-unit-run-996a06b88265` (exit `1`) and first GREEN `971034f37f73` are historical. Independent final review: Spec Compliance `PASS`, Task quality `Approved`, no findings.

No API/OpenAPI or runtime claim follows. The Ent recovery-target facts source and narrow static-only Ent pager exist and later participate in the passive recovery graph; interval scheduling/process ownership, coordinator/repository production installation, writer cutover, service/server wiring, and live PostgreSQL/Mongo/NATS/WS/browser/integration remain `UNBOUND/NOT_RUN`; Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`, product not ready.

## Phase 151 Ent recovery-target facts coverage (2026-08-28)

Commit 4b100ee6 is covered only for compile-time Phase151PendingRecoveryTargetFactSource satisfaction and pre-query fail-closed guards in EntRepository.ResolvePendingRecoveryTargetFacts: receiver/client, nil or typed-nil/cancelled context, and tenant/channel UUID inputs. RED worktree-master-backend-unit-run-e14c5e925814 failed only because the method was absent. Final GREEN worktree-master-backend-unit-run-546c91d7105f at 2026-08-28T20:26:28+03:00 recorded target/build/vet/unit/race PASS, exit 0, and four empty stderr logs; independent review is PASS/Approved with no findings and separate predicate review is PASS.

Successful query, NotFound, driver/generated SQL, projection, and live PostgreSQL semantics remain NOT_RUN. No API, runtime, or readiness conclusion follows.

## Phase 151 Ent pending-recovery target pager coverage (2026-08-28)

Commit `748f4997` adds the narrow `EntRepository.ListPendingRecoveryTargets` pager, not a full scanner. Valid RED `worktree-master-backend-unit-run-35a8f1ba0967` had target `FAIL`, first exit `1`, build `PASS`, and vet/unit/race missing-method failures; retained RED stderr lengths were `0/461/0/0`. Root-read GREEN `worktree-master-backend-unit-run-a95ff242957f` recorded target/build/vet/unit/race `PASS`, first exit `0`, and four stderr lengths `0`. Specification and query reviews are `PASS`/`APPROVED`, `C0/I0/M0`.

The unit cases prove compile/interface satisfaction and fail-closed guards before singleton/query access: dependencies/context, `ctx.Err()`, positive/`MaxInt`-safe limit, and structural cursor. They do not prove cursor-tenant mismatch, singleton cardinality, generated Ent predicates for live `public|private|dm|group` plus active-project or projectless-creator tenant predicates, optional IDGT, ascending order, safe `limit + 1`, query execution, mapping/pagination/Next, error chains, or live PostgreSQL. Those are static-only or `NOT_RUN`. The later passive holder composes the pager path into the recovery graph, but scheduler/process ownership, production installation, writer cutover, service/server runtime, integration, and product readiness remain `UNBOUND/NOT_RUN`.

## Messenger standalone pending-recovery interval-gate regression (2026-08-28)

Focused deterministic unit/race tests for `cb995211` and cancellation fix `3632c37` cover narrow-interface satisfaction; valid/nil/typed-nil clock construction and nil receiver; nil/typed-nil/cancelled contexts before any clock read; zero/negative intervals; first admission, below-minimum denial, exact equality admission; zero/backward time without state mutation; cancellation while queued and while `Now()` is channel-blocked (no sleeps); exactly one fixed-time concurrent admission; two independent gates both admitting; and immediate interval consumption.

The contract is explicitly a per-constructed-instance best-effort throttle with a clock distinct from `Phase151PostCreateRecoveryClock`. Its mutex covers only admission check/read/write, not recovery I/O. It is not a scheduler, coordinator, scanner, singleton, distributed/global gate, or exactly-once mechanism. The focused cases assert cancellation before mutex acquisition, after acquisition, and after `Now()` before mutation.

Valid missing-contract RED `9611a72a031b` preceded initial GREEN `323c19983456` at `2026-08-28T21:01:26+03:00`. Final-fix root-read receipts at `2026-08-28T21:55:49+03:00` report target/build/vet/unit/race PASS, exit 0, four stderr lengths 0, and **container ID not retained**. Scoped re-review: ADDRESSED/PASS/APPROVED; new breakage 0/0/0. Cross-store scanner/coordinator production migration, scheduler/process-lifetime owner/wiring, runtime/live stores/transports/browser, and integration remain UNBOUND/NOT_RUN; Phase 151 is `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.
## Calls configured credential-service composition evidence (2026-08-28)

Commit `07f2fb5c` has a focused contract for `NewCallServiceWithConfiguredProvider`: LiveKit configuration flows through factory, typed issuer, projector, and facade; the issued credential is projected from server-loaded call/member/name facts; and Jitsi factory rejection plus typed-nil source failure are fail-closed. Valid RED `46b078978d66` preceded GREEN `4d2f71546884` at `2026-08-28T22:39:54+03:00`; target/build/vet/unit/race are root-read `PASS` with exit `0` and empty stderr. The `qa-build` and swagger check `PASS` are retained implementer receipts, not independently refreshed or root-verified receipts. Review found the implementation conformant; its only finding was this required documentation sync.

These are compile/unit/race and static-review facts only. No `cmd/server` runtime, public `JoinCall`/route/DTO/frontend/recording change, legacy-token-path exercise, provider cutover, Jitsi deployment, integration, or browser/media execution is proven; all remain `UNBOUND/NOT_RUN`. LiveKit remains active and Jitsi is rejected before Phase 152.


## Calls confirmed participant-presence reducer evidence (2026-08-28)

Commit `f405304f` has an external unit/race contract for the pure reducer: invalid previous state, zero call/user IDs, and invalid kind return the exact sentinel and zero transition; the table covers unknown/left-to-connected, duplicate connected, connected-to-left, unknown-to-left, and duplicate left. It also verifies that credential issuance is not participant presence or a joined effect.

Valid RED `worktree-master-backend-unit-run-3234f1a2a70b` had target `FAIL`, first exit `1`, build `PASS`, and vet/unit/race missing only the exported reducer symbols; retained implementer RED stderr lengths were `0/204/0/0`. Root-read GREEN `worktree-master-backend-unit-run-af958086c4a2` had target/build/vet/unit/race `PASS`, first exit `0`, and four stderr lengths `0`. Specification and quality reviews are `PASS/APPROVED`, `C0/I0/M0`.

This is pure unit/race evidence, not a provider/runtime claim: no raw webhook verification, provider selection, persistence, NATS, inbox/outbox, transactional application, dedupe, clocks, delivery IDs, multi-session, guest, refresh/revoke, API/service/repository/schema/server/frontend binding, deployment, browser, media, or cutover was exercised. The legacy `/join`, DTO/token/frontend/recording path remains unchanged; LiveKit remains active and Jitsi remains `UNBOUND/NOT_RUN`.

## Calls trusted-provider participant-presence mapper evidence (2026-08-29)

The mapper suite asserts the contract before the reducer boundary: trusted provider-normalized `connected` and `disconnected` facts map respectively to confirmed `Connected` and `Left` observations while preserving resolved call and Hub-user IDs. Zero call/user IDs and zero or unknown kinds return the zero observation and the existing invalid-observation sentinel. Tests encode and exercise the typed-contract precondition and mapping shape only; real upstream provider-event authentication/normalization, room-to-call resolution, and identity-to-Hub-user resolution are `UNBOUND/NOT_RUN`. The mapper takes no raw event, room, identity, JWT, provider, token, or delivery ID.

Valid RED `worktree-master-backend-unit-run-e8e44e800277` had target `FAIL`, first exit `1`, build `PASS`, and vet/unit/race missing exports; retained stderr lengths were `0/222/0/0`. Root-read GREEN `worktree-master-backend-unit-run-72b909a17791` and the Godoc-fix GREEN `worktree-master-backend-unit-run-80fe0c688984` each record target/build/vet/unit/race `PASS`, first exit `0`, and four empty stderr logs. Final specification and quality review is `PASS/APPROVED`, `C0/I0/M0`.

This proves only a pure mapping contract. Reducer invocation, webhook application, persistence/NATS, durable identity/order/dedupe, service/server/runtime wiring, integration/browser/media, and Jitsi cutover remain `UNBOUND/NOT_RUN`; LiveKit remains active.

## Phase 151 recovery scanner adapter evidence (2026-08-29)

TDD RED `worktree-master-backend-unit-run-fad4016a5f63` had target `FAIL`, first exit `1`, build `PASS`, and expected missing-symbol vet/unit/race failures; retained stderr lengths were build `0`, vet `254`, unit `0`, race `0`. Initial GREEN `worktree-master-backend-unit-run-d2860a8fb723` was root-read target/build/vet/unit/race `PASS`, exit `0`, with four empty stderr logs. Review required nil-receiver plus cancelled-context precedence coverage and a comment correction; commit `99501bc3` addressed both. Final GREEN `worktree-master-backend-unit-run-7f9add122b9b` is root-read target/build/vet/unit/race `PASS`, first exit `0`, with four empty stderr logs; final spec/quality review is `PASS`/`APPROVED`, `C0/I0/M0`.

Coverage is bounded unit/race evidence: invalid structural receiver/context/dependencies before cancellation, exact cancelled-context return after structural validity, one exact forwarding call per matching port, exact value/error/alias preservation including nonzero-plus-error, and no synthesized recovery behavior. Docker was not rerun for this documentation-only update. No live Ent/Mongo, scheduler, coordinator/server construction, integration, or runtime proof is claimed.

## Calls confirmed participant connection-key coverage (2026-08-29)

The focused suite for commit `b7695ddc` covers pure Hub-owned `(CallID, UserID, Epoch)` construction, exact canonical grammar `hub.call-connection.v1/<call>/<user>/<epoch>`, every zero-component rejection through `ErrConfirmedParticipantConnectionKeyInvalid` and a zero result, and empty canonical output for invalid keys. It also asserts that neither the value nor pointer satisfies `fmt.Stringer`; this only guards accidental stringification and is explicitly not redaction/security coverage because UUID fields remain exported and formattable. The suite assumes an already trusted upstream fact and covers no epoch generation, parser/serializer, provider/webhook, persistence/refcount, service/server/runtime, or Jitsi path.

Historical RED had missing symbols, exit `1`, and stderr length `972`; no RED container name was retained. Intermediate failed gates were contaminated by a concurrent Messenger RED, are not product failures, and provide no `PASS` authority. The only canonical final container is `worktree-master-backend-unit-run-01f901840085`: target/build/vet/unit/race `PASS`, exit `0`, and four empty stderr logs; stdout was not inspected. Independent specification review is `PASS` and quality review `APPROVED`, `C0/I0/M0`. This is unit/race evidence only; legacy behavior remains unchanged, LiveKit remains active, and Jitsi is `UNBOUND/NOT_RUN`.

## Phase 151 pending-recovery runtime-config coverage (2026-08-29)

Focused unit/race coverage for commit `75d26e8e` verifies exact acceptance of a nonzero owner UUID, positive lease, and positive `MaxPages`, `MaxTargets`, `PendingScanLimit`, and `MinimumInterval`; each invalid shape returns nil. It also proves nil-accessor zero values, exact retained values, and copy isolation by mutating the supplied execution config after construction and an accessor result after retrieval.

Valid RED `worktree-master-backend-unit-run-8bc84ea0e92f` recorded target `FAIL`, first exit `1`, build `PASS`, and expected missing-constructor vet/unit/race with stderr lengths `0/248/0/0`. The canonical GREEN receipt `worktree-master-backend-unit-run-8f2a97825507` records target/build/vet/unit/race `PASS`, exit `0`, and four empty stderr logs. Specification review is `PASS`, quality review `APPROVED`, `C0/I0/M0`. Two accidental concurrent containers/receipts are explicitly invalid orchestration evidence and do not constitute a product failure or PASS authority.

This config-only suite does not prove owner generation, lease renewal, singleton/distributed ownership, caps/defaults, or environment/deployment parsing. The later passive holder composes reauthorizer -> scanner -> process-local gate -> claimed coordinator, but proves no production factory/installer/server/scheduler/executor/process binding, live stores/writer, runtime, or integration. Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`; runtime is `UNBOUND/NOT_RUN`.

## Calls connection-aware refcount reducer regression (2026-08-29)

Focused units for commit `519d14b1` cover the accepted Hub key plus trusted observation, per-user aggregate edges (`0 -> 1` joined and `1 -> 0` left), duplicate connected/left no-ops, unknown-left reconciliation request, rejection of active same-epoch `Left -> Connected`, structural-invalid precedence before ended-state handling, and valid ended-call descriptive no-op without an individual-left storm. They also retain the boundary that this reducer performs no provider/webhook, persistence, event, or runtime work.

Valid RED `worktree-master-backend-unit-run-4ceca6a1e260` recorded target `FAIL`, first exit `1`, build `PASS`, and expected missing-symbol vet/unit/race failures with retained stderr lengths `0/232/0/0`. Canonical GREEN `worktree-master-backend-unit-run-5f5fbdac76bd` records target/build/vet/unit/race `PASS`, exit `0`, and four empty stderr logs. Specification review is `PASS`, quality review `APPROVED`, `C0/I0/M0`.

This is unit/race evidence only: it does not prove I/O, logging, persistence, CAS/transactions, event delivery, reconciliation execution, provider/webhook integration, lifecycle/server wiring, self-hosted Jitsi runtime, cutover, or browser/media behavior. The earlier simpler reducer is unchanged; LiveKit remains active and Jitsi stays `UNBOUND/NOT_RUN`.

## Phase 151 pending-recovery composition coverage (2026-08-29)

Focused unit/race coverage for commit `fc80c9e8` verifies passive composition of a valid immutable runtime config with one complete post-create-owned Combined Ent/Mongo graph and a plain nonnil clock; typed-nil clock and every incomplete graph/config shape fail closed. The suite proves the holder returns one opaque repository identity, copies the retained execution config, and forwards fixed owner/lease values. Apart from exactly `OwnerID()`, `LeaseDuration()`, and `ExecutionConfig()` reads on the config value, it invokes no clock, delegate, persistence, or transport method.

Valid RED `worktree-master-backend-unit-run-e0417e335e61` had target `FAIL`, first exit `1`, build `PASS`, and expected missing-constructor vet/unit/race failures with stderr lengths `0/243/0/0`. `f86bc87afa18` is not canonical because only a test-side typed-nil dereference failed unit/race. Canonical GREEN `worktree-master-backend-unit-run-6dad11197ae4` records target/build/vet/unit/race `PASS`, exit `0`, and four empty stderr logs. Specification review is `PASS`; quality review is `APPROVED`, `C0/I0/M0`.

This proves no I/O, recovery execution, executor, `RunOnce`, scheduler/goroutine, process ownership, server/startup, deployment installation, live Ent/Mongo behavior, or runtime/integration flow. Production installation and live execution remain `UNBOUND/NOT_RUN`; Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

## Calls confirmed participant connection state-change evidence - 2026-08-29

Commit `8274eb0c` unit-tests the pure builder over every permitted persisted transition (`Unknown -> Connected` increment and `Connected -> Left` decrement), including the exact joined `0 -> 1` and left `1 -> 0` publication edges. It also covers descriptive no-op and reconciliation zero results; malformed keys/states/counts, counter overflow, partial persistence, wrong publication edges, mixed publication flags, invalid reconciliation, unknown-left persistence, and same-epoch reconnect all return the exact sentinel with zero payload and `shouldApply=false`. Value and pointer are verified not to implement Stringer/text/JSON output interfaces; this is an accidental-interface test, not redaction evidence.

Valid RED `worktree-master-backend-unit-run-6b5378a1ffd1` recorded target `FAIL`, first exit `1`, build `PASS`, and expected missing-builder vet/unit/race failures with retained stderr lengths `0/245/0/0`. Initial GREEN `worktree-master-backend-unit-run-54f5dc4fa055` preceded review hardening. Final canonical GREEN `worktree-master-backend-unit-run-568e22efb42b` records target/build/vet/unit/race `PASS`, exit `0`, and four empty stderr logs. Final specification review is `PASS`; quality review is `APPROVED`, `C0/I0/M0`.

This receipt proves neither a store adapter/schema/transaction nor actual atomic persistence, outbox/event publication, retries, reloads, context handling, provider/webhook processing, service/server/runtime wiring, deployment, or browser/media behavior. Conflict/non-application semantics are a private future port contract only. LiveKit remains active; Jitsi and all runtime/cutover claims remain `UNBOUND/NOT_RUN`; the Phase 150 verdict is unchanged.
## Phase 151 pending-recovery once adapter evidence (2026-08-29)

Focused units for commit `8ba30eed` cover plain and typed-nil repository rejection, nonpositive bounded config rejection, retained config copy, structural invalidity before cancellation, exact cancelled-context return before delegation, and one exact direct executor invocation for valid input without an added wrapper or retry. The spy path proves gate denial finishes before scanning, while a gate error remains testable with `errors.Is` through existing executor wrapping.

RED `worktree-master-backend-unit-run-8aa008302863` had target `FAIL`, first exit `1`, build `PASS`, expected missing-constructor vet/unit/race failures, and retained stderr lengths `0/245/0/0`. Initial GREEN `worktree-master-backend-unit-run-74986c6628b2` preceded review hardening. Final GREEN `worktree-master-backend-unit-run-ea5443e146bf` is target/build/vet/unit/race `PASS`, exit `0`, with four empty stderr logs. Final specification/quality review is `PASS`/`APPROVED`, `C0/I0/M0`.

This is bounded unit/race evidence only. No scheduler/ticker/goroutine, process owner, server lifecycle, production wiring, live-store runtime, or integration was run or added. Production installation remains `UNBOUND/NOT_RUN`; Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

## Phase 151 pending-recovery leader lease evidence (MPR-01, 2026-08-31)

Commit `9c2019fc` binds the external oracle `backend/tests/messenger/phase151_pending_recovery_leader_lease_contract_test.go` (SHA-256 `07F7B07B1344ADD91B34741E02507BBDE07E2210BCAA3677CB20A6E366911C6E`) to the single production file `backend/internal/messenger/pending_recovery_leader_lease.go`. The exact Docker launcher was `docker compose -f docker-compose.local.yml run --rm backend-unit`; no selector, pull or build flag was used.

Canonical RED container `worktree-master-backend-unit-run-15ae2c7f6e17` exited `1`: build passed and vet/unit/race failed only on the four absent MPR-01 symbols; there was no product assertion failure, panic, data race or unrelated package build failure. Canonical GREEN container `worktree-master-backend-unit-run-97be80074322` exited `0`: target/build/vet/unit/race are all `PASS`, `first_exit=0`, and all four stderr logs are empty. Both used local image `sha256:51a4cc1bf7302a4632d0e71a5bec7cb177dc8744f993b1f52b13f138f7b03035` with no RepoDigest; the disposable GREEN container was removed. Independent post-GREEN acceptance is `PASS C0/I0/M0`.

This is package-level unit/race evidence for the private port/value contract only. Receipt-to-index binding is temporal rather than cryptographic because ignored report files contain no staged blob digest. PostgreSQL/Mongo durability, scheduler/process lifecycle, migration, adapter, NATS/WS, integration, deployment and product/runtime behavior remain `NOT_RUN/UNBOUND`.

## Phase 151 MPR-09 deployment and live-acceptance matrix (2026-08-31)

This section freezes future evidence; every row is `NOT_RUN`. It does not supersede the sole queued MPR-01 oracle or turn static review into product acceptance. Host and ad-hoc execution are forbidden. Applicable product evidence may use only `backend-unit`, `frontend-unit`, `python-api`, `python-integration`, `python-e2e-smoke` and `python-e2e-full`. `python-contract` remains package/framework-only.

| ID | Required behavior | Canonical service/evidence | Current |
|---|---|---|---|
| `M09-01` | Recovery config, PG/Mongo TLS fail-start/session truth, secret-file resolution, CIDR, non-root process, liveness/readiness/standby/shutdown and accepted OCI/running `RepoDigest` identity | `backend-unit` plus authorized immutable image/Compose receipt | `NOT_RUN/UNBOUND_IMAGE` |
| `M09-02` | Expand migration once, constraints/indexes/seed, nullable legacy compatibility, old server/mail behavior, no down migration | `python-integration` | `NOT_RUN` |
| `M09-03` | Enrolled canonical create/retry/conflict, one message, exact reply/follow, one PG receipt/outbox set before Mongo completion, `Created=false` zero effects | `python-api` + `python-integration` | `NOT_RUN` |
| `M09-04` | Exact nested channel event, identifier-only unread, authoritative refetch, durable ordinary-user hook and excluded producer matrix | `frontend-unit` + `python-integration` + `python-e2e-full` | `NOT_RUN` |
| `M09-05` | Crash before/after PG commit, PG/Mongo ACK unknown, two materializers, receipt corruption, authorized empty and revoke-before-first-commit | `python-integration` | `NOT_RUN` |
| `M09-06` | NATS outage, publish/mark and consumer/ACK-loss windows, deterministic delivery headers, dead-letter and bounded redrive without exactly-once claim | `python-integration` | `NOT_RUN` |
| `M09-07` | Poison identity/redelivery/quarantine, invalid metadata retry, immutable mismatch rejection and content-free artifacts | `python-integration` | `NOT_RUN` |
| `M09-08` | Fresh recipient/provider/hook authorization, no stored preview, provider closed states and honest `outcome_unknown` duplicate risk | `python-integration` | `NOT_RUN` |
| `M09-09` | Two recovery processes, stale fence rejection, server-time takeover, two epochs with insertion/wrap/hot target, restart and bounded weak fairness | `python-integration` | `NOT_RUN` |
| `M09-10` | Enrollment-generation race, one-writer/no-bypass, sticky canonical history, normal rollback drain and separate emergency-stop report | `backend-unit` + `python-api` + `python-integration` | `NOT_RUN` |
| `M09-11` | Real API/WS/browser send/reply/unread/reconnect convergence with no duplicate visible effects | `python-e2e-smoke` then `python-e2e-full` | `NOT_RUN` |
| `M09-12` | Messenger recovery has no Mail/Calendar or Calls/Jitsi/LiveKit dependency, config mutation, subject traffic or container lifecycle effect | static manifest diff + `python-integration` | `NOT_RUN` |

Every executable receipt records Git SHA and dirty state, exact Compose digest, service/command, start/end/exit, accepted OCI manifest/platform-child digest and running `RepoDigest` when available, sanitized enrollment generation, before/after PostgreSQL/Mongo state, NATS stream sequence/delivery count, fixed-label metrics, exact fault boundary and cleanup. URL userinfo/query credentials are redacted before persistence; raw DSNs, tokens, message content, provider payloads and persistent unsalted user identifiers are forbidden. Fault injection is limited to task-owned QA databases, streams and processes; real providers, production redrive, store corruption, container kill, restore and K3s apply require separate authority. Missing evidence stays `BLOCKED` or `NOT_RUN`.

## Phase 150 JITSI-02A provider-event HMAC coverage (2026-09-01)

The `hubtest` oracle covers nil and 31/32/33-byte secret boundaries; raw-envelope lengths 0/1/65536/65537; nil, 31/32/33-byte, hex and base64 MAC shapes; digest and raw-byte mutation; correctly signed whitespace/NUL, invalid-JSON-looking and gzip-looking opaque envelopes; exact non-nil sentinel and zero rejected value; unchanged inputs; success/input isolation; fresh `Bytes()` copies; a private struct with no exported fields; and absence of value/pointer `Stringer`, `TextMarshaler` and `json.Marshaler`. Exact validation order, explicit `hmac.Equal` and absence of logging are source-review gates.

Canonical RED `worktree-master-backend-unit-run-c18eaa65d772` records target `FAIL`/first exit `1`, build `PASS`, and vet/unit/race failure only on the three absent frozen symbols. Canonical GREEN `worktree-master-backend-unit-run-6582c1e66044` records target/build/vet/unit/race `PASS`/exit `0`, four empty launcher-owned stderr logs and no fail/panic/data-race pattern. Both used image `sha256:51a4cc1bf7302a4632d0e71a5bec7cb177dc8744f993b1f52b13f138f7b03035` with no RepoDigest. Source and oracle SHA-256 are `0766FC2AA8628F5F49C92822D9951B696A9EB698F2C3A20B9748D312FF9F48F6` and `E947651F31EF297514C7AB01DECB26CCA9A7DC60E58D097C283F0032CB848D2B`; three static post-implementation reviews are `PASS C0/I0/M0`. Stale 2026-08-23 `qa-build`/`swagger-check` files in the shared report directory are excluded.

This is package-level unit/race evidence for a pure verifier only. Ingress/routes/configuration, provider compatibility, mTLS, delivery ID, replay/durability, Jitsi deployment/images/SSO/branding/recording/load/media/browser/cutover and product runtime remain `NOT_RUN/UNBOUND`.

## Phase 150 CALL-DUR-01 schema-source coverage (2026-09-01)

The external `hubtest` oracle checks exactly one descriptor for each field, exact enum languages and defaults, immutable required non-nillable/non-unique shape, absence of an update default and clean descriptor state. It reads no migration, source file or filesystem and proves no generated Ent or PostgreSQL behavior. Schema, migration and oracle SHA-256 are `E5C45C4DC2CB9336D21BDF5E300B9CE305BFD3D411E009CCD315090968BE6F5A`, `AE7F670E06C24130994653298CBE3EEE35DE86316D7AE1356A0F19529B0D69D4` and `56CB8686C9C42D5579723454825AE176B96B19D177D40362A82444E0764E032A`.

Canonical RED `worktree-master-backend-unit-run-3edf91c8ac59` had build/vet `PASS` and the intended missing-descriptor unit/race failure. First GREEN `worktree-master-backend-unit-run-70a339f09db9` is rejected: the image-contained stale generated Ent runtime produced broad `ent.init.0` panics after fields were inserted in the middle of the list. Moving the descriptors to the end of `Call.Fields()` preserved existing indices. Final canonical GREEN `worktree-master-backend-unit-run-2cbaf3f11662` has target/build/vet/unit/race `PASS`, four empty stderr logs and one named-oracle pass in unit and race. The image is `sha256:51a4cc1bf7302a4632d0e71a5bec7cb177dc8744f993b1f52b13f138f7b03035` with no RepoDigest.

Runs were serialized with the founder-approved `2 GiB` free-RAM floor, `250 GiB` C: floor, three disk samples and no concurrent Mail/Calendar build or other `backend-unit`; idle BuildKit was left untouched. This is schema-source/unit evidence only. Ent generation/image refresh, SQL parse/application/rollback atomicity, real PostgreSQL defaults/backfill/compatibility, provider selection/routing, durable presence behavior, ingress and Jitsi runtime remain `NOT_RUN` or `BLOCKED_GENERATOR`.

## Phase 150 CALL-DUR-03 neutral participant-identity coverage (2026-09-01)

The external `hubtest` oracle covers construction, UUID accessors and the separate exact connection/participant canonical languages; zero call/user/epoch components; non-v4/variant and repeated nonzero UUIDs; constructor type separation from `ConfirmedParticipantConnectionKey`; exact 134-byte identity round-trip; more than 30 malformed namespace, version, UUID, case, whitespace, NUL, non-ASCII, segment and length cases; zero safety; private fields; and absence of value/pointer Stringer, Text and JSON interfaces. Key, identity and oracle SHA-256 are `986A32F99CD21F0A5C65681DA6D5098C81CFE4CC10BA80687198CDFE6628F2F1`, `31D2EECC78CD5786490A134E4A550045758B55A54371030D647E2A725C5AF222` and `980C5982D91344A67D35A3646DD22A385AB3E08F19F3693832929011D88C0DDE`.

Canonical RED `worktree-master-backend-unit-run-a5248a2468d0` had overall `FAIL`/exit `1`, build `PASS`, and vet/unit/race failure only on the absent frozen symbols; fail-action, panic and data-race counts were zero. Canonical GREEN `worktree-master-backend-unit-run-e44a00d790b5` has target/build/vet/unit/race `PASS`/exit `0`, four empty launcher-owned stderr logs, 12 top-level CALL-DUR-03 tests passing in unit and race, and zero fail-action/panic/data-race markers. Stale `qa-build` and `swagger-check` files are excluded from both receipts.

Both runs were serialized behind the founder-approved `2 GiB` free-RAM floor while retaining the `250 GiB` C: floor, three disk samples, no concurrent Mail/Calendar build or other `backend-unit`, and idle-BuildKit-only rule. This is source/unit/race evidence only. Provider event/room/JWT behavior, authentication/authorization/confirmation, service/store/router binding, persistence, Jitsi/LiveKit runtime, deployment, browser/media and current image RepoDigest provenance remain `NOT_RUN/UNBOUND`.

## Phase 150 CALL-DUR-02A participant-connection schema coverage (2026-09-01)

The external `hubtest` oracle instantiates only `schema.CallParticipantConnection` and verifies exact nine-field order and descriptor shape; UUID defaults/immutability; closed state enum/default; optional/nillable timestamps; UTC create/update defaults; the exact existing `VersionMixin`; direct non-inverse Call/User edge ownership and delete annotations; unique identity, partial terminal-retention and non-unique user indexes; and absence of provider/delivery/token/raw-event/aggregate/outbox fields. It uses no generated Ent client, SQL, filesystem, database, network, transition or retention worker. Schema and oracle SHA-256 are `F03A139DC6CE15B43BFB8AEDA3D456FC3B35CD52DB77AE5178F4FDE70610BDE4` and `41A468D2172B11E8BF4A86CF6FFDD0B8203D2790B95844CF3E84AC69577784A0`.

Canonical RED `worktree-master-backend-unit-run-6bb1463ea563` had target `FAIL`/exit `1`, build `PASS`, and vet/unit/race failure only on the absent schema type. Canonical GREEN `worktree-master-backend-unit-run-fe3460f692e5` has target/build/vet/unit/race `PASS`/exit `0`, four empty launcher-owned stderr logs, and one named-oracle pass in unit and race with no fail/panic/data-race marker. Stale 2026-08-23 `qa-build`/`swagger-check` files in the flat report directory are excluded.

RED preflight recorded `4177220 KiB` free RAM, `311948279808` C: bytes and disk/queue `1/0, 0/0, 0/0`; GREEN recorded `6066180 KiB`, `311858552832` bytes and `0/0, 0/0, 0/0`. Both used the founder-approved `2 GiB` free-RAM floor, retained the `250 GiB` C: floor, ran with no other `backend-unit` or Mail/Calendar build, and left idle BuildKit untouched. This is only `SOURCE_ACCEPTED` descriptor evidence. Generated Ent API/runtime, current image provenance, PostgreSQL timestamp constraints, persistence, transitions, retention, routing and Jitsi/LiveKit runtime remain `NOT_RUN/BLOCKED_GENERATOR/UNBOUND`.

## Phase 150 CALL-DUR-02B participant-presence schema coverage (2026-09-01)

The external `hubtest` oracle instantiates only `schema.CallParticipantPresence` and verifies no mixins; exact seven-field order/types/defaults/mutability; generated nonzero UUID ID; immutable required call/user UUIDs; signed-int64 zero counter defaults; validators that accept zero, one and max-int64 while rejecting negatives; UTC create/update defaults; direct non-inverse Call/User edge ownership and delete annotations; exact unique call/user plus non-unique user indexes; and absence of connection/provider/delivery/legacy/outbox/event fields. It uses no generated Ent client, SQL, filesystem, database, network, adapter or transition. Schema and oracle SHA-256 are `026914D81E224C2B0D3492FEB7AA9FE7927BE643BA9D3388336C932533275B61` and `96231C8C178699B2967DD8FD6295C89D3000DFD279C1516ED7BDEB666584131C`.

Canonical RED `worktree-master-backend-unit-run-ff7b47d55060` had target `FAIL`/exit `1`, build `PASS`, and vet/unit/race failure only on the absent schema type. Canonical GREEN `worktree-master-backend-unit-run-ac7663bf7081` has target/build/vet/unit/race `PASS`/exit `0`, four empty launcher-owned stderr logs, and one named-oracle pass in unit and race with no fail/panic/data-race marker. Stale 2026-08-23 `qa-build`/`swagger-check` files are excluded.

RED preflight recorded `5782524 KiB` free RAM, `311847276544` C: bytes and disk/queue `0/0, 0/0, 0/0`; GREEN recorded `3064168 KiB`, `325693759488` bytes and `5/0, 0/0, 0/0`. Both used the founder-approved `2 GiB` free-RAM floor, retained the `250 GiB` C: floor, ran with no other `backend-unit` or Mail/Calendar build, and left idle BuildKit untouched. This is only `SOURCE_ACCEPTED` descriptor evidence. Counter/sequence projection truth, atomic connection-plus-aggregate mutation, underflow/lost-update prevention and monotonic replay handling belong to later adapter/integration tests. Generated Ent/runtime, PostgreSQL behavior, routing and Jitsi/LiveKit runtime remain `NOT_RUN/BLOCKED_GENERATOR/UNBOUND`.

## Phase 150 CALL-DUR-02C provider-delivery schema coverage (2026-09-01)

The external `hubtest` oracle instantiates only `schema.CallProviderDelivery` and verifies exact ten-field order/types/defaults/mutability/nullability/storage; exact enum values; generated UUID ID; immutable provider delivery facts; exactly NotEmpty plus MaxLen(255) delivery-ID validators; nullable mutable GDPR user; UTC received time; nullable terminal time; no mixin/edges/scalar unique fields; safe field-name denylist; and exact dedupe/cleanup/user indexes. Root source review separately confirms no field annotation. It uses no generated Ent client, SQL, filesystem, database, network, provider verifier or transition worker. Schema and corrected oracle SHA-256 are `41219FA765D92692B5664E9C4B5FD69E169748B114FD933D65C98E7C3E9631FE` and `918597670D4D9A962312F5612BE22021E163CFFF17BAF8148408E3E3573EAC11`.

Canonical RED `worktree-master-backend-unit-run-39b4a4226df2` had target `FAIL`/exit `1`, build `PASS`, and vet/unit/race failure only on the absent schema type. First GREEN `worktree-master-backend-unit-run-f8915f185a7a` is rejected oracle evidence: the enum helper falsely prohibited the required result default. After the one-line discovered/draft correction, final GREEN `worktree-master-backend-unit-run-21218f718a4e` has target/build/vet/unit/race `PASS`/exit `0`, four empty launcher-owned stderr logs, and one named-oracle pass in unit and race with no fail/panic/data-race marker. Stale shared-directory `qa-build`/`swagger-check` files are excluded.

RED preflight recorded `6186668 KiB` free RAM and `325600043008` C: bytes; after one queue-10 sample, three repeated disk samples were `0/0` before launch. Rejected GREEN recorded `3937196 KiB`, `325582229504` bytes and `34/1, 28/0, 0/0`; final GREEN recorded `8561656 KiB`, `325761441792` bytes and `4/0, 0/0, 0/0`. Every run retained the `2 GiB` RAM and `250 GiB` C: floors, no other `backend-unit` or Mail/Calendar build, and no idle-BuildKit intervention. This is only `SOURCE_ACCEPTED`. Later tests must cover authenticated ingress, immutable-fact collision rejection, tombstone no-relink, conditional result/user/terminal transitions, 02M DB checks and replay/GDPR concurrency. Generated Ent/runtime, SQL application and product runtime remain `NOT_RUN/BLOCKED_GENERATOR/UNBOUND`.

## Phase 150 CALL-DUR-02M owner-migration source coverage (2026-09-01)

Migration commit `e9a81ba4` and SHA-256 `CBD4A269F86F18A57884FD00B0F61B69A9A5A9B0499D5BC6C919C3BED3E484CB` are accepted only as `MIGRATION_SOURCE_ACCEPTED`. Lightweight static inspection found one `BEGIN`, one `COMMIT`, three tables, three unique indexes, five non-unique indexes, no forbidden existence/destructive guard, no delivery FK and no delivery `call_id` index. Independent SQL, schema-parity and security transcript reviews returned `C0/I0/M0`; direct reviewer filesystem provenance was unavailable because of ACL denial.

No Docker command or unit oracle was run because 02M is a migration-only atom. PostgreSQL parse/apply/rollback/replay, clean-database generated parity, catalog columns/FK actions/index semantics, constraint enforcement, query/index use, User-first GDPR transaction/rollback, ingress-versus-delete concurrency, opaque byte-exact provider identity and runtime remain `NOT_RUN`. These require the later authorized integration gates; they must not be inferred from static source review.

`CALL-DUR-01/02A/02B/02C` remain `BLOCKED_GENERATOR`. `ENT-CALL-GEN-01` is the next serial gate and remains `BLOCKED_AUTHORITY_AND_RESOURCES` until a founder-authorized single image rebuild binds every accepted schema/generator input to generation sentinels, image identity and a later canonical `backend-unit` receipt.

## Phase 150 ENT-CALL-GEN-01 generation and backend-unit coverage (2026-09-01)

`ENT-CALL-GEN-01` is `ACCEPTED` at clean source HEAD `5b08e9aeea05251b83e62a20b0fc85296c40711e`. The 96 declared schema/generator/module/Docker/Compose inputs had stable aggregate SHA-256 `7AF9CDA7D975F3D74E3A438C9CA65599FEC6AC429E4B89460FD7A87F7A41D5B7`. Host `backend/ent` contained only 91 source files and no generated output before or after execution.

The exact Compose image build ran `go generate ./ent` and the three Dockerfile sentinels successfully, producing local image `sha256:cbd0b9f06bc7e499f9dd1a8a9973e4981b3e75b87f207fda259118c62101b0b9`. Exact disposable container `worktree-master-backend-unit-run-e0bb8f6add37` then exited `0` with target/build/vet/unit/race `PASS`, first exit `0`, four empty stderr logs and zero fail/panic/data-race markers. The four accepted Phase-150 schema descriptor tests passed under both unit and race; result SHA-256 is `622DFACDD94271805EB680200E153F7C2614D591B1447B8A18A25C972279B1E2`.

Pre-build, pre-run and post-run resource windows passed the `2 GiB` RAM, `250 GiB` C:, two-of-three disk and exclusivity guards; no retry, second build or concurrent QA runner occurred. Independent build/provenance/resource reviews are `C0/I0/M0`. The generation blocker is removed from 01/02A/02B/02C and the queue may advance to `CALL-DUR-02G`. No `RepoDigest` exists; remote/CI publication, generated clean-DB parity, SQL apply/rollback/replay, live PostgreSQL, ingress, GDPR/concurrency and product runtime remain `NOT_RUN`.

## Phase 150 CALL-DUR-02G transactional GDPR serialization coverage (2026-09-01)

Commit `64e68da443ad0e842aeec300b4c206e6eb89e102` and the frozen external oracle cover the Audit transaction coordinator plus generated Ent query shape. The service test proves the in-transaction User lock is first, founder is rechecked from the locked role, one nonzero UTC terminal time feeds one durable erase before the unchanged cascade, every pre-commit failure stops and rolls back before Mongo, and commit failure returns without synthetic rollback or Mongo. Fake-dialect query tests prove exact connection delete, presence delete, pending/reconcile terminalization, applied/ignored user clearing, predicates/bindings/order and immutable replay facts.

Canonical RED `worktree-master-backend-unit-run-1a1f47028d19` exited `1`: build passed and vet/unit/race failed only on the absent `LockUserForGDPR` and `EraseUserDurableCallData` methods. The flat launcher directory was later overwritten by the required GREEN; the canonical verification receipt preserves the RED source binding, diagnostics and report SHA-256 values without claiming retained raw logs. Canonical GREEN `worktree-master-backend-unit-run-14226df38bc9` exited `0` with target/build/vet/unit/race PASS, empty stderr logs, all focused tests/subtests passing under unit/race and zero fail/panic/data-race markers. Oracle SHA-256 is `6A02B50AFB647A9DF75F56ACD2ED3017F90B352B246D3C6AEAC2B4DE18554C00`.

Both exact no-selector/no-build/no-pull runs reused local image `sha256:cbd0b9f06bc7e499f9dd1a8a9973e4981b3e75b87f207fda259118c62101b0b9`. Resource windows passed the `2 GiB` free-RAM and `250 GiB` C: floors, the two-of-three disk rule and exclusivity guard; no Mail/Calendar build or other QA runner overlapped. Independent code, security/concurrency/privacy and evidence reviews are `C0/I0/M0`, yielding `SOURCE_ACCEPTED`. Real PostgreSQL lock/deadlock/commit behavior, constraints, ingress/GDPR concurrency, Mongo integration, product runtime and remote/CI provenance remain `NOT_RUN/UNBOUND`.

## Phase 150 CALL-DUR-08A provider-room resolver coverage (2026-09-01)

Commit `eae39e7df9f6160b591e3dd36a798564a19bd7c6` and frozen oracle SHA-256 `E4622E9BEE982FBB9024EAFD3095E829E273FC41D66CD2DD96229504A0C6AAD3` cover the closed persisted provider type and exact provider-plus-room Ent resolver. Invalid/no-query cases include nil and typed-nil context, nil repository/client, empty/unknown/case-drift/padded provider, and empty/whitespace/padded room. Valid cases prove exact provider AND room predicates/bindings, both providers across active/ended Calls, generated row mapping, Ent not-found/not-singular and unwrapped driver-error identity.

Initial RED `worktree-master-backend-unit-run-9824f75c8109` is rejected test-triage evidence because the oracle contained two independent return-arity defects. Canonical RED `worktree-master-backend-unit-run-1fc98ce666ad` exited `1` with build PASS and vet/unit/race failing only on absent 08A API symbols. Canonical GREEN `worktree-master-backend-unit-run-46c76cc67313` exited `0` with target/build/vet/unit/race PASS, four empty stderr logs, zero global fail/panic/data-race markers and 24 focused PASS actions with zero FAIL in each of unit and race.

Both accepted runs reused local image `sha256:cbd0b9f06bc7e499f9dd1a8a9973e4981b3e75b87f207fda259118c62101b0b9` without selector/build/pull. RED, GREEN and post-GREEN resource windows passed the `2 GiB` RAM, `250 GiB` C:, disk and exclusivity guards; no Mail/Calendar build or other QA runner overlapped. Final code, security/authority and evidence reviews are `C0/I0/M0`, yielding `SOURCE_ACCEPTED`.

This evidence is isolated repository source/unit/race only. Historical migration/default/backfill persistence, real PostgreSQL planner/cardinality/error behavior, trusted provider ingress/authentication, delivery identity/replay, provider selection, routing, callbacks, service/handler/wiring, Jitsi activation/deployment/media/browser and product runtime remain `NOT_RUN/UNBOUND`.

## Phase 150 CALL-DUR-08B provider capability registry coverage (2026-09-01)

The frozen external `hubtest` oracle covers exact `livekit|jitsi` identities, empty/nil/typed-nil/duplicate/invalid registrations, missing and unknown capability bits, per-capability bidirectional interface mismatch, both insertion orders, invalid receiver state, identity-before-capability validation, known omitted versus unknown IDs, exact sentinel identity, no fallback and no provider calls after construction. It proves LiveKit exposes all five capabilities and the Jitsi wrapper exposes only call-aware credential issuance, including absence of legacy lifecycle, guest, participant and room-string surfaces. It performs no credential issuance, DB, filesystem, network or generated-provider access.

Canonical RED `worktree-master-backend-unit-run-df5898408ca3` had build PASS and only missing 08B symbols in vet/unit/race. The first exit-zero run `worktree-master-backend-unit-run-71806db24507` is rejected because Mail `backend-unit` overlapped its race stage. Canonical clean GREEN `worktree-master-backend-unit-run-9af81ee5d32d` exited `0` with target/build/vet/unit/race PASS, four empty stderr logs, zero fail/panic/data-race markers and 32 focused PASS actions in both unit and race. Frozen oracle SHA-256 is `E0913BECB2D46C00EC1E4EC467741CFAE8911869ECC6CFC7B9562525CC67D918`; implementation commit is `b66dada8`.

The accepted run respected the `2 GiB` free-RAM and `250 GiB` C: floors, three disk samples and exclusivity rule; only idle BuildKit coexisted. This is `SOURCE_ACCEPTED` source/unit/race evidence only. Provider selection/routing, persistence mapping, guest denial/HTTP behavior, config activation, service/server wiring, integration/PostgreSQL, deployment, media/browser and product runtime remain `NOT_RUN/UNBOUND`.

## Phase 150 CALL-DUR-08C1 durable Jitsi issuer coverage (2026-09-01)

The sole external oracle SHA-256 `26E92F146FAD1AEEE6A5D82CA064E5234EAAE4730C089982B51AF351FEE2417F` freezes exact interface shape, Jitsi issuer and wrapper satisfaction, durable provider participant identity in the JWT, exact claim allowlist, active/canonical Call validation, identity-call binding, canonical server-name validation, fixed zero-on-error behavior, legacy UUID/padded-name parity, and absence of broader Jitsi provider capabilities. It performs no DB, filesystem, network, provider runtime or generated-Ent access.

Canonical RED `worktree-master-backend-unit-run-7a9333829acc` retained build PASS and failed vet/unit/race only on absent 08C1 symbols. The first exit-zero run is rejected as `INVALID_MAIL_OVERLAP`. Founder-authorized container `worktree-master-backend-unit-run-5561b7d8ac95` is qualified `USER_AUTHORIZED_CONCURRENT_MAIL_DEV_INFRA_GREEN`: target/build/vet/unit/race passed, stderr was empty, focused durable cases recorded 16 PASS and zero FAIL in each of unit/race, and global fail/panic/data-race markers were zero. The Mail development stack was active, no Mail QA/test/build runner overlapped, and one RAM sample dipped just below 2 GiB before immediate recovery; the qualified label must be retained.

Independent code/spec and security reviews are `C0/I0/M0`; the raw evidence review's packaging-only M1 is closed by the separate launcher receipt in `150-VERIFICATION.md`. Commit `7b1d63fa` is `SOURCE_ACCEPTED` for source/unit/race only. Allocation/transaction/Ent/PostgreSQL, provider routing, guest/lifecycle, service/server/HTTP/client, Jitsi deployment/media/browser/cutover and product runtime remain `UNBOUND/NOT_RUN`; 08C is not complete and 08C2 is next.

## Phase 150 CALL-DUR-08C2 durable allocator coverage (2026-09-01)

The sole external oracle SHA-256 `0AE525C86A75369BDB5790F26B1308F42574FB545883328DA24011D4150B03A9` covers the exact allocator API and sentinels; nil/typed-nil dependencies, context and transaction; zero IDs; canonical user-name and Call/origin validation; exact seven-operation trace and arguments; one observable fresh neutral key/provider identity; channel and instant authorization outcomes; every dependency stop point; cause suppression and rollback ownership; zero and partial credentials; final cancellation; and ambiguous commit without rollback/retry/reissue. It exposes no Ent, DB, network, provider runtime, lifecycle, presence/refcount or outbox surface.

Canonical RED `worktree-master-backend-unit-run-6d1badf07a01` failed only on absent 08C2 symbols. No post-implementation GREEN is accepted: `worktree-master-backend-unit-run-2fae431c8890` produced unrelated compiler kills and race `137`, while serialized attempts `6310a3f3cb18`, `a3d8cc4728bf`, and `28cae114b93b` were stopped after `OOMKilled=true` under the required 2 GiB ceiling. Parallelism and soft-heap experiments retained the full fixed build/vet/unit/race entrypoint but are resource diagnostics, not canonical receipts.

Three final independent reviews are `C0/I0/M0`; commit `af20766b` is `STATIC_ACCEPTED`. Executable build/vet/unit/race remain `NOT_RUN / BLOCKED_INFRA_2GIB_OOM`, so no PASS, source/runtime GREEN, PostgreSQL, provider activation or deployment claim is made. 08C3 and Phase-150 completion remain blocked.

## hard2g backend-unit observer candidate (2026-09-05)

This follow-up is **STATIC_ACCEPTED / EXECUTABLE_NOT_RUN**, not a new GREEN or
replacement for the historical 2 GiB baselines above. The candidate fixes a
Docker-only envelope of 2 GiB/no swap, 6 CPU, `/tmp` 64 MiB, `GOCACHE` 768 MiB,
`GOTMPDIR` 384 MiB, and a 512 MiB clean threshold. Its only read-only source
roots are `backend/cmd`, `backend/internal`, `backend/tests`,
`backend/ent/schema`, `backend/migrations`, and `backend/testframework`; reports
remain writable.

The fixed observer invocation is `python testframework/observers/backend_unit/observe.py --compose-project worktree-master`. Final choreography is root
handoff `prebuild`, the single authorized build, one `python-contract` GREEN,
root handoff `postgreen`, then start the separately managed observer. It does
**not** launch Compose: wait for `READY`, then separately issue exactly once
`docker compose -f docker-compose.local.yml run --rm backend-unit`. Normal closure is `COMPLETE`. Only safe bootstrap/execution TERM handling
may close `ABORTED`, with one guarded stop at most and a matching same-session
`die` proving exit `143`, no OOM, and no error. Cold, planning, or uncertain
interruption closes no receipt and returns `75`.
