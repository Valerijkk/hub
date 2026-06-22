# Hub — Конкурентный аудит и backlog фич

> **Обновлено:** 2026-06-23  
> **Связанные документы:** [competitive_analysis_2026.md](./competitive_analysis_2026.md) · [user_needs_analysis_2026.md](./user_needs_analysis_2026.md) · [.planning/PLATFORM-VISION-v4.md](../.planning/PLATFORM-VISION-v4.md)  
> **Цель:** best practices от лидеров рынка + приоритизированный backlog для Hub v2.5–v4.5+.

---

## 1. Текущая позиция Hub (2026-06-23)

Hub — **self-hosted corporate super-app** для команд 10–50 человек: dev-suite (трекер, мессенджер, TMS, звонки) **+** корпоративные модули v4 (интранет, календарь, CRM-lite, HR, учёт, no-code базы, AI hub).

| Показатель | Значение |
|------------|----------|
| Milestone v4 Super-app | ✅ Phases 39–58 (2026-06-25) |
| Milestone v2.5 Trust / v2.6 Automate / v3.0 AI | ✅ Phases 15–19 |
| Wave C final audit FA-01–FA-12 | ✅ (кроме FA-06 manual) |
| Покрытие backend (`internal/`) | **83.0%** (Wave C) |
| Security-critical пакеты ≥80% | **6/6** ✅ |
| REST API endpoints | **286+** |
| Product score (AUDIT-SCORECARD) | **93/100** |
| Ветка | `master` · HEAD `bbdaf16` |
| CI | ✅ green (coverage-gate, e2e-full 20 specs, lighthouse artifact) |

**Стратегия:** не конкурировать с Jira в enterprise workflow; побеждать по **TCO + интеграция + self-host + RU-friendly AI** для команд с data residency.

---

## 2. Матрица конкурентов

| Категория | Конкуренты | Сильные стороны | Gap в Hub (остаточный) |
|-----------|------------|-----------------|------------------------|
| Task Tracker | Jira, Linear, Asana | JQL deep, cycles, roadmap, automations | Automation UI есть (v2.6); triage inbox — polish |
| Messenger | Slack, Discord, Teams | Threads, bots, marketplace | Bot framework ограничен (Telegram); AI v3.0 ✅ |
| TMS | TestRail, Allure TestOps | Flaky ML, CI gates, run diff | Allure import ✅; flaky detection — partial |
| Calls | Zoom, Meet, Discord | Breakout, noise cancel, webcams | LiveKit voice+screen ✅; webcams — Phase 16 IN |
| Dashboards | Grafana, Datadog | Custom panels, alerting | Статичные виджеты; OBS stack ✅ (Prometheus/Loki) |
| Corp super-app | Mighty, Bitrix24, Lark | News, HR, CRM, calendar | **v4 закрыто** — intranet, HR, CRM, bases |
| Admin/Observability | Datadog, Sentry, Auth0 | APM, audit, ban | ✅ Audit UI + HTTP timing + OTel + Alertmanager |

---

## 3. Best Practices — статус внедрения

### Security & Admin (P0 — закрыто в v2.0.1–v2.5)

| Практика | Источник | Статус Hub |
|----------|----------|------------|
| Centralized audit log с фильтрами | Auth0, AWS CloudTrail | ✅ `/admin/audit` + NATS consumer |
| HTTP request timing в audit | Datadog APM lite | ✅ `HTTPAuditMiddleware` |
| User ban с причиной + session revoke | Discord, Slack | ✅ `POST /admin/users/:id/ban` |
| Security headers (CSP, HSTS) | OWASP | ✅ `SecurityHeaders` middleware |
| Refresh token hash at rest | Auth0 | ✅ SHA-256 в Redis/PG |
| Single-flight refresh (FE) | Spotify, Auth0 SDK | ✅ `client.ts` mutex |
| Multi-tab auth sync | Slack web | ✅ BroadcastChannel |
| Error boundaries per module | React docs | ✅ `ErrorBoundary` в AppShell |
| Secret rotation runbook | Production ops | ⚠️ PG-03 — manual ops |

### Developer Experience (P1 — v2.6 закрыто)

| Практика | Источник | Статус Hub |
|----------|----------|------------|
| OpenTelemetry traces | Grafana Tempo | ✅ OTLP export (Phase 32) |
| Structured slog + request_id | Google SRE | ✅ |
| `Timed()` wrapper для сервисов | Internal SRE | ✅ `platform/observability/timing.go` |
| Prometheus `/metrics` | Industry standard | ✅ Phase 31 |
| Loki + Promtail + Alertmanager | Grafana stack | ✅ Phases 33–36 |
| Feature flags | LaunchDarkly, Linear | 📋 Redis-backed — backlog |
| API changelog / deprecation headers | Stripe | 📋 backlog |

### UX (P1 — v2.5 закрыто частично)

| Практика | Источник | Статус Hub |
|----------|----------|------------|
| Command palette (⌘K) | Linear, Notion | ✅ `command-palette.tsx` (2026-06-20) |
| Optimistic UI updates | Linear | ⚠️ tracker ✅; messenger reactions — partial |
| Keyboard shortcuts overlay | GitHub | ✅ `/help` route + skip-link |
| Empty states с CTA | Notion | ⚠️ partial |
| Skeleton loaders | Slack | ⚠️ partial — заменить «Загрузка...» |
| RHF+Zod forms | Industry standard | ⚠️ onboarding ✅; issue-form — partial (UX-1) |

### Collaboration & AI (P2 — v3.0 закрыто)

| Практика | Источник | Статус Hub |
|----------|----------|------------|
| AI summarize / @hub-ai | Slack AI, Notion AI | ✅ AI gateway BYOK + Ollama (Phase 17) |
| Smart notifications (digest) | Slack | ⚠️ quiet hours ✅; digest — backlog |
| Issue ↔ message linking | Linear + Slack sync | ⚠️ deep links ✅; bidirectional — partial |
| Presence indicators | Discord | ✅ WS presence |
| Workflow automation lite | Jira Automation | ✅ Phase 16 PROD-01 |

---

## 4. Feature Backlog (приоритизированный)

### Must Have (production trust — в основном закрыто)

1. ~~12 CRITICAL security blockers~~ — ✅ v2.0.1
2. ~~Test coverage ≥50% overall / ≥80% security~~ — ✅ **83.0%** / **6/6 ≥80%**
3. Secret rotation в git history — ⏳ **FA-06 manual** — единственный prod-блокер
4. ~~Vitest smoke tests auth + admin~~ — ✅ 99/99 tests
5. ~~E2E critical paths~~ — ✅ 20 Playwright specs
6. ~~Test framework migration~~ — ✅ 488 testframework / 0 internal
7. ~~Squash PR v2.5-trust~~ — ✅ merged PR #1 2026-06-22

### Should Have (v4.5 / Wave C audit — в основном закрыто)

1. ~~Test framework migration~~ — ✅ TF-M01–M10 COMPLETE
2. ~~tms/project micro-gap~~ — ✅ FA-11 (82.0%)
3. **Webhook delivery log polish** — PROD-04 базово есть; admin UI depth
4. **Flaky test detection** — PROD-02 partial
5. **Notification digest** — backlog
6. **Lighthouse ≥90 blocking + WCAG sign-off** — artifact CI ✅ (PR #27); full audit pending

### Could Have (v4.5+ / 2027)

1. Custom fields builder UI — schema есть, UI нет
2. Timeline roadmap / initiatives view
3. Triage inbox UX polish
4. Mobile PWA offline depth
5. Feature flags (Redis-backed beta)
6. Accounting v4.5+ (расширенный P&L, интеграции)

### Won't Have (per ADR / `.cursor/rules/60-tz-constraints.mdc`)

- Native mobile apps (PWA достаточно)
- SSO/LDAP/OAuth провайдеры (до v5 ADR)
- Multi-tenant SaaS billing
- Multi-org на одном инстансе (v5)

**IN SCOPE by phase (PO 2026-06-21):** Gantt (Phase 16), webcams in calls (Phase 16), hub-mail marketing (Phase 15.2/16), full JQL automation (Phase 17), no-code bases + accounting (v4).

---

## 5. Observability Stack (prod — реализовано)

```
Hub API → slog (JSON) → Loki (Promtail)
       → audit_log table → Admin UI
       → NATS events → audit consumer
       → OTel OTLP → Tempo/Jaeger (optional)
       → Prometheus /metrics → Grafana
       → Alertmanager → webhook → audit log
```

**DEV_MODE** остаётся для локального Docker: magic link в логах, relaxed CORS, полный HTTP audit.

---

## 6. Метрики успеха

| KPI | Target | Текущее (2026-06-23) |
|-----|--------|----------------------|
| Statement coverage (`internal/`) | ≥82% | **83.0%** ✅ |
| Security packages ≥80% | 6/6 | **6/6** ✅ |
| Mean API latency (p95) | < 200ms | Измеряется через `duration_ms` в audit |
| Audit coverage admin + auth | 100% | ✅ |
| Security findings CRITICAL open | 0 | ✅ (12/12 fixed) |
| Product score | ≥80 | **93/100** ✅ |

---

## 7. Следующие шаги

1. **FA-06 PG-03** — ротация секретов на staging/prod (`bash scripts/fa06.sh`)
2. **Release v2.5.0** — `release-gate` → git tag после FA-06 + CTO
3. **NFR-16/17** — WCAG full sign-off · Lighthouse blocking gate
4. **SSH deploy** — GitHub Secrets + firewall :22

Полный аудит: `.planning/AUDIT-2026-06-23.md`

---

*См. также: [FEATURE-BACKLOG-2026.md](../.planning/FEATURE-BACKLOG-2026.md) · [COVERAGE.md](../.planning/COVERAGE.md)*
