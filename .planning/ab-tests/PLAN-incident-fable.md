# PLAN — Прод-инцидент: WS-реконнекты + лаг уведомлений + flaky e2e + 500 на создании задач

> **A/B-тест оркестрации (вариант Fable 5, max effort)** · team-lead · 2026-07-02
> Read-only диагностический план. Код не менялся. Основано на реальном коде репозитория (файлы и строки указаны).
> Референсы: `.cursor/playbooks/observability.md` (§4, §9.5), `.cursor/skills/test-triage/SKILL.md`, `.cursor/playbooks/team-workflow.md`, `docs/architecture.md`.

---

## 1. Триаж и приоритезация симптомов

| # | Симптом | Severity | Impact | Почему такой приоритет |
|---|---------|----------|--------|------------------------|
| S1 | WS-реконнекты каждые ~30–60с у части пользователей; «send on closed channel»-подобные паники в ws | **P0** | Весь real-time (messenger, board, notifications UI) деградирует; паника в горутине **без recover роняет весь процесс** (playbook observability §3.5) | Вероятный драйвер остальных симптомов |
| S2 | Лаг уведомлений (in-app/Telegram) 5+ мин; NATS consumer pending растёт | **P1** | Пользователи пропускают упоминания/назначения; SLA доставки нарушен | Может быть следствием S1 (рестарты процесса) или самостоятельной деградацией consumer/DB |
| S4 | Изредка 500 при создании задачи («не создаются с первого раза») | **P1** | Прямая потеря пользовательского действия; ретрай вручную | Может быть следствием S1 (запрос попал на падение процесса) или lock-contention в `CreateIssue` |
| S3 | Flaky e2e `tracker/project-lifecycle.spec.ts` (~1 из 4) после недавнего мержа | **P2** | Блокирует CI-доверие, но не пользователей напрямую | Каналарейка: тот же мерж — та же временная точка, что и S1/S2/S4. Обязательный test-triage (§7) |

### Связность: единая корневая причина?

Да, есть кандидат, который связывает все четыре симптома:

**Гипотеза H0 (unifying): паника в ws-горутине роняет весь backend-процесс → оркестратор (Docker/K3s) его рестартует по циклу.**

Механика:
- Паника «send on closed channel» (или родственная) в producer-горутине WS **не ловится** HTTP-`Recover` middleware — паника в самостоятельной горутине убивает процесс целиком (см. предупреждение в `observability.md` §3.5).
- Каждый рестарт = **массовый WS-disconnect → реконнекты у «части пользователей»** (тех, кто был на упавшей реплике/в момент падения) — S1.
- Пока процесс лежит/поднимается, `notification-dispatcher` не разгребает JetStream → **pending растёт**, доставка запаздывает — S2. Плюс in-memory `debounce`-карта consumer'а обнуляется на рестарте.
- HTTP-запросы, попавшие на окно падения/старта, получают **обрыв/500** — «задачи иногда не создаются с первого раза» — S4.
- e2e-тест с 30-секундными ожиданиями редирект/URL попадает на окно рестарта → **flaky ~1/4** — S3.

Периодичность «каждые 30–60с» подозрительно совпадает с константами `pongWait = 60s` / `pingPeriod = 54s` (`backend/internal/ws/handler.go:22-31`) и `AckWait: 30s` consumer'а (`backend/internal/notification/consumer.go:66`) — это надо подтвердить или опровергнуть данными (§3), а не постулировать.

H0 проверяется за ~15 минут (шаг D1 ниже) и, если подтверждается, сводит инцидент к одному фиксу. Если нет — симптомы разбираются по независимым гипотезам §2.

---

## 2. Гипотезы корневых причин по симптомам

### S1 — WS-реконнекты и паники

Контекст кода: SEC-05 уже задокументирован и реализован — канал `send` **никогда не закрывается**, сигнал завершения только через `done`:

- `backend/internal/ws/hub.go:18-23` (контракт SEC-05), `hub.go:356-361` (`removeClient` закрывает только `done`);
- producer'ы проверяют `done` через select: `BroadcastToRoom` (`hub.go:272-284`), `SendToUser` (`hub.go:303-315`), pong-путь `readPump` (`handler.go:256-260`).

Значит «send on closed channel» **не должен** воспроизводиться на этих путях — если паника реальна, её внёс недавний мерж. Гипотезы:

| ID | Гипотеза | Что и где проверить |
|----|----------|---------------------|
| H1.1 | Мерж добавил новый producer, пишущий в `client.send` (или новый канал) без select-на-`done`, либо вернул `close(send)` | `git log --first-parent --since=<дата инцидента>` + `git diff <pre-merge>..HEAD -- backend/internal/ws/`; grep `close(` и все callsites `client.send <-` / `.Send() <-` в ws и вызывающих модулях |
| H1.2 | Паника не «send on closed», а родственная (nil deref в hijacked conn, panic в Bridge-callback) — «подобные» в симптоме | Точный стектрейс из логов процесса: Loki/`docker logs` по `panic:` — **файл:строка обязательны**, не гадать |
| H1.3 | Реконнекты без паник: клиент не получает ping (writePump завис/вышел) или сервер не получает pong → `ReadDeadline(pongWait=60s)` истекает → disconnect | `handler.go:204-218` (`readPump` read deadline), `handler.go:272-299` (`writePump` ping ticker). Смотреть Warn `ws read error` в Loki: если ошибка = timeout — это deadline-путь, не паника |
| H1.4 | Backpressure: broadcast-шторм переполняет буфер `send` (256, `hub.go:43`) → drops → клиентская логика считает соединение мёртвым и переподключается | Частота Warn `ws client send buffer full, dropping message` (`hub.go:280-281, 311-312`) в Loki; корреляция по userID с жалобами |

### S2 — Лаг уведомлений, рост NATS pending

| ID | Гипотеза | Что и где проверить |
|----|----------|---------------------|
| H2.1 | Следствие H0: consumer лежит вместе с процессом при рестартах | Сопоставить график pending (`nats consumer info`) с временем рестартов контейнера |
| H2.2 | Consumer — **однопоточный**: одна горутина последовательно обрабатывает сообщения (`consumer.go:77-103`, `PullMaxMessages(10)`), а каждый event порождает 3–6 SQL-запросов (`IsEnabled` ×3 + `CreateNotification` + `GetUser` ×2 в `service.go:73-127`, `GetChannelMembers`/`GetIssueAssignee` в handlers). Деградация Postgres → throughput consumer падает ниже входного потока | p95 SQL-времени; `nats consumer info EVENTS notification-dispatcher`: NumPending растёт при NumAckPending малом = медленная обработка, не redelivery |
| H2.3 | Обработка события > `AckWait: 30s` → redelivery (MaxDeliver 5) → дубликаты усиливают очередь | NumRedelivered в consumer info; длительность `processEvent` (сейчас не замеряется — см. D5) |
| H2.4 | Telegram-очередь: буфер 100, rate-limit 35мс (~28 msg/s) (`telegram.go:45,64-68`) — при burst'е сообщения дропаются с Warn `очередь telegram переполнена` (`telegram.go:96-98`), а доставленные запаздывают | Частота этого Warn в Loki. NB: это объясняет лаг **Telegram**, но не in-app |
| H2.5 | Outbox-стадия: воркер публикует пачками ≤50 каждые 100мс (`outbox/worker.go:28-40,51-58`); если запрос `status=pending AND attempts<10 ORDER BY created_at LIMIT 50` замедлился (разросшаяся таблица published-строк без cleanup/индекса) — лаг возникает ещё **до** NATS | SQL: `SELECT status, count(*), min(created_at) FROM outbox_events GROUP BY status`; EXPLAIN запроса воркера; есть ли cleanup published-строк |

⚠️ Попутная находка (не причина лага, но фиксировать): комментарий в `consumer.go:130-131,160` обещает «ACK только при успешной обработке», а фактически `msg.Ack()` вызывается **безусловно** — handlers проглатывают ошибки внутри (log-and-continue). Ошибка Dispatch = уведомление потеряно без redelivery. Это расхождение код↔контракт → отдать be2 + docs-writer.

### S4 — Изредка 500 на создании задачи

Путь: `handler_issues.go` → `service_issues.go:18-90` (`tracker.CreateIssue`, обёрнут в `observability.Timed` → метрика есть) → `repo_issues.go:25-126`.

| ID | Гипотеза | Что и где проверить |
|----|----------|---------------------|
| H4.1 | Следствие H0: запрос попал на рестарт процесса | Корреляция timestamps 500-ответов (access-лог `status=500`) с рестартами |
| H4.2 | **Row-lock contention на строке проекта**: транзакция начинается с `AddTaskCounter(1)` (`repo_issues.go:32-38`) и держит блокировку строки `project` до `Commit` (`:120`) — включая запрос дефолтного статуса, insert задачи и outbox-insert. Конкурентные создания в одном проекте сериализуются; при медленной DB — lock wait → timeout → 500 «инкремент счётчика задач» | По `request_id` из 500-ответа найти лог ошибки: какой именно `fmt.Errorf`-шаг падает; pg_stat_activity / lock waits в окно инцидента |
| H4.3 | Optimistic-lock конфликт: по ТЗ `version`-поле в каждой изменяемой схеме — если Ent-hook на Project превращает конкурентный `UpdateOneID` в конфликт версий, интермиттентные 500 воспроизводятся на ровно двух параллельных create | Прочитать `ent/schema/project.go` (hook/version); repro-тест: 10 параллельных CreateIssue в один проект |
| H4.4 | `findDefaultStatusTx` использует `Only()` (`repo_issues.go:128-136`) — 0 или 2 default-статуса = ошибка → 500. Могло сломаться данными после мержа (provisioner) | Ошибка «получение дефолтного статуса» в логах; SQL-проверка дублей `is_default=true` |

### S3 — Flaky e2e `project-lifecycle.spec.ts`

Файл: `testframework/stages/e2e/domains/tracker/project-lifecycle.spec.ts`. Точки хрупкости:

| Точка | Строки | Риск |
|-------|--------|------|
| Ожидание редиректа после «Создать проект» | :38 (`toHaveURL`, 30s) | Если backend отдал 500/упал (S4/H0) — тест падает именно здесь; это **сигнал прод-бага, а не теста** |
| Двойной клик «Удалить» / «Восстановить» через `.last()` | :53-54, :63-64 | Гонка с анимацией диалога подтверждения; `.last()` хрупок при изменении DOM после мержа |
| Строка в корзине после soft-delete | :58-61 | Если список корзины обновляется через WS/инвалидацию — деградация S1 делает появление строки недетерминированным |

Классификация — строго через test-triage (§7), не чинить тест вслепую.

---

## 3. План диагностики: какие данные собрать, в каком порядке

Рабочий цикл — по runbook `observability.md` §9.5: метрика → окно времени → логи по модулю → `request_id` → трейс.

| Шаг | Что собрать | Как (конкретно) | Кто | Подтверждает/опровергает |
|-----|-------------|-----------------|-----|--------------------------|
| **D1** | Рестарты процесса + стек паники | `docker ps`/`kubectl get pods` (restart count); `docker logs`/Loki по `panic:` — вытащить **точный stacktrace файл:строка**; uptime процесса vs период 30–60с | debugger + devops | **H0**, H1.1/H1.2 |
| **D2** | Diff недавнего мержа | `git log --first-parent -15`; `git diff <pre-merge>..HEAD -- backend/internal/ws backend/internal/notification backend/internal/platform/outbox backend/internal/tracker frontend/src` (NB: в working tree уже есть незакоммиченные правки `auth/handler.go`, `audit/repository.go`, `messenger/gif-picker.tsx` — проверить, не задеплоено ли что-то из них) | debugger | H1.1, S3-корреляция |
| **D3** | NATS consumer lag | `nats consumer info EVENTS notification-dispatcher`: NumPending, NumAckPending, NumRedelivered, динамика за час | devops | H2.1/H2.2/H2.3 |
| **D4** | Outbox depth | SQL: `SELECT status, count(*), min(created_at), max(attempts) FROM outbox_events GROUP BY status`; EXPLAIN ANALYZE запроса воркера | devops | H2.5 |
| **D5** | Метрики операций | Prometheus: `hub_operation_calls_total{operation="tracker.CreateIssue",result="error"}` rate; p95 `hub_operation_duration_milliseconds{operation="tracker.CreateIssue"}`; p95 HTTP; 5xx-rate по route | devops | H4.2/H4.3, SLO p95<200мс |
| **D6** | WS-сигналы в Loki | rate Warn `ws client send buffer full` (module=ws); `ws read error` (timeout vs abnormal closure); `ws bridge: malformed payload` | debugger | H1.3/H1.4 |
| **D7** | 500-цепочки | Access-лог `status=500` → `request_id` → полный лог цепочки: какой шаг CreateIssue падает (`инкремент счётчика` / `дефолтный статус` / `фиксация транзакции`) | debugger | H4.2/H4.3/H4.4 |
| **D8** | Flaky repro | `npx playwright test stages/e2e/domains/tracker --repeat-each=20 --workers=1` (изолированно) + trace/video артефакты падений; для backend-гонки: `go test -tags hubtest -race -count=20 ./backend/testframework/tracker/...` | qa-auto2 | S3, H4.3 (если падает на :38) |
| **D9** | Telegram-очередь | Loki: Warn `очередь telegram переполнена`; частота отправок vs 28 msg/s потолок | be2 | H2.4 |
| **D10** | Временно `LOG_LEVEL=debug` на 15 мин (если введён по playbook §8.3) для `ws client registered/unregistered` — фактическая частота реконнектов | devops | подтверждение цикличности S1 |

**Порядок:** D1+D2 немедленно и параллельно (15 мин, решают H0) → D3–D7 параллельно (ещё 30 мин) → D8–D10 фоном. Решение о ветке фикса — после D1/D2.

---

## 4. Делегирование (ростер `.cursor/playbooks/team-workflow.md`, волнами, без конфликтов по файлам)

### Волна 0 — диагностика (параллельно, ~1 час, read-only)

| Агент | Задача | Выход |
|-------|--------|-------|
| `debugger` | D1, D2, D6, D7: стек паники → файл:строка; merge-diff по ws/notification/tracker; классификация H0 vs независимые причины | Root-cause отчёт с evidence (`.planning/debug/`) |
| `devops-engineer` | D3, D4, D5, D10: consumer lag, outbox depth, метрики p95/error-rate, restart count; снапшоты графиков за окно инцидента | Данные + график «pending vs рестарты» |
| `qa-auto2` | D8: изолированный repro flaky 20×, trace-артефакты, точка падения (какой expect) | FAIL-профиль спеки: где и как падает |

Точка синхронизации: **team-lead собирает волну 0 → выбирает ветку** (H0 подтверждена / не подтверждена) → запускает волну 1.

### Волна 1 — фиксы (параллельно, домены не пересекаются)

| Агент | Домен (файлы) | Задача |
|-------|----------------|--------|
| `be1` | `backend/internal/ws/*`, `backend/internal/tracker/*` | Фикс паники по стектрейсу (восстановить SEC-05-инвариант / recover в новой горутине по образцу `observability.md` §3.5); при H4.2/H4.3 — вынести `AddTaskCounter` из длинной транзакции или ретрай на конфликте версии. + regression-тесты |
| `be2` | `backend/internal/notification/*`, `backend/internal/platform/outbox/*` | Устранить bottleneck consumer по данным D3/D9 (параллелизм обработки/батчинг SQL); привести ACK-семантику к контракту (Ack только при успехе / осознанный Ack с Warn); cleanup published outbox-строк если H2.5. + тесты |
| `qa-auto2` | `testframework/stages/e2e/domains/tracker/*` | По результату test-triage (§7): стабилизация спеки **или** тест не трогать (падение = прод-баг из волны 1) |
| `docs-writer` | `docs/`, `.planning/STATE.md` | Docs-sync по факту фиксов (rule 55): system_specification (если менялась ACK-семантика), STATE session note |

Условные подключения: `security-auditor` — только если стек паники/diff указывает на `ws/authorizer.go`, auth-путь WS-upgrade или незакоммиченный `auth/handler.go`; `devops-engineer` доп. задача — алёрты §9.2 playbook (WS buffer drops, NATS pending, 5xx-rate), если их не было.

### Волна 2 — верификация и приёмка

| Агент | Задача |
|-------|--------|
| `be-ver` | Верификация фиксов be1/be2: layering, SEC-05-инвариант, отсутствие data race, DoD-чеклист observability §10.1 |
| `qa-auto-ver` | Детерминизм стабилизированной спеки, отсутствие sleep/flaky-паттернов, привязка TC↔автотест |
| `code-reviewer` | Pre-merge review итогового diff |
| `team-lead` | Финальная приёмка по критериям §5 + решение о закрытии инцидента |

Кросс-ревью be1↔be2 по team-workflow сохраняется.

---

## 5. План верификации фикса (что должно стать зелёным)

| # | Проверка | Команда / критерий |
|---|----------|--------------------|
| V1 | Регрессия backend | `cd backend && go build -tags release ./... && go test -race -count=1 -timeout=600s ./...` — green |
| V2 | Canonical Docker-регрессия | `go test -tags hubtest -race -count=1 ./testframework/...` (docker-compose.test.yml) — green |
| V3 | **Нагрузочный WS-soak** | ≥200 конкурентных клиентов, broadcast-шторм + агрессивные connect/disconnect 10 мин: **0 паник процесса**, restart count не растёт, reconnect-rate фоновый; Warn `send buffer full` — эпизодический, не устойчивый |
| V4 | Consumer lag | `nats consumer info EVENTS notification-dispatcher`: NumPending → **0** в steady-state; доставка in-app < 5 сек e2e (событие → WS) |
| V5 | Outbox | pending-глубина ~0 между тиками; publish-lag (created_at → published_at) < 1 сек p95 |
| V6 | Создание задач под конкуренцией | Тест: 10 параллельных CreateIssue в один проект × `-count=20` с `-race` — 0 ошибок, issue_key без дублей; `hub_operation_calls_total{operation="tracker.CreateIssue",result="error"}` = 0 в проде за 24ч |
| V7 | Flaky-спека | `npx playwright test stages/e2e/domains/tracker --repeat-each=20` — **20/20 green** локально и в CI (e2e-full) |
| V8 | Полный гейт | `/qa-gate` перед handoff; frontend: `npx tsc --noEmit && npm run lint && npm run test:run` |
| V9 | Прод-наблюдение | 24ч после деплоя: p95 API < 200мс, 5xx-rate < 1%, NATS pending = 0, ноль `panic:` в Loki |

`verified: true` только при green regression + приложенный лог (rule 56).

---

## 6. Обязательный test-triage для flaky (skill `.cursor/skills/test-triage/SKILL.md`)

Порядок строгий — функционал сначала:

1. **Зафиксировать симптом:** спека `project-lifecycle.spec.ts`, точка падения из D8 (какой expect, какой timeout).
2. **Воспроизвести функционал, не тест:** вручную/через curl пройти цепочку create project → board → soft-delete → корзина → restore по `docs/system_specification.md` (soft delete 30 дней — ТЗ-инвариант). Ответить: «как ДОЛЖНО работать по спецификации?»
3. **Классификация:**

| Наблюдение из D8 | Вердикт | Действие |
|------------------|---------|----------|
| Падает на `:38` (редирект после создания) и в этот момент в логах 500/рестарт | **ROOT=code** (это S4/H0, не тест) | Тест НЕ трогать; фикс в волне 1, спека станет зелёной сама |
| Падает на `.last()`-кликах диалога или строке корзины при живом backend'e | **ROOT=test** (хрупкий селектор/гонка UI) | qa-auto2: детерминированные локаторы, ожидание состояния диалога вместо `.last()` |
| Корзина не показывает проект при корректном API-ответе | **ROOT=code (frontend)** | fe1/fe2 через отдельную задачу |
| Спецификация restore-flow неоднозначна | **Stop** → уточнить у `sa2`, обновить docs, потом код/тест |

4. **Запрещено:** менять assertion «чтобы прошло», `t.Skip`/`test.skip` без ticket, удалять тест.
5. **После фикса:** `--repeat-each=20` green (V7) + запись в skill-лог: `TRIAGE: project-lifecycle — ROOT=... | FIX: файл:строка | VERIFIED: ...`.

---

## 7. Риски и эскалация

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| Паника невоспроизводима локально (только под прод-нагрузкой) | Средняя | Стектрейс из прод-логов обязателен (D1) до любого фикса; soak-тест V3 как ворота |
| Фикс consumer'а (параллелизм) внесёт гонки в debounce-карту | Средняя | `debounceMu` уже есть; `-race` обязателен (V1/V2); be-ver проверяет data race |
| Незакоммиченные правки в working tree (`auth/handler.go`, `audit/repository.go`) смешаются с hotfix | Средняя | Hotfix-ветка от чистого прод-тега; текущие правки — stash/отдельная ветка |
| Outbox cleanup затронет GDPR/аудит-цепочки | Низкая | Только `status=published` старше N дней; ревью security-auditor при сомнении |
| Диагностика затягивается, пользователи страдают | — | **Таймбокс волны 0 = 2 часа**, дальше решение об откате |

### Нужен ли откат мержа?

**Критерий (решает team-lead после D1/D2):**
- Если merge-diff содержит изменения в `backend/internal/ws/*` или producer-путях к `client.send`, и стек паники указывает на этот код → **revert мержа немедленно** (`git revert -m 1`), фикс — спокойно поверх. Паника, роняющая процесс, дешевле откатом, чем hotfix под давлением.
- Если паника в коде, существовавшем до мержа (латентная гонка, проявленная нагрузкой) → откат бесполезен, идём hotfix-веткой волны 1.
- Если через 2 часа волны 0 нет уверенного root-cause → превентивный откат мержа + продолжение диагностики без прод-давления.

Эскалация: подтверждённая порча данных (дубли issue_key, потерянные уведомления с бизнес-эффектом) → инцидент повышается, `security-auditor` + ручной отчёт владельцу. Напоминание: FA-06/PG-03 (secret rotation) — открытый прод-блокер, не смешивать с этим инцидентом.

---

## Саммари

Наиболее вероятная единая корневая причина — **паника в WebSocket-горутине (producer-путь к `client.send`, внесённый или разбуженный недавним мержем), которая убивает весь backend-процесс и запускает цикл рестартов**: массовые WS-реконнекты (S1), рост NATS pending пока consumer лежит (S2), 500 на запросах в окна падения (S4) и flaky e2e, попадающий на эти окна (S3). SEC-05-инвариант в `ws/hub.go` формально закрывает известные пути «send on closed channel», поэтому подозрение — на новый код из мержа. Альтернатива, если рестартов нет: три независимые деградации — однопоточный notification-consumer упирается в медленную DB, row-lock contention на `project.task_counter` в `CreateIssue`, хрупкие селекторы e2e.
**Первое действие:** debugger + devops параллельно за 15 минут достают restart count процесса и точный stacktrace `panic:` из прод-логов и сверяют с diff мержа — это сразу выбирает между «один фикс + revert» и тремя независимыми ветками.

*Файл создан в рамках A/B-теста. Ничего кроме этого файла не менялось; тесты/сборки не запускались.*
