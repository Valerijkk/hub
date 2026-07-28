# Portable Hub Architecture (расширенный)

> Шаблон корпоративного хаба: **AI-команда + урезанный GitHub-flow + микросервисы (ADR)**.
> Не код — переносимая архитектура kit. База: `ADR-ARCH-MICROSERVICES-2026`, стек Hub, TZ.
> Связанные: [`05-PORTABLE-KIT.md`](05-PORTABLE-KIT.md), [`06-REDUCED-GITHUB-MODULE.md`](06-REDUCED-GITHUB-MODULE.md),
> skill `gsd-ru/microservices-strangler`, roadmap `.planning/archive/roadmaps/MICROSERVICES-ROADMAP.md`.

- **Статус:** Architecture template / docs-only
- **Владелец:** arch · **Верифицирует:** arch-ver
- **Назначение:** скопировать границы и инварианты в другой monorepo, адаптировав стек, **сохранив** ADR-границы

---

## 1. Принципы переносимости

1. **Границы важнее фреймворка** — Fiber/NATS можно заменить только с ADR; границы core/ai/notification/realtime — нет.
2. **Schema-per-service** с первого выноса; shared tables запрещены.
3. **Async-first** (NATS JetStream + outbox); sync ≤1 hop.
4. **Обратимый strangler** (Caddy route-flip), не big-bang.
5. **AI Ops слой копируется первым** (agents/skills/PE) — продукт-код опционален (слои kit L1–L6).

---

## 2. C4 Level 1 — System Context

```
                    ┌─────────────────────────────────────┐
 [Команда 10–50] ──►│         Portable Hub                 │
                    │  Web SPA · API · Realtime · AI Ops   │
                    └──────────────┬──────────────────────┘
           ┌───────────────┬───────┴────────┬──────────────┐
           ▼               ▼                ▼              ▼
    [Email IdP]     [Git Forge]      [SMTP/Resend]   [SFU LiveKit]
    Magic Link       optional          optional        external
    (+ SSO later)    GitHub/GitLab
```

**Внешние акторы:** разработчик, тимлид, admin/founder, CI runner (webhooks), AI-оператор (Cursor/Claude с kit).

**Доверие:** 1 org = 1 инстанс (multi-org → v5 / отдельный ADR). Telegram — только notify bind, не login (`ADR-AUTH-01`).

---

## 3. C4 Level 2 — Containers (целевые)

| Контейнер | Технология (Hub) | Ответственность | В kit |
|-----------|------------------|-----------------|-------|
| **edge** | Caddy | TLS, route-flip, static, `/api` `/ws` `/lk` | deploy skeleton |
| **hub-core** | Go Fiber monorepo module | auth, rbac, project, **tracker**, sprint, tms?, dashboard, search-agg, github/gitlab, **devflow**, v4, audit, storage | module map |
| **hub-notification** | Go service (пилот) | consumers событий → in-app/TG/push | strangler checklist |
| **hub-ai** | Go + LiteLLM-pattern | gateway, MCP/agents, BYOK | ADR + prompts |
| **hub-ws** | Go Fiber WS edge | Hub/Bridge/Authorizer, membership RM | ADR timeline |
| **hub-messenger** | Go + Mongo | channels, messages | ADR |
| **hub-calls** | Go + LiveKit | signaling; SFU внешний | ADR; mediasoup REJECTED |
| **hub-mail** | already extracted | transactional email | pattern docs |
| **meilisearch-indexer** | already extracted | index worker | pattern docs |
| **PostgreSQL 17** | schemas per service | SoT транзакционных данных | TZ + ADR §1 |
| **MongoDB 8** | per concern DB/collections | messages, Allure, call-chat | уже «разделён» |
| **Redis 7** | ACL users / prefix | sessions, rate limit, revoke | TZ refresh |
| **NATS JetStream** | subjects `hub.*` | единственный pub/sub | mandatory |
| **SeaweedFS** | S3 aws-sdk-go-v2 | files, buckets per-service | TZ |
| **AI Ops** | Cursor/Claude | agents, skills, PE, GSD-RU | **copy first** |

```
Browser ⇄ Caddy
          ├─ /api/*  → hub-core (и route-flip → split services)
          ├─ /ws     → hub-ws (после выноса) / core WS
          ├─ /lk/*   → LiveKit
          └─ /*      → SPA static
```

---

## 4. C4 Level 3 — Components (hub-core, сжато)

```
hub-core
├── auth (Magic Link, JWT 15m, refresh Redis 30d)
├── rbac
├── project / tracker / sprint / tms?
├── github / gitlab (forge adapters)
├── devflow (reduced GitHub — Later; см. 06)
├── storage / audit / backup / imex / apigateway
├── search aggregator (indexer уже снаружи)
├── platform/ (config, outbox, nats, middleware, safedialer, obs)
└── wiring (cmd/server) — composition root
```

**Frontend (portable map):** TanStack Router · Query · Zustand · Shadcn className-only · modules/* зеркалят домены.

**AI Ops (обязательный слой kit):**

```
.cursor/agents|playbooks|skills/gsd-ru|rules
prompt-engineering/   # project-agnostic
.planning/            # STATE, ROADMAP, decisions, ai-expansion
```

---

## 5. Sequence — Strangler Fig (пилот notification)

Согласовано с ADR §2–§3 и skill `microservices-strangler`.

### 5.1 Precondition (Фаза 0 / M1 Hardening)

1. Закрыть утечки `*ent.Client` в service/handler (TD-02) и `GetDB()` (TD-03) в затрагиваемых пакетах.
2. Repository interfaces + outbox publish in-TX.
3. OTel trace-context в NATS headers — **gate до пилота**.
4. Path-filtered CI (бюджет ~3000 мин).
5. Contract-тесты producer↔consumer заготовлены.

### 5.2 Sequence (happy path выноса notification)

```
[M2a] Выделить schema `notification` + DB user GRANT only
[M2b] Dual-write / dual-consume: монолит ещё шлёт локально + NATS
[M2c] Поднять hub-notification (свой Ent/Atlas)
[M2d] Contract tests green
[M2e] Caddy route-flip: /api/notifications* → new service (canary → 100%)
[M2f] Наблюдение ≥2 недели (Go/No-Go композит)
[M2g] Удалить старый код из монолита (ADR-patch), бэкапы N схем
```

### 5.3 Sequence diagram (логика запроса после flip)

```
Client → Caddy → hub-notification
                      ├─ validate JWT (HS256 pilot → JWKS до hub-ai)
                      ├─ own schema only
                      └─ publish/consume NATS
NATS ← hub-core outbox (hub.tracker.*, hub.devflow.*, …)
hub-notification consumer → delivery (in-app / TG / push)
     push URL → ValidateOutboundURL
```

### 5.4 Rollback

- Caddy flip назад на монолит; схема notification сохраняется (данные не теряем).
- Зафиксировать вывод в ADR-follow-up; **не** форсировать следующий вынос.

### 5.5 Порядок выноса (не менять без пользователя)

1. **hub-notification** (пилот)
2. **hub-ai** (после JWKS)
3. **realtime-трио:** hub-ws · hub-messenger · hub-calls (раздельно)
4. tms — условно M6 после Go пилота
5. **tracker — исключён** из горизонта

---

## 6. Взаимодействие и данные

| Паттерн | Правило |
|---------|---------|
| Async | NATS JetStream + outbox; idempotent `event_id` |
| Sync | внутренний REST; gRPC не вводить на старте; >1 hop = неверная граница |
| Read чужого домена | CQRS projection через события |
| Auth inter-service | HS256 (пилот) → EdDSA/RS256 + JWKS до hub-ai |
| Refresh/revoke | централизованно Redis (TZ) |
| Orphans UUID | consumer `*.deleted` + reconcile-job |
| Rate limits | auth в core; доменные per-service; 100/с на Caddy |

---

## 7. TZ-инварианты (таблица для kit)

| Категория | Mandatory (нести в kit) | Forbidden / caution |
|-----------|-------------------------|---------------------|
| Auth | Magic Link default; JWT 15m; refresh Redis 30d; rate 3/10min | OAuth соцсетей; Telegram login; MFA свое |
| SSO | Только после `ADR-AUTH-SSO` implement | Реализация «на глаз» без ADR |
| Org | 1 org / instance; invite; founder один | Multi-org (v5) |
| Data | UUID v4; UTC; `version`; soft-delete 30d | string PK; hard delete без Admin/GDPR |
| API | `{"error","request_id"}`; pagination rules | — |
| Realtime | WS через Fiber (fasthttp); NATS only pub/sub | gorilla/net/http WS; Redis pub/sub |
| Storage | SeaweedFS + aws-sdk-go-v2; pinned version | minio-go; `:latest` |
| Frontend | TanStack Router; Tailwind4+Shadcn className-only | другой роутер/UI kit; `!important` |
| Infra | Compose dev; K3s prod; **Caddy** | nginx; `DEV_MODE=true` в prod |
| Security | webhook signature always; WS RBAC; search user-scope; safedialer | skip signature; open rooms |
| SFU | LiveKit | mediasoup без нового триггера (SEED-001 REJECTED) |
| Split | schema-per-service; strangler | shared tables; big-bang |

**Что ADR меняет в TZ:** deployment = много процессов; HS256→asymmetric+JWKS. Остальное — сохранить.

---

## 8. Deployment skeleton

### 8.1 Окружения

| Env | Compose / orch | Proxy | Notes |
|-----|----------------|-------|-------|
| Dev full | `docker-compose.local.yml` | Caddy :8080 | DEV_MODE ok |
| Dev infra | `deployments/docker-compose.yml` | direct | PG/Mongo/Redis/NATS/S3/LiveKit |
| Test | `docker-compose.test.yml` | — | hubtest |
| Prod | `docker-compose.prod.yml` / K3s | Caddy | secrets via env; no DEV_MODE |
| Monitoring | `docker-compose.monitoring.yml` | — | optional (TD-05) |

### 8.2 Prod target (ADR)

- Одна K3s-нода + local-path PVC — ок для старта SMB.
- Namespaces per service group.
- Caddy ingress: route-flip по path.
- После каждого выноса: **dev compose profile = split** того сервиса.
- Бэкапы: решить на пилоте M2 при N схемах.

### 8.3 CI

- Path-filtered jobs обязательны (M0 ADR).
- Бюджет ~3000 мин/мес — не плодить матрицы.
- Contract tests в CI до flip.

### 8.4 Секреты

- Не копировать в kit: `.env`, livekit prod keys, Resend/Telegram tokens.
- FA-06 rotation — ops, не блокирует архитектурный шаблон.

---

## 9. Урезанный GitHub в архитектуре

Детали — [`06-REDUCED-GITHUB-MODULE.md`](06-REDUCED-GITHUB-MODULE.md).

| Слой | Где | Статус |
|------|-----|--------|
| Process (agents/CI/PR templates) | AI Ops kit | **сейчас** |
| Runtime `devflow` | hub-core | Later M1–M3 |
| Notifications on review | hub-notification | после пилота / events |
| AI review | hub-ai | после JWKS |

MVP kit = процесс + агенты + CI-паттерны, **не** клон GitHub UI.

---

## 10. Слой AI-команды (что переносится обязательно)

```
.cursor/
  agents/          # Hub-specific ИЛИ PE corporate (выбрать один стиль)
  playbooks/       # team-workflow, prompt-contract, …
  skills/gsd-ru/   # RU GSD-lite
  rules/           # style, regression, docs-sync, tz (адаптировать)
prompt-engineering/ # project-agnostic
.planning/          # STATE, ROADMAP, decisions, ai-expansion refs
```

Порядок копирования — [`05-PORTABLE-KIT.md`](05-PORTABLE-KIT.md).
DoD переноса: context-intake → агент по шаблону делегирования → 10-этап verification → STATE ведётся.

---

## 11. Go/No-Go после пилота (композит ADR)

| Метрика | Порог |
|---------|-------|
| Lead-time деплоя домена | ↓ vs монолит |
| p95 API | не хуже +10% |
| Инциденты из-за split | 0 за 2 недели |

Если пользы нет — **остановиться** на модульном монолите + точечные hub-ai/hub-realtime; Фаза 0/1 всё равно ценна.

---

## 12. Риски переносимости

| Risk | Mitigation |
|------|------------|
| Hub-only пути в PE | context-intake; corporate agents без `backend/internal/...` |
| Толстый kit | слои L1 AI-only → L6 full-stack map |
| Чужой стек (не Go/Fiber) | «адаптируй стек, сохрани границы и TZ-таблицу» |
| Попытка shared DB «для скорости» | Явный anti-pattern в kit ADR stub |
| Повтор mediasoup | SEED-001 REJECTED — не поднимать без драйвера |
| Big-bang split в целевом репо | Только strangler checklist skill |

---

## 13. Техдолг, который kit должен помнить (TD)

| ID | Суть | Влияние на split |
|----|------|------------------|
| TD-02/03 | ent.Client / GetDB leaks | Блокер Фазы 0 |
| TD-04 | K3s manifests partial | DevOps backlog |
| TD-05 | Observability optional | Включить до пилота |
| TD-09 | notification/push на Ent | Чинить в пилоте |

---

## 14. Verification (arch-ver)

- [ ] C4 L1/L2 согласованы с ADR catalog юнитов
- [ ] Sequence strangler содержит dual-write, route-flip, rollback, Go/No-Go
- [ ] TZ table полная; нет противоречий mandatory
- [ ] Deployment: Caddy, не nginx; Compose+K3s; path-filtered CI
- [ ] tracker не в списке выноса; LiveKit stay
- [ ] DevFlow/github metadata отнесены к hub-core
- [ ] Portable layers L1–L6 ссылаются на 05; reduced GitHub — на 06

**Handoff →** `arch-ver` → `docs-writer` (при sync architecture.md) → `devops-engineer` (K3s/Caddy when implementing).
