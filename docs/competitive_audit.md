# Hub — Competitive Audit & Feature Backlog

> Обновлено: 2026-06-21. См. также [competitive_analysis_2026.md](../docs/competitive_analysis_2026.md) (Mighty, Bitrix, Lark) и [PLATFORM-VISION-v4.md](./PLATFORM-VISION-v4.md).
> Цель: best practices от лидеров рынка + backlog нововведений для Hub v2.5–v3.0.

---

## 1. Матрица конкурентов

| Категория | Конкуренты | Сильные стороны | Gap в Hub |
|-----------|------------|-----------------|-----------|
| Task Tracker | Jira, Linear, Asana | JQL, cycles, roadmap, automations | Нет workflow automation UI |
| Messenger | Slack, Discord, Teams | Threads, reactions, search, bots | Нет bot framework |
| TMS | TestRail, Allure TestOps | Test plans, flaky detection, CI gates | Allure import есть, flaky — нет |
| Calls | Zoom, Meet, Discord | Breakout rooms, noise cancel | LiveKit screen share есть |
| Dashboards | Grafana, Datadog | Custom panels, alerting | Статичные виджеты |
| Admin/Observability | Datadog, Sentry, Auth0 | APM, audit trail, user ban | **Частично закрыто** (audit UI + HTTP log) |

---

## 2. Best Practices (внедрить)

### Security & Admin (P0 — v2.0.1/v2.5)

| Практика | Источник | Статус Hub |
|----------|----------|------------|
| Centralized audit log с фильтрами | Auth0, AWS CloudTrail | ✅ `/admin/audit` + NATS consumer |
| HTTP request timing в audit | Datadog APM lite | ✅ middleware `HTTPAuditMiddleware` |
| User ban с причиной + session revoke | Discord, Slack | ✅ `POST /admin/users/:id/ban` |
| Security headers (CSP, HSTS) | OWASP | ✅ `SecurityHeaders` middleware |
| Refresh token hash at rest | Auth0 | ✅ SHA-256 в Redis/PG |
| Single-flight refresh (FE) | Spotify, Auth0 SDK | ✅ `client.ts` mutex |
| Multi-tab auth sync | Slack web | ✅ BroadcastChannel |
| Error boundaries per module | React docs | ✅ `ErrorBoundary` в AppShell |

### Developer Experience (P1 — v2.6)

| Практика | Источник | Рекомендация |
|----------|----------|--------------|
| OpenTelemetry traces | Grafana Tempo | Добавить OTel exporter в `observability/` |
| Structured slog + request_id | Google SRE | ✅ уже есть |
| `Timed()` wrapper для сервисов | Internal SRE | ✅ `platform/observability/timing.go` |
| Feature flags | LaunchDarkly, Linear | Redis-backed flags для beta |
| API changelog / deprecation headers | Stripe | `Sunset` header на deprecated routes |

### UX (P1 — v2.5 Phase 17)

| Практика | Источник | Рекомендация |
|----------|----------|--------------|
| Command palette (⌘K) | Linear, Notion | Глобальный поиск + actions |
| Optimistic UI updates | Linear | Расширить на messenger reactions |
| Keyboard shortcuts overlay | GitHub | `?` hotkey help modal |
| Empty states с CTA | Notion | Шаблоны для новых проектов |
| Skeleton loaders | Slack | Заменить "Загрузка..." в таблицах |

### Collaboration (P2 — v3.0)

| Практика | Источник | Рекомендация |
|----------|----------|--------------|
| AI summarize thread | Slack AI, Notion AI | M3-001..M3-030 roadmap |
| Smart notifications (digest) | Slack | Quiet hours есть — добавить digest |
| Issue ↔ message linking | Linear + Slack sync | Deep links есть, bidirectional sync — нет |
| Presence indicators | Discord | WS presence частично |

---

## 3. Feature Backlog (приоритизированный)

### Must Have (до production)

1. ~~12 CRITICAL security blockers~~ — закрыты в v2.0.1
2. Test coverage 50% overall / 80% security-critical (Phase 18)
3. Secret rotation в git history (manual ops)
4. Vitest smoke tests для auth + admin audit

### Should Have (v2.5)

1. **Command palette** — глобальный поиск (уже есть search module, нужен UI)
2. **Audit log retention policy** — 90 дней hot, archive в S3
3. **Slow query dashboard** — топ N медленных `duration_ms` из audit
4. **Rate limit visibility** — admin видит кто упёрся в лимит
5. **Webhook delivery log** — для GitHub/GitLab integrations

### Could Have (v2.6–v3.0)

1. Workflow automation (if X then Y) — как Jira Automation lite
2. Custom fields builder UI — schema есть, UI нет
3. Flaky test detection в TMS
4. Recording retention dashboard
5. Mobile PWA polish (уже web-only по scope)
6. AI chat hub (MILESTONE-3)

### Won't Have (out of scope per PROJECT.md)

- Native mobile apps
- SSO/LDAP
- Multi-tenant SaaS billing
- Video calls with webcams (voice only)
- Gantt charts

---

## 4. Observability Stack (рекомендация для prod)

```
Hub API → slog (JSON) → Loki/ELK
       → audit_log table → Admin UI
       → NATS events → audit consumer
       → (future) OTel → Tempo/Jaeger
```

**DEV_MODE** остаётся для локального Docker: magic link в логах, relaxed CORS, полный HTTP audit.

---

## 5. Метрики успеха

| KPI | Target | Как измерять |
|-----|--------|--------------|
| Mean API latency (p95) | < 200ms | `duration_ms` в audit, entity_type=http |
| Audit coverage | 100% admin + auth + errors | COUNT by action |
| Time to ban abusive user | < 30s | Admin UI flow |
| Security findings open | 0 CRITICAL | `docs/security_guidelines.md` |

---

## 6. Следующие шаги

1. ~~Phase 17: command palette~~ → **Ctrl+K UI shipped** 2026-06-20 (`command-palette.tsx`); extend with actions (create issue, navigate modules)
2. Phase 18: E2E-06 Ctrl+K Playwright + a11y audit (target WCAG AA)
3. OTel spike в `platform/observability/`
4. UX backlog: UserPicker, inline validation, Kanban keyboard nav
