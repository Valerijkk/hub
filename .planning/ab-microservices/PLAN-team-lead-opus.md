# PLAN — Вынос монолита Hub в доменные микросервисы

> **Роль автора:** `team-lead` (оркестратор + архитектор), Opus 4.8 @ 1M.
> **Режим:** read-only по коду; единственный создаваемый артефакт — этот файл.
> **Дата:** 2026-07-02 · **A/B-тест качества оркестрации** (вариант Opus).
> **Источники:** `.planning/july-2026/EXPANSION-ANALYSIS-2026-07-02.md`, `docs/architecture.md`, `.cursor/rules/60-tz-constraints.mdc`, `.cursor/playbooks/{team-workflow,backend}.md`, `.cursor/rules/20-backend.mdc`, фактический код `backend/internal/**`, `backend/internal/platform/nats/client.go`, `platform/outbox/publisher.go`.
>
> **Дисклеймер приоритета:** это план ПО ЗАПРОСУ пользователя (вынос в микросервисы). Он НЕ противоречит выводу `EXPANSION-ANALYSIS`, где микросервисный split помечен как **F2 / P3 / v4.5, «только по нагрузке»**. Данный план — это подготовленный «run-book на случай, когда решение принято», с честной рекомендацией **НЕ начинать split сейчас** (см. §0 и §12).

---

## 0. TL;DR для занятого читателя

1. **Сейчас Hub — качественный модульный монолит** (Go/Fiber, 83 Ent-схемы, единый NATS JetStream stream `EVENTS`, outbox-паттерн уже есть). Это идеальная стартовая позиция, но она же означает, что **насильный split сегодня даёт отрицательный ROI**: нагрузка (10–50 человек на инстанс, «одна org = один инстанс») не требует независимого масштабирования.
2. **Рекомендация по БД (главный вопрос):** **НЕ одна общая БД с общими таблицами** (это shared-database антипаттерн — он превращает микросервисы в распределённый монолит). Правильный компромиссный путь для Hub — **schema-per-service внутри одного PostgreSQL-кластера сначала** (data ownership per service, кросс-доступ запрещён), с опцией отделения в физически отдельные БД для сервисов с иным профилем нагрузки (messenger→Mongo, calls, search). Полный DB-per-instance — только там, где он оправдан.
3. **Стратегия миграции:** **Strangler Fig поверх укреплённого модульного монолита**, НЕ big-bang split. Сначала фаза «Modular Monolith Hardening» (границы, запрет кросс-импортов, всё общение через интерфейсы/события), потом вынос по одному сервису, начиная с наименее связанных.
4. **Порядок выноса (по возрастанию риска):** `ai-gateway` → `search/indexer` → `notification` → `calls (signaling)` → `messenger` → `tms` → `tracker/sprint` (последними, т.к. ядро связности). `auth` НЕ выносим как отдельный рантайм — превращаем в **общий JWKS/identity-контракт**.
5. **Транспорт:** асинхронно — уже имеющийся **NATS JetStream + outbox** (source of truth для межсервисных событий); синхронно — **gRPC** для internal service-to-service, REST только на кромке через Caddy/apigateway. Идемпотентность + саги для кросс-сервисных мутаций.
6. **ADR обязателен** — затрагивает mandatory-инварианты TZ (deployment, «один инстанс = одна org», единый NATS, JWT-модель).

---

## 1. Текущее состояние (baseline, факт из кода)

| Аспект | Факт | Значение для split |
|---|---|---|
| Рантайм | Один Go/Fiber v3 процесс, `cmd/server/main.go` — точка композиции (~629 LOC) | Единый deployment; wiring — узкое место при параллельной работе |
| Модули | 33 top-level в `backend/internal/`: `auth, tracker, sprint, messenger, tms, calls, dashboard, search, notification, rbac, ws, github, gitlab, ai, intranet, calendar, crm, hr, accounting, bases, marketplace, status, project, audit, apigateway, backup, imex, mail, storage, integrations/onec, dev, platform` | Готовые доменные границы = кандидаты в сервисы |
| Данные | PostgreSQL 17 (83 Ent-схемы, общий client), MongoDB 8 (messages, Allure, call chat), Redis 7 (session/refresh/cache), SeaweedFS (файлы) | Общий `*ent.Client` — главный барьер к data ownership |
| События | **NATS JetStream, единый stream `EVENTS`** (subjects `project.>`, `tracker.>`, `messenger.channel.>`, `tms.>`, `calls.*`, `auth.session.>`, `sprint.>`, `notification.>`), retention Interest, MaxAge 7д; durable consumers: audit, webhook-dispatcher, notification | **Уже event-driven** — фундамент для async межсервисного обмена почти готов |
| Outbox | `platform/outbox/publisher.go` — at-least-once (PG outbox → worker → NATS) | Ключ к надёжной eventual consistency; уже реализован |
| Real-time | WebSocket через Fiber; `ws/bridge.go` мостит NATS→клиент, `ws/authorizer.go` — RBAC на room | WS — отдельный масштабируемый слой (sticky sessions/fan-out) |
| Auth | Magic Link, JWT access 15м, refresh 30д в Redis, instant revoke | Валидация JWT должна стать разделяемой (JWKS) |
| API-gateway | `internal/apigateway` (API-key auth, outbound webhook), Caddy — reverse proxy/TLS/rate-limit | Готовый edge; расширяется до маршрутизации по сервисам |
| Observability | slog + OTel + Prometheus + Loki + Grafana + Alertmanager | Трейсинг есть — нужно пробросить trace-context между сервисами |
| Layering-долг (TD-02) | `*ent.Client` протекает в 7 пакетах (`auth`, `dashboard`, `backup`, `imex`, `storage`, `apigateway`, `rbac`) | Эти пакеты нельзя выносить, пока не изолированы за repository-интерфейсом |
| Deployment | Docker Compose (dev), K3s (partial, TD-04), Caddy | K3s-манифесты неполные — предпосылка для DevOps-фазы |
| Нагрузка | 10–50 человек / инстанс, «одна org = один инстанс» (TZ) | **Нет бизнес-драйвера масштабирования** — ключевой аргумент «не сейчас» |

**Вывод baseline:** архитектурно Hub уже «микросервис-ready по событиям», но «монолит по данным и рантайму». 80% ценности split достигается **укреплением границ модулей БЕЗ разнесения по процессам**.

---

## 2. Рекомендация по БД (ключевой вопрос пользователя)

> Вопрос дословно: «одна общая БД если норм, если нет — как лучше и правильнее».

### 2.1 Почему «одна общая БД» (shared tables) — антипаттерн

Shared database = несколько сервисов читают/пишут **одни и те же таблицы**. Последствия:

- **Скрытая связанность через схему.** Любая миграция колонки в `issue` ломает N сервисов сразу. Теряется главная выгода микросервисов — независимый деплой.
- **Нет инкапсуляции инвариантов.** Сервис B пишет в таблицы сервиса A в обход его бизнес-правил (валидаций, optimistic lock `version`, soft-delete). Данные деградируют.
- **Блокировки и contention.** Общие «горячие» таблицы (issue, message) создают кросс-доменную конкуренцию за локи/пул соединений.
- **Невозможность разной технологии хранения.** messenger хочет Mongo/оптимизацию под запись, tracker — реляционные джойны; общая БД навязывает компромисс.
- **Результат:** «распределённый монолит» — вся операционная сложность микросервисов без единого их плюса. Это худший из миров.

### 2.2 Спектр вариантов и выбор для Hub

| Вариант | Что это | Плюсы | Минусы | Вердикт для Hub |
|---|---|---|---|---|
| **A. Shared DB, shared tables** | Все сервисы в общих таблицах | Ноль миграции данных | Все минусы §2.1 | ❌ **Запретить** |
| **B. Schema-per-service (один кластер PG)** | У каждого сервиса своя PG-schema; кросс-доступ запрещён политикой + отдельные роли/grants | Data ownership, независимые миграции, дёшево операционно (один кластер), легко откатить | Общий физический ресурс (blast radius кластера), нужна дисциплина «не ходить в чужую схему» | ✅ **Рекомендуемый первый шаг** |
| **C. Database-per-service (физически раздельные)** | Отдельная БД/инстанс на сервис | Полная изоляция сбоев и масштабирования, разные движки | Дорого операционно, distributed data, бэкапы×N | ✅ **Точечно** — для сервисов с иным профилем (messenger→Mongo уже, search→Meili, calls) |
| **D. CQRS + read-replicas** | Пишущий сервис-владелец, остальные читают проекции через события | Развязка чтения/записи, нет прямых кросс-запросов | Eventual consistency, дублирование данных | ✅ **Целевое состояние** для «нужен чужой домен на чтение» |

### 2.3 Явная рекомендация

**Гибрид B → (C точечно) + D для чтения.**

1. **Data ownership:** каждый домен владеет своими таблицами. Никакой сервис не делает SQL/Ent-запрос к чужим таблицам — **только** через API (gRPC) или через события (проекции).
2. **Старт — schema-per-service в одном PostgreSQL 17:** физически один кластер, логически N схем + N ролей БД с grant только на свою схему (СУБД принудительно ловит нарушения). Это даёт 90% выгоды data ownership при 10% операционных затрат. Полностью соответствует «один инстанс = одна org».
3. **Физическое отделение — по профилю нагрузки, а не «ради красоты»:** `messenger` (уже Mongo для сообщений), `search` (Meilisearch — уже отдельно), `calls` (эфемерные данные), `ai` (может жить с внешними провайдерами) — кандидаты на собственную БД раньше остальных.
4. **Кросс-доменное чтение — через события + локальные проекции (D):** напр. сервису `notification` нужен `user.display_name` — он держит read-модель, обновляемую по `auth.user.updated`, а не джойнит таблицу `users`. Единый источник истины на запись остаётся у владельца.
5. **Общие «референсные» данные (users, rbac, org):** остаются за `identity`-доменом; выдаются другим как (а) claims в JWT (роли/org в токене), (б) gRPC-lookup, (в) кэш-проекция. Никаких прямых джойнов к `users` из других сервисов.
6. **Согласованность:** отказ от кросс-сервисных ACID-транзакций → **саги + идемпотентность + outbox** (§5).

> **Одной фразой:** общая БД — нет; общий **кластер** с раздельными схемами и строгим data ownership — да, как безопасный первый шаг; физически раздельные БД и CQRS-проекции — по мере отделения конкретных сервисов.

---

## 3. Стратегия миграции: Modular Monolith Hardening → Strangler Fig

### 3.1 Варианты

| Стратегия | Суть | Риск | Подходит Hub? |
|---|---|---|---|
| **Big-bang split** | Разрезать всё сразу на N сервисов | Очень высокий: одновременно данные + рантайм + сеть + деплой | ❌ Нет |
| **Прямой split по сервису без укрепления** | Резать модуль как есть | Высокий: TD-02 (`*ent.Client` протекает), нечёткие границы | ❌ Нет |
| **Modular Monolith Hardening → Strangler Fig** | Сначала укрепить границы В монолите, затем выносить по одному, оставляя фасад | Управляемый, обратимый | ✅ **Да** |

### 3.2 Почему именно так для Hub

- **Состояние кода это позволяет и требует:** границы модулей уже есть, но `*ent.Client` протекает в 7 пакетах (TD-02) — их нельзя выносить, пока доступ к данным не за интерфейсом. Hardening это чинит и **даёт ценность даже если split не случится** (чистая архитектура, тестируемость).
- **Обратимость.** Каждый шаг Strangler Fig можно откатить (фасад маршрутизирует либо в in-process модуль, либо в новый сервис — feature-flag). Big-bang откатить нельзя.
- **AI-команда работает волнами по ≤4 агента (team-workflow §4).** Мелкие обратимые шаги идеально ложатся на модель «атомарный план = атомарный коммит»; big-bang не декомпозируется в атомарные волны.
- **Экономия CI-минут (TZ: ~3000 мин/мес).** Постепенный вынос не взрывает матрицу CI разом.
- **Strangler Fig** = вокруг монолита ставится фасад (Caddy + apigateway); по одному эндпоинту/домену трафик перенаправляется на новый сервис, монолит «удушается» постепенно, пока не останется тонкое ядро.

### 3.3 Целевая топология (после полного выноса — v-next, не обязательна)

```
                    ┌─────────────┐
   Browser ───────► │    Caddy    │  TLS, edge rate-limit, routing
                    └──────┬──────┘
                           │
                    ┌──────▼───────┐
                    │  apigateway  │  API-key, JWT verify (JWKS), fan-out
                    └──────┬───────┘
        ┌──────────┬───────┼────────┬──────────┬───────────┐
        ▼          ▼       ▼        ▼          ▼           ▼
   ┌─────────┐┌────────┐┌──────┐┌────────┐┌────────┐┌──────────┐
   │identity ││tracker ││messg ││ calls  ││  tms   ││ai-gateway│  ... (сервисы)
   │(auth/   ││+sprint ││+ws   ││signaling││        ││          │
   │ rbac)   ││        ││      ││        ││        ││          │
   └────┬────┘└───┬────┘└──┬───┘└───┬────┘└───┬────┘└────┬─────┘
        │ own schema/db per service (data ownership)      │
        └──────────┴───────┴────────┴─────────┴───────────┘
                           │ async events
                    ┌──────▼──────┐
                    │NATS JetStream│  EVENTS stream + per-service outbox
                    └─────────────┘
   Shared infra: PostgreSQL (schemas), MongoDB, Redis, SeaweedFS, LiveKit, Meilisearch
   Cross-cutting: OTel trace-context propagation, Prometheus, Loki
```

---

## 4. Границы сервисов и владение данными (порядок выноса)

### 4.1 Карта доменов → сервисы

Группируем 33 модуля в **связные bounded contexts** (нельзя выносить каждый `internal/*` отдельно — некоторые сильно связаны):

| Сервис (кандидат) | Модули внутри | Владеет данными | Связность с ядром | Приоритет выноса |
|---|---|---|---|---|
| **ai-gateway** | `ai` | свои настройки/аудит; провайдеры внешние | Низкая (событийная) | **1 (первый)** |
| **search** | `search` (+ Meilisearch indexer) | индексы (Meili уже отдельно) | Низкая (читает события) | **2** |
| **notification** | `notification`, часть `mail` | свои настройки/очереди | Низкая (consumer событий) | **3** |
| **calls** | `calls` (signaling; LiveKit внешний), call-chat (Mongo) | эфемерные call-данные | Низкая-средняя | **4** |
| **messenger** | `messenger`, `ws`(частично), presence | messages (Mongo), channels | Средняя | **5** |
| **tms** | `tms` | test cases/runs/Allure (Mongo) | Средняя (ссылки на tracker) | **6** |
| **tracker-core** | `tracker`, `sprint`, `project` | issues, sprints, projects | **Высокая (ядро)** | **7 (последним)** |
| **identity** (НЕ отдельный рантайм на старте) | `auth`, `rbac`, org | users, roles, sessions (Redis) | Все зависят | **Особый — §4.3** |
| **platform-monolith (остаток)** | `dashboard, github, gitlab, intranet, calendar, crm, hr, accounting, bases, marketplace, status, audit, backup, imex, storage, apigateway, integrations/onec` | свои таблицы | разное | Остаются в монолите/выносятся по нагрузке позже |

### 4.2 Критерии порядка (почему так)

Порядок = **min(связность) + min(разделяемые данные) + max(независимость нагрузки) + min(риск)**:

1. **`ai-gateway` первым** — почти чистый outbound-адаптер к внешним LLM, минимум общих данных, отдельный профиль нагрузки (долгие стримы, BYOK-лимиты), уже логически изолирован (`ai/adapter_*.go`). Идеальный «пилот» для отработки шаблона (Dockerfile, gRPC-контракт, JWKS-verify, трейсинг).
2. **`search`** — уже полу-отдельный (Meilisearch + indexer), питается событиями, только чтение для остальных.
3. **`notification`** — чистый event-consumer (durable consumer уже есть), исходящие каналы (in-app/Telegram/WebPush/mail).
4. **`calls`** — эфемерные данные, LiveKit уже внешний; требует sticky/WS-масштабирования (отдельный аргумент за вынос).
5. **`messenger`** — высокая запись, Mongo, WS; выносится после отработки WS-масштабирования на calls.
6. **`tms`** — крупный (6834 LOC), но относительно автономный; ссылки на tracker через события.
7. **`tracker-core` последним** — ядро связности (все ссылаются на issues/projects). Выносить только когда шаблон отлажен на всех остальных.

### 4.3 Особый случай — auth/identity

`auth` **не выносим как первый сервис** и, возможно, не как отдельный рантайм вообще на старте. Вместо этого:

- Превращаем в **identity-контракт**: издатель JWT + **JWKS endpoint** (публичные ключи для верификации), + gRPC `Introspect`/`GetUserRoles`.
- Остальные сервисы **верифицируют JWT локально по JWKS** (без синхронного вызова на каждый запрос) — роли/org кладём в claims.
- Refresh/revocation остаётся централизованно в Redis (instant revoke — mandatory TZ).
- Это снимает «auth — single point of sync» и сохраняет TZ-инвариант (Magic Link, 15м/30д).

---

## 5. Межсервисное взаимодействие

### 5.1 Транспорты и когда какой

| Паттерн | Транспорт | Когда | Реализация в Hub |
|---|---|---|---|
| **Domain events (async, факт свершился)** | **NATS JetStream + outbox** | Всё, что «случилось»: issue.created, message.sent, run.completed | Уже есть: stream `EVENTS`, `platform/outbox`, durable consumers |
| **Команда/запрос (sync, нужен ответ сейчас)** | **gRPC** (internal) | «дай user roles», «валиден ли project» | Новое: `.proto`-контракты, buf |
| **Внешний API** | **REST/JSON** через Caddy→apigateway | Браузер, публичные webhooks | Уже есть |
| **Real-time push** | WS (Fiber) + NATS bridge | presence, typing, live-updates | Уже есть `ws/bridge.go` |

**Правило:** предпочитать **async события** (слабая связанность). gRPC — только когда нужен немедленный ответ для запроса пользователя и данные принадлежат другому сервису.

### 5.2 Избегание распределённого монолита

- **Anti-pattern чек:** если для одного пользовательского действия нужно синхронно дёрнуть >2 сервиса цепочкой — граница проведена неверно (пересмотреть bounded context).
- **Никаких синхронных цепочек A→B→C→D.** Оркестрация — через саги на событиях.
- **Данные для чтения дублируем проекциями**, а не ходим синхронно за ними (§2.3).

### 5.3 Согласованность: саги, идемпотентность, eventual consistency

- **Саги (choreography-based на событиях)** для кросс-сервисных мутаций. Пример: «CRM: won deal → создать invoice в accounting» — событие `crm.deal.won` → `accounting` создаёт draft invoice → публикует `accounting.invoice.created`; при ошибке — компенсирующее событие.
- **Идемпотентность обязательна** (JetStream = at-least-once, дубли неизбежны): каждое событие несёт `event_id` (UUID); consumer хранит processed-set (Redis/таблица) и дедуплицирует. Уже соответствует стилю Hub (UUID v4 везде).
- **Optimistic locking (`version`)** сохраняется на уровне владельца данных — не нарушать при проекциях.
- **Порядок и версионирование событий:** схемы событий версионируем (`v1` в subject или payload); consumer терпим к неизвестным полям (forward-compat).

### 5.4 Контракты

- **Схемы событий** — в общем репозитории контрактов (`contracts/events/*.json` или protobuf), single source of truth; и продюсер, и консюмер импортируют.
- **gRPC .proto** — в `contracts/proto/`, кодоген через `buf`; версионирование пакетов.
- **Contract tests** между продюсером и консюмером (§6, §9).

---

## 6. Сквозные заботы (cross-cutting)

| Забота | Сейчас | Что делаем при split |
|---|---|---|
| **Auth/JWT** | JWT подписывается auth, валидируется в монолите одним секретом | **JWKS**: auth публикует публичные ключи; сервисы верифицируют локально; ротация ключей через kid. Общий секрет НЕ шарить между процессами. Refresh/revoke — централизованно в Redis. |
| **API-gateway** | Caddy + `internal/apigateway` (API-key, webhooks) | Caddy — routing по path/host на сервисы; apigateway — единая точка JWT-verify + rate-limit на кромке; internal-трафик — mTLS/сетевая политика |
| **Service discovery** | n/a (один процесс) | K3s DNS (`svc.namespace.svc.cluster.local`); для dev — Compose service names. Никакого внешнего Consul на старте (KISS) |
| **Config/secrets** | env + Compose | K3s Secrets/ConfigMaps per namespace; **PG-03 ротация (уже prod-блокер!)** становится ещё критичнее — секреты на N сервисов. Vault — кандидат, но не на старте |
| **Observability** | slog + OTel + Prometheus + Loki | **Пробросить W3C trace-context** через NATS-заголовки и gRPC-метадату (distributed tracing сквозь сервисы); per-service dashboards; correlation по `request_id`+`trace_id`. Это критично для отладки — main-риск split |
| **Тестирование** | testcontainers, testframework (unit/e2e), 495 backend тестов | **Contract tests** (Pact-стиль) продюсер↔консюмер; **e2e через несколько сервисов** (docker-compose с N сервисами); сохранить пирамиду |
| **CI/CD и бюджет** | GitHub Actions ~3000 мин/мес (TZ: экономить!) | **Path-based CI** (собирать/тестить только изменённый сервис — monorepo с фильтрами); отдельные образы per service; матрица не должна взорвать бюджет. **Явный риск — §10** |
| **Data migrations** | Atlas, единый | Per-service Atlas миграции (своя schema); порядок применения оркеструется devops |

---

## 7. Deployment на K3s

| Элемент | Решение |
|---|---|
| **Namespaces** | `hub-core` (identity, tracker, tms), `hub-comms` (messenger, calls, notification, ws), `hub-ai` (ai-gateway, search), `hub-platform` (остаток монолита), `hub-infra` (PG, Mongo, Redis, NATS, SeaweedFS, LiveKit, Meili) |
| **Ingress** | **Caddy** (mandatory TZ — не nginx) как ingress-контроллер/edge; routing по host/path |
| **Сервисы** | Каждый — Deployment + Service (ClusterIP); HPA по CPU/custom-метрикам |
| **WS/calls масштабирование** | Отдельный аргумент за вынос: WS-сервис — **sticky sessions** (session affinity) или stateless fan-out через NATS; calls — LiveKit внешний + signaling реплики; HPA по числу соединений |
| **NATS** | StatefulSet, JetStream persistence (PVC); один кластер на инстанс (TZ: единый pub/sub) |
| **PostgreSQL** | Managed/StatefulSet; **schema-per-service** внутри (§2); бэкапы per schema |
| **Config** | ConfigMap + Secret per namespace; **pinned образы** (TZ: no `:latest`) |
| **K3s текущее** | TD-04: манифесты частичны — DevOps-фаза достраивает (см. roadmap) |

---

## 8. TZ / ADR

### 8.1 Затрагиваемые инварианты (из `.cursor/rules/60-tz-constraints.mdc`)

| Инвариант TZ | Статус при split | Действие |
|---|---|---|
| Docker Compose (dev) / K3s (prod), Caddy | **Меняется топология** (много сервисов) | **ADR обязателен** |
| «Один инстанс = одна org» (multi-org → v5) | Сохраняется (сервисы всё равно на одну org) | Подтвердить в ADR; НЕ открывать multi-org |
| NATS JetStream — единственный pub/sub | Сохраняется и усиливается | ✅ соответствует |
| WebSocket через Fiber (не gorilla/net/http) | Сохраняется в WS-сервисе | ✅ |
| SeaweedFS/aws-sdk-go-v2, presigned | Сохраняется (общий storage) | ✅ |
| JWT 15м + refresh 30д Redis, Magic Link, instant revoke | Сохраняется; добавляется JWKS | Уточнить в ADR (JWKS не меняет модель, а распределяет верификацию) |
| UTC/UUID/optimistic-lock/soft-delete | Сохраняются per service | ✅ (усилить — data ownership) |
| `*ent.Client` не в service/handler | **Усиливается** (предусловие выноса) | Hardening-фаза чинит TD-02 |
| CI ~3000 мин/мес | **Под угрозой** (N сервисов) | ADR + path-based CI (§6) |

### 8.2 Требуемый ADR

**`ADR-ARCH-MICROSERVICES-2026`** должен зафиксировать:
1. Триггер выноса (SLA/нагрузка), а не «архитектурная мода» — с метрикой (напр. >500 concurrent WS, p95 latency, независимый релиз-каденс домена).
2. Решение по БД (§2): schema-per-service → точечно DB-per-service; запрет shared tables.
3. Транспорт: NATS+outbox (async) / gRPC (sync) / REST (edge).
4. JWKS-модель auth (совместимость с TZ auth-инвариантами).
5. Порядок выноса (§4) и обратимость (Strangler Fig).
6. Обновление `docs/architecture.md`, `docs/stack_and_tools.md`, `docs/tz_constraints.md`, `deployments/`.

> Без принятого ADR **ни одна фаза кода не стартует** (TZ + REGULATION).

---

## 9. Риски и митигация

| # | Риск | Вероятн. | Влияние | Митигация |
|---|---|---|---|---|
| R1 | **Распределённый монолит** (плохие границы → синхронные цепочки) | Высокая | Критич. | Границы по bounded context (§4); async-first (§5); чек «>2 sync-хопов = редизайн» |
| R2 | **Распределённые транзакции / рассинхрон данных** | Высокая | Высокое | Саги + outbox + идемпотентность (§5.3); проекции для чтения; никаких кросс-сервисных ACID |
| R3 | **Рост latency** (сетевые хопы вместо in-process) | Средняя | Среднее | gRPC (не REST) internal; кэш-проекции; батчинг; измерять p95 до/после |
| R4 | **Операционная сложность** (N деплоев, секретов, БД) | Высокая | Высокое | Начать со schema-per-service (§2); Strangler Fig (обратимость); сильная observability до выноса |
| R5 | **Сложность отладки** (трейс через сервисы) | Высокая | Высокое | **Distributed tracing (OTel context) — предусловие, а не «потом»** (§6); correlation id |
| R6 | **CI-бюджет взрывается** (TZ ~3000 мин) | Средняя | Высокое | Path-based CI, кэш, per-service образы; контроль до масштабирования |
| R7 | **Auth рассинхрон / дыра в RBAC** | Средняя | Критич. | JWKS + локальная верификация; RBAC-проверки сохранить в каждом сервисе (WS room = уже проверяется); security-audit каждой границы |
| R8 | **Регрессия при вынесении** | Средняя | Высокое | Фасад/feature-flag (трафик old vs new); shadow-traffic; полный `/qa-gate` + e2e через сервисы |
| R9 | **Откат невозможен** | Низкая (при Strangler) | Критич. | Каждый шаг обратим (фасад маршрутизирует назад в монолит); БД — expand/contract миграции |
| R10 | **Команда AI-агентов не удержит контекст N сервисов** | Средняя | Среднее | ≤4 агента/волна; один сервис = один поток be1/be2; контракты первым коммитом (team-workflow §4) |
| R11 | **Преждевременная оптимизация** (нет драйвера нагрузки) | **Высокая** | Высокое | **Главная митигация: не начинать split без ADR-триггера; сделать только Hardening (Фаза 0), остальное — по нагрузке** |

---

## 10. Поэтапный roadmap (фазы → планы → волны)

> Формат по `team-workflow §6`. Оценки: **S** 1–2д · **M** 3–5д · **L** 1–2нед · **XL** >2нед.
> Исполнители из ростера (30 агентов). Каждая фаза заканчивается `*-ver` + `team-lead` приёмкой + docs-sync.
> **Фазы 1+ стартуют ТОЛЬКО после принятого ADR (Фаза −1).**

### Фаза −1 — Решение и ADR (gate)

**Цель:** зафиксировать «делаем ли вообще» + ADR со всеми инвариантами.

| # | Задача | Исполнитель | Depends | Оценка |
|---|---|---|---|---|
| 1 | Бизнес-триггер выноса (метрики нагрузки, релиз-каденс) | `product` | — | S |
| 2 | ADR-ARCH-MICROSERVICES-2026 (§8) | `arch` | 1 | M |
| 3 | Верификация ADR на TZ-инварианты | `arch-ver` + `security-auditor` | 2 | S |
| 4 | Целевая C4-диаграмма (context/container) | `arch` | 2 | S |

**Волны:** W1: #1 → W2: #2,#4 ∥ → W3: #3.
**Success:** ADR approved; триггер измерим; TZ-инварианты подтверждены; `product` sign-off.
**Если триггера нет → СТОП на Фазе 0 (Hardening), остальное — backlog.**

---

### Фаза 0 — Modular Monolith Hardening (делать в любом случае — ценно само по себе)

**Цель:** укрепить границы модулей ВНУТРИ монолита; это предусловие любого выноса и чистая архитектура сама по себе.

| # | Задача | Исполнитель | Файлы (границы) | Depends | Оценка |
|---|---|---|---|---|---|
| 1 | Устранить TD-02: `*ent.Client` за repository-интерфейс в `auth` | be1 | `auth/*repository*.go` | — | M |
| 2 | То же для `dashboard`, `storage` | be2 | соотв. пакеты | — | M |
| 3 | То же для `backup`, `imex`, `apigateway`, `rbac` | be1 | соотв. пакеты | 1 | M |
| 4 | Явные модульные API (package-facade) + запрет кросс-импортов (lint: `depguard`) | be2 | `internal/*/api.go`, CI lint | — | M |
| 5 | Каталог событий: описать текущие NATS subjects как версионированные контракты | sa1 | `contracts/events/*` (спека) | — | S |
| 6 | Идемпотентность consumer'ов (event_id dedup) где нет | be1 | `notification/consumer.go`, `audit/consumer.go` | 5 | M |
| 7 | Регрессия + coverage не ниже 82% | qa-auto1/2 | testframework | 1–6 | M |

**Волны:** W1: #1,#2,#4,#5 (≤4) → W2: #3,#6 → W3: #7 → приёмка.
**Success:** 0 `*ent.Client` в service/handler (TD-02 closed); depguard запрещает кросс-модульные импорты; event-контракты задокументированы; идемпотентность; `/qa-gate` green; COV ≥82%.
**Verifier:** `be-ver`, `sa-ver`, `qa-auto-ver`, `arch-ver`.

---

### Фаза 1 — Платформа выноса (шаблон сервиса + инфра)

**Цель:** «пустой» шаблон микросервиса + gRPC/JWKS/tracing/CI, чтобы вынос конкретных сервисов был механическим.

| # | Задача | Исполнитель | Depends | Оценка |
|---|---|---|---|---|
| 1 | Service template (Go/Fiber + graceful + config + slog/OTel) | be1 | Ф0 | M |
| 2 | gRPC-инфра: `buf`, `contracts/proto/`, кодоген | be2 | — | M |
| 3 | JWKS: auth публикует ключи; shared verify-middleware | be1 + security-auditor | Ф0 | M |
| 4 | OTel trace-context через NATS-headers + gRPC-metadata | devops-engineer | 1 | M |
| 5 | Caddy routing + apigateway edge JWT-verify | devops-engineer | 3 | M |
| 6 | Path-based CI (per-service build/test), бюджет-контроль | devops-engineer | 1 | M |
| 7 | K3s base manifests (namespaces, secrets, NATS StatefulSet) — закрыть TD-04 | devops-engineer | — | L |
| 8 | Contract-test harness (Pact-стиль) | qa-auto1 | 2,5 | M |

**Волны:** W1: #1,#2,#3,#7 → W2: #4,#5,#6 → W3: #8 → приёмка.
**Success:** «hello» сервис поднимается в K3s, верифицирует JWT по JWKS, шлёт/принимает NATS-событие с трейсингом, CI собирает только его; contract-test проходит.
**Verifier:** `arch-ver`, `be-ver`, `security-auditor`, `qa-auto-ver`.

---

### Фаза 2 — Пилот: вынос `ai-gateway` (Strangler Fig)

**Цель:** первый реальный сервис по шаблону; отработать весь цикл с наименьшим риском.

| # | Задача | Исполнитель | Depends | Оценка |
|---|---|---|---|---|
| 1 | Выделить `ai` в сервис по шаблону (свои таблицы → своя schema) | be1 | Ф1 | L |
| 2 | Atlas-миграция: перенести ai-схемы в отдельную PG-schema | be2 | 1 | M |
| 3 | gRPC/REST-контракт ai (chat/stream/settings) | sa1 | 1 | S |
| 4 | Фасад в монолите: флаг routing ai → сервис или in-process | be1 | 1 | M |
| 5 | Проекции: ai читает нужные данные проекта через события | be2 | Ф0#5 | M |
| 6 | e2e через 2 сервиса (монолит+ai) + contract tests | qa-auto1/2 | 3,4 | M |
| 7 | K3s deploy ai-gateway (namespace hub-ai, HPA) | devops-engineer | Ф1#7 | M |
| 8 | Security-review границы (JWKS, RBAC, SSRF на провайдеров) | security-auditor | 1 | S |

**Волны:** W1: #1,#3 → W2: #2,#4,#5 → W3: #6,#7,#8 → приёмка.
**Success:** ai работает как отдельный сервис; трафик переключается флагом и откатывается; данные изолированы в своей schema; e2e+contract green; latency p95 в пределах; security-review PASS.
**Verifier:** `be-ver`, `sa-ver`, `qa-auto-ver`, `security-auditor`, `arch-ver`.

> **GATE после Фазы 2:** `team-lead` + `product` ретроспектива. Если ROI/операционная сложность не оправданы — **остановиться здесь** (ai вынесен, остальное — по нагрузке). Это осознанная точка выхода.

---

### Фаза 3 — Вынос низкосвязных сервисов (`search`, `notification`)

**Цель:** повторить шаблон на ещё двух event-driven сервисах (параллельно двумя потоками be).

| # | Задача | Исполнитель | Depends | Оценка |
|---|---|---|---|---|
| 1 | `search` → сервис (Meili уже отдельно; indexer как consumer) | be1 | Ф2 | L |
| 2 | `notification` → сервис (durable consumer + каналы) | be2 | Ф2 | L |
| 3 | Проекции user/project для обоих | be1/be2 | Ф0#5 | M |
| 4 | Contract + e2e (3–4 сервиса) | qa-auto1/2 | 1,2 | M |
| 5 | K3s deploy (hub-ai/search, hub-comms/notification) | devops-engineer | 1,2 | M |

**Волны:** W1: #1 ∥ #2 → W2: #3 → W3: #4,#5.
**Success:** оба сервиса независимо деплоятся; события доставляются идемпотентно; регрессия green.
**Verifier:** `be-ver`, `qa-auto-ver`, `arch-ver`.

---

### Фаза 4 — Real-time домены (`calls`, затем `messenger`)

**Цель:** отработать WS/масштабирование — самая нетривиальная часть.

| # | Задача | Исполнитель | Depends | Оценка |
|---|---|---|---|---|
| 1 | `calls` signaling → сервис (LiveKit внешний; call-chat Mongo) | be1 | Ф3 | L |
| 2 | WS-масштабирование: sticky/fan-out через NATS (валидация RBAC room сохранена) | be2 + arch | 1 | L |
| 3 | `messenger` → сервис (Mongo messages, presence, реакции) | be1 | 2 | XL |
| 4 | Проекции + saga (issue↔message link через события) | be2 | Ф0#5 | M |
| 5 | Нагрузочный тест WS (>500 concurrent — ADR-триггер) | qa-auto2 | 2 | M |
| 6 | K3s deploy hub-comms (HPA по соединениям) | devops-engineer | 1,3 | M |
| 7 | Security-review WS-границ (room RBAC, JWKS) | security-auditor | 2,3 | S |

**Волны:** W1: #1 → W2: #2 → W3: #3,#4 → W4: #5,#6,#7.
**Success:** звонки и мессенджер работают распределённо; WS масштабируется горизонтально; RBAC на room не ослаб; нагрузочный тест подтверждает триггер.
**Verifier:** `be-ver`, `arch-ver`, `security-auditor`, `qa-auto-ver`.

---

### Фаза 5 — `tms`, затем ядро `tracker-core` (последним)

**Цель:** вынести крупный автономный tms, и в самом конце — ядро связности.

| # | Задача | Исполнитель | Depends | Оценка |
|---|---|---|---|---|
| 1 | `tms` → сервис (Allure/Mongo; ссылки на tracker через события) | be1 | Ф4 | XL |
| 2 | `tracker-core` (`tracker`+`sprint`+`project`) → сервис | be1+be2 | 1 | XL |
| 3 | Саги: issue↔tms run, sprint↔issue через события | be2 | 1,2 | L |
| 4 | Проекции для остатка монолита (dashboard читает issues через события) | be1 | 2 | L |
| 5 | Полный e2e через все сервисы; регрессия | qa-auto1/2 | 1–4 | L |
| 6 | K3s hub-core deploy + финальная топология | devops-engineer | 2 | M |
| 7 | Финальный security-audit всех границ | security-auditor | 5 | M |

**Волны:** W1: #1 → W2: #2 → W3: #3,#4 → W4: #5,#6,#7.
**Success:** ядро вынесено; монолит «удушен» до тонкого platform-остатка; вся система работает распределённо; полный `/qa-gate` + e2e green; security-audit 0 CRITICAL.
**Verifier:** все `*-ver` + `team-lead` финальная приёмка.

---

### Что остаётся в монолите (осознанно)

`dashboard, github, gitlab, intranet, calendar, crm, hr, accounting, bases, marketplace, status, audit, backup, imex, storage, integrations/onec` — **v4-модули-скелеты и низконагруженные** — остаются в «platform-monolith» и выносятся **только по индивидуальному драйверу нагрузки**. Выносить их сейчас = чистые затраты без выгоды (см. EXPANSION-ANALYSIS: это MVP-скелеты, приоритет — их углубление, а не split).

---

## 11. Открытые вопросы к продукту

1. **Есть ли реальный драйвер?** Каков триггер (concurrent users, p95 latency, потребность в независимом релиз-каденсе домена)? Без него split — преждевременная оптимизация (R11). Какие метрики нагрузки сейчас по факту?
2. **«Один инстанс = одна org» остаётся?** Если да — независимое масштабирование доменов нужно только для очень крупных инсталляций. Это меняет ROI split.
3. **Целевой объём:** вынести ВСЁ (полная микросервисная архитектура) или прагматично 2–4 сервиса с иным профилем нагрузки (ai, calls, messenger, search) и оставить ядро монолитом? (Рекомендация — второе.)
4. **Приоритет vs расширение функционала:** EXPANSION-ANALYSIS говорит, что главная ценность сейчас — **углубление v4-модулей**, не split. Что важнее для бизнеса в этом квартале?
5. **Операционная готовность:** есть ли команда/бюджет на эксплуатацию N сервисов (мониторинг, дежурства, N бэкапов, ротация секретов ×N)? PG-03 (ротация) уже prod-блокер на монолите.
6. **CI-бюджет:** ~3000 мин/мес — приемлемо ли его перераспределение на path-based per-service pipeline, или это риск?
7. **БД-инфраструктура:** согласны на schema-per-service в одном PG как первый шаг (рекомендация §2), или сразу нужен managed PG-per-service (стоимость)?
8. **Кто владелец identity:** подтвердить JWKS-модель (§4.3) — она не нарушает TZ auth-инварианты, но меняет распределение верификации.

---

## 12. Итоговая рекомендация (саммари)

1. **Общая БД с общими таблицами — нет.** Это shared-database антипаттерн → распределённый монолит.
2. **Правильный путь по БД:** data ownership; старт — **schema-per-service в одном PostgreSQL** (дёшево, обратимо), физически отдельные БД — точечно по профилю нагрузки, кросс-доменное чтение — через события/проекции (CQRS).
3. **Не делать big-bang.** Только **Modular Monolith Hardening → Strangler Fig**, по одному сервису, обратимо.
4. **Фаза 0 (Hardening) ценна независимо** от split — делать её стоит в любом случае (чинит TD-02, укрепляет границы).
5. **Порядок:** ai → search → notification → calls → messenger → tms → tracker-core (последним); auth → JWKS-контракт, не отдельный первый сервис.
6. **Транспорт готов на 80%:** NATS JetStream + outbox уже есть; добавить gRPC (sync) и distributed tracing (предусловие).
7. **ADR обязателен** (затрагивает deployment/NATS/auth/CI-инварианты TZ); без принятого ADR код не стартует.
8. **Честный вывод team-lead:** при текущей нагрузке (10–50 чел/инстанс, одна org) **экономически оправдан вынос максимум 2–4 сервисов с особым профилем** (ai, calls, messenger, search). Полный split — держать как обратимый run-book, запускать по измеримому триггеру, а не «потому что микросервисы».

---

## Приложение — соответствие требованиям задания

| Требование | Раздел |
|---|---|
| 1. Рекомендация по БД | §2, §12 |
| 2. Стратегия миграции | §3 |
| 3. Границы сервисов и владение | §4 |
| 4. Межсервисное взаимодействие | §5 |
| 5. Сквозные заботы | §6 |
| 6. Deployment K3s | §7 |
| 7. TZ/ADR | §8 |
| 8. Риски | §9 |
| 9. Roadmap (фазы/волны/оценки/success) | §10 |
| 10. Открытые вопросы | §11 |
