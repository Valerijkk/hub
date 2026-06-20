# Hub — Анализ продукта и потребностей пользователей 2026

> Обновлено: 2026-06-21 · Основано на vision_and_scope.md, PROJECT.md, AUDIT-SCORECARD (93/100), competitive_analysis_2026.md.

---

## 1. Что такое Hub сегодня

### 1.1 Суть продукта

**Hub** — self-hosted единая платформа для команд разработки и QA (10–50 человек), объединяющая:

| Модуль | Зрелость | Ключевые возможности |
|--------|----------|----------------------|
| **Auth** | Production | Magic Link + Telegram, JWT 15m, refresh 30d, sessions |
| **Projects + RBAC** | Production | Шаблоны, invite, custom roles, workflow |
| **Tracker** | Production | 9 типов задач, Kanban, Scrum, time tracking, Git links |
| **Messenger** | Production | Каналы, треды, reactions, polls, voice msgs, #hub-ci |
| **TMS** | Production | Classic/BDD cases, runs, Allure CI, traceability matrix |
| **Calls** | Production | LiveKit voice, screen share, recordings, scheduled |
| **Dashboards** | Mature | Personal/project widgets, burndown, velocity, health |
| **Admin** | Production | Audit log, trash 30d, backups, limits, ban |
| **Search** | Mature | Ctrl+K command palette, user-scope API |
| **Integrations** | Mature | GitHub, GitLab, Jira import, webhooks, hub-cli |

### 1.2 Статус релиза

- **v2.0.2 SHIP-READY** — product score **93/100** (commit `90042eb`)
- Coverage backend **~56.6%**, security-critical **≥80%**
- Vitest **49**, Playwright e2e **7 specs**
- 12 CRITICAL security blockers — **закрыты в коде**
- Потолок без v2.6: automation, flaky TMS, full E2E, Lighthouse sign-off

### 1.3 Ценностное предложение (one-liner)

> «Один self-hosted инстанс вместо Jira + Slack + TestRail + Meet — с трассировкой от CI-теста до бага и PR.»

---

## 2. Персоны → Jobs-to-be-Done → Боли → Незакрытые needs

### 2.1 Алексей — Developer

| JTBD | Боли в Hub сегодня | Незакрытые US |
|------|-------------------|---------------|
| Взять задачу, закодить, залить PR | Context switch между табами OK | US-C06: CI workflow status в sidebar |
| Быстро найти задачу/канал | Ctrl+K ✅ | US-C04: shared JQL filters |
| Не терять время на формы | UX-08 toastApiError ✅ | RHF+Zod на issue-form (partial) |
| Созвониться по задаче | Call from context ✅ | — |

**Приоритет фич:** CI status in issue → keyboard shortcuts overlay → optimistic reactions

### 2.2 Мария — QA Lead

| JTBD | Боли | Незакрытые US |
|------|------|---------------|
| Вести кейсы + получать Allure из CI | Parser ✅ | US-C02: flaky + run comparison |
| Traceability requirement→test | Matrix ✅ | Bidirectional issue↔case sync polish |
| Создать баг из failed | ✅ | Bulk bug from run |
| Отчёты для релиза | Basic trends | Allure TestOps-level dashboards |

**Приоритет:** PROD-02 flaky detection → environment comparison UX → PDF export runs

### 2.3 Дмитрий — Team Lead

| JTBD | Боли | Незакрытые US |
|------|------|---------------|
| Планировать спринт, видеть burndown | ✅ | Roadmap timeline (нет) |
| Velocity / retro | ✅ | Sprint goals tracking |
| Автоматизировать рутину | ❌ **нет automation** | PROD-01 |
| Видеть загрузку команды | Dashboard widgets ✅ | Capacity planning (нет) |

**Приоритет:** PROD-01 automation → roadmap view → digest notifications

### 2.4 Виктор — Project Owner

| JTBD | Боли | Незакрытые US |
|------|------|---------------|
| Настроить workflow | UI ✅ | Automation rules |
| RBAC / invites | ✅ | Custom fields UI (schema есть) |
| GitHub/GitLab connect | ✅ | Webhook delivery log (PROD-04) |
| Шаблоны проектов | ✅ | — |

**Приоритет:** PROD-01 → PROD-04 webhooks → custom fields builder

### 2.5 Елена — Admin

| JTBD | Боли | Незакрытые US |
|------|------|---------------|
| Users, audit, GDPR | ✅ | PG-03 secret rotation (manual) |
| Backups restore | ✅ | Backup verify schedule |
| Resource limits | ✅ | Rate limit visibility admin UI |
| Registration mode | ✅ | — |

**Приоритет:** slow query dashboard → OTel → audit retention S3 archive

### 2.6 Игорь — Member (messenger-heavy)

| JTBD | Боли | Незакрытые US |
|------|------|---------------|
| Общаться в каналах | ✅ | Issue↔thread sync (PROD-03) |
| Уведомления Telegram/Push | ✅ | Digest / quiet hours polish |
| Быстрый поиск | Ctrl+K ✅ | Mobile PWA polish US-C05 |
| AI помощник | ❌ | v3.0 @hub-ai |

**Приоритет:** PROD-03 → notification digest → PWA offline

### 2.7 Сергей — DevOps

| JTBD | Боли | Незакрытые US |
|------|------|---------------|
| hub-cli push-results | ✅ | CLI docs in-product |
| JUnit webhook | ✅ | Webhook retry dashboard |
| #hub-ci read-only | ✅ | Pipeline failed → auto issue (automation) |
| Observability | HTTP audit ✅ | OTel export |

**Приоритет:** PROD-04 → PROD-01 CI triggers → OTel

---

## 3. MoSCoW приоритизация по impact

### Must (удержание + production trust)

| ID | Feature | Personas | Rationale |
|----|---------|----------|-----------|
| PROD-01 | Workflow automation lite | PO, TL, DevOps | #1 competitive gap vs Jira |
| E2E-02..05 | Critical path e2e | All | Regression confidence |
| Lighthouse/WCAG sign-off | a11y NFR-16 | All | G-05 completion |
| PG-03 | Secret rotation exec | Admin | Production ops |

### Should (differentiation v2.6)

| ID | Feature | Personas |
|----|---------|----------|
| PROD-02 | Flaky test detection | QA |
| PROD-03 | Issue↔message sync | Dev, Member |
| PROD-04 | Webhook delivery log | PO, DevOps |
| US-C06 | GitHub Actions in issue | Dev |
| US-C04 | Shared JQL filters | Dev, TL |
| RHF+Zod forms | Validation UX | All |

### Could (v2.7+)

| ID | Feature | Personas |
|----|---------|----------|
| Roadmap timeline | TL, PO |
| Custom fields UI | PO |
| Triage inbox | TL |
| Notification digest | Member |
| PWA polish | Member |

### Won't (constraints)

SSO, email notif (except magic link), webcams, native apps, Gantt, multi-tenant, billing — per `.cursor/rules/60-tz-constraints.mdc`.

---

## 4. Противоречия scope vs ожидания пользователей

| Ожидание пользователя | Ограничение Hub | Mitigation |
|-----------------------|-----------------|------------|
| «Хочу как в Slack AI» | AI только v3.0 | Roadmap MILESTONE-3, @hub-ai |
| «Нужен SSO для compliance» | No SSO v1–v2 | Self-host + audit; SCIM в backlog 999.x |
| «Email когда упал прод» | No email notif | Telegram + Web Push + urgent bypass |
| «Видео на созвонах» | Voice-only | Screen share as primary |
| «Как Jira Automation 50 rules» | Lite 1–3 rules MVP | Phase PROD-01 scope honest in UI |

---

## 5. Voice of Customer — почему перейдут и останутся

### Switch triggers (FROM Jira+Slack+TestRail)

1. **Счёт за 50 человек > $30k/год** на SaaS — Hub = infra only
2. **Data must stay in RU/on-prem** — self-hosted K3s
3. **QA устала синкать Allure вручную** — hub-cli + auto match
4. **Dev устал Alt-Tab** — один UI для task+chat+call
5. **Admin хочет один audit** — не 3 admin panels

### Retain triggers (чтобы не вернулись)

1. **Automation** — «Done → post to #releases» без Zapier
2. **Flaky reports** — QA видит ценность TMS vs Excel
3. **Reliability** — e2e green, no WS panics, fast search
4. **AI @hub-ai** — уникально для RU команд
5. **Keyboard UX** — скорость как Linear

### Churn risks

- PO уходит в Jira если **нет automation** и **custom fields UI**
- QA уходит в TestOps если **нет flaky + run diff**
- Members уходят в Slack если **messenger lag** или **нет mobile UX**

---

## 6. Сводная карта persona → feature (heatmap)

| Feature | Dev | QA | TL | PO | Admin | DevOps | Member |
|---------|-----|-----|-----|-----|-------|--------|--------|
| PROD-01 Automation | ○ | ○ | ● | ● | ○ | ● | ○ |
| PROD-02 Flaky | ○ | ● | ○ | ○ | ○ | ○ | ○ |
| PROD-03 Issue↔msg | ● | ○ | ○ | ○ | ○ | ○ | ● |
| PROD-04 Webhook log | ○ | ○ | ○ | ● | ○ | ● | ○ |
| v3.0 @hub-ai | ● | ● | ● | ○ | ● | ● | ● |
| E2E expansion | ○ | ● | ● | ○ | ● | ● | ○ |
| Roadmap timeline | ○ | ○ | ● | ● | ○ | ○ | ○ |

● = high impact · ○ = medium/low

---

*См. также: [competitive_analysis_2026.md](./competitive_analysis_2026.md) · [.planning/FEATURE-BACKLOG-2026.md](../.planning/FEATURE-BACKLOG-2026.md)*
