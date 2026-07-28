# Hub Platform Vision v4 — Corporate Super-App

> **Codename:** Hub One · **Horizon:** 2026–2028  
> **Positioning:** Self-hosted «Майти / Bitrix24 / Lark» для **IT-команд и mid-market**, с фокусом на **разработку + QA + коммуникации**, расширяемый до **corp intranet + AI hub**.

---

## 1. Эволюция продукта

| Версия | Фокус | Аналог |
|--------|-------|--------|
| **v1–v2** ✅ | Dev suite: tracker + messenger + TMS + calls | Jira + Slack + TestRail + Meet |
| **v2.5–v2.6** | Production + automation + flaky TMS | Linear + Jira Automation lite |
| **v3.0** | AI Chat Hub (RU providers + OpenAI-compatible) | Slack AI + Notion AI |
| **v4.0** | **Corporate super-app** | Tinkoff Mighty + Bitrix24 + Viva |

**Принцип:** один инстанс = одна организация (single-tenant). Не SaaS billing, не multi-tenant — но **полный контроль** для employer.

---

## 2. Целевая архитектура (логические домены)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Hub Portal (React SPA)                          │
│  Dev · QA · Comms · Calendar · Intranet · CRM-lite · AI · Admin        │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │ HTTPS / WSS
┌───────────────────────────────▼─────────────────────────────────────────┐
│                    API Gateway (Fiber / future Kong)                     │
│         JWT · RBAC · rate limit · audit · request_id                    │
└───┬─────────┬─────────┬─────────┬─────────┬─────────┬─────────┬─────────┘
    │         │         │         │         │         │         │
    ▼         ▼         ▼         ▼         ▼         ▼         ▼
┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐
│ CORE  │ │ COMMS │ │QUALITY│ │ MEDIA │ │INTRANET│ │  AI   │ │  CRM  │
│monolith│ │module │ │  TMS  │ │calls  │ │ news  │ │gateway│ │ lite  │
│tracker │ │messenger│      │ │storage│ │people │ │providers│     │
│project │ │notif  │ │       │ │       │ │calendar│ │       │ │       │
└───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘
    │         │         │         │         │         │         │
    └─────────┴─────────┴──── NATS JetStream ─────────┴─────────┘
                              │
                    PostgreSQL · MongoDB · Redis · SeaweedFS
```

**Фаза 1 (сейчас):** модульный монолит — все домены в `backend/internal/`.  
**Фаза 2 (v4):** выделение **hub-ai-gateway**, **hub-intranet** при нагрузке >500 concurrent WS.  
**Фаза 3:** optional **hub-integrations** (1C, Bitrix, HH) — connector pattern, не встроенная бухгалтерия.

---

## 3. Новые модули v4 (предложение)

### 3.1 Intranet / «Майти-lite» (MOD-INTRA)

| Фича | Описание | Mighty / Bitrix аналог |
|------|----------|------------------------|
| News feed | Лента новостей компании, pin, reactions | MyTinkoff лента |
| Org directory | Дерево отделов, профили, день рождения | HR directory |
| Kudos / thanks | «Спасибо» коллеге → лента + уведомление | Random Coffee / kudos |
| Polls company-wide | Опросы (reuse messenger polls) | Surveys |
| Internal marketplace | Доска объявлений (moderated) | T2 барахолка |
| Onboarding paths | Чеклисты для новичков beyond dev | HR onboarding |

**Не делаем:** payroll, больничные в 1C — только **заявки** + webhook в 1C/ZUP.

### 3.2 Calendar (MOD-CAL)

| Фича | Описание |
|------|----------|
| Unified calendar | Спринты, созвоны, отпуска (read), deadlines |
| iCal export | Subscribe in Outlook/Google (read-only feed) |
| Room booking | Переговорки (optional) |
| Team availability | Из presence + calendar blocks |

**Связь:** расширение `modules/calls/call-calendar.tsx` → first-class `/calendar` route.

### 3.3 CRM-lite (MOD-CRM)

| Фича | In scope | Out of scope |
|------|----------|--------------|
| Contacts / companies | ✅ | — |
| Deals pipeline (kanban) | ✅ simple | Salesforce parity |
| Link deal → project | ✅ | — |
| Invoicing / accounting | ❌ | **1C connector** |
| Email campaigns | ❌ | no email notif rule |

**Use case:** IT outsourcing / product company — клиенты и контракты рядом с проектами.

### 3.4 HR-lite (MOD-HR)

| Фича | Описание |
|------|----------|
| Leave requests | Submit → approver workflow (reuse PROD-01) |
| Org chart | Visual from `department` field |
| Employee handbook | Wiki pages (reuse help + pages) |
| 1C sync | Export approved leave via API |

**Не делаем:** расчёт зарплаты, НДФЛ — интеграция only.

### 3.5 AI Provider Hub (MOD-AI) ⭐

**Killer feature без VPN:** единый **OpenAI-compatible proxy** + native adapters.

| Provider | API | RU без VPN | Use case |
|----------|-----|------------|----------|
| DeepSeek | OpenAI-compatible | ✅ | Code, cheap |
| GigaChat | OAuth | ✅ | Enterprise RU |
| YandexGPT | IAM | ✅ | RU compliance |
| OpenRouter | Gateway | ⚠️ | 100+ models one key |
| Ollama | Self-host | ✅ | Air-gap |
| Anthropic / OpenAI | Official | ⚠️ VPN/proxy | Admin choice |

**UX:**
- Admin: Settings → AI Providers → add key, enable models
- User: Profile → default model; `@hub-ai` or AI Panel (Ctrl+Shift+A)
- **Model picker:** dropdown «GPT-4o / DeepSeek-V3 / Claude / …» — whatever admin enabled
- Developer API: `POST /api/ai/v1/chat/completions` (OpenAI format) for internal tools

**Security (mandatory):**
- Keys encrypted at rest (AES-256, `platform/crypto`)
- `ValidateOutboundURL` on provider base URL
- Rate limit per user/day; audit log **metadata only** (tokens, model, no prompt body)
- RBAC: `PermAIUse`, `PermAIManageProviders`

Spec: [AI-PROVIDER-HUB-SPEC.md](./AI-PROVIDER-HUB-SPEC.md)

### 3.6 Integrations hub (MOD-INT)

| Connector | Direction | Priority |
|-----------|-----------|----------|
| 1C:Enterprise | Leave, org sync | P2 |
| Bitrix24 | Tasks/contacts import | P3 |
| HH.ru | Vacancies (read) | P3 |
| Google Calendar | iCal sync | P2 |
| LDAP read-only | **Future 999.x** — not v4 (scope constraint) |

---

## 4. Конкурентное позиционирование v4

| | Hub v4 | Mighty | Bitrix24 | Lark |
|---|--------|--------|----------|------|
| Dev tracker + TMS | ⭐⭐⭐ | ❌ | ⭐ | ⭐ |
| Messenger + calls | ⭐⭐⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| Self-host / data control | ⭐⭐⭐ | ❌ | ⭐ optional | ❌ |
| Corp news + directory | ⭐ (v4) | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| CRM + accounting | ⭐ lite + 1C | ⭐ | ⭐⭐⭐ | ⭐ |
| AI multi-provider RU | ⭐⭐⭐ (v3–v4) | ⭐ | ⭐ | ⭐⭐ |
| TCO 50 users | **Low** | N/A internal | Medium | Medium |

**Стратегия:** не «убить Bitrix24», а **заменить связку Jira+Slack+TestOps+Meet** и **добавить intranet** для удержания сотрудников в одном приложении.

---

## 5. Roadmap milestones

| Milestone | Duration | Deliverables |
|-----------|----------|--------------|
| **v2.5** | Q3 2026 | E2E, WCAG, RHF, PG-03 |
| **v2.6** | Q4 2026 | PROD-01..04 automation, flaky, webhooks |
| **v3.0** | Q1–Q2 2027 | AI gateway MVP + @hub-ai + 3 RU providers |
| **v3.5** | Q3 2027 | OpenRouter + model picker + dev API |
| **v4.0** | Q4 2027–2028 | Intranet + Calendar + CRM-lite + HR-lite |
| **v4.5** | 2028 | Selective microservices split if scale demands |

---

## 6. Success metrics v4

| KPI | Target |
|-----|--------|
| DAU/MAU ratio | >40% |
| Modules used per user/week | ≥3 |
| AI requests/user/day (median) | 5–20 |
| Leave request SLA | <24h approve |
| News feed engagement | >60% monthly readers |
| Still 0 CRITICAL security | Always |

---

## 7. Constraints (не negotiable без ADR)

Из `.cursor/rules/60-tz-constraints.mdc`:

- ❌ Multi-tenant SaaS · SSO/LDAP v1–v4 · Email notif · Native apps · Gantt · Webcams
- ✅ Magic Link + Telegram · Self-host · NATS · SeaweedFS · Fiber WS

**Accounting:** только через **1C connector**, не built-in ledger.

---

*См. также: [MICROSERVICES-ROADMAP.md](../../archive/roadmaps/MICROSERVICES-ROADMAP.md) · [FEATURE-BACKLOG-2026.md](../../archive/roadmaps/FEATURE-BACKLOG-2026.md) · [EXECUTIVE-REPORT-2026.md](../../archive/roadmaps/EXECUTIVE-REPORT-2026.md)*
