# PLAN: Вынос монолита Hub в доменные микросервисы (A/B: team-lead-fable)

> **Дата:** 2026-07-02 · **Автор:** team-lead-fable (оркестратор+архитектор, A/B-тест)
> **Статус:** ПРОЕКТ ПЛАНА — требует ADR (`ADR-MICROSERVICES-2026`) и приёмки пользователем/product.
> **Источники:** `.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md`, `docs/architecture.md`, `docs/stack_and_tools.md`,
> `docs/tz_constraints.md` / `.cursor/rules/60-tz-constraints.mdc`, `.cursor/playbooks/team-workflow.md`,
> `.cursor/rules/20-backend.mdc`, фактический код `backend/internal/**`, `backend/cmd/**`, `deployments/**`.

---

## 0. Исходное состояние (факты, проверено по коду)

| Факт | Значение | Следствие для плана |
|---|---|---|
| Стиль | Модульный монолит, Clean Architecture (handler→service→repo), score ~85/100 | Хорошая база для strangler fig — границы модулей уже есть |
| Модули | 31 domain-пакет + `platform/` в `backend/internal/`, wiring в `cmd/server/main.go` (~820 LOC) | main.go — единая точка композиции, узкое место при split |
| Данные | 83 Ent-схемы в **одном** PG 17; MongoDB 8 (messages, Allure, call chat); Redis 7 (refresh, RBAC cache, presence); SeaweedFS | Ent — один клиент на всю БД; messenger-данные УЖЕ отделены (Mongo) |
| События | NATS JetStream + **Outbox** (`platform/outbox`: Publisher/Worker) уже в проде | Event-driven фундамент готов, ничего изобретать не надо |
| Микросервисные прецеденты | `cmd/hub-mail` (transactional email), `cmd/meilisearch-indexer` — уже отдельные процессы | Паттерн выноса в репо уже обкатан |
| Auth | Magic Link, JWT **HS256** (golang-jwt v5), access 15м + refresh 30д в Redis | HS256 = общий секрет → для микросервисов нужен JWKS/asymmetric или контролируемый shared secret |
| Real-time | WS через Fiber, `ws.Hub` + NATS Bridge + `ws.Authorizer` (membership/RBAC) | WS-слой уже слушает NATS → готов быть отдельным сервисом |
| Инфра | Docker Compose dev/prod; **K3s: манифестов нет** (TD-04, только networkpolicy-metrics.yaml); Caddy | K3s-фундамент придётся строить с нуля — это отдельная фаза |
| Layering-долг | TD-02: `*ent.Client` в 7 доменных пакетах (auth, dashboard, backup, imex, storage, apigateway, rbac); TD-03: `GetDB()` escape hatch в tracker; TD-08/09 | Прямой split невозможен без укрепления границ — модули лезут в чужие таблицы через общий Ent-клиент |
| Наблюдаемость | slog + OTel v1.40 + Prometheus + Loki/Grafana (optional stack) | Trace propagation через HTTP/NATS нужно распространить |
| CI | GitHub Actions, ~3000 мин/мес (студенческий) | Матрица сервисов взорвёт бюджет → path-filtered jobs обязательны |
| Нагрузка | 1 org на инстанс, 10–50 человек | Мотивация split — организационная (изоляция, независимый деплой AI-командой) и точечное масштабирование (WS/calls/AI), НЕ throughput |

**Вывод из фактов:** Hub — редкий случай, когда монолит уже «модульный по-настоящему» (repository-интерфейсы, outbox, NATS, два вынесенных процесса). Но общий Ent-клиент + 7 пакетов с layering-утечками означают: **границы данных на бумаге есть, на уровне БД — нет.** Это главный технический риск и первый объект работы.

---

## 1. Ключевой вопрос: общая БД или database-per-service

### 1.1 Почему «одна общая БД на все сервисы» — анти-паттерн

Shared database между микросервисами (все сервисы читают/пишут одни таблицы) ломает главную причину, ради которой вообще делают микросервисы:

1. **Скрытая связность (coupling через схему).** Любой `ALTER TABLE issue` требует координации деплоя всех сервисов, которые её читают. Независимый деплой — миф.
2. **Нет владения данными.** Кто отвечает за инвариант `issue.status ∈ workflow`? Все и никто. Багфиксы превращаются в археологию.
3. **Скрытые JOIN'ы.** Сервис A джойнит таблицы сервиса B «потому что можно» → распределённый монолит с сетевыми затратами, но без изоляции.
4. **Миграции.** Atlas/Ent-миграции из нескольких сервисов на одну схему = гонки и конфликты версий схемы.
5. **Блокировки и шумные соседи.** Тяжёлый отчёт TMS лочит таблицы трекера.
6. **Безопасность.** Один credential ко всему → компрометация одного сервиса = компрометация всех данных.

### 1.2 Спектр вариантов для Hub

| Вариант | Изоляция | Оперсложность | Стоимость для Hub (1 нода, 10–50 юзеров) | Вердикт |
|---|---|---|---|---|
| **A. Shared DB, общие таблицы** | ❌ нет | низкая | нулевая, но это распределённый монолит | ❌ отвергнуть |
| **B. Один кластер PG, schema-per-service + отдельные DB-users, запрет cross-schema** | ✅ логическая | низкая-средняя | 1 PG-инстанс как сейчас; RAM/бэкапы не растут | ✅ **рекомендация** |
| **C. Database-per-service (отдельные PG-инстансы/кластеры)** | ✅✅ физическая | высокая (N бэкапов, N пулов, N мониторингов) | избыточно для одной ноды K3s и 50 юзеров | ⚠️ точечно, только там где уже естественно (Mongo для messenger) |

### 1.3 РЕКОМЕНДАЦИЯ (явная)

**Вариант B: один физический PostgreSQL 17, но `schema-per-service` с жёстким владением данными.** Это канонический компромисс (тот же «database-per-service» на логическом уровне — у каждого сервиса своя схема, свой миграционный конвейер, свой DB-user с грантами ТОЛЬКО на свою схему), без операционного налога N инстансов.

Правила (внести в ADR как инварианты):

1. **Один сервис = одна PG-схема** (`tracker.*`, `tms.*`, `calls.*`, `core.*`…) + свой DB-user (`hub_tracker` и т.д.) с `GRANT` только на свою схему. Cross-schema `SELECT/JOIN` **запрещён на уровне грантов**, не только code-review.
2. **Никаких FK между схемами.** Ссылки на чужие сущности — только UUID-значение + событийная синхронизация или API-запрос. (FK `issue.author_id → core.user.id` заменяется на UUID без constraint; консистентность — через события `user.deleted` + soft-delete политику.)
3. **Каждый сервис — свой Ent-проект и свои Atlas-миграции** (свой `ent/schema/` поднабор, свой versioned migration dir). Общий `backend/ent` распиливается по мере выноса.
4. **Данные читаются только у владельца:** synchronous API (внутренний REST) для редких запросов, **локальные read-model копии через NATS-события** для частых (например, notification держит свой кэш display_name юзеров по событиям `core.user.updated`).
5. **Mongo/Redis/SeaweedFS:** Mongo уже де-факто database-per-service (messenger messages, TMS Allure, call chat — разнести по отдельным Mongo-database при выносе); Redis — отдельные key-prefix/logical DB + отдельные ACL-users (Redis 7 ACL); SeaweedFS — бакеты per-service.
6. **Escape hatch:** если через 2–3 вынесенных сервиса обнаружим реальную потребность в физической изоляции (нагрузка, compliance) — переезд schema→dedicated instance тривиален, потому что связность уже разорвана. Вариант B — это **мост к C, а не тупик.**

Почему не C сразу: prod Hub — один хост/K3s-нода, 10–50 пользователей, бэкап-конвейер (`internal/backup`) рассчитан на 1×PG+1×Mongo+S3. N физических БД умножает бэкапы, мониторинг, connection-пулы и потребление RAM без единого выигрыша при такой нагрузке. Вся ценность (владение, независимые миграции, независимый деплой) достигается вариантом B.

---

## 2. Стратегия миграции: модульный монолит → strangler fig (НЕ big-bang)

### 2.1 Решение

**Двухтактная стратегия:**

- **Такт 1 — «Monolith hardening»:** укрепить границы внутри монолита (закрыть TD-02/TD-03/TD-08/TD-09, распилить Ent-схемы по доменным группам, ввести запрет cross-domain импортов через lint-правило, перевести все cross-domain вызовы на события/интерфейсы).
- **Такт 2 — strangler fig:** выносить по **одному** сервису за фазу; Caddy/gateway маршрутизирует `/api/<domain>/*` на новый сервис, монолит («hub-core») продолжает обслуживать остальное. Каждый вынос атомарен и **обратим** (rollback = вернуть route на монолит).

### 2.2 Обоснование для Hub (почему не прямой split)

1. **Реальное состояние кода:** 7 пакетов с `*ent.Client` в service-слое и `GetDB()` в tracker означают, что «чистые» границы — частично иллюзия. Прямой split размажет этот долг по N репозиториев/сервисов, где чинить его дороже.
2. **AI-команда — сильна в атомарных задачах** (план ≤200–300 строк = 1 коммит, волны ≤4 агентов, верификаторы). Strangler fig идеально ложится на этот процесс: «вынести notification» — одна фаза с 6–10 атомарными планами и проверяемым success criteria («все /api/notifications обслуживает новый под, регрессия green, latency ≤ x»). Big-bang split на 10+ сервисов невозможно верифицировать по частям — против всей нашей модели ворот.
3. **Обратимость.** По TZ смена deployment требует ADR; strangler даёт rollback на каждом шаге (route flip), big-bang — нет.
4. **CI-бюджет:** ~3000 мин/мес. Поэтапный вынос добавляет по одному path-filtered job за раз; матрица из 10 сервисов сразу — сожжёт бюджет на первой же неделе.
5. **Монорепо сохраняем** (не multi-repo): один go.mod workspace / общие shared-библиотеки `platform/*` версионируются вместе, PR-flow и агентские волны не меняются.

### 2.3 Что НЕ выносим вообще (и почему)

| Остаётся в hub-core | Причина |
|---|---|
| auth + org + onboarding | Источник идентичности; JWT-эмитент; трогать последним, если вообще |
| project + rbac + audit | Сквозные зависимости всех доменов (membership, permission bits, audit trail); RBAC-проверки нужны везде — выносить = latency на каждый запрос |
| storage (presign) | Тонкий слой над SeaweedFS, размазан по доменам |
| v4-скелеты (intranet, calendar, crm, hr, accounting, bases, marketplace, status) | Скелеты по 500–1500 строк (EXPANSION-ANALYSIS §2) — выносить нечего; сначала углубить функционал, потом решать. Микросервис из скелета = оверхед без ценности |
| dashboard, search-агрегатор, imex, backup, github/gitlab | Либо read-агрегаторы поверх чужих данных, либо тонкие интеграции — связность выше пользы выноса |

---

## 3. Границы сервисов, владение данными, порядок выноса

### 3.1 Целевая карта сервисов (горизонт плана)

| # | Сервис | Модули из монолита | Владеет данными | Почему выносим / когда |
|---|---|---|---|---|
| 0 | **hub-core** (монолит-остаток) | auth, project, rbac, audit, storage, dashboard, imex, backup, github, gitlab, dev, v4-модули | `core.*` (PG), Redis auth/rbac | Не сервис, а «ствол» strangler'а |
| 1 | **hub-notification** | notification | `notification.*` (PG: настройки, каналы, digest) | Уже event-consumer по природе (JetStream consumer). Минимальный sync-API. **Пилот — учимся на нём** |
| 2 | **hub-ai** | ai | `ai.*` (PG: settings, BYOK-ключи, лимиты, audit) | Внешние вызовы провайдеров = непредсказуемая latency, изоляция сбоев; BYOK-секреты в отдельном security-контуре; независимое масштабирование. Ранее уже помечен кандидатом (EXPANSION F2) |
| 3 | **hub-calls** | calls | `calls.*` (PG: rooms, scheduled, recordings-мета), Mongo call-chat | Соседствует с LiveKit, своя модель масштабирования (эгресс, запись); чат звонка уже в Mongo |
| 4 | **hub-messenger** | messenger (+ его часть ws) | Mongo messages (уже!), `messenger.*` (PG: channels, membership) | Данные УЖЕ в Mongo — самый «готовый» крупный домен; главная сложность — WS/presence (см. §3.3) |
| 5 | **hub-tms** | tms | `tms.*` (PG), Mongo Allure | 6834 строки, самый крупный и самый автономный домен: связи наружу — user-id и project-id, входящие — Allure upload + hub-cli |
| 6 | **hub-tracker** | tracker + sprint | `tracker.*` (PG: issues, workflow, boards, sprints) | Ядро продукта, максимальная связность (rbac, project, search, github, ai, dashboard). **Выносим последним**, только если такты 1–5 подтвердили ценность |
| — | **hub-ws** (edge) | ws (Hub, Bridge, Authorizer) | stateless (presence в Redis) | Отдельный WS-edge: терминация соединений, подписки, fan-out из NATS. Выносится вместе/до messenger |
| — | Уже есть | hub-mail, meilisearch-indexer | — | Прецеденты; привести к общим стандартам (§5) |

### 3.2 Критерии порядка (почему именно так)

Ранжирование по трём осям: **(а)** независимость данных (свои таблицы, мало FK наружу), **(б)** событийная природа (уже общается через NATS), **(в)** риск/blast radius при откате.

- `notification` — минимум по всем осям риска: consumer событий, почти нет sync-вызовов от других модулей, отказ = деградация уведомлений, не продукта. **Идеальный пилот**, на нём обкатываем весь конвейер (schema split, deploy, tracing, contract-тесты).
- `ai` — уже архитектурно шлюз; секреты BYOK выигрывают от изоляции; отказ изолирован (@hub-ai недоступен, ядро живёт).
- `calls` — чёткая граница (LiveKit-домен), уже частично Mongo.
- `messenger` — большой выигрыш (данные в Mongo), но тянет WS/presence — сложнее, идёт после обкатки на 1–3.
- `tms` — большой, но автономный; выигрыш — независимые релизы самого быстрорастущего домена.
- `tracker` — глубоко вплетён в rbac/project/search/dashboard/github; вынос оправдан только доказанной необходимостью. Держим как **опциональную** фазу.

### 3.3 Два трудных места (решить в ADR)

1. **WS и подписки.** Сегодня `ws.Authorizer` проверяет membership канала/проекта/звонка синхронно по БД. В целевой схеме hub-ws — тонкий edge: авторизация подписки — вызов внутреннего API владельца комнаты (messenger/calls/tracker) **или** локальный read-model membership, наполняемый событиями (`channel.member.added/removed`). Рекомендация: read-model в Redis (уже есть presence) — O(1) на подписку, инвалидация событиями. Инвариант NF-02 (нет подписки без RBAC-проверки) сохраняется, меняется только источник истины проверки.
2. **RBAC.** Выносить RBAC в сервис нельзя (latency на каждый запрос) и не нужно: permission-биты кешируются в Redis уже сейчас. Решение: hub-core остаётся владельцем ролей; сервисы читают **подписанный снапшот прав из JWT-claims (короткий TTL 15м покрывает большинство случаев) + Redis-кэш** с инвалидацией по событию `rbac.role.changed`. Это фиксируем в ADR как «RBAC остаётся централизованным, распределяется кэшем, не сервисом».

---

## 4. Межсервисное взаимодействие

| Тип | Механизм | Правила |
|---|---|---|
| **Домен-события** (основной путь) | NATS JetStream + существующий Outbox (`platform/outbox`) | Каждый сервис: своя outbox-таблица в своей схеме → свой Worker → JetStream stream per-domain (`HUB_TRACKER`, `HUB_MSG`…). Subjects как сейчас: `hub.<domain>.<entity>.<verb>` |
| **Синхронные вызовы** (минимизировать) | **Внутренний REST/JSON** (Fiber), через K8s ClusterIP DNS (`http://hub-tms.hub.svc`) | gRPC НЕ вводим на старте: новый инструмент в стек = ADR + обучение агентов + кодоген; выигрыш при наших RPS ~нулевой. Пересмотреть после 3 сервисов, если типизация контрактов станет болью. Паттерн внутренней авторизации уже есть — `mail/internal_auth.go` |
| **Идемпотентность** | JetStream `Nats-Msg-Id` (dedup window) + идемпотентные consumers (upsert по event id, таблица `processed_events` в схеме сервиса) | At-least-once — уже принятая семантика outbox; consumer обязан быть идемпотентным — внести в `20-backend.mdc` |
| **Саги / eventual consistency** | Хореография через события (без оркестратора саг) | Пример: CRM `deal.won` → accounting создаёт invoice-draft → событие `invoice.created` → уведомление. Компенсация — обратные события. Оркестрируемые саги (temporal и т.п.) НЕ вводим — масштаб не тот |
| **Анти-распределённый-монолит** | Правило: **синхронный вызов другого сервиса в hot-path запрещён без записи в ADR-приложение**; каждый sync-вызов — таймаут ≤2с, retry с jitter, circuit breaker (обёртка в `platform/http`) | Если сервису для ответа на свой запрос нужно >1 синхронного вызова соседа — граница нарезана неправильно, возвращаем в arch |
| **Контракты** | Версионируемые: события — JSON-схемы в `contracts/events/*.json`; внутренние API — OpenAPI в `contracts/api/*` | Contract-тесты в testframework (§5.6): producer проверяет «пишу по схеме», consumer — «читаю по схеме». Изменение контракта = отдельный атомарный коммит первым (правило волн §4.3 team-workflow) |

---

## 5. Сквозные заботы

### 5.1 Auth/JWT между сервисами

- **Сейчас:** HS256 общий секрет — для монолита нормально, для N сервисов = секрет в N местах, ротация болезненна.
- **Целевое:** hub-core (auth) переходит на **асимметричную подпись (EdDSA/RS256) + JWKS-endpoint** (`/api/auth/.well-known/jwks.json`). Сервисы валидируют access-токен **локально** по публичному ключу (кэш JWKS, TTL), zero сетевых вызовов на запрос. Ротация ключей — публикация нового kid в JWKS.
- **Переходно (фаза M1):** общий HS256-секрет через K8s Secret — допустимо для пилота (notification), миграция на JWKS в фазе M2 до выноса ai (BYOK-секреты требуют строгого контура).
- **S2S-вызовы:** отдельные service-токены (client-credentials у auth, aud=имя сервиса) либо K8s ServiceAccount + NetworkPolicy как первый рубеж. Инварианты TZ (magic link only, 15м/30д, revocation в Redis) — не затрагиваются: refresh-flow остаётся только в hub-core.

### 5.2 API Gateway и маршрутизация

- **Caddy остаётся единственным edge** (инвариант TZ). Стратегия strangler — на нём: `/api/notifications* → hub-notification`, остальное `/api/* → hub-core`. Внешний контракт API для фронтенда **не меняется** — пути стабильны.
- Модуль `apigateway` (API-keys, outbound webhooks) остаётся в hub-core; API-key-аутентификация внешних клиентов — на edge-уровне hub-core, для вынесенных доменов — заголовок identity пробрасывается (подписанный) или сервис валидирует сам по JWKS.
- Никакого отдельного gateway-продукта (Kong/Traefik) не вводим — Caddy покрывает потребность, замена запрещена TZ без ADR.

### 5.3 Discovery, конфиги, секреты

- **Discovery:** K8s DNS (`<svc>.hub.svc.cluster.local`) — ничего дополнительно (no consul/etcd).
- **Конфиг:** текущий `platform/config` (env) переиспользуется; per-service ConfigMap. Общие настройки (NATS URL, OTel endpoint) — общий ConfigMap `hub-shared`.
- **Секреты:** K8s Secrets (позже — sealed-secrets), per-service: свой PG-user/password, свой Redis ACL, JWT public key (все) / private key (только hub-core). Прямо связано с PG-03 rotation — раскладка секретов per-service делает будущие ротации локальными.

### 5.4 Observability

- **Tracing:** OTel уже в стеке. Обязательно: (а) W3C traceparent через HTTP-мидлварь во всех сервисах, (б) **прокидывание trace-контекста через NATS headers** в outbox Worker и consumers — иначе следы рвутся на асинхронной границе (главная жалоба всех микросервисных отладок). Один trace = login→REST→outbox→NATS→notification→WS push.
- **Метрики:** `/metrics` в каждом сервисе, Prometheus scrape по аннотациям; RED-дашборд per-service + межсервисная latency-панель.
- **Логи:** slog JSON → Loki; обязательные поля `service`, `request_id`, `trace_id` (request_id уже сквозной).
- **Алерты:** per-service liveness + межсервисные (лаг JetStream consumer'а, глубина outbox, error-rate sync-вызовов).

### 5.5 CI/CD и бюджет минут

- **Path-filtered jobs:** `dorny/paths-filter` — тесты/сборка сервиса только при изменении его каталога + `platform/` (общий код триггерит всех — стимул держать `platform/` стабильным).
- **Образы:** один multi-stage Dockerfile с `--build-arg SERVICE=<name>` (все сервисы — `cmd/<name>/main.go`), GHCR-теги `hub-<svc>:<sha>`.
- **Deploy:** поэтапно per-service (K8s rollout, health-gate), не «весь стек разом».
- **Бюджет:** оценка — full-CI сегодня ~25–30 мин; после split при типичном PR в 1 сервис — ~8–12 мин. Экономия, а не рост, ЕСЛИ path-filter внедрён в фазе M0. e2e-full остаётся общим (master-only).

### 5.6 Тестирование

| Уровень | Что | Где |
|---|---|---|
| Unit/integration per-service | Как сейчас (testcontainers, `-tags hubtest`) | `backend/testframework/<domain>/` — уже разложено по доменам, split почти бесплатный |
| **Contract-тесты (новое)** | Producer: события соответствуют `contracts/events/*.json`; Consumer: обрабатывает эталонные события; внутренние API — golden request/response по OpenAPI | `testframework/stages/contracts/` (новая стадия pipeline) |
| e2e cross-service | Существующие 20 Playwright-спеков поверх compose-стека, где вынесенные сервисы подняты отдельными контейнерами | `testframework/stages/e2e/` — сценарии НЕ меняются (внешний API стабилен) — это и есть главный регрессионный щит strangler'а |
| Coverage-gate | Per-service (≥82% сохраняем), суммарный отчёт в COVERAGE.md | quality stage |

---

## 6. Deployment на K3s

> TD-04: манифестов сейчас нет — фаза M0 строит фундамент.

- **Namespaces:** `hub` (приложение), `hub-infra` (PG, Mongo, Redis, NATS, SeaweedFS, LiveKit — StatefulSets + PVC local-path), `hub-observability` (Prometheus, Loki, Grafana). NetworkPolicy: app→infra по потребности, deny по умолчанию между app-сервисами кроме объявленных (первый рубеж S2S-безопасности).
- **Ingress:** Caddy как Deployment+Service LB (сохраняем Caddyfile-модель, НЕ переходим на nginx-ingress — TZ). Маршруты strangler в Caddyfile → ConfigMap.
- **Масштабирование:**
  - `hub-ws` — 2+ реплики сразу: соединения шардируются, fan-out через NATS (архитектура Bridge уже это умеет); presence в Redis — общий. Sticky-sessions не нужны (WS = 1 длинное соединение, upgrade балансируется на установке).
  - `hub-calls` — HPA по CPU (egress/запись); LiveKit — отдельный под, `/lk/*` через Caddy как сейчас.
  - `hub-ai` — HPA по in-flight запросам к провайдерам; лимиты RPM уже в модуле.
  - Остальные — по 1 реплике (нагрузка 10–50 юзеров), но манифесты пишем сразу с PodDisruptionBudget/readiness, чтобы масштабирование было «крутилкой», а не проектом.
- **Rollout:** RollingUpdate + readiness-gate; rollback сервиса = `kubectl rollout undo` + (для strangler-шага) flip маршрута в Caddy ConfigMap.
- **Dev-режим:** Docker Compose остаётся (TZ-инвариант); compose-профили `--profile split` поднимают вынесенные сервисы отдельными контейнерами, по умолчанию — монолит. Dev-петля агентов не ломается.

---

## 7. TZ / ADR

**ADR обязателен: `ADR-MICROSERVICES-2026`** (владелец arch, верифицирует arch-ver, принимает пользователь+product). Содержание — решения §1–§6 этого плана. Затрагиваемые инварианты:

| Инвариант TZ | Статус |
|---|---|
| Docker Compose dev / K3s prod, Caddy | ✅ сохраняется (K3s наконец реализуется; смена deployment-модели = сам предмет ADR) |
| NATS JetStream — единственный pub/sub | ✅ сохраняется и становится хребтом |
| WebSocket через Fiber | ✅ сохраняется (hub-ws — тоже Fiber) |
| JWT 15м + refresh Redis 30д, Magic Link only | ✅ сохраняется; **изменение**: HS256 → asymmetric+JWKS — зафиксировать в ADR отдельным пунктом (алгоритм подписи не был инвариантом, но связан с auth-контуром) |
| 1 org = 1 инстанс | ✅ не затрагивается (микросервисы ≠ multi-tenant) |
| UUID PK, UTC, optimistic locking, soft delete | ✅ переносятся в каждую сервисную схему без изменений |
| Rate limiting (100/сек API, 3/10м auth…) | ⚠️ уточнить в ADR: лимиты на edge (Caddy/hub-core) vs per-service — рекомендация: auth-лимиты в hub-core, доменные — в сервисе |
| «`*ent.Client` в Service/Handler запрещён» | ✅ ужесточается: + запрет cross-schema доступа на уровне DB-грантов |

Новые инварианты в ADR: schema-per-service + гранты; запрет cross-schema FK/JOIN; идемпотентные consumers; sync-вызов в hot-path — только по записи в ADR-приложении; contracts/ версионируются.

---

## 8. Риски и митигация

| # | Риск | Вероятность/Влияние | Митигация |
|---|---|---|---|
| R1 | Распределённые транзакции (данные рвутся между схемами: won deal → invoice) | Средняя / Высокое | Outbox уже даёт атомарность «данные+событие»; саги-хореография + компенсации; НЕ размазывать один агрегат по сервисам (граница = агрегат целиком) |
| R2 | Latency от sync-вызовов (RBAC, membership) | Высокая / Среднее | JWT-claims + Redis read-model (§3.3); правило «0–1 sync-вызов на запрос»; замер p95 в success criteria каждой фазы |
| R3 | Оперсложность (N подов, N миграций, K3s с нуля) | Высокая / Высокое | Такт M0 строит фундамент ДО первого выноса; по одному сервису за фазу; devops-engineer + runbook per-service; общий шаблон манифестов |
| R4 | Отладка через границы (обрыв trace на NATS) | Средняя / Среднее | Trace-контекст в NATS headers с фазы M0 (до выноса!); request_id сквозной уже есть; Grafana cross-service дашборд |
| R5 | Откат неудачного выноса | Низкая / Высокое | Strangler: код остаётся в монорепо, маршрут flip-ается обратно за минуты; данные: на переходный период (1 фаза) schema-миграция обратно-совместима (expand-contract) |
| R6 | Дрейф контрактов событий | Средняя / Среднее | contracts/ + contract-тесты в CI (§5.6); изменение схемы события = versioned subject (`.v2`) |
| R7 | CI-бюджет 3000 мин | Средняя / Среднее | Path-filter с M0; e2e-full только на master; замер минут в success criteria M0 |
| R8 | Распил общего `backend/ent` (83 схемы) сломает кодоген/тесты | Средняя / Высокое | Поэтапно: сначала выносимая доменная группа в отдельный Ent-проект, остальное не трогаем; testframework уже разложен по доменам |
| R9 | AI-агенты запутаются в multi-service репо | Низкая / Среднее | Обновить playbooks/rules (`20-backend.mdc`, `00-project.mdc`) в той же фазе, что и структура; карта «модуль→сервис» в docs/architecture.md |
| R10 | Split ради split'а: ценность не подтверждается | Средняя / Высокое | Go/No-Go gate после M2 (пилот): если независимый деплой/изоляция не дали измеримой пользы — останавливаемся на модульном монолите + hub-ws/hub-ai, без потерь (M0–M1 полезны сами по себе) |

---

## 9. Поэтапный roadmap (фазы → планы → волны)

> Оценки: S ≈ 1–2 дня · M ≈ 3–5 дней · L ≈ 1–2 недели · XL > 2 недель. Волны ≤4 агентов (team-workflow §4). Каждая фаза заканчивается: регрессия green + `/qa-gate` + docs-sync + приёмка team-lead.

### Фаза M0 — Решение и фундамент (ADR + K3s + observability) — **M/L**

**Depends on:** ответы на открытые вопросы (§10), приёмка этого плана.

| План | Содержание | Исполнитель | Волна |
|---|---|---|---|
| M0-01 | `ADR-MICROSERVICES-2026`: решения §1–§7, инварианты, Go/No-Go критерии | arch → arch-ver | 1 |
| M0-02 | Спека контрактов: формат `contracts/events/*.json` + реестр текущих NATS-subjects по коду | sa1 → sa-ver | 1 |
| M0-03 | Спека внутренней auth: JWKS-дизайн, S2S-токены, план миграции HS256→EdDSA | sa2 + security-auditor → sa-ver | 1 |
| M0-04 | K3s фундамент: namespaces, infra StatefulSets (PG/Mongo/Redis/NATS/SeaweedFS/LiveKit), Caddy ingress, NetworkPolicy, шаблон Deployment | devops-engineer | 2 |
| M0-05 | CI: path-filtered jobs + параметризованный Dockerfile (`SERVICE` build-arg), замер минут | devops-engineer | 2 (после M0-04) |
| M0-06 | OTel: trace-контекст через NATS headers (outbox Worker + consumers), W3C propagation middleware | be1 → be-ver | 2 |
| M0-07 | Docs: architecture.md (целевая C4 L2), tz_constraints/60-rule update, playbooks | docs-writer | 3 |

**Success criteria:** ADR принят пользователем; монолит крутится в K3s (или готовые манифесты проверены в kind/k3d) параллельно compose-prod; трейс REST→NATS→consumer непрерывен; CI-минуты замерены (baseline).

### Фаза M1 — Monolith hardening (границы модулей) — **L**

**Depends on:** M0-01. Полезна независимо от split (закрывает TD-02/03/08/09).

| План | Содержание | Исполнитель | Волна |
|---|---|---|---|
| M1-01 | Repository-интерфейсы вместо `*ent.Client`: auth, rbac | be1 → be-ver | 1 |
| M1-02 | То же: dashboard, backup, imex, storage, apigateway | be2 → be-ver | 1 |
| M1-03 | Убрать `GetDB()` escape hatch (tracker), TD-08 audit/retention, TD-09 notification/push | be1 → be-ver | 2 |
| M1-04 | Lint-правило: запрет cross-domain импортов `internal/<a>` → `internal/<b>` (кроме platform), белый список текущих нарушений → выжигание | be2 + qa-auto1 | 2 |
| M1-05 | Каталогизация cross-domain вызовов: что станет событием, что sync-API, что read-model (вход для фаз выноса) | sa1 → sa-ver | 2 |
| M1-06 | Группировка Ent-схем по доменным «данным-владельцам» (пока в одном проекте, но с манифестом владения `contracts/data-ownership.md`) | sa2 + be1 | 3 |
| M1-07 | Регрессия full + coverage не упал; docs-sync | qa-auto2 → qa-auto-ver | 3 |

**Success criteria:** 0 `*ent.Client` вне `ent_repository.go`; lint-запрет в CI green; манифест владения данными покрывает 83 схемы; COV ≥82%.

### Фаза M2 — Пилот: hub-notification (strangler #1) — **L**

**Depends on:** M0 (вся), M1-05/06. **Это фаза-образец: её планы становятся шаблоном для M3–M6.**

| План | Содержание | Исполнитель | Волна |
|---|---|---|---|
| M2-01 | PG-схема `notification` + DB-user + Atlas-миграция переноса таблиц (expand-contract) | be1 | 1 |
| M2-02 | Отдельный Ent-проект/подмодуль notification + `cmd/hub-notification` (Fiber, health, metrics, JWKS-валидация) | be2 | 1 |
| M2-03 | Contract-тесты: события, которые consumer'ит notification (issue.*, message.*, sprint.* …) | qa-auto1 | 2 |
| M2-04 | K3s Deployment + Caddy route flip `/api/notifications*` + compose `--profile split` | devops-engineer | 2 |
| M2-05 | Read-model пользовательских display-данных (по событиям core.user.*) вместо cross-schema чтения | be1 | 2 |
| M2-06 | E2E: полный existing-набор поверх split-топологии + новый спек «уведомление сквозь 2 сервиса» | qa-auto2 → qa-auto-ver | 3 |
| M2-07 | Runbook + docs-sync + замер p95/CI-минут | devops-engineer + docs-writer | 3 |
| M2-08 | **Go/No-Go review** по критериям ADR | team-lead + product + arch | 4 |

**Success criteria:** все `/api/notifications*` обслуживает отдельный под; монолит не содержит notification-кода в рантайме; e2e 20/20 green; p95 деградация ≤10%; rollback-процедура отрепетирована (flip назад ≤5 мин); Go/No-Go зафиксирован.

### Фаза M3 — hub-ai — **L** (Depends on: M2 Go)

Планы по шаблону M2 (схема+user, cmd, contracts, deploy, e2e) + специфика: M3-05 BYOK-секреты в отдельный K8s Secret + аудит контура (security-auditor); M3-06 messenger-bot интеграция через события/внутренний API вместо inprocess-вызова. Волны: be1∥be2 (сервис/данные) → devops∥qa-auto1 → security-auditor∥qa-auto2.
**Success criteria:** @hub-ai и /ai-команды работают через сервис; отказ hub-ai не влияет на ядро (chaos-тест kill pod); ключи BYOK недоступны из других подов.

### Фаза M4 — hub-calls — **L** (Depends on: M2 Go; параллелизуема с M3 при свободных агентах)

Специфика: M4-04 LiveKit-токены выпускает hub-calls (FA-02-03 constant-time compare закрыть здесь же); M4-05 call-chat Mongo — отдельная database; M4-06 WS-события звонков через NATS (уже так). Recordings-мета + SeaweedFS bucket `calls`.
**Success criteria:** звонок c записью и чатом end-to-end через отдельный сервис; e2e calls-спеки green; HPA-манифест проверен.

### Фаза M5 — hub-ws + hub-messenger — **XL** (Depends on: M3 или M4 — опыт 2 выносов; ADR-решение §3.3 по авторизации подписок)

| План | Содержание | Исполнитель |
|---|---|---|
| M5-01 | hub-ws edge: вынос ws.Hub+Bridge, membership read-model в Redis (события `*.member.*`), Authorizer на read-model | be1 |
| M5-02 | Нагрузочный тест WS ×2 реплики (fan-out через NATS, отсутствие двойной доставки) | qa-auto1 |
| M5-03 | hub-messenger: PG-схема (channels/membership) + Mongo database split | be2 |
| M5-04 | Rate-limit 60/мин на сервисе; presence в Redis — общий контракт с hub-ws | be2 |
| M5-05..07 | contracts / deploy+route / e2e messenger-домен | qa-auto2, devops, qa-auto1 |

**Success criteria:** WS-переподключение при рестарте пода ≤5с; NF-02 инвариант доказан тестом (подписка без membership отклоняется); messenger e2e green; 2 реплики hub-ws без дублей сообщений.

### Фаза M6 — hub-tms — **L/XL** (Depends on: M2 Go; независима от M5 — параллелизуема)

Специфика: Allure-приёмка (hub-cli push-results) — прямой upload в hub-tms; связь с tracker (дефекты из прогонов) — через события `tms.run.finished` + UUID-ссылки.
**Success criteria:** импорт Allure и полный TMS-флоу через сервис; hub-cli работает без изменений контракта; tms e2e green.

### Фаза M7 (ОПЦИОНАЛЬНАЯ) — hub-tracker(+sprint) — **XL**

**Depends on:** M5+M6 в проде ≥2–4 недели без инцидентов; отдельное product-решение. Вынос ядра оправдан только если подтверждена боль независимого деплоя ядра. По умолчанию — **не делаем**, tracker остаётся сердцем hub-core.

### Сводная карта зависимостей

```
M0 (ADR+K3s+CI+OTel) ──► M1 (hardening) ──► M2 (пилот notification) ──► Go/No-Go
                                                    │
                          ┌─────────────────────────┼───────────────────┐
                          ▼                         ▼                   ▼
                     M3 (hub-ai)              M4 (hub-calls)      M6 (hub-tms)
                          └────────────┬───────────┘
                                       ▼
                             M5 (hub-ws + messenger)
                                       ▼
                             M7 (tracker) — опционально, отдельное решение
```

Параллелизм: M3∥M4 и M6 — в разных волнах разными парами (be1/be2 не пересекаются по модулям); внутри фазы — волны по шаблону M2. Оценка всего горизонта M0–M6: ~8–12 недель календарно при текущем ростере.

---

## 10. Открытые вопросы к продукту (до старта M0)

1. **Мотивация split:** что болит сейчас — изоляция сбоев, независимый деплой доменов, масштабирование WS/calls/AI, или «хочется правильной архитектуры»? От ответа зависит Go/No-Go критерий M2 и судьба M7. (Если мотивация только организационная — возможно, M0+M1 + вынос лишь hub-ws/hub-ai закрывает 80% ценности за 30% стоимости.)
2. **Прод-топология:** одна K3s-нода или планируется несколько? Влияет на ценность HPA/реплик и на выбор storage-class.
3. **Приоритет vs EXPANSION-бэклог:** углубление v4-модулей (эпик B) и split конкурируют за be/fe-агентов. Что раньше? (Рекомендация: FA-06 → M0/M1 [полезны всегда] → решение по M2 vs эпик B.)
4. **FA-06/PG-03:** ротация секретов — до M0 (желательно: раскладку секретов per-service делать уже чистыми ключами).
5. **Допустим ли переходный период HS256** (общий секрет) для пилота M2, или JWKS — блокер первого же выноса?
6. **Бюджет минут CI:** остаёмся в студенческом пакете или есть готовность к платному тарифу при росте матрицы?
7. **SLA на время миграции:** допустимы ли короткие (минуты) окна недоступности домена при flip маршрута, или нужен полностью бесшовный переход (blue-green на уровне Caddy)?

---

## Саммари-рекомендация

1. **БД:** один PostgreSQL 17, но **schema-per-service** с отдельными DB-user'ами, запретом cross-schema FK/JOIN на уровне грантов и собственными миграциями каждого сервиса. Shared-таблицы — отвергнуть; отдельные физические БД — избыточны для 1 ноды и 10–50 юзеров, мост к ним остаётся открытым.
2. **Путь:** не big-bang, а **модульный монолит-hardening (M1) → strangler fig по одному сервису** за фазу, с flip-rollback на Caddy.
3. **Порядок:** notification (пилот) → ai → calls → ws+messenger → tms; tracker — только по доказанной необходимости; v4-скелеты не выносим.
4. **Взаимодействие:** NATS JetStream + существующий outbox как хребет; sync — внутренний REST по минимуму; идемпотентные consumers; contract-тесты в CI.
5. **Обязательные предпосылки:** ADR-MICROSERVICES-2026, K3s-фундамент (TD-04), JWKS вместо HS256, OTel-трейсы через NATS, path-filtered CI.
6. **Go/No-Go gate после пилота** защищает от «split ради split» — M0/M1 полезны проекту при любом исходе.
