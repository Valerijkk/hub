# Hub — Тестовая документация

## Canonical frontend-unit: предел workers — 2026-09-06

В service-local environment зафиксированы MAX=2/MIN=1 для Vitest threads
и forks. Память остаётся2GiB/no-swap, CPU6, image и exact Compose command
не изменены. Набор/assertions/таймауты сохранены. Составное дерево с edit/
composer кандидатами прошло TypeScript и197файлов/758тестов за114.95s,
без failed/pending/todo и whole-run unhandled errors. Это не runtime
доказательство каждого промежуточного коммита. Два фактических workers и
engine exit0 наблюдались; все8 receipts сохранены отдельно. Ошибочные ранние
прогоны не удалены и не переименованы в PASS.
[Контракт, provenance и ограничения](superpowers/plans/2026-09-06-frontend-unit-worker-budget.md).

## Messenger: поиск, удаление и действия в тредах — 2026-09-06

Принятый составной Docker frontend-unit: TypeScript PASS;195 файлов,
728/728 тестов PASS за64.12s;2GiB/no-swap,6CPU,readonly,Compose exit0.
62 новых genuine unit-случая:35 поиск,20 удаление,7 передача capabilities.
Независимые source reviews C/I/M0/0/0; исходники/oracles не менялись в прогоне.
Предыдущий RED684PASS/32FAIL из716 и3 uncaught DELETE errors сохранён,
ошибки исправлены без ослабления assertions. В GREEN нет whole-run
unhandled-error failure, но остаются прежние React act/description warnings.
Отчёты: testframework/reports/frontend-unit-messenger-lifecycle-green-20260906/.
Start/image/resource limits наблюдались; post-run Docker event history
не вернул события, поэтому отдельные die/OOM timestamps NOT_CAPTURED.
Browser/API/storage/media и готовность модулей этим прогоном не подтверждены.
Предыдущая ошибка загрузки oracle из-за отсутствующего user-event исправлена
через установленный RTL fireEvent и не считается поведенческим RED.
[Контракт и точные отчёты](superpowers/plans/2026-09-06-search-dialog-lifecycle.md).

## Messenger thread read-through: ветка calls-feature-codex, 2026-09-06

Последний принятый для этого атома Docker frontend-unit: TypeScript PASS;191 файлов /666
тестов PASS,58.20s,2GiB/no-swap/6CPU/readonly, engine/Compose exit0, контейнер
удалён.36 новых genuine случаев в thread-read-receipts.test.tsx проверяют
видимый конец ответа, защиту scope/прав, дедупликацию, ошибки и явный retry.
Предшествующий RED662PASS/4FAIL сохранён: два ROOT=code исправлены; два
ROOT=test требовали дождаться нового наблюдателя после focus/refetch.
Ассерты, таймауты и runner не ослаблялись. Независимое review C/I/M0/0/0.
[Хеши, RED/GREEN и ограничения](superpowers/plans/2026-09-05-thread-read-receipts.md).
Реальный browser/API/storage/media остаётся NOT_RUN; Jitsi не активирован.

## Messenger follow UI: ветка calls-feature-codex, 2026-09-05

Предыдущий frontend-unit: TypeScript PASS, 190 файлов / 630 тестов PASS за 54.44s через
`docker compose -f docker-compose.local.yml run --rm frontend-unit`, 2 GiB
без swap. 29 новых сценариев в `frontend/tests/messenger/thread-follow-control.test.tsx`;
подробные RED/GREEN, hashes и границы — [в контракте](superpowers/plans/2026-09-05-thread-follow-control.md).
Четыре таймаута по 15s предыдущего RED не повторились без изменения тестов или
профиля; тот отчёт сохранён. Browser/API/media не проверялись этим target.
Backend ниже — исторический candidate: актуальные Docker-only изменения
runner и подтверждённый Ent compiler OOM при 2 GiB описаны в
[backend continuation](superpowers/plans/2026-09-05-docker-only-backend-packet.md).
Полный framework snapshot 529 PASS / 71 FAIL остаётся NOT_ACCEPTED; frontend GREEN
не делает backend/framework или модули готовыми.

## hard2g `backend-unit` candidate

Status: **STATIC_ACCEPTED / EXECUTABLE_NOT_RUN**. This is not a new `PASS` and
does not replace historical baselines. The Docker-only candidate fixes 2 GiB
with no swap, 6 CPU, `/tmp` 64 MiB, `GOCACHE` 768 MiB, `GOTMPDIR` 384 MiB, and a
512 MiB clean threshold. Read-only source binds are exactly `backend/cmd`,
`backend/internal`, `backend/tests`, `backend/ent/schema`,
`backend/migrations`, and `backend/testframework`; reports are the only writable
bind.

Normal terminal evidence is `COMPLETE`. TERM can seal `ABORTED` only from safe
`bootstrap` or `execution`; cold, planning, and uncertain paths return `75`
without a receipt. The abort proof allows at most one guarded stop and requires
the matching same-event-session `die` event with exit `143`, no OOM, and no
error. Root handoff owns `prebuild` then `postgreen`; the observer alone adds
`observer_snapshot` before its final run choreography.

> **Обновлено:** 2026-08-24 · Стратегия: `docs/test_plan.md` · Coverage: `.planning/COVERAGE.md` · Quality audit: `.planning/audits/2026-06/AUDIT-2026-06-22-quality-processes.md`
> **Edge/Corner backlog:** [`.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`](../.planning/archive/july-2026/EDGE-CORNER-CATALOG.md) — Happy/Neg/Edge/Corner по доменам (источник будущих TC; CAL-SYNC C1–C15)
>
> **Current QA authority:** legacy commands and paths below are inventory, not executable canon.
> **Minimal slice status (2026-08-23):** the 14-path target-unit curation and
> 241 hook-dependent deletion batches have accepted receipts. Full legacy-root
> reconciliation is 349 relocations + 232 net component deletions + 241
> hook-dependent deletions = 822 removed source paths. Protected Mail/Calendar
> paths remain outside this merge. Runnable genuine
> units live under `backend/tests/**` and `frontend/tests/**`; historical
> legacy paths/counts below are inventory only and never an execution fallback.
> Единственная публичная точка — `docker compose -f docker-compose.local.yml run --rm <exact-service>`.
> Допустимы только восемь declared QA services; host test processes запрещены.

---

## Стратегия тестирования

### Уровни

| Уровень | Инструменты | Расположение | CI job |
|---------|-------------|--------------|--------|
| Unit (Backend) | Go testing + testify | genuine product units: `backend/tests/{domain}/` | `backend-unit`: PASS (build/vet/unit/race) |
| Unit (Frontend) | Vitest + RTL | `frontend/tests/{domain}/`, `frontend/tests/support/` | `frontend-unit`: TypeScript PASS; Vitest 162/162 files, 427/427 tests |
| Framework | pytest | `testframework/selftests/**` | `python-contract`: PASS (182/182 framework selftests; no SUT product tests) |
| API / integration / E2E | pytest + httpx/Playwright | outside the minimal slice | `NOT_RUN`; not a merge gate for this slice |
| Load | Locust | outside the minimal slice | `NOT_RUN`; not a merge gate for this slice |

### Принципы

1. **Execution:** exact service только через root Compose `run --rm`; PASS относится только к фактически запущенному target и не переносится на `NOT_RUN` targets.
2. **Historical metrics:** statement coverage и legacy counts — inventory, не текущий migration/release gate.
3. **Auth тесты:** Magic Link only (v2.5) — Telegram login endpoints **удалены**
4. **E2E:** legacy TypeScript inventory удаляется, greenfield target — Python suite.
5. **Triage:** spec-first — `.cursor/skills/test-triage/SKILL.md`

### `backend-unit`: current-source, low-disk profile

For the current hard2g candidate, direct `docker compose ... run --rm
backend-unit` is not an executable canon. Follow the exact observer
choreography declared by the hard2g candidate above: root handoff `prebuild`,
the one authorized build, `python-contract` GREEN, root handoff `postgreen`,
then observer snapshot and its one final `backend-unit` run. Its status remains
**STATIC_ACCEPTED / EXECUTABLE_NOT_RUN** until that choreography has produced
new evidence.

`backend-unit` uses a **partial source overlay**: exactly these current checkout
roots are mounted read-only at matching `/opt/hub/source/backend/**` targets:
`backend/cmd`, `backend/internal`, `backend/tests`, `backend/ent/schema`,
`backend/migrations`, and `backend/testframework`. It deliberately does not
mount either the `backend/` or `backend/ent/` root, so generated Ent packages
from the QA image remain authoritative while ordinary edits in those six roots
are visible without rebuilding the image.

The image retains the resolved Go module cache; `/reports` remains the only
writable bind for receipts and logs. The hard2g envelope uses a `noexec`
`GOCACHE` tmpfs at `/var/cache/hub-go` capped at 768 MiB and an executable
`GOTMPDIR` tmpfs at `/var/tmp/hub-go` capped at 384 MiB. Both are within a
2 GiB, no-swap container limit (6 CPU); the required clean threshold is
512 MiB. Earlier 3 GiB/4 GiB/10 GiB figures are historical topology evidence,
not current candidate configuration or execution authority.

An explicit, separately authorized rebuild is required for `backend/go.mod` or
`backend/go.sum` changes, generated Ent output (including a schema change that
requires regeneration), or edits outside the six overlaid roots. Do not add
`--build` for ordinary source or test edits within the partial overlay.

### `frontend-unit`: current-source, low-disk profile

Use the canonical command exactly as declared above:

```text
docker compose -f docker-compose.local.yml run --rm frontend-unit
```

`frontend-unit` mounts the current `frontend/src` and `frontend/tests` roots
read-only at their matching `/opt/hub/source/frontend/**` image targets. The
`/reports` host bind is explicit for receipts and logs; its
`/reports/frontend-unit/home` and `/reports/frontend-unit/tmp` children are
separate RAM tmpfs mounts. The common 64 MiB `/tmp` mount is retained first;
the two report children are bounded to 512 MiB and 4 GiB respectively, with
`noexec,nosuid,nodev`, UID/GID 10001, and mode 0700. The service has a
6.0 CPU and 10 GiB memory ceiling, which also bounds those tmpfs mounts.

The resolved npm dependency tree and non-overlaid frontend inputs remain image
owned. An explicit, separately authorized rebuild is required for
`frontend/package.json`, `frontend/package-lock.json`, or a source change
outside the two mounted roots; ordinary `src` or `tests` edits do not add
`--build`. On 2026-08-24, the latest canonical `frontend-unit` command
executed this exact overlay with an existing local image (no build, pull, or
package install) and exited 0. Its latest receipt records TypeScript `PASS` and
Vitest `PASS`: 162 files / 427 tests in 47.28s; Vitest JSON records 340/340
suites and `result.json` preserves `first_exit:0`. This supersedes the prior
Task 3 receipt of 162 files / 425 tests in 45.30s, which remains historical,
and the earlier pre-Task3 receipt of 161 files / 419 tests in 45.83s. This
evidence is limited to this frontend unit target and topology;
it does not imply a deployed runtime, API/integration, or Jitsi result.

### `python-contract`: current-framework source profile

Use the canonical command exactly as declared above:

```text
docker compose -f docker-compose.local.yml run --rm python-contract
```

`python-contract` mounts only the current `testframework/src` and
`testframework/selftests` roots read-only at the matching image targets
`/opt/hub/testframework/src` and `/opt/hub/testframework/selftests`, plus
read-only copies of `docker-compose.local.yml` and the Phase 130 plan at
`/opt/hub/docker-compose.local.yml` and
`/opt/hub/.planning/phases/130-tf-traceability-gates/PLAN.md`. `/reports`
remains an explicit writable host bind for receipts. The resolved virtual
environment, package metadata/lock, policies, schemas, Docker assets, and
other image-owned configuration remain authoritative in the image; no SUT
product source is mounted by this profile.

An explicit, separately authorized rebuild is required for
`testframework/pyproject.toml`, `testframework/uv.lock`,
`testframework/.python-version`, `testframework/policies/**`,
`testframework/schemas/**`, `testframework/docker/**`, or another
image-owned configuration change. Ordinary edits under the two mounted roots
do not add `--build`. On 2026-08-24, the canonical `python-contract` command
executed this exact partial overlay with an existing local image (no build,
pull, or package install), exited 0, and passed 182 framework selftests in
9.53s. This evidence is limited to the framework contract target and topology;
it does not imply product-SUT, deployed-runtime, API/integration, or Jitsi
results.

### Исторический inventory snapshot (2026-07-29)

| Слой | Метрика |
|------|---------|
| Backend test files (`backend/testframework/`) | **812** `_test.go` (0 в `internal/` после TF-M08; measured 2026-07-29) |
| Backend `export_hooks.go` | **39** (`//go:build hubtest`; measured 2026-07-29) |
| Backend `internal/*_test.go` | **0** |
| Frontend Vitest | **179** files (`stages/unit/frontend-test/`; measured 2026-07-29; 61 `.ts` + 118 `.tsx`) |
| Playwright | **20** spec files (`stages/e2e/domains/`) |
| Overall `internal/` coverage | **83.0%** (2026-06-22 Docker baseline) |

---

## Исполнение и evidence

Backend/Frontend genuine units принадлежат целевым roots из таблицы выше.
Legacy `backend/testframework/**`, static-contract Vitest и TypeScript
Playwright удалены; их historical inventory не запускается как fallback.

Применимый target запускается только через root Compose `run --rm` с exact service.
QA services disposable, а integration/e2e используют изолированную QA-инфраструктуру.
Исторические legacy QA entrypoints архивны; отсутствие прогона через `docker compose -f docker-compose.local.yml run --rm <service>` не заменяется green-claim.

---

## API-тесты по модулям

### Формат: Ожидаемые HTTP-коды

#### Auth (Magic Link only — v2.5+)

> Telegram login (`/auth/telegram*`) **удалён** (ADR-AUTH-01). Telegram — только notify bind.

| Endpoint | 200 | 400 | 401 | 404 | 429 |
|----------|-----|-----|-----|-----|-----|
| POST /auth/magic-link | Код отправлен | Невалидный email | — | — | Rate limit (3/10m) |
| POST /auth/magic-link/verify | Токены выданы | Неверный код | — | — | Rate limit |
| POST /auth/refresh | Новый access token | Невалидный refresh | 401 expired | — | — |
| GET /auth/me | Профиль | — | Не авторизован | — | — |
| PUT /auth/me | Обновлён | Невалидные данные | Не авторизован | — | — |
| POST /auth/logout | Выход | — | Не авторизован | — | — |
| GET /auth/sessions | Список сессий | — | 401 | — | — |
| DELETE /auth/sessions/:id | Сессия отозвана | — | 401 | — | — |

#### Projects

| Endpoint | 200/201 | 400 | 401 | 403 | 404 |
|----------|---------|-----|-----|-----|-----|
| POST /projects | Создан | Невалидные данные | Нет токена | - | - |
| GET /projects | Список | - | Нет токена | - | - |
| GET /projects/:id | Проект | - | Нет токена | Не участник | Не найден |
| PUT /projects/:id | Обновлён | Невалидные данные | Нет токена | Нет прав | Не найден |
| DELETE /projects/:id | Удалён | - | Нет токена | Нет прав (Owner) | Не найден |
| POST /projects/:id/members | Добавлен | Невалидные данные | Нет токена | Нет прав | Не найден |

#### Tracker

| Endpoint | 200/201 | 400 | 401 | 403 | 404 | 409 |
|----------|---------|-----|-----|-----|-----|-----|
| POST /projects/:id/issues | Создана | Невалидные данные | Нет токена | RBAC | Проект не найден | - |
| PUT /projects/:id/issues/:key | Обновлена | Невалидные данные | Нет токена | RBAC | Задача не найдена | Version conflict |
| DELETE /projects/:id/issues/:key | Удалена | - | Нет токена | RBAC | Не найдена | - |

#### Messenger

| Endpoint | 200/201 | 400 | 401 | 403 |
|----------|---------|-----|-----|-----|
| POST /messenger/channels | Создан | Невалидные данные | Нет токена | - |
| POST /messenger/channels/:id/messages | Отправлено | Пустое сообщение | Нет токена | Не участник |
| POST /messenger/channels/:id/join | Присоединён | - | Нет токена | Приватный канал |
| POST /messenger/dm | DM создан/получен | Невалидный userId | Нет токена | - |

##### Phase 149 Task 2A: message persistence contract

- `backend/tests/messenger/message_idempotency_persistence_test.go` covers DTO/response nonce fields, sender-and-nonce isolation, the empty-nonce transitional insert, and concurrent mock create-or-return with one canonical ID.
- Both Messenger repository implementations are compile-checked through `backend-unit`; the test mock exposes a synchronized nonce index, create-call counter, and deterministic result override for Task 2C service units.
- Real MongoDB unique-index and upsert concurrency behavior is `NOT_RUN`: mock/unit evidence must not be treated as Mongo integration evidence.
- Runtime verdict for the current source is recorded only from the canonical Docker `backend-unit` receipt.

##### Phase 149 Task 2B: mandatory Messenger index startup

- If `messengerMongoRepo.EnsureIndexes(d.ctx)` fails, `wireMessenger` preserves the structured error log and exits with status 1 before constructing the Messenger combined repository, services, or routes.
- The sole canonical `backend-unit` Docker regression run completed `PASS` for build/vet/unit/race (`result.json`: `first_exit:0`; aggregate receipt `2026-08-24T17:31:34.1795517Z`). It does not exercise a running server.
- A real failing-startup process test is `NOT_RUN` because this direct `os.Exit` path has no genuine isolated unit seam; MongoDB integration execution of the required-index failure is also `NOT_RUN`.

##### Phase 149 Task 2C: message send idempotency orchestration

- `backend/tests/messenger/service_messages_test.go` covers canonical retry before parent revalidation, retry after parent deletion, concurrent create-or-return with one reply increment, channel and payload fingerprint conflicts, deleted canonical rejection, `Created=false`/deleted-canonical side-effect suppression, and the empty-nonce transitional create path.
- The payload table covers deterministic JSON number serialization, attachment order and fields, voice URL/duration, semantic mention and mention-group set normalization, and nil-versus-empty collection compatibility. It also proves that authorization occurs before nonce-derived canonical content is returned.
- Fresh canonical Docker `backend-unit` evidence is `PASS` for build, vet, unit, and race (`result.json`: `first_exit:0`; aggregate receipt `2026-08-24T18:11:52.4071814Z`). It covers both Task 1 thread-parent and all Task 2C service units.
- Unit assertions observe only the accepted repository mock seams (canonical map, create counter, reply count). NATS/unread are not instrumented as fake runtime evidence; MongoDB index/upsert concurrency and startup-process behavior remain `NOT_RUN` pending phase-owned integration/process coverage.

##### Phase 149 Task 3: composer nonce retry lifecycle

- `frontend/tests/messenger/message-composer-retry-nonce.test.tsx` covers optional nonce API types, one immutable text/attachment, GIF, sticker, or voice snapshot per logical submit, same-nonce retry, new-draft preservation, cancel/new-nonce behavior, and owned-toast cleanup on unmount.
- The test observes one persistent, non-dismissible retry toast per failed snapshot, optimistic text and attachment clearing, retry success invalidation and owned-toast dismissal. It does not claim persisted outbox, reload/offline/reconnect replay, or a runtime delivery guarantee.
- Canonical Docker `frontend-unit` RED evidence was captured before this implementation. Final Task 3 Docker GREEN is `PASS`: `first_exit:0`, TypeScript `PASS`/exit 0, Vitest `PASS`/exit 0, stdout 162/162 files and 425/425 tests in 45.30s; Vitest JSON records 340/340 suites and 425/425 tests. This historical Task 3 receipt remains unit-target evidence only, not browser E2E, Mongo/integration, persisted outbox, reload/offline/reconnect replay, or a runtime-delivery guarantee.

##### Phase 149 Task 4: realtime event/cache foundation

- `backend/tests/messenger/message_sent_producers_test.go` exercises normal send (including idempotent retry), forward, bot, poll, and promote via public service APIs. The hubtest capture publisher asserts exact NATS subjects and literal nested canonical `MessageSentEvent` JSON, including rich persisted response fields; it rejects legacy top-level event shapes and proves the retry does not publish.
- The backend receipt is `PASS`: `result.json` `first_exit:0`; build, vet, unit, and race each `PASS` with exit 0. These are genuine isolated units. Real NATS, MongoDB, WebSocket, and deployed runtime execution are `NOT_RUN`.
- `frontend/tests/messenger/use-messages-cache-events.test.tsx` covers strict nested-message acceptance, root/thread cache routing, durable-ID dedupe, fail-closed edit/delete, malformed-cache guards, room reference-count lifecycle, and subscribe-before-exact-key invalidation after reconnect. `frontend/tests/messenger/reaction-toggled-guard.test.tsx` covers immutable, idempotent reaction add/remove deltas in both caches and malformed/foreign-event guards.
- The final frontend receipt is `PASS`: `result.json` `first_exit:0`, TypeScript `PASS` exit 0, Vitest 162/162 files and 427/427 tests in 47.28s, and JSON 340/340 suites. This is not browser E2E or a live WebSocket claim.
- Task 5 user events/unread and persisted outbox, reload, and offline behavior remain out of scope/`NOT_RUN`.
##### Phase 149 Task 5: authorised user-channel delivery

- Backend unit coverage exercises exact self-only `user:<uuid>` authorization; strict user subjects and payloads; all lifecycle and five message producer paths; nil/instant Calls handling; delivery-time current-members filtering for Messenger and channel-scoped Calls; current call-chat audience handling; and fail-closed resolver errors, timeout, nil, and former-member cases. An instant call chat targets active participants plus its creator, not channel membership.
- Frontend unit coverage proves that `useChannels` observes only cache invalidation while `useMessengerChannelsRealtime` owns the one subscription, uses exact `user:<uuid>`, subscribes before reconnect invalidation, and resets WebSocket/query/Messenger state at a real session boundary without resetting on token refresh.
- Final Docker receipts, 2026-08-25: `backend-unit` `PASS` (`first_exit:0`; build, vet, unit, race `PASS`) and `frontend-unit` `PASS` (`first_exit:0`; TypeScript and Vitest `PASS`). Full independent static review is `C0/I0/M0`.
- MongoDB, a real NATS broker, live external WebSocket/browser execution, Ent direct-DB audience behavior, and `python-integration` are `NOT_RUN`. The phase is therefore `PARTIAL`; these units are not integration or deployment evidence.

#### TMS

| Endpoint | 200/201 | 400 | 401 | 403 | 404 |
|----------|---------|-----|-----|-----|-----|
| POST /projects/:id/tms/cases | Создан | Невалидные данные | Нет токена | RBAC | Проект |
| POST /projects/:id/tms/runs | Создан | Невалидные данные | Нет токена | RBAC | Проект |
| POST /projects/:id/tms/allure/import | Импортировано | Невалидный JSON | Нет токена | RBAC | Проект |
| POST /projects/:id/tms/webhook/allure | Принято | Невалидный формат | API Key | - | Проект |

#### Calls

| Endpoint | 200/201 | 400 | 401 | 404 |
|----------|---------|-----|-----|-----|
| POST /calls | Создан | Нет channelId | Нет токена | Канал не найден |
| POST /calls/instant | Создан | - | Нет токена | - |
| POST /calls/:id/join | LiveKit token | - | Нет токена | Звонок не найден |
| POST /calls/guest-join | LiveKit token | Нет invite_code/guest_name | - | Звонок не найден |
| POST /calls/:id/recording/start | Запись начата | - | Нет токена | Звонок не найден |

---

## Тест-кейсы по модулям

> Каждый тест-кейс (TC-*) связан с автоматизированными unit-тестами. Колонка **Автотест** указывает Go test function, покрывающую данный сценарий.

### AUTH — Авторизация (Magic Link)

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-AUTH-01 | Magic Link: запрос кода | Email в org | POST /auth/magic-link {email} | 200, код отправлен | handler validation tests |
| TC-AUTH-02 | Magic Link: верификация | Код из email | POST /auth/magic-link/verify | 200, access + refresh | `TestVerifyMagicLink_*` |
| TC-AUTH-03 | Magic Link: неверный код | — | verify с wrong code | 400/401 | `ConstantTimeCompare` path |
| TC-AUTH-04 | Refresh: обновление токена | Валидный refresh | POST /auth/refresh | 200, новый access | JWT round-trip tests |
| TC-AUTH-05 | Refresh: истёкший токен | Просроченный refresh | POST /auth/refresh | 401 | `TestJWTExpired` |
| TC-AUTH-06 | Профиль: получение | Авторизован | GET /auth/me | 200 | handler tests |
| TC-AUTH-07 | Профиль: обновление | Авторизован | PUT /auth/me | 200 | handler validation |
| TC-AUTH-08 | Сессии: список | Авторизован | GET /auth/sessions | 200 | — (E2E/manual) |
| TC-AUTH-09 | Сессии: отзыв | 2+ сессии | DELETE /auth/sessions/:id | 200 | — (E2E/manual) |
| TC-AUTH-10 | Rate limit | — | 4 auth запроса за 10 мин | 429 на 4-м | manual / middleware test |
| TC-AUTH-11 | Онбординг | onboarding=false | POST /auth/me/onboarding | 200 | `onboarding-profile.spec.ts` |
| TC-AUTH-12 | Logout | Авторизован | POST /auth/logout | 200, refresh revoked | `logout.spec.ts` |
| TC-AUTH-13 | Magic Link: OTP не в slog (A14-006) | Dev/Resend paths | SendMagicLink + VerifyCode; capture slog | raw OTP и `"code"` отсутствуют; email masked | `TestMagicLink_OTPNotLogged_A14_006`, `TestMaskEmailMinimized_A14_006` |

### PROJ — Проекты

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-PROJ-01 | Создание проекта | Авторизован | POST /projects {name, key} | 201, проект + workflow + доска + канал | `TestCreateProject_Success` |
| TC-PROJ-02 | Дубликат ключа | Проект HUB существует | POST /projects {key: "HUB"} | 400/409, ключ занят | `TestCreateProject_DuplicateKey` |
| TC-PROJ-03 | Список проектов | Участник 2 проектов | GET /projects | 200, 2 проекта | — (ручной) |
| TC-PROJ-04 | Архивирование | Owner проекта | POST /projects/:id/archive | 200, status=archived | `TestArchiveProject_SetsStatusArchived` |
| TC-PROJ-05 | Удаление в корзину | Owner проекта | DELETE /projects/:id | 200, deleted_at set | `TestDeleteProject_SoftDelete` |
| TC-PROJ-06 | Восстановление | Проект в корзине | POST /projects/:id/restore | 200, deleted_at null | — (ручной) |
| TC-PROJ-07 | Invite: создание | Owner/Developer | POST /projects/:id/invites | 201, код + ссылка | `TestJoinByInvite_ValidCode` |
| TC-PROJ-08 | Invite: присоединение | Валидный invite-код | POST /invites/join {code} | 200, участник добавлен | `TestJoinByInvite_ValidCode`, `TestJoinByInvite_ExpiredCode` |
| TC-PROJ-09 | Участник: смена роли | Owner | PUT /projects/:id/members/:userId {role: "tester"} | 200 | — (ручной) |
| TC-PROJ-10 | Участник: удаление | Owner | DELETE /projects/:id/members/:userId | 200, удалён | `TestRemoveMember_LastOwner` |
| TC-PROJ-11 | Team: last-member remove → 409 | Команда с 1 участником | DELETE /teams/:teamId/members/:userId | 409 Conflict, count=1 | `TestRemoveTeamMember_RejectsLastMember_TASK184` |
| TC-PROJ-12 | Team: remove non-last OK, then last → 409 | Команда ≥2 участников | DELETE member; затем last | 200 затем 409, count=1 | `TestRemoveTeamMember_AllowsWhenMultiple_TASK184` |

### TRCK — Трекер задач

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-TRCK-01 | Создание задачи | Участник проекта | POST issues {title, type, priority} | 201, issue_key = HUB-1 | `TestCreateIssue` |
| TC-TRCK-02 | Автоинкремент | HUB-1 существует | POST issues {title, type} | 201, issue_key = HUB-2 | `TestCreateIssueAllTypes` |
| TC-TRCK-03 | Обновление с version check | version=1 | PUT issues/:key {title, version: 1} | 200, version=2 | `TestUpdateIssueOptimisticLocking` |
| TC-TRCK-04 | Конфликт версий | version=2 в БД | PUT issues/:key {title, version: 1} | 409, conflict | `TestUpdateIssueOptimisticLocking` |
| TC-TRCK-05 | Комментарий с @mention | Задача существует | POST comments {body, mentions: [userId]} | 201, уведомление отправлено | — (ручной) |
| TC-TRCK-06 | Связь задач | 2 задачи | POST issues/:key/links {target, type: "blocks"} | 201 | `TestCreateIssueLink` |
| TC-TRCK-07 | Time tracking: manual | Задача существует | POST issues/:key/time {minutes: 60} | 201, time_spent += 60 | — (ручной) |
| TC-TRCK-08 | Time tracking: timer | Задача существует | POST timer/start -> POST timer/stop | 201, лог с длительностью | — (ручной) |
| TC-TRCK-09 | JQL поиск | Задачи существуют | POST issues/jql {query: "type = bug AND priority = high"} | 200, отфильтрованные задачи | `TestParseJQLSimple`, `TestParseJQLMultipleConditions`, `TestToEntPredicates` |
| TC-TRCK-10 | Workflow: смена статуса | Переход разрешён | PUT issues/:key {status_id} | 200, статус изменён | `TestMoveIssueValidTransition` |
| TC-TRCK-11 | Workflow: запрещённый переход | Переход не задан | PUT issues/:key {status_id} | 400, переход не разрешён | `TestMoveIssueInvalidTransition` |
| TC-TRCK-12 | Мягкое удаление | Задача существует | DELETE issues/:key | 200, deleted_at set | — (ручной) |

### MSG — Мессенджер

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-MSG-01 | Создание публичного канала | Авторизован | POST channels {type: "public", name} | 201 | `TestCreateChannel_WithProjectID`, `TestCreateChannel_ReadOnlyFlag` |
| TC-MSG-02 | DM дедупликация | DM между A и B существует | POST /dm {userId: B} | 200, тот же канал | `TestCreateDM_Dedup` |
| TC-MSG-03 | Отправка сообщения | Участник канала | POST messages {content} | 201, NATS event | `TestSendMessage_MemberCheck` |
| TC-MSG-04 | Реакция toggle | Сообщение существует | POST reactions {emoji: "thumbsup"} x2 | Добавлена, затем удалена | `TestToggleReaction_AddRemove` |
| TC-MSG-05 | Pin сообщения | Участник канала | POST pin | 200, pinned | — (ручной) |
| TC-MSG-06 | Тред | Сообщение существует | POST messages {reply_to: msgId} | 201, привязано к треду | `TestThreadReply_IncrementsReplyCount`, `TestDeleteMessage_ThreadReplyDecrementsCount` |
| TC-MSG-07 | Read-only канал | is_readonly=true, не Owner | POST messages | 403 | `TestSendMessage_ReadOnlyBlocked`, `TestSendMessage_ReadOnlyBlocked_RegularUser`, `TestSendMessage_ReadOnlyAllowed_Admin` |
| TC-MSG-08 | Mute канала | Участник | POST mute {until: "2026-04-01"} | 200, muted_until set | — (ручной) |
| TC-MSG-09 | Unread counts | Новые сообщения | GET unread | 200, count > 0 | `TestIncrement_AllMembers`, `TestIncrement_SkipsSender`, `TestMarkRead_ResetsToZero`, `TestGetAll_ReturnsAllChannels` |
| TC-MSG-10 | Поиск сообщений | Сообщения существуют | GET search?q=keyword | 200, результаты | `TestSearchMessages_TextQuery` |

### TMS — Управление тестированием

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-TMS-01 | Создание тест-кейса (classic) | Участник проекта | POST cases {title, case_type: "classic", steps} | 201, case_key = HUB-TC-1 | `TestCreateTestCase_Classic` |
| TC-TMS-02 | Создание тест-кейса (BDD) | Участник проекта | POST cases {case_type: "bdd", steps: [{keyword: "Given"}]} | 201 | `TestCreateTestCase_BDD` |
| TC-TMS-03 | Версионирование | Кейс существует | PUT cases/:key {title: "new"} -> GET versions | Версия 2 создана | `TestUpdateTestCase_CreatesVersion` |
| TC-TMS-04 | Откат версии | Версия 1 и 2 | POST rollback/1 | 200, данные версии 1 | `TestRollbackTestCase` |
| TC-TMS-05 | Создание прогона | Кейсы существуют | POST runs {title} + POST runs/:id/cases | 201, кейсы добавлены | `TestAddTestCasesToRun` |
| TC-TMS-06 | Выполнение результата | Прогон in_progress | PUT results/:id {status: "passed"} | 200, passed_count++ | `TestExecuteTestResult_UpdatesStats` |
| TC-TMS-07 | Allure import | Allure JSON | POST allure/import {multipart} | 200, прогон создан, matched | `TestParseAllureResults_ValidJSON`, `TestMatchTestCase_ByAnnotation`, `TestMatchTestCase_ByASIDLabel` |
| TC-TMS-08 | JUnit webhook | JUnit XML | POST webhook/junit | 200, результаты импортированы | — (ручной) |
| TC-TMS-09 | Создание бага из failed | Результат failed | POST results/:id/bug | 201, задача Bug создана | `TestCreateBugFromFailedResult` |
| TC-TMS-10 | Flaky report | Тесты с разными результатами | GET reports/flaky | 200, список flaky тестов | `TestGetCoverageReport` |
| TC-TMS-11 | Матрица трассируемости | Кейсы привязаны к задачам | GET reports/traceability | 200, матрица | `TestLinkTestCaseToIssue` |
| TC-TMS-12 | Экспорт PDF | Прогон completed | GET runs/:id/export | 200, PDF файл | `TestExportToCSV` |

### CALL — Созвоны

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-CALL-01 | Создание звонка | Канал существует | POST calls {channelId} | 201, room_name, LiveKit token | `TestGenerateToken` |
| TC-CALL-02 | Instant-звонок | Авторизован | POST calls/instant | 201, invite_code | `TestGenerateToken_DifferentRooms` |
| TC-CALL-03 | Join по invite-коду | Звонок active | GET calls/invite/:code | 200, информация + token | — (ручной) |
| TC-CALL-04 | Guest join (без аккаунта) | Звонок active, invite_code | POST calls/guest-join {invite_code, guest_name} | 200, LiveKit token (гостевой) | — (ручной) |
| TC-CALL-05 | Leave | В звонке | POST calls/:id/leave | 200, participant_count-- | — (ручной) |
| TC-CALL-06 | End call | В звонке | POST calls/:id/end | 200, status=ended | — (ручной) |
| TC-CALL-07 | Start recording | В звонке | POST calls/:id/recording/start | 200, egress started | `TestMaxActiveRecordingsLimit` |
| TC-CALL-08 | Stop recording | Запись идёт | POST calls/:id/recording/stop | 200, status=processing | `TestRecordingStatusValues` |
| TC-CALL-09 | Chat в звонке | В звонке | POST calls/:id/chat {text} | 200, NATS event | `TestSendMessage`, `TestSendMessage_TextTooLong`, `TestWSAuthorizer_CallRoom_CanSubscribe_A2_007`, `TestReadPump_CallAuthorizer_StrangerDenied_A2_007`, `TestReadPump_CallAuthorizer_ParticipantAllowed_A2_007` |
| TC-CALL-10 | Raise hand | В звонке | POST calls/:id/raise-hand | 200, NATS event | — (ручной) |
| TC-CALL-11 | Webhook: participant_joined | LiveKit event | POST calls/webhook | 200, participant_count++ | `TestVerifyWebhookSignature_EmptyAuth`, `TestVerifyWebhookSignature_InvalidJWT` |

### SPRINT — Спринты

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-SPRINT-01 | Создание спринта | Участник проекта | POST sprints {name, start_date, end_date} | 201, status=planning | `TestCreateSprint`, `TestCreateSprint_EndDateBeforeStartDate` |
| TC-SPRINT-02 | Запуск спринта | Спринт planning | POST sprints/:id/start | 200, status=active | `TestStartSprint_Success` |
| TC-SPRINT-03 | Двойной active | Спринт уже active | POST sprints/:id2/start | 400, уже есть active | `TestStartSprint_RejectsWhenActiveExists` |
| TC-SPRINT-04 | Burndown | Спринт active | GET sprints/:id/burndown | 200, точки по дням | `TestGetBurndown` |
| TC-SPRINT-05 | Velocity | 3+ спринтов completed | GET velocity | 200, planned vs completed | `TestGetVelocity` |
| TC-SPRINT-06 | Planning Poker | Спринт planning | POST poker -> vote -> reveal | 200, голоса раскрыты | — (ручной) |
| TC-SPRINT-07 | Завершение спринта | Спринт active | POST sprints/:id/complete | 200, status=completed | `TestCompleteSprint_MovesIncompleteIssues` |

### NOTIF — Уведомления

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-NOTIF-01 | Назначение задачи -> уведомление | Задача существует | PUT issue {assignee: userId} | Уведомление type=assigned | `TestSubjectToType`, `TestNotificationConsumer_NakOnBadJSON_A2_010`, `TestNotificationConsumer_AckOnSuccess_A2_010`, `TestNotificationConsumer_NakOnDispatchError_A2_010` |
| TC-NOTIF-02 | @mention -> уведомление | Комментарий с mention | POST comment {mentions: [userId]} | Уведомление type=mentioned | `TestSubjectToType` |
| TC-NOTIF-03 | Mark as read | Уведомление unread | PATCH notifications/:id/read | 200, is_read=true | — (ручной) |
| TC-NOTIF-04 | Mark all read | 5 unread | PATCH notifications/read-all | 200, все is_read=true | — (ручной) |
| TC-NOTIF-05 | Quiet hours | quiet_hours set, не urgent | Событие в quiet hours | Уведомление отложено | `TestShouldDispatch_Debounce` |
| TC-NOTIF-06 | Urgent пробивает quiet hours | quiet_hours set, is_urgent=true | Событие в quiet hours | Уведомление доставлено | — (ручной) |
| TC-NOTIF-07 | Web Push subscribe | Авторизован | POST push/subscribe {endpoint, keys} | 200 | — (ручной) |

### ADMIN — Администрирование

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-ADMIN-01 | Аудит-лог | Admin | GET audit?action=login | 200, логи входов | `TestAuditLogFiltersDefaults` |
| TC-ADMIN-02 | Деактивация юзера | Admin | POST users/:id/deactivate | 200, status=deactivated | — (ручной) |
| TC-ADMIN-03 | GDPR удаление | Admin | DELETE users/:id/gdpr | 200, все данные удалены | — (ручной) |
| TC-ADMIN-04 | Бэкап | Admin | POST /admin/backups | 200, backup started | `TestCreateZIP` |
| TC-ADMIN-05 | Restore | Admin, backup exists | POST /admin/backups/:id/restore | 200, восстановлено | `TestCreateZIPEmptyData` |
| TC-ADMIN-06 | Restore/dump errors без секретов (A7-007) | Admin; pg_restore/mongorestore fail | Restore / ExportRestore* | ошибка без DSN/CombinedOutput | `TestRestorePostgres_DoesNotEmbedDumpOutput_A7_007`, `TestRestoreMongo_DoesNotEmbedDumpOutput_A7_007`, `TestRestoreHandler_DoesNotLeakSecrets_A7_007` |
| TC-ADMIN-07 | Mail broadcast только active (A7-008) | Admin; pending+deactivated users | POST broadcast dry-run | recipients = active only | `TestAdminService_Broadcast` |
| TC-HR-01 | Leave approve/cancel version≥1 (A6-006) | Member leave pending | approve/cancel version=0 / bad JSON | 400 | `TestHandler_ApproveLeave_VersionZero_A6_006`, `TestHandler_CancelLeave_BindAndVersion_A6_006` |

### Дополнительные модули

| ID | Название | Автотест |
|----|----------|----------|
| TC-RBAC-01 | Проверка наличия permission | `TestPermission_Has` |
| TC-RBAC-02 | Добавление permission | `TestPermission_Add` |
| TC-RBAC-03 | Удаление permission | `TestPermission_Remove` |
| TC-RBAC-04 | Idempotent операции | `TestPermission_AddRemoveIdempotent` |
| TC-RBAC-05 | Дефолтные роли | `TestDefaultRolePermissions` |
| TC-GITLAB-01 | Issue key regex | `TestIssueKeyRegex` |
| TC-GITLAB-02 | Slugify | `TestSlugify` |
| TC-GITLAB-03 | Truncate | `TestTruncate` |
| TC-GITHUB-01 | Issue key regex | `TestIssueKeyRegex` |
| TC-GITHUB-02 | Webhook signature | `TestValidateWebhookSignature` |
| TC-GITHUB-03 | Slugify | `TestSlugify` |
| TC-GITHUB-04 | Truncate | `TestTruncate` |
| TC-IMEX-01 | Field mapping | `TestBuildFieldMap` |
| TC-IMEX-02 | Get field | `TestGetField` |
| TC-IMEX-03 | Bounds check | `TestGetField_OutOfBounds` |
| TC-STORAGE-01 | Upload URL TTL | `TestUploadURLTTL` |
| TC-STORAGE-02 | Download URL TTL | `TestDownloadURLTTL` |
| TC-STORAGE-03 | Result fields | `TestUploadResult_Fields` |
| TC-STORAGE-04 | ConfirmUpload already confirmed → 409 (A4-006) | `TestHandler_ConfirmUpload_AlreadyConfirmed_409_A4_006`, `TestConfirmUpload_RejectsAlreadyConfirmed_A4_006` |
| TC-WS-01 | Message type constants | `TestMessageTypeConstants` |
| TC-WS-02 | Client/Server JSON | `TestClientMessage_JSON`, `TestServerMessage_JSON` |
| TC-WS-03 | Ping protocol | `TestClientMessage_PingHasNoRoom` |
| TC-WS-04 | NewClient | `TestNewClient` |
| TC-APIKEY-01 | Hash key | `TestHashKey`, `TestHashKeyDeterministic`, `TestHashKeyUniqueness` |
| TC-APIKEY-02 | Key generation | `TestKeyGeneration` |
| TC-APIKEY-03 | Event filtering | `TestContainsEvent`, `TestContainsEventWildcard` |
| TC-APIKEY-04 | Webhook signing | `TestSignPayload` |
| TC-NATS-01 | Stream config | `TestStreamNameConstant`, `TestStreamSubjects`, `TestClientStructFields` |
| TC-PLATFORM-01 | Fiber recover EnableStackTrace == DevMode (A14-007) | `TestFiberRecoverEnableStackTrace_FollowsDevMode_A14_007` |
| TC-SEED-01 | Demo data validity | `TestDemoUsersValid`, `TestDemoStatusesValid`, `TestDemoIssuesValid`, `TestDemoTestCasesValid`, `TestDemoChannelsValid` |
| TC-DASH-01 | Layout serialization | `TestLayoutItemSerialization`, `TestLayoutDTOEmptyItems` |
| TC-DASH-02 | Data defaults | `TestPersonalDataDefaults`, `TestProjectStatsDefaults` |
| TC-DASH-03 | Admin stats | `TestAdminStatsServiceHealth` |
| TC-DASH-04 | Save validation | `TestSaveLayoutRequestValidation` |
| TC-SEARCH-01 | Merge results | `TestMergeResultsByScore`, `TestMergeResultsWithLimit` |
| TC-SEARCH-02 | UUID array | `TestUUIDSliceToPgArray`, `TestUUIDSliceToPgArrayEmpty` |
| TC-SEARCH-03 | Saved filters | `TestSavedFilterToDTO` |
| TC-SEARCH-04 | Defaults | `TestSearchRequestDefaultLimit`, `TestSearchResultTypes` |
| TC-SEARCH-05 | Soft-delete exclusion Ctrl+K (A4-002) | `TestPostgresSearch_ExcludesSoftDeleted_A4_002`, `TestSearchSQL_ExcludesSoftDeleted_A4_002` |
| TC-JQL-01 | JQL parsing | `TestParseJQLSimple` .. `TestParseJQLOR` (13 тестов) |
| TC-JQL-02 | JQL to Ent predicates | `TestToEntPredicates`, `TestToEntPredicatesIN` |
| TC-ALLURE-01 | Parse valid JSON | `TestParseAllureResults_ValidJSON` |
| TC-ALLURE-02 | Missing/empty/malformed | `TestParseAllureResults_MissingFields`, `TestParseAllureResults_EmptyInput`, `TestParseAllureResults_MalformedJSON` |
| TC-ALLURE-03 | Nil pointers | `TestParseAllureResults_NilPointers` |
| TC-ALLURE-04 | Status mapping | `TestAllureStatusToTMS`, `TestExtractLabelValue` |
| TC-ALLURE-05 | Matching strategies | `TestMatchTestCase_ByAnnotation` .. `TestMatchTestCase_PriorityOrder` (7 тестов) |
| TC-PRESENCE-01 | Heartbeat | `TestHeartbeat_SetsScore` |
| TC-PRESENCE-02 | Status detection | `TestGetStatus_Online`, `TestGetStatus_Away`, `TestGetStatus_Offline` |
| TC-PRESENCE-03 | Cleanup | `TestCleanupStale` |
| TC-PRESENCE-04 | Custom status | `TestSetGetCustomStatus` |
| TC-PRESENCE-05 | Bulk status | `TestGetBulkStatus` |

---

## Статистика покрытия

### Автоматизация тест-кейсов

| Модуль | TC всего | Автоматизировано | Только ручные | % автоматизации |
|--------|----------|-----------------|---------------|-----------------|
| AUTH | 11 | 7 | 4 | 64% |
| PROJ | 10 | 6 | 4 | 60% |
| TRCK | 12 | 7 | 5 | 58% |
| MSG | 10 | 8 | 2 | 80% |
| TMS | 12 | 10 | 2 | 83% |
| CALL | 11 | 6 | 5 | 55% |
| SPRINT | 7 | 6 | 1 | 86% |
| NOTIF | 7 | 3 | 4 | 43% |
| ADMIN | 5 | 2 | 3 | 40% |
| Доп. модули | 36 | 36 | 0 | 100% |
| **Итого** | **121** | **91** | **30** | **75%** |

---

## Regression-тесты

### Критический путь (после каждого деплоя / CI e2e-full)

1. **Авторизация**: Magic Link или dev-login → GET /auth/me
2. **Проект**: Создать проект → workflow + канал
3. **Задача**: CRUD issue → смена статуса → комментарий
4. **Мессенджер**: Отправить сообщение → реакция → тред
5. **TMS**: Кейс → прогон → результат passed
6. **Ctrl+K**: Глобальный поиск
7. **Logout**: POST /auth/logout → redirect login
8. **a11y**: axe на login + authenticated routes

### WebSocket Regression

1. Подключение /ws через ticket subprotocol (не ?token=JWT)
2. Subscribe to project:{id}
3. Создать задачу через API
4. Получить event issue_created через WS
5. Unsubscribe

---

## Performance-тесты

### Целевые метрики

| Метрика | Целевое значение |
|---------|-----------------|
| API response time (p95) | < 200ms |
| WebSocket message latency | < 100ms |
| Kanban board load (100 задач) | < 500ms |
| Messenger: загрузка канала (1000 сообщений) | < 300ms |
| Глобальный поиск | < 500ms |
| Allure import (1000 результатов) | < 5s |
| Одновременных WS-подключений | 100+ |
| Участников в звонке | до 50 |

### Rate Limiting

| Тест | Ожидание |
|------|----------|
| 4 запроса auth за 10 мин | 4-й -> 429 |
| 101 API запрос за 1 сек | 101-й -> 429 |
| 11 file uploads за 1 мин | 11-й -> 429 |
| 61 сообщение за 1 мин | 61-й -> 429 (лимит **60/мин**, `MessageSendRateLimit`) |

---

## Ручное тестирование — чеклист по фичам

### Авторизация
- [ ] Magic Link: email → код → вход
- [ ] Dev-login (только DEV_MODE)
- [ ] Профиль: имя, timezone, язык
- [ ] Сессии: список, отзыв
- [ ] Logout: POST /auth/logout
- [ ] Онбординг: 4 шага (профиль → тур → тема → проект)
- [ ] Telegram: **только** привязка уведомлений (не login)

### Проекты
- [ ] Создание проекта с ключом
- [ ] Invite-ссылка: генерация, присоединение
- [ ] Смена ролей участников
- [ ] Архивирование и восстановление
- [ ] Удаление (корзина) и восстановление

### Трекер задач
- [ ] Kanban-доска: drag & drop задач между колонками
- [ ] Создание задачи всех типов
- [ ] Комментарии с @mention
- [ ] Time tracking: таймер start/stop
- [ ] JQL поиск
- [ ] Workflow: настройка статусов и переходов

### Мессенджер
- [ ] Публичные/приватные каналы
- [ ] DM (личные сообщения)
- [ ] Отправка сообщений с форматированием
- [ ] Реакции, pin, forward
- [ ] Треды (ответы)
- [ ] Поллы
- [ ] Typing indicator
- [ ] Unread counts

### TMS
- [ ] Создание тест-кейса (classic + BDD)
- [ ] Прогон: создание, добавление кейсов, выполнение
- [ ] Allure import через API
- [ ] Создание бага из failed результата
- [ ] Отчёты: flaky, trends, traceability

### Созвоны
- [ ] Создание звонка из канала
- [ ] Instant-звонок с invite-кодом
- [ ] Гостевой вход по invite-коду (без аккаунта)
- [ ] Screen share
- [ ] Запись звонка (start/stop)
- [ ] Эфемерный чат звонка
- [ ] Raise hand

### Дашборды
- [ ] Личный дашборд: виджеты с данными
- [ ] Проектный дашборд: burndown, velocity
- [ ] Админский дашборд: статистика пользователей
- [ ] Drag & drop виджетов, изменение размеров

### Уведомления
- [ ] In-app уведомления (колокольчик)
- [ ] Telegram бот уведомления
- [ ] Web Push подписка
- [ ] Quiet hours
- [ ] Настройка по типам

---

## Accessibility-тесты

### Стандарт

WCAG 2.1 AA (целевой для v2.0 Redesign)

### Проверки

| Проверка | Метод | Статус |
|----------|-------|--------|
| Контраст текста (4.5:1) | Automated (axe-core) | Проверяется при UI review |
| Контраст UI элементов (3:1) | Automated (axe-core) | Проверяется при UI review |
| Keyboard navigation | Manual | Tab, Enter, Escape на всех interactive элементах |
| Screen reader labels | Manual | aria-label на кнопках без текста |
| Focus indicators | Visual | Все focusable элементы имеют видимый focus ring |
| Color не единственный индикатор | Visual | Иконки + текст в дополнение к цвету |

---

*Последнее обновление: 2026-07-29 — inventory **812** BE `*_test.go` / **39** export_hooks / Vitest **179** files; COV 83% остаётся baseline прогона 2026-06-22*
