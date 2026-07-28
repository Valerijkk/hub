# ADR-ARCH-MICROSERVICES-2026: Поэтапный вынос Hub в доменные микросервисы

- **Status:** accepted (2026-07-02, продуктовое решение пользователя)
- **Владелец:** arch · **Верифицирует:** arch-ver + security-auditor · **Оркестрация:** team-lead
- **Источники:** A/B-планы `.planning/decisions/ab-microservices/PLAN-team-lead-opus.md` и `PLAN-team-lead-fable.md`, `.planning/research/MARKET-AND-NEEDS-2026-07-02.md`, `.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md`.
- **Заменяет/уточняет:** `docs/tz_constraints.md` (deployment), помечавший микросервисный split как F2/v4.5.

---

## Контекст

Hub — качественный **модульный монолит** (Go/Fiber, 83 Ent-схемы, единый NATS JetStream + outbox, уже вынесены `hub-mail`, `meilisearch-indexer`, внешний LiveKit). Продуктовое решение — двигаться к доменным микросервисам по блокам. Ключевой вопрос: «одна общая БД или нет».

Два независимых A/B-плана (Opus 4.8 и Fable 5) и рыночное исследование **сошлись** в выводах — это и есть основание ADR.

## Решение

### 1. База данных
- **Shared-таблицы между сервисами — ЗАПРЕЩЕНЫ** (shared-database антипаттерн → распределённый монолит).
- **Старт: schema-per-service в одном PostgreSQL 17** — у каждого сервиса своя схема + свой DB-user с `GRANT` только на неё; cross-schema `JOIN/FK` запрещён **на уровне грантов**, не только ревью.
- Физически отдельная БД — **точечно по профилю нагрузки** (Mongo уже де-факто разделён: messages/Allure/call-chat; Redis — ACL-users/prefix; SeaweedFS — бакеты per-service).
- Кросс-доменное чтение — **CQRS-проекции через NATS-события**, не прямые запросы к чужим данным. Это мост к DB-per-service, а не тупик.

### 2. Стратегия
- **НЕ big-bang.** Фаза 0 «Modular Monolith Hardening» (чинит утечку `*ent.Client` TD-02/03, укрепляет границы — **ценна независимо от split**) → **Strangler Fig** по одному сервису, обратимо (route-flip в Caddy).
- Пилот выноса — **`notification`** (чистый event-consumer, минимальный blast-radius).
- Монорепо сохраняем; per-service Ent-проекты и Atlas-миграции по мере выноса.

### 3. Целевой каталог (уточнено 2026-07-03 по решениям gray-zones)
| Юнит | Состав | Статус |
|------|--------|--------|
| hub-core | auth, rbac, project, **tracker** (исключён из выноса), sprint, tms (условно), dashboard, search-agg, github/gitlab, v4-модули, audit/backup/imex/storage | ядро (вместе) |
| hub-ai | AI-gateway (LiteLLM-паттерн: универсальный OpenAI-совместимый адаптер, virtual keys/бюджеты) + agent-native/MCP, BYOK-изоляция | вынос №1 |
| hub-ws | WS-edge (Hub/Bridge/Authorizer, membership read-model) | вынос №2 (realtime-трио) |
| hub-messenger | messenger (Mongo messages, channels) | вынос №2 (realtime-трио) |
| hub-calls | calls-signaling (LiveKit внешний, call-chat Mongo) | вынос №2 (realtime-трио) |
| hub-notification | пилот strangler-fig | выносим первым (обучение) |
| hub-mail, meilisearch-indexer | — | уже вынесены |
| LiveKit | WebRTC | внешний |

**Решения пользователя (2026-07-03):** realtime — **три отдельных сервиса** (независимое масштабирование calls/egress), не единый юнит. **Scope горизонта: пилот notification → hub-ai → realtime-трио.** tms — условная фаза M6 (только после Go по пилоту); **tracker явно исключён** из горизонта (возврат — отдельным ADR при доказанном драйвере). «search» в порядке = стандартизация уже вынесенного indexer, агрегатор остаётся в ядре. `auth` → JWKS/identity-контракт, не отдельный первый сервис.

**Go/No-Go метрика после пилота (композит):** lead-time деплоя домена ↓, p95 API не хуже +10%, 0 инцидентов из-за split за 2 недели.

**Прочие зафиксированные детали:** JWKS — HS256 допустим для пилота, JWKS обязателен до hub-ai; OTel-трейсинг через NATS-headers — жёсткий gate до пилота; rate-limits слоями (auth в core, доменные per-service, 100/с на Caddy); path-filtered CI обязателен в M0 (бюджет 3000 мин); переключения — Caddy route-flip; одна K3s-нода + local-path PVC; dev-профиль compose = split после приёмки каждого выноса; сироты UUID — consumer на `*.deleted` + reconcile-job; бэкапы при N схемах — решить на пилоте M2. Полный реестр: `.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`.

### 4. Взаимодействие
- Async — **NATS JetStream + outbox** (уже есть), идемпотентные consumers (`event_id` dedup), саги-хореография + компенсации.
- Sync — минимизировать; внутренний REST (gRPC не вводим на старте — избыточно для нагрузки). Правило: **>1 sync-хопа на запрос = граница нарезана неверно**.

### 5. Auth между сервисами
- HS256 (общий секрет) → **асимметричная подпись (EdDSA/RS256) + JWKS-endpoint**; сервисы валидируют JWT локально. Refresh/revoke остаётся централизованно в Redis (TZ-инвариант сохраняется: Magic Link, 15м/30д, instant revoke).

### 6. Сквозное
- OTel trace-context через NATS-headers (предусловие, не «потом»). Path-filtered CI (бюджет ~3000 мин). Contract-тесты (producer↔consumer). K3s namespaces + Caddy ingress (не nginx).

## Затрагиваемые инварианты TZ
Deployment-топология (много сервисов) — **этот ADR и есть их изменение**. Сохраняются: NATS единственный pub/sub, WS через Fiber, «1 org = 1 инстанс», UUID/UTC/optimistic-lock/soft-delete, SeaweedFS, Caddy. Изменение: HS256→asymmetric+JWKS (зафиксировано здесь).

## Go/No-Go gate
После пилота (`notification`) — ретроспектива team-lead + product: если независимый деплой/изоляция не дали измеримой пользы, **остановиться на модульном монолите + hub-ai/hub-realtime** без потерь (Фаза 0/1 полезны всегда).

## Последствия
- **+** изоляция сбоев/секретов (BYOK), независимое масштабирование WS/AI, независимый деплой доменов, чистые границы данных.
- **−** операционная сложность (N деплоев/секретов/миграций), риск распределённого монолита, latency, сложнее отладка → митигируется постепенностью, observability-предусловием, обратимостью.

## Следствия для документации
Обновить `docs/architecture.md` (целевая C4), `docs/stack_and_tools.md`, `docs/tz_constraints.md` (deployment-раздел), `deployments/` (K3s-манифесты, TD-04). Детальный roadmap — `.planning/archive/roadmaps/MICROSERVICES-ROADMAP.md`.
