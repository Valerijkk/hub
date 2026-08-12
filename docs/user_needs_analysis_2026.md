# Hub — Анализ продукта и потребностей пользователей 2026

> **Текущая product-truth (founder, 2026-08-12):** Hub остаётся
> незавершённым MVP. Ни один продуктовый модуль не принят как `complete`,
> `stable` или `production-ready`. Метки `SHIPPED`/`COMPLETE`, milestone/audit
> `PASS`, зрелость и галочки в датированных срезах ниже являются только
> историческим planning/technical evidence; они не подтверждают готовность
> продукта, модуля или production. PG-03/FA-06 — отдельный manual ops-блокер
> production cutover, а не индикатор готовности MVP.
>
> **Обновлено:** 2026-08-12; исходный capability snapshot датирован 2026-06-21.
> **Основано на:** `docs/vision_and_scope.md`, `.planning/PROJECT.md`, `AUDIT-SCORECARD` / `AGENTS.md` (**89/100**), `competitive_analysis_2026.md`, `.planning/archive/tasks/TASKS-PHASES-121-FINAL-AUDIT.md`.

---

## 1. Текущая product-truth и исторический capability snapshot

### 1.1 Суть продукта

**Hub** проектируется как self-hosted **corporate super-app** для команд
разработки, QA и сотрудников компании (10–50 человек). Таблицы ниже фиксируют
наличие заявленных code/UI surfaces, но не приёмку их модулей.

#### Dev-suite (v1.0–v2.0)

| Модуль | Зрелость | Ключевые возможности |
|--------|----------|----------------------|
| **Auth** | Не принят | Magic Link + Telegram bind, JWT 15m, refresh 30d, org + founder |
| **Projects + RBAC** | Не принят | Шаблоны, invite, custom roles, workflow, teams |
| **Tracker** | Не принят | 9 типов задач, Kanban, Scrum, JQL, automation lite, Git links |
| **Messenger** | Не принят | Каналы, треды, reactions, polls, voice msgs, #hub-ci, @hub-ai |
| **TMS** | Не принят | Classic/BDD, runs, Allure CI, traceability, reports, env compare |
| **Calls** | Не принят | LiveKit voice/video (`CallControls` camera), screen share, recordings, scheduled |
| **Dashboards** | Не принят | Personal/project/admin widgets, burndown, velocity, health |
| **Search** | Не принят | Ctrl+K, Meilisearch, user-scope API, saved filters |
| **Integrations** | Не принят | GitHub, GitLab, Jira import, webhooks, hub-cli, 1C connector |

#### Corporate super-app (v4, Phases 39–58)

| Модуль | Зрелость | Ключевые возможности |
|--------|----------|----------------------|
| **Intranet** | Не принят | News feed, kudos, Random Coffee |
| **Org directory** | Не принят | MOD-02, Ctrl+K people search |
| **Calendar** | Не принят | Unified events, iCal export |
| **CRM-lite** | Не принят | Contacts, deals kanban, won→invoice hook |
| **HR-lite** | Не принят | Leave requests, calendar sync, 1C webhook |
| **Marketplace** | Не принят | Classifieds, mod queue |
| **Status page** | Не принят | Public `/status` |
| **Bases (no-code)** | Не принят | Lark-style tables `/bases` |
| **Accounting MVP** | Не принят | Ledger, journal, invoices, P&L |
| **AI Hub** | Не принят | BYOK + Ollama, composer suggest, external AI consent |

#### Platform & Admin

| Модуль | Зрелость | Ключевые возможности |
|--------|----------|----------------------|
| **Admin** | Не принят | Audit log, trash 30d, backups, limits, ban, GDPR |
| **Observability** | Не принят | Prometheus, Loki, OTel, Alertmanager → audit |
| **Notifications** | Не принят | In-app + Web Push + Telegram, quiet hours |
| **API Gateway** | Не принят | API keys, webhook subscriptions, delivery history |

### 1.2 Исторический technical snapshot (2026-06-23; не release acceptance)

| Показатель | Значение |
|------------|----------|
| Historical milestone labels | v1.0 · v2.0 · v2.5 Trust · v2.6 Automate · v3.0 AI · **v4 Super-app** |
| Historical audit product score | **89/100** |
| Backend coverage (`internal/`) | **83.0%** (2026-06-22 Docker baseline; Wave C, `-tags hubtest`) |
| Security-critical ≥80% | **6/6** (auth, ws, rbac, audit, calls, platform/middleware) |
| REST API (OpenAPI paths) | **252** (`swagger.json`, 2026-07-28); route inventory ~286 |
| Frontend Vitest | **179** files (`stages/unit/frontend-test/`, measured 2026-07-29; 61 `.ts` + 118 `.tsx`) |
| Playwright E2E | **20** specs |
| 12 CRITICAL security blockers | **Закрыты** (v2.0.1) |
| Ветка | `cursor-grok-4.5` · HEAD `d868ba77` |
| Separate ops cutover blocker | **FA-06/PG-03** secret rotation (manual); не proxy готовности MVP |

### 1.3 Ценностное предложение

> «Один self-hosted инстанс вместо Jira + Slack + TestRail + Meet + Bitrix — с трассировкой от CI-теста до бага, PR и счёта клиенту, плюс интранет и AI для всей компании.»

---

## 2. Персоны → Jobs-to-be-Done → Боли → Незакрытые needs

> `✅` в таблицах персон означает только исторически зафиксированный capability
> surface; это не `complete`, `stable` или production acceptance модуля.

### 2.1 Алексей — Developer

| JTBD | Статус в Hub | Незакрытые needs |
|------|--------------|------------------|
| Взять задачу, закодить, залить PR | ✅ Единый UI | US-C06: CI workflow status в sidebar (partial) |
| Быстро найти задачу/канал | ✅ Ctrl+K + Meilisearch | Shared JQL filters polish |
| Не терять время на формы | ⚠️ toastApiError ✅ | RHF+Zod на issue-form (UX-1) |
| Созвониться по задаче | ✅ Call from context | — |
| AI помощник в чате | ✅ @hub-ai v3.0 | — |

**Приоритет:** CI status sidebar → keyboard shortcuts depth → RHF forms

### 2.2 Мария — QA Lead

| JTBD | Статус в Hub | Незакрытые needs |
|------|--------------|------------------|
| Кейсы + Allure из CI | ✅ Parser + hub-cli | Flaky ML-level (vs TestOps) |
| Traceability requirement→test | ✅ Matrix | Bidirectional issue↔case polish |
| Баг из failed run | ✅ | Bulk bug from run |
| Отчёты для релиза | ✅ trends + env compare | PDF depth, run diff UX |

**Приоритет:** flaky detection depth → run comparison UX → TestRail enum fix (FA-12)

### 2.3 Дмитрий — Team Lead

| JTBD | Статус в Hub | Незакрытые needs |
|------|--------------|------------------|
| Спринт, burndown, velocity | ✅ | Timeline roadmap (v4.5+) |
| Retro, planning poker | ✅ | Sprint goals tracking |
| Автоматизация рутины | ✅ automation lite (v2.6) | Rule depth vs Jira |
| Загрузка команды | ✅ Dashboard widgets | Capacity planning |

**Приоритет:** automation rule UX → roadmap view → notification digest

### 2.4 Виктор — Project Owner

| JTBD | Статус в Hub | Незакрытые needs |
|------|--------------|------------------|
| Workflow + RBAC | ✅ | Custom fields UI (schema есть) |
| GitHub/GitLab connect | ✅ | Webhook delivery admin UX |
| Шаблоны проектов | ✅ | — |
| CRM + accounting hook | ✅ v4 won→invoice | Accounting v4.5 depth |

**Приоритет:** custom fields builder → webhook log UX → invoice workflow polish

### 2.5 Елена — Admin / HR

| JTBD | Статус в Hub | Незакрытые needs |
|------|--------------|------------------|
| Users, audit, GDPR | ✅ | PG-03 secret rotation (manual) |
| Backups restore | ✅ | Backup verify schedule |
| Resource limits | ✅ | Rate limit visibility admin UI |
| HR leave + 1C | ✅ v4 | 1C connector prod wiring (FA-05) |
| OBS stack | ✅ Prometheus/Loki/AM | — |

**Приоритет:** product gaps развиваются отдельно; PG-03 rotation относится к
manual ops production cutover и не заменяет product acceptance.

### 2.6 Игорь — Member (messenger + intranet)

| JTBD | Статус в Hub | Незакрытые needs |
|------|--------------|------------------|
| Каналы, news, kudos | ✅ v4 intranet | Issue↔thread full sync |
| Telegram/Push | ✅ | Digest notifications |
| Поиск | ✅ Ctrl+K | Mobile PWA polish |
| AI | ✅ @hub-ai | — |

**Приоритет:** notification digest → PWA offline → thread↔issue sync

### 2.7 Сергей — DevOps

| JTBD | Статус в Hub | Незакрытыe needs |
|------|--------------|------------------|
| hub-cli push-results | ✅ | In-product CLI docs |
| JUnit webhook | ✅ | Webhook retry dashboard UX |
| #hub-ci | ✅ | Pipeline failed→issue (automation) |
| Observability | ✅ OTel + metrics | hub-mail prod compose (FA-05) |
| Test framework | исторический TF-M01–M10 technical slice | активен rebase; package/launcher `NOT_READY`, runtime `NOT_RUN` |

**Приоритет:** product/quality acceptance; FA-06/PG-03 ведётся отдельно как
manual ops production-cutover blocker.

### 2.8 Анна — Employee (non-dev, v4 persona)

| JTBD | Статус в Hub | Незакрытые needs |
|------|--------------|------------------|
| Новости компании | ✅ intranet feed | — |
| Отпуск | ✅ HR-lite | Calendar polish |
| Random Coffee | ✅ | — |
| Marketplace | ✅ classifieds | — |
| Directory | ✅ MOD-02 | Org chart viz (v4.5+) |

---

## 3. MoSCoW приоритизация (актуально 2026-06-23)

### Historical technical handoff gates (не product acceptance)

| ID | Feature | Статус |
|----|---------|--------|
| COV ≥82% | Backend coverage | ✅ **83.0%** (2026-06-22 Docker baseline) |
| Security ≥80% | 6 critical packages | ✅ **6/6** |
| TF-M04–M08 | Historical migration slice | historical COMPLETE label; новый package/launcher `NOT_READY`, runtime `NOT_RUN` |
| FA-10 | Full `/qa-gate` Docker | ✅ PASS |
| PG-03 | Secret rotation execute | ⏳ отдельный manual ops production-cutover blocker; не readiness proxy |
| v2.5.0 tag | Historical release-plan item | ⏳ требует отдельного founder sign-off; FA-06 сам по себе готовность не доказывает |

### Should — differentiation polish

| ID | Feature | Personas | Статус |
|----|---------|----------|--------|
| FA-11 | tms/project → 82% | QA, PO | ✅ |
| UX-1 | Command palette actions | All | ✅ PR #15 |
| NFR-16/17 | WCAG + Lighthouse | All | ⚠️ partial (LH artifact PR #27) |
| PROD-02 depth | Flaky detection | QA | backlog |

### Could — v4.5+

| ID | Feature |
|----|---------|
| Roadmap timeline | TL, PO |
| Custom fields UI | PO |
| Notification digest | Member |
| Org chart visualization | HR, Admin |
| Accounting v4.5 advanced | PO, Finance |

### Won't (constraints)

SSO, multi-org, native apps, SaaS billing — per `.cursor/rules/60-tz-constraints.mdc`.  
Email notifications (кроме magic link + hub-mail marketing) — Telegram + Push primary.

---

## 4. Противоречия scope vs ожидания

| Ожидание | Ограничение Hub | Mitigation |
|----------|-----------------|------------|
| «Slack AI из коробки» | BYOK / Ollama setup | historical v3.0 capability slice; AI-модуль не принят |
| «SSO для compliance» | No SSO v1–v4 | Self-host + audit; SCIM backlog 999.x |
| «Email alerts on prod down» | No generic email notif | Telegram + Push + status page + AM webhook |
| «Jira Automation 50 rules» | Lite automation MVP | Honest UI limits; roadmap depth |
| «Allure TestOps flaky ML» | Basic trends only | hub-cli + reports; ML — v4.5+ |

---

## 5. Voice of Customer

### Switch triggers (FROM Jira+Slack+TestRail+Bitrix)

1. **TCO > $30k/год** на 50 человек — Hub = infra only
2. **Data residency RU/on-prem** — K3s self-host
3. **QA устала ручной Allure sync** — hub-cli + auto match
4. **Dev устал Alt-Tab** — один UI task+chat+call+AI
5. **HR/Admin хочет Mighty без отдельного app** — v4 intranet+HR
6. **Finance хочет CRM→invoice** — v4 accounting hook

### Retain triggers

1. **Automation lite** — «Done → #releases» без Zapier ✅
2. **Reliability evidence (historical, not readiness)** — 82.6% cov (2026-06-21 phase-125 exit; superseded by 83.0% 2026-06-22 Docker baseline), e2e 20 specs, no WS panics
3. **AI @hub-ai RU providers** — GigaChat/YandexGPT/Ollama
4. **Keyboard UX** — Ctrl+K как Linear
5. **Single audit trail** — dev + corp actions в одном месте

### Churn risks (остаточные)

- PO → Jira если **custom fields UI** затянется
- QA → TestOps если **flaky ML** не улучшится
- Members → Slack если **mobile PWA** отстанет
- Admin → cloud если **PG-03 rotation** не выполнится к production cutover; это отдельный ops-риск, не полный readiness gate

---

## 6. Heatmap: persona × feature (2026)

| Feature | Dev | QA | TL | PO | Admin | DevOps | Member | Employee |
|---------|-----|-----|-----|-----|-------|--------|--------|----------|
| Automation lite ✅ | ○ | ○ | ● | ● | ○ | ● | ○ | ○ |
| @hub-ai ✅ | ● | ● | ● | ○ | ● | ● | ● | ○ |
| v4 intranet/HR ✅ | ○ | ○ | ○ | ○ | ● | ○ | ● | ● |
| CRM→invoice ✅ | ○ | ○ | ○ | ● | ○ | ○ | ○ | ○ |
| Flaky depth ⚠️ | ○ | ● | ○ | ○ | ○ | ○ | ○ | ○ |
| TF migration 🔄 | ○ | ● | ○ | ○ | ● | ● | ○ | ○ |
| Custom fields UI 📋 | ○ | ○ | ○ | ● | ○ | ○ | ○ | ○ |
| PG-03 rotation ⏳ | ○ | ○ | ○ | ○ | ● | ● | ○ | ○ |

● = высокий impact · ○ = средний/низкий · ✅ историческое implementation evidence, не acceptance · ⚠️ partial · 🔄 in progress · 📋 backlog · ⏳ ops

---

*См. также: [competitive_analysis_2026.md](./competitive_analysis_2026.md) · [.planning/archive/roadmaps/FEATURE-BACKLOG-2026.md](../.planning/archive/roadmaps/FEATURE-BACKLOG-2026.md) · [.planning/COVERAGE.md](../.planning/COVERAGE.md)*
