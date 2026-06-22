# Microservices Roadmap — Hub

> **Principle:** Evolve from modular monolith → **selective extraction** when metrics justify ops cost.  
> **Not:** microservices for microservices' sake.

---

## 1. Current state (v2.0.2)

| Property | Value |
|----------|-------|
| Style | Modular monolith |
| Entry | `cmd/server/main.go` |
| Modules | 20 domain + platform |
| Deploy | Single Docker image / K3s Deployment |
| Messaging | NATS JetStream (already async boundary) |

**When to split:** p95 latency >500ms on hot path, WS connections >2k, or AI GPU isolation needed.

---

## 2. Target services (v4)

| Service | Responsibility | Extract when | Storage |
|---------|----------------|--------------|---------|
| **hub-core** | auth, rbac, project, tracker, sprint, search | Default monolith | PG |
| **hub-comms** | messenger, notification, ws hub | WS >2k conn | PG + Mongo + Redis |
| **hub-quality** | tms, imex test imports | CI ingest >100/min | PG + Mongo |
| **hub-media** | calls, storage, recordings | Egress bandwidth | PG + S3 |
| **hub-ai-gateway** | AI provider proxy, usage | GPU/network isolation | PG + Redis |
| **hub-intranet** | news, directory, kudos, marketplace | v4 module size | PG |
| **hub-integrations** | 1C, Bitrix connectors | ≥3 connectors | PG + queue |

**Shared:** PostgreSQL (schema per service or prefix), NATS, Redis, Caddy ingress.

---

## 3. Communication patterns

```
                    ┌─────────────┐
  Sync REST         │   Caddy     │
  (user requests)   └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
    hub-core          hub-comms        hub-ai-gateway
         │                 │                 │
         └──────── NATS (domain events) ─────┘
              issue.created, message.posted,
              ai.usage.recorded, news.published
```

| Pattern | Use |
|---------|-----|
| Sync REST | User-facing CRUD via gateway |
| NATS events | Cross-module side effects (already used) |
| Outbox | Transactional publish (already used) |
| gRPC internal | Optional v4.5 between services |

---

## 4. Extraction order (recommended)

| Step | Service | Trigger | Risk |
|------|---------|---------|------|
| 1 | **hub-ai-gateway** | v3.0 AI launch | Low — stateless proxy |
| 2 | **hub-integrations** | First 1C connector | Low — async workers |
| 3 | **hub-intranet** | v4 news feed | Medium — new domain |
| 4 | **hub-comms** | Scale WS | High — split brain |
| 5 | **hub-quality** | TMS load | Medium |
| 6 | **hub-media** | Recording volume | Medium |

---

## 5. Deployment (K3s)

```yaml
# deployments/k3s/hub-core.yaml — existing
# deployments/k3s/hub-ai-gateway.yaml — new v3
# deployments/k3s/hub-intranet.yaml — new v4
```

**Compose dev:** single `docker compose` with optional `--profile ai` for Ollama sidecar.

---

## 6. Data ownership

| Service | Owns tables |
|---------|-------------|
| hub-core | users, projects, issues, workflows, sprints |
| hub-comms | channels, messages (Mongo), notifications |
| hub-quality | test_cases, runs, allure_reports |
| hub-ai-gateway | ai_provider, ai_usage_daily |
| hub-intranet | news_posts, kudos, marketplace_listings |

**Rule:** No cross-service SQL joins — use API or read models (search index).

---

## 7. Security per service

- Same JWT validation middleware (shared `platform/auth`)
- Service-to-service: mTLS or NATS credentials (v4.5)
- AI gateway: no DB access to messages content

---

*Parent: [PLATFORM-VISION-v4.md](./PLATFORM-VISION-v4.md)*
