# PLAN — Prod-инцидент: WS-реконнекты + лаг уведомлений + flaky трекер + 500 при создании задач

> **Автор:** `team-lead` (Opus 4.8, max effort) · A/B оркестрация под неопределённостью
> **Дата:** 2026-07-02 · **Режим:** read-only диагностика (код НЕ меняется в рамках этого плана)
> **Скоуп:** root-cause план + делегирование, без фиксов. Фиксы — отдельными атомарными PR после подтверждения гипотез.
> **Базы знаний:** `.cursor/playbooks/observability.md`, `.cursor/playbooks/testing.md`, `.cursor/skills/test-triage/SKILL.md`, `.cursor/playbooks/team-workflow.md`, `docs/architecture.md`

---

## 0. TL;DR (для дежурного)

Четыре симптома пришли одновременно — это **сильный сигнал общей корневой причины**, а не четырёх независимых багов. Наиболее вероятная единая причина: **насыщение/деградация общего PostgreSQL-пула соединений** (вероятно, из-за медленного/локирующего запроса, попавшего в недавний мерж — кандидаты `backend/internal/audit/repository.go`, `backend/internal/auth/handler.go`). Всё в бэкенде делит один `*ent.Client` (один pool). Деградация БД → синхронный consumer уведомлений не успевает за `AckWait=30s` → **шторм редоставок JetStream + повторные публикации outbox** → лаг очереди; та же деградация БД → спорадические 500 при создании задач; давление на хост + возможная регрессия закрытия канала → залипание `writePump` → таймаут ping/pong (54s/60s) → реконнекты ~30–60s. Flaky-тест трекера, скорее всего, — то же 500-на-создание под нагрузкой либо тайминг теста (обязателен test-triage).

**Первое действие:** снять метрики (`§4`) — глубина outbox, JetStream `num_pending`/`num_redelivered`, DB pool `in_use/wait_count`, p95 по `operation`, и грепнуть логи на WS-panic stack — параллельно repro flaky с `-count=20`.

---

## 1. Триаж и приоритизация симптомов

### 1.1 Матрица приоритетов

| # | Симптом | Влияние на пользователя | Масштаб | Severity | Приоритет |
|---|---------|-------------------------|---------|----------|-----------|
| S1 | WS реконнекты каждые ~30–60s + panic «send on closed channel»-подобные в `ws` | Пропадает real-time (чат, доски, звонки-сигналинг); паника может ронять горутины | Часть пользователей | **CRITICAL** (паника в проде — запрет ТЗ) | P0 |
| S2 | Лаг доставки уведомлений 5+ мин, растёт consumer lag NATS | Уведомления бесполезны (in-app/Telegram) | Все получатели | **HIGH** | P1 |
| S4 | «Задачи иногда не создаются с первого раза» (изредка 500) | Прямая потеря работы, фрустрация | Изредка, все | **HIGH** | P1 |
| S3 | Flaky e2e трекера (`stages/e2e/domains/tracker`), ~1/4 после мержа | Блокирует CI/релиз, маскирует регрессию | CI | **MEDIUM** (но блокер мержа) | P2 |

Ранжирование по «бьёт сильнее»: **S1 (паника+real-time) → S2/S4 (равнозначны, оба HIGH) → S3 (CI-гейт)**. Но по **скорости диагностики** первым запускаем сбор метрик, покрывающий S2/S4 (быстро отвечает «жива ли БД»), т.к. это, вероятно, общий корень.

### 1.2 Гипотеза общей причины (связаны ли симптомы?)

**Архитектурный факт (по коду):** весь бэкенд использует **один `*ent.Client`** → один пул соединений Postgres. На этом пуле одновременно висят:

- HTTP-путь создания задачи: `tracker.CreateIssue` → `repo.CreateIssue` + `LogActivity` (`service_issues.go:77,83`);
- consumer уведомлений: `Dispatch` делает **5–6 последовательных запросов** к БД на КАЖДОЕ событие (`IsEnabled`×3, `GetUser`×2 в `isQuietHours`+`sendTelegram`, `CreateNotification`) — синхронно, в **одной горутине**, ACK только после (`consumer.go:85-102`, `service.go:73-127`);
- outbox worker: `Query` + `UpdateOneID` каждые **100 мс** (`worker.go:29,51,85,99`);
- audit consumer, push-подписки, RBAC-проверки WS-subscribe.

**Причинно-следственная цепочка (единый корень → 4 симптома):**

```
[медленный/локирующий SQL из мержа  ИЛИ  всплеск нагрузки]
        │
        ▼
насыщение пула Postgres  ← общий *ent.Client
        ├─► CreateIssue/LogActivity таймаутит/лочит  ──► спорадические 500  (S4)
        ├─► Dispatch (5-6 DB-запросов) > AckWait(30s)
        │        └─► JetStream redelivery-шторм (MaxDeliver=5)
        │        └─► + outbox republish (Save published упал → pending → повторный Publish)
        │                └─► рост num_pending → лаг 5+ мин  (S2)
        └─► CPU/GC/goroutine-давление на хосте
                 └─► writePump залипает на WriteMessage(writeWait=10s) / ping не уходит вовремя
                          └─► pong не приходит < pongWait(60s), pingPeriod(54s) ──► реконнект ~30-60s  (S1)
                                   └─► (возможная РЕГРЕССИЯ: reintroduced close(send)) ──► panic (S1)

S3 flaky = проявление S4 (500 при создании под нагрузкой)  ИЛИ  чистый тайминг теста → test-triage решает
```

**Вывод триажа:** рабочая гипотеза — **общий корень (БД-деградация)**, но её нужно **подтвердить метриками**, а не принять на веру. Панику WS (S1) и flaky (S3) нельзя списывать на общий корень без отдельной проверки: паника может быть **самостоятельной регрессией** канала, flaky — тайминговым тестом.

---

## 2. Гипотезы корневых причин (что и где смотреть)

> Правило: **не гадать** — каждая гипотеза привязана к файлу/строке и к конкретному сигналу (метрика/лог/slog-поле/трейс). См. `observability.md` §4, §9.

### 2.1 S2 — Лаг уведомлений / рост consumer lag

| H | Гипотеза | Где смотреть (код) | Чем подтвердить/опровергнуть |
|---|----------|--------------------|------------------------------|
| H2.1 | **Синхронный однопоточный consumer** не успевает: 5-6 DB-запросов на событие в одной горутине | `consumer.go:77-103` (одна горутина, `iter.Next()`→`processEvent`), `service.go:73-127` (`Dispatch`: `IsEnabled`×3, `CreateNotification`, `GetUser`×2) | JetStream consumer `num_pending` растёт монотонно; `num_ack_pending`≈1; p95 `Dispatch` (нет `Timed`-обёртки — **пробел наблюдаемости**, см. §2.5) |
| H2.2 | **Redelivery-шторм**: `Dispatch` > `AckWait=30s` → NATS передоставляет то же сообщение, `MaxDeliver=5` | `consumer.go:62-67` (`AckWait:30s`, `MaxDeliver:5`) | JetStream `num_redelivered` > 0 и растёт; в логах повторные `handleAssigned` для одного `entity_id` |
| H2.3 | **Outbox republish амплификация**: publish ок, но `SetStatus(published).Save` упал под нагрузкой БД → запись остаётся `pending` → повторный `Publish` следующим тиком | `worker.go:83-108` (publish ДО mark-published; на ошибке Save — лог, запись не published) | Метрика глубины outbox (кол-во `status=pending`) не спадает; в логах `outbox: failed to mark event published`; дубли событий в NATS |
| H2.4 | **`time.LoadLocation` на каждый Dispatch** (tzdata I/O) добавляет латентность | `service.go:218` (`isQuietHours`→`LoadLocation(user.Timezone)`) | p99 `Dispatch` с хвостом; профиль CPU показывает `LoadLocation` |
| H2.5 | Telegram/Push воркеры блокируют? (проверить, что НЕ они) | `telegram.go:90-100`, `push.go:107-117` — оба **неблокирующие** (буфер+drop) | Опровергается кодом: `Send` = `select default drop`. Consumer НЕ ждёт сеть Telegram/Push. Лаг — не здесь |

**Первичный подозреваемый S2:** H2.1 + H2.2 (синхронный consumer + короткий AckWait) в связке с H2.3 (амплификация). Общий множитель — медленная БД (см. §2.4).

### 2.2 S4 — Спорадические 500 при создании задач

| H | Гипотеза | Где смотреть | Чем подтвердить |
|---|----------|--------------|-----------------|
| H4.1 | **Пул Postgres исчерпан/лочится** → `CreateIssue` таймаутит → `NewInternal` 500 | `service_issues.go:77` (`repo.CreateIssue` возвращает err → 500); `tracker.CreateIssue` обёрнут в `Timed` → есть метрика | `hub_operation_calls_total{operation="tracker.CreateIssue",result="error"}` всплеск; DB pool `wait_count`/`wait_duration` растут; логи `create issue failed` с `error` = context deadline / pool timeout |
| H4.2 | **Лок-контеншн / долгая транзакция** из недавнего мержа (`audit/repository.go`, `auth/handler.go`) держит строки/соединения | git diff мержа; `audit/repository.go` (аудит на hot-path записи); `auth/handler.go` | `pg_stat_activity` показывает долгие/idle-in-transaction; корреляция времени 500 с деплоем мержа |
| H4.3 | **Генерация issue_number** гонка (два create → уникальный конфликт) | `repo_issues.go` `CreateIssue` (проверить стратегию нумерации/sequence/retry) | 500 несут `duplicate key`/`unique constraint`; воспроизводится при параллельных create |
| H4.4 | `LogActivity` падает и роняет ответ? (проверить: НЕ роняет) | `service_issues.go:83-86` — ошибка активности лишь логируется, не возвращается | Опровергается кодом: 500 не от активности. Смотреть только `repo.CreateIssue` |

**Первичный подозреваемый S4:** H4.1/H4.2 (общий корень БД) или H4.3 (гонка нумерации — самостоятельный баг). Разводит их характер ошибки в логе (`deadline`/`pool` vs `unique constraint`).

### 2.3 S1 — WS реконнекты + panic

| H | Гипотеза | Где смотреть | Чем подтвердить |
|---|----------|--------------|-----------------|
| H1.1 | **Таймаут ping/pong под нагрузкой**: `writePump` — одна горутина и на `client.send`, и на ping-ticker; `WriteMessage` с `writeWait=10s`; медленный клиент/бэкпрешер задерживает ping → pong не успевает < `pongWait=60s` | `handler.go:22-31` (таймауты), `handler.go:272-299` (`writePump` совмещает send+ping) | Реконнект-каденс совпадает с `pingPeriod=54s`/`pongWait=60s`; `ws read error` в логах; всплеск `ws client send buffer full` (`hub.go:280,311`) |
| H1.2 | **РЕГРЕССИЯ: reintroduced `close(send)`** где-то → «send on closed channel» panic (текущий код SEC-05 канал `send` НЕ закрывает) | grep `close(` по `ws/*` и по местам, писавшим в `Client.send`; `hub.go:356-361` (комментарий: send намеренно не закрывается) | Стек паники в логах указывает на `close`/`send`; наличие нового `close(client.send)` в diff мержа |
| H1.3 | **Флуд событий в комнату** (тот же лаг/redelivery/duplicate из S2) забивает буфер `send`(256) → дропы + залипание writePump | `bridge.go` (все `BroadcastToRoom`/`SendToUser`), `hub.go:259-316` | Рост `ws send buffer full, dropping message`; корреляция WS-дропов с ростом NATS-трафика |
| H1.4 | Массовые дисконнекты из-за деактивации/`context.Canceled` при деплое (ложный сигнал) | `handler.go:173-179` (проверка deactivated), `hub.go:114-141` (shutdown) | Реконнекты только вокруг рестартов/деплоя, а не постоянно |

**Первичный подозреваемый S1:** H1.1 (таймаут под нагрузкой — связь с общим корнем) **плюс** обязательная проверка H1.2 (регрессия закрытия канала — потенциально независимый P0). Паника обрабатывается? В горутинах WS **нет** явного `recover` в `writePump`/`readPump` (в отличие от outbox `worker.go:72`) — паника в горутине уронит процесс (`observability.md` §3.5). Это усиливает P0.

### 2.4 Общий корень — БД (кросс-симптомный)

| H | Гипотеза | Где смотреть | Чем подтвердить |
|---|----------|--------------|-----------------|
| HC.1 | **Насыщение общего пула** `*ent.Client` (один пул на HTTP+consumers+outbox+audit) | `cmd/server/*` wiring пула; все repo | `sql.DBStats`: `InUse`≈`MaxOpenConns`, `WaitCount`/`WaitDuration` растут |
| HC.2 | **Медленный/локирующий SQL из мержа** | git diff: `audit/repository.go`, `auth/handler.go` | `pg_stat_statements` топ по total_time; корреляция с деплоем |
| HC.3 | Отсутствие индекса/полное сканирование на выросшей таблице (outbox/notification/activity_log) | `ent/schema/*` для outbox/notification; `EXPLAIN` тяжёлых запросов | `EXPLAIN ANALYZE` показывает seq scan; рост latency с ростом строк |

### 2.5 Пробелы наблюдаемости (фиксируем как отдельный вывод)

- **`notification.Dispatch` НЕ обёрнут в `observability.Timed`** → нет p95/метрики консьюмера. Слепая зона именно там, где лаг. (Рекомендация фикс-фазы; `observability.md` §4.2.)
- **Нет метрики глубины outbox** и **нет экспорта JetStream `num_pending`/`num_redelivered`** в Prometheus. (`observability.md` §9.2 требует «Очереди NATS» — проверить, есть ли scrape.)
- **WS-горутины (`writePump`/`readPump`) без `recover`** — паника валит процесс (`observability.md` §3.5).

---

## 3. План диагностики (что собрать, в каком порядке)

> Порядок: сначала дешёвые агрегаты (метрики), затем логи по окну, затем трейсы/repro. Runbook-корреляция — `observability.md` §9.5.

### Wave D0 — Сбор данных (параллельно, read-only, ≤4)

| Шаг | Данные | Источник/команда | Отвечает |
|-----|--------|------------------|----------|
| D0.1 | Глубина outbox: `SELECT status, count(*) FROM outbox_event GROUP BY status` + возраст самого старого `pending` | psql (read-only) | `devops-engineer` |
| D0.2 | JetStream: `num_pending`, `num_ack_pending`, `num_redelivered` для consumer `notification-dispatcher` | `nats consumer info EVENTS notification-dispatcher` | `devops-engineer` |
| D0.3 | DB pool: `sql.DBStats` (InUse/Idle/WaitCount/WaitDuration) + `pg_stat_activity` (idle in transaction, long queries) + `pg_stat_statements` топ-10 | `/metrics` (если экспортится) + psql | `devops-engineer` |
| D0.4 | p95/p99 и error-rate: `tracker.CreateIssue`, HTTP `POST /api/.../issues` (5xx) | Prometheus: `hub_operation_*`, `hub_http_requests_total{status=~"5.."}` (`observability.md` §4.6) | `debugger` |
| D0.5 | Логи WS: грепнуть `panic`, стек «send on closed channel», `ws read error`, `ws client send buffer full` за окно инцидента | Loki `{module="ws"}` / stdout | `debugger` |
| D0.6 | Repro flaky: `stages/e2e/domains/tracker` с `-count`/повторами (20×), собрать частоту и точку падения | Playwright (см. §6) | `qa-auto2` |
| D0.7 | git diff недавнего мержа: изменения в `audit/repository.go`, `auth/handler.go`, `ws/*`, `tracker/repo_issues.go` | `git log`/`git diff` мержа | `debugger` |

**Критерий готовности D0:** есть числа по outbox depth, consumer pending/redelivered, DB pool wait, p95 CreateIssue, стек WS-паники (да/нет), частота flaky, содержимое мержа.

### Wave D1 — Разбор по корню (после D0, зависит от данных)

- Если D0.1/D0.2 растут **и** D0.3 показывает WaitCount↑/long queries → подтверждён **общий корень HC.1/HC.2** → root-cause у `debugger`, БД-запрос у `be` по домену.
- Если D0.2 `num_redelivered`↑ при умеренном pending → подтверждён **H2.2** (AckWait), фикс — вынести доставку/укрупнить параллелизм consumer.
- Если D0.5 показывает **стек с `close`** → подтверждён **H1.2** (регрессия) → отдельный P0, независимо от БД.
- Если D0.6 flaky падает на шаге создания/ожидания с 500 → **S3=проявление S4**; иначе — тайминг теста.

---

## 4. Метрики/пороги для подтверждения (справочно)

| Сигнал | Норма | Инцидент | PromQL/источник |
|--------|-------|----------|-----------------|
| outbox pending | ~0, спадает за тик | растёт/держится | psql count |
| JetStream num_pending | ~0 | 5+ мин работы = тысячи | `nats consumer info` |
| JetStream num_redelivered | 0 | >0 растёт | `nats consumer info` |
| DB pool WaitCount | 0 | растёт | `sql.DBStats` / `/metrics` |
| CreateIssue error-rate | ~0 | всплеск | `hub_operation_calls_total{operation="tracker.CreateIssue",result="error"}` |
| HTTP 5xx rate | <1% | >1% | `observability.md` §9.3 |
| WS buffer drops | 0 | >0 устойчиво | лог `ws client send buffer full` |
| WS panic | 0 | ≥1 | Loki `panic` в `module=ws` |

---

## 5. Делегирование (волнами, без конфликтов файлов)

> Правила: ≤4 агента/волна, разведены по доменам, `team-lead` — финальная приёмка (`team-workflow.md` §4, §5). Здесь диагностика — параллельна; фиксы — атомарные PR по доменам.

### Wave 1 — Диагностика (параллельно, read-only)

| Агент | Задача | Файлы/область (только чтение) |
|-------|--------|-------------------------------|
| `debugger` | Root-cause: свести D0.4/D0.5/D0.7, построить причинную цепочку, подтвердить/отклонить HC.1–HC.3, H1.2 | `ws/*`, `tracker/repo_issues.go`, diff мержа, логи |
| `devops-engineer` | Инфра-метрики: outbox depth, JetStream consumer info, DB pool + `pg_stat_*` | psql, NATS CLI, `/metrics` |
| `qa-auto2` | Repro flaky 20×, зафиксировать частоту и точку падения, подготовить test-triage-вход | `stages/e2e/domains/tracker/*` |
| `security-auditor` | Проверить, не связан ли `auth/handler.go` из мержа с деградацией/локом сессий (S4/S1 через `IsUserDeactivated` на WS-upgrade `handler.go:173`) | `auth/handler.go`, `ws/handler.go` |

### Wave 2 — Локализация фикса (после Wave 1, по подтверждённому корню)

| Агент | Домен фикса (если гипотеза подтверждена) |
|-------|------------------------------------------|
| `be1` | **notification consumer** (H2.1/H2.2): параллелизм обработки, вынести доставку за ACK, увеличить AckWait/`Dispatch` в `Timed`. Файлы: `notification/consumer.go`, `notification/service.go` |
| `be2` | **outbox + БД** (H2.3/HC): mark-published до/после с идемпотентностью, индексы, устранить медленный SQL мержа. Файлы: `platform/outbox/worker.go`, затронутый `audit/repository.go`, `ent/schema/*` |
| `be1`/`be2` (доп.) | **WS** (H1.1/H1.2): `recover` в `writePump`/`readPump`, разнести ping и send, найти/убрать регрессию `close(send)`. Файл: `ws/handler.go`, `ws/hub.go` |
| `qa-auto2` | **flaky** (S3): по итогам test-triage — стабилизировать тест ИЛИ подтвердить баг кода |

**Разведение по файлам:** `be1` — `notification/*` и `ws/handler.go`; `be2` — `outbox/*`, `audit/*`, `ent/schema/*`. WS `hub.go` отдаётся **одному** агенту (точка конкуренции), как в `team-workflow.md` §4.3. Общая точка композиции (`cmd/server/*` wiring пула) — только один агент, если понадобится тюнинг MaxOpenConns.

### Wave 3 — Верификация (после фиксов)

`be-ver` (layering/security/регрессия) ∥ `qa-auto-ver` (flaky-детерминизм, coverage) → `code-reviewer` → `team-lead` приёмка.

### Эскалация

- **Паника WS в проде** (S1/H1.2) → немедленно `team-lead` → пользователь: это запрет ТЗ (`panic()` в prod-путях), кандидат на **откат мержа** (см. §7).
- **PG-03/секреты** не трогаем; если корень — инфра БД (диск/OOM), → `devops` + пользователь.
- Конфликт файлов между `be1/be2` → `team-lead` определяет базовый коммит.

---

## 6. Обязательный test-triage для flaky (S3)

> Протокол `.cursor/skills/test-triage/SKILL.md` + `testing.md`. **Запрещено** делать тест зелёным правкой assert до проверки функционала.

1. **Симптом:** `stages/e2e/domains/tracker/project-lifecycle.spec.ts`, падение ~1/4 после мержа. Тест: create project → board → delete → restore from trash.
2. **Repro функционала (не теста):** прогнать `create project`+`create issue` под лёгкой нагрузкой; проверить по `docs/system_specification.md`, как ДОЛЖНО работать создание/восстановление; снять `request_id` из ответа при падении.
3. **Классификация:**

| Наблюдение при repro | ROOT | Действие |
|----------------------|------|----------|
| Падение на создании/переходе с 500 (совпадает с S4) | **code** | Чинит `be` (это S4), тест оставить строгим; ROOT=code |
| Код по спеке ок, тест падает на гонке навигации/`toHaveURL`/timeout | **test** | Стабилизировать: детерминированные ожидания вместо фикс. timeout, `expect.poll`; ROOT=test |
| Спека неясна (поведение restore) | **stop** | Уточнить у SA/BA, обновить docs, потом фикс |

4. **Repro-команда:** `-count`-эквивалент для e2e — повтор спека 20× (`--repeat-each=20` Playwright) на изолированном стенде; для backend-подозрения `go test -race -count=20 ./testframework/tracker/...`.
5. **Запись triage:** `ROOT=code|test`, файл:строка, обоснование — в skill-лог (`test-triage` §5).

---

## 7. План верификации фикса (что должно стать зелёным)

| Проверка | Инструмент | Критерий «зелёного» |
|----------|-----------|---------------------|
| Backend регрессия | `cd backend && go build -tags release ./... && go vet ./... && go test -race -count=1 -timeout=600s ./...` | PASS, 0 race |
| Consumer lag | нагрузочный прогон N событий → `nats consumer info` | `num_pending → 0`, `num_redelivered = 0` устойчиво |
| Outbox | наблюдение под нагрузкой | `pending → 0`, нет `failed to mark published` |
| CreateIssue 500 | нагрузочный create (parallel), метрика | error-rate `tracker.CreateIssue` ≈ 0; p95 < SLO (200 мс HTTP, `observability.md` §4.6) |
| WS стабильность | нагрузочный WS: держать N клиентов, флуд событий | нет реконнектов вне деплоя; **0 panic**; `send buffer full` в норме; ping/pong стабилен |
| WS panic | `-race` + сценарий remove+broadcast гонки | 0 «send on closed channel» |
| Flaky | `--repeat-each=20` (e2e) / `-count=20 -race` (backend) | **20/20 green** (детерминизм, `qa-auto-ver`) |
| Полный гейт | `/qa-gate` | build ✅ test-race ✅ lint ✅ security ✅ e2e-smoke ✅ |
| Docs-sync | rule 55 | STATE + `system_specification`/`observability` обновлены |

**Правило приёмки:** `verified: true` только при green регрессии + приложенном логе (`team-workflow.md` §7).

---

## 8. Риски и эскалация. Нужен ли откат мержа?

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| Паника WS роняет процесс (нет `recover` в горутинах) | Средняя | P0-фикс `recover`; при подтверждении регрессии `close(send)` — **rollback мержа** |
| Фикс параллелизма consumer вносит дубли/гонки | Средняя | Идемпотентность по `entity_id`+debounce; `-race` |
| Тюнинг пула БД маскирует медленный SQL, не лечит | Средняя | Сначала устранить медленный запрос (HC.2), потом пул |
| Откат мержа теряет полезные изменения | Низкая | Откат — временная мера до точечного фикса; сохранить ветку |

**Решение об откате:** если Wave 1 подтверждает, что корень (медленный SQL или `close(send)`-регрессия) внесён недавним мержем **и** идёт паника/потеря данных в проде — **рекомендовать откат мержа немедленно** (эскалация пользователю, `team-workflow.md` Прил. G), а точечный фикс делать отдельным PR. Если корень — не-мерж (рост нагрузки/инфра БД), откат не нужен; идём в тюнинг+параллелизм.

---

## 9. Саммари (наиболее вероятная единая причина + первое действие)

1. Четыре симптома, пришедшие одновременно, указывают на **общий корень, а не 4 независимых бага**.
2. Единый множитель — **общий пул PostgreSQL** (`*ent.Client` один на HTTP + все consumers + outbox + audit).
3. Наиболее вероятная причина: **деградация БД** (медленный/локирующий SQL из недавнего мержа — `audit/repository.go`/`auth/handler.go` — или всплеск нагрузки), насыщающая пул.
4. Каскад: медленная БД → синхронный однопоточный consumer уведомлений превышает `AckWait=30s` → **шторм редоставок JetStream + повторные публикации outbox** → лаг 5+ мин (S2); та же БД → спорадические 500 при `CreateIssue` (S4).
5. Давление на хост + возможная **регрессия `close(send)`** → залипание `writePump`, таймаут ping/pong (54s/60s) → реконнекты ~30–60s и паника (S1).
6. Flaky трекера (S3) — вероятно проявление S4 под нагрузкой либо тайминг теста; решает **обязательный test-triage**.
7. Отдельный вывод: **пробелы наблюдаемости** (нет `Timed` на `Dispatch`, нет метрик outbox/JetStream, нет `recover` в WS-горутинах) — их закрыть в фикс-фазе.
8. **Первое действие:** Wave D0 — снять метрики (outbox depth, JetStream `num_pending`/`num_redelivered`, DB pool `WaitCount`, p95/error `tracker.CreateIssue`) и грепнуть логи на WS-panic stack, параллельно repro flaky 20× — чтобы подтвердить/опровергнуть общий БД-корень до любых правок.
