# Executive Report — Hub Platform Strategy 2026

> **Дата:** 2026-06-21 · **Audience:** Product / CTO / Stakeholders  
> **Commits:** v2.0.2 shipped (`90042eb`), analytics (`1ff00f1`)

---

## TL;DR

Hub сегодня — **лучший self-hosted dev suite** (tracker+messenger+TMS+calls) для 10–50 человек.  
Следующий скачок — **v3 AI Provider Hub** (любая LLM по API без VPN) + **v4 corporate super-app** (intranet, calendar, CRM-lite, HR-lite).  
Бухгалтерия — **только интеграция с 1C**, не встроенный модуль.

---

## 1. Что изучили

| Источник | Insight |
|----------|---------|
| **Tinkoff Mighty / MyTinkoff** | Единое app: новости, HR, документы, коммуникации — **удержание через intranet** |
| **Bitrix24** | CRM+tasks+chat в одном; слаб в dev/TMS depth |
| **Lark** | Super-app + no-code automation; cloud-only |
| **Microsoft Viva** | Employee experience поверх Teams — learning, news, OKR |
| **Linear / Jira** | Automation + AI — table stakes для tracker |
| **Allure TestOps** | Flaky ML — QA retention |

---

## 2. Что хотят пользователи Hub (7 персон)

| Приоритет | Need | Module |
|-----------|------|--------|
| P0 | Workflow automation | PROD-01 (v2.6) |
| P0 | Reliable E2E / WCAG | v2.5 |
| P0 | **AI в одном месте, RU models** | AI Hub v3 |
| P1 | Flaky tests | PROD-02 |
| P1 | Corp news + directory | Intranet v4 |
| P1 | Unified calendar | Calendar v4 |
| P2 | CRM для клиентов аутсорса | CRM-lite v4 |
| P2 | Leave requests | HR-lite v4 |
| P3 | 1C sync | Integrations |
| ❌ | Full accounting in Hub | Out — 1C connector |

---

## 3. Предложение: новые модули

| ID | Module | Microservice | Milestone | Effort |
|----|--------|--------------|-----------|--------|
| MOD-AI | AI Provider Hub | hub-ai-gateway | v3.0–v3.5 | XL |
| MOD-INTRA | Intranet (news, kudos, directory) | hub-intranet | v4.0 | L |
| MOD-CAL | Corporate calendar | hub-core → split later | v4.0 | M |
| MOD-CRM | CRM-lite (contacts, deals) | hub-core | v4.0 | L |
| MOD-HR | HR-lite (leave, org chart) | hub-core | v4.0 | M |
| MOD-INT | 1C / Bitrix connectors | hub-integrations | v4.5 | L |

**Existing modules to deepen:** tracker automation, TMS analytics, messenger↔issue sync.

---

## 4. Killer feature: AI Provider Hub

**Проблема:** ChatGPT/Claude требуют VPN; команда хочет **выбрать модель в UI**.

**Решение:**
- Admin подключает DeepSeek, GigaChat, YandexGPT, OpenRouter, Ollama
- Пользователь выбирает модель в AI Panel / `@hub-ai`
- OpenAI-compatible API `POST /api/ai/v1/chat/completions` для внутренних tools
- Keys server-side, audit без content, rate limits

**Spec:** `.planning/specs/AI-PROVIDER-HUB-SPEC.md`

---

## 5. Security improvements (roadmap)

| ID | Improvement | Priority |
|----|-------------|----------|
| SEC-R1 | PG-03 secret rotation execute | P0 |
| SEC-R2 | AI gateway SSRF + key encryption | P0 (with AI) |
| SEC-R3 | mTLS service-to-service (v4.5) | P2 |
| SEC-R4 | CSP report-uri endpoint | P2 |
| SEC-R5 | SBOM in CI (syft) | P2 |
| SEC-R6 | CodeQL enabled (pr-guard extended) | P2 |
| SEC-R7 | Audit retention → S3 cold storage | P2 |
| SEC-R8 | AI prompt injection guardrails | P1 |

---

## 6. Infra / DevOps updates (this session)

| Change | Path |
|--------|------|
| `.gitignore` | `.cursor/`, `.claude/`, agent artifacts |
| GitHub | dependabot, PR template, issue templates, pr-guard, labeler |
| GitLab | `.gitlab-ci.yml` mirror |

**Note:** `.cursor/` already in git history — для новых клонов rules копировать локально или `git rm -r --cached .cursor` при следующем commit.

---

## 7. Recommended execution order

```
Q3 2026  v2.5  E2E + WCAG + RHF
Q4 2026  v2.6  PROD-01 automation + flaky + webhooks
Q1 2027  v3.0  AI-1..3 (DeepSeek, Ollama, GigaChat, @hub-ai)
Q2 2027  v3.5  OpenRouter + model picker + dev API
Q3 2027  v4.0  Intranet + Calendar
Q4 2027  v4.0  CRM-lite + HR-lite
2028     v4.5  Microservices split + 1C connector
```

---

## 8. Score projection

| Milestone | Product score | Notes |
|-----------|---------------|-------|
| v2.0.2 (now) | 93/100 | Dev suite ship-ready |
| v2.6 | 95/100 | Automation |
| v3.0 | 97/100 | AI differentiation |
| v4.0 | 98/100 | Corporate super-app niche |

---

## 9. Document index

| Document | Purpose |
|----------|---------|
| [../specs/PLATFORM-VISION-v4.md](../specs/PLATFORM-VISION-v4.md) | Full vision |
| [MICROSERVICES-ROADMAP.md](./MICROSERVICES-ROADMAP.md) | Service split plan |
| [../specs/AI-PROVIDER-HUB-SPEC.md](../specs/AI-PROVIDER-HUB-SPEC.md) | AI technical spec |
| [FEATURE-BACKLOG-2026.md](./FEATURE-BACKLOG-2026.md) | FB-01..FB-28 + MOD-* |
| [docs/competitive_analysis_2026.md](../../docs/competitive_analysis_2026.md) | Competitors |
| [docs/user_needs_analysis_2026.md](../../docs/user_needs_analysis_2026.md) | Personas |

---

*Prepared inline after research subagents API-limited; web sources: Mighty, Bitrix24, Lark, Viva, Linear/Jira, Allure TestOps.*
