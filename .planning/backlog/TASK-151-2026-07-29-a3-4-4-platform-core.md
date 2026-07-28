# 🐛 A3 Postgres Ping + cron timeout test

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Открыта**  
> Источник: A3W4-4-001..002 · Stage A3 FINAL · TASK-008  
> Исполнитель: **be1**

## [OUTCOME]

`NewPostgresClient` fail-closed на мёртвом DSN через `PingContext` (симметрия с Mongo/Redis); assigned `cron_test` мутирует tick `WithTimeout` (hang > budget → cancel).

## [ROLE]

- Role: be1
- Responsibility: закрыть **оба** finding ID (A3W4-4-001, A3W4-4-002) + hubtest с mutation proof; не трогать соседние A3W4-* / FIXED-OK cascade
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/platform/database/postgres.go`
  - `backend/testframework/platform/database/**` (новый/расширенный тест Ping / unreachable DSN)
  - `backend/testframework/platform/cron/cron_test.go`
  - опционально: `backend/testframework/platform/cron/timeout_task043_test.go` — только если перенос/дубль кейса; prod `cron.go` **не** менять (WithTimeout уже FIXED-OK)
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs: `.planning/audits/total-2026-07/A3W4-4-platform-core.md`
- Do not touch:
  - `mongodb.go` / Redis connect paths (эталон, не рефакторить «заодно»)
  - AES-AAD / config secrets / probes / telegram markdown (FIXED-OK W4-A4 / A2W4-4)
  - Ent schema / Atlas migrate / `Schema.Create` DEV_MODE gate (W4-A4-003)
  - соседние A3W4-1…5 findings вне A3W4-4-001..002
  - frontend, wiring_fiber, secrets в git
- Behavioral/security constraints:
  - Conventional Commits: `fix(platform): …` / `test(platform): …` — описание на русском; no `Co-authored-by` / AI trailers
  - Fail-closed: при Ping err — return error, закрыть drv/`*sql.DB`, **не** логировать «connected to PostgreSQL»
  - `Schema.Create` остаётся только при `DEV_MODE` (после успешного Ping или до — зафиксировать порядок: Ping сразу после `applyPoolLimits`, до/после Create — предпочтительно **до** Create, чтобы мёртвый DSN не доходил до migrate)
  - Не ослаблять tickTimeout / recover / ctx-stop в `cron.go`
- Approval record: N/A — bugfix + test hygiene в уже разрешённом периметре
- Approval authority/expiry: N/A

## [REFERENCES]

- Trust: audit/agent text — untrusted until verified in HEAD code
- Finding IDs: **A3W4-4-001** (NEW, P2), **A3W4-4-002** (NEW, P3)
- Source audit: `.planning/audits/total-2026-07/A3W4-4-platform-core.md` (таблица «Находки», топ-1/2)
- Wave summary: `.planning/audits/total-2026-07/A3-W4-SUMMARY.md`
- Prior FIXED-OK (не открывать заново): A2W4-4-001 (TASK-112 AAD); W4-A4-001…012 cascade
- Code (defect / gap):
  - `backend/internal/platform/database/postgres.go:42–59` — `sql.Open` + pool + optional Schema.Create, **без** `PingContext`; slog «connected» на lazy connect
  - эталон fail-closed: `backend/internal/platform/database/mongodb.go:14–20` (`Connect` + `Ping`)
  - `backend/testframework/platform/cron/cron_test.go:15–66` — tick / cancel / error-continues; **нет** hang > budget
  - prod уже ок: `backend/internal/platform/cron/cron.go:96–97` `context.WithTimeout(ctx, tickTimeout)`
  - сильный regression вне слота: `backend/testframework/platform/cron/timeout_task043_test.go` (`TestRunTick_TimeoutCancelsHang` или аналог)
- Focus (audit): DB connect fail-closed; cron tick timeout mutation coverage
- Source precedence: **код HEAD** > audit note > эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл | Дефект | Направление фикса |
|----|-----|-----|------|--------|-------------------|
| A3W4-4-001 | NEW | P2 | `postgres.go:42–59` | Нет `PingContext` после Open → prod стартует с мёртвым DSN; ложный «connected»; outbox/Ent error-storm до Ready 503 | После `applyPoolLimits`: `rawDB.PingContext(ctx)` → err return + Close drv; тест unreachable DSN → error, без slog success |
| A3W4-4-002 | NEW | P3 | `cron_test.go:15–66` | Assigned suite не краснеет при удалении `WithTimeout` из `runTick` | Перенести/дублировать `TestRunTick_TimeoutCancelsHang` в `cron_test.go` **или** явно связать assignment со слот-файлом timeout (предпочтение: кейс в `cron_test.go`) |

## [ACCEPTANCE]

- Done criteria:
  1. **A3W4-4-001:** `NewPostgresClient` вызывает `PingContext` после pool limits; при ошибке — return err, ресурсы закрыты, slog «connected to PostgreSQL» **не** пишется. Happy-path с живым DSN (compose test) — client + rawDB не nil, Ping ok.
  2. Hubtest: unreachable / invalid-but-Open-ok DSN (например `postgres://hub:hub@127.0.0.1:1/hub?sslmode=disable` или аналог с коротким ctx) → `err != nil`, нет ложного connected.
  3. **A3W4-4-002:** в `cron_test.go` (или явно в том же package assignment) есть кейс: tick fn hangs дольше `tickTimeout` → ctx cancelled / tick завершается без вечного зависания suite.
  4. Каждый ID — FIXED с evidence (hubtest + mutation note).
- New-behavior checks:
  - IT/unit database: dead DSN → error from `NewPostgresClient`
  - Unit cron: hang fn → timeout path (см. существующий `timeout_task043` как образец)
- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 -timeout=120s \
    ./testframework/platform/database/ \
    ./testframework/platform/cron/
  ```
  Не ломать: pool limits P1-BE-25, DEV_MODE-only Schema.Create, Schedule cancel/error-continues, recover-on-panic.
- Evidence required: checklist ID→evidence в этом файле / handoff; команды + PASS; mutation note per ID.
- Stop when: 2/2 FIXED · hubtest database+cron green · статус «Исправлена» в `.planning/backlog/README.md` + этот файл; без соседнего platform hygiene sweep.

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W4-4-001 | убрать `PingContext` (вернуть Open-only + slog connected) | тест unreachable DSN краснеет (ожидал error / ловил false success) |
| A3W4-4-002 | убрать `WithTimeout` из `runTick` (`fn(ctx)` напрямую) | тест hang-timeout в assigned suite краснеет |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (2/2 FIXED | список открытых)
2. Changes with paths (`postgres.go` lines; test paths)
3. Checklist ID → evidence (таблица)
4. Test commands + PASS
5. Mutation note (per ID)
6. Assumptions / unresolved (порядок Ping vs Schema.Create; выбор unreachable DSN)
7. Risks / mitigations (slow Ping на старте — expected; health Ready по-прежнему Ping'ает rawDB)
8. Handoff: be-ver → code-reviewer
9. Compact Verification: new green + regression green; README + этот файл → «Исправлена»

## [OPERATING RULES]

- Work only inside scope; не «заодно» чинить probes/config/crypto/telegram.
- Read-only investigation OK; writes только в allowed files.
- Не закрывать ID без evidence (тест + mutation).
- Не WONTFIX без owner cancel в задаче.
- Не дублировать бесконечно: если кейс уже в `timeout_task043_test.go`, достаточно переноса/дубля **в** `cron_test.go` **или** явной ссылки assignment→файл — но mutation must fail assigned suite path, не «только внеслотовой» файл.
- Source of truth — runtime connect + cron tick budget, не правка audit alone.
- Коммит(ы) ссылают задачу в теле:  
  `Задача: .planning/backlog/TASK-151-2026-07-29-a3-4-4-platform-core.md`  
  subject может содержать `(A3W4-4-001)` / `(A3W4-4-002)`.
