# Hub — Конкурентный анализ 2026

> Обновлено: 2026-06-21 · Источники: docs/competitive_audit.md, vision_and_scope.md, web research (Linear, Jira, Plane, Allure TestOps, Slack).
> Целевой сегмент: **self-hosted команды 10–50 человек**, замена стека Jira + Slack + TestRail/Allure + Meet + Grafana.

---

## 1. Матрица конкурентов по категориям

| Категория | Лидеры рынка | Open-source / self-hosted | Hub покрывает? |
|-----------|--------------|---------------------------|----------------|
| **Task Tracker** | Jira, Linear, Asana, ClickUp | Plane, OpenProject, Taiga | ✅ Kanban+Scrum, workflow, JQL partial |
| **Messenger** | Slack, Discord, Teams | Mattermost, Element, Rocket.Chat | ✅ Каналы, треды, DM, reactions |
| **TMS** | TestRail, Allure TestOps, Xray | — (Allure Report OSS) | ✅ Кейсы, Allure CI, traceability partial |
| **Calls** | Zoom, Meet, Discord | LiveKit (embed) | ✅ Voice + screen share, без webcams |
| **Dashboards** | Grafana, Datadog, monday | — | ⚠️ Статичные виджеты, нет alerting |
| **Unified suite** | Atlassian (Jira+Confluence), GitLab Ultimate | Plane (tracker only), OpenProject | ✅ **Единый UX + deep links** |

### Прямые «заменители стека» (ближайшие аналоги Hub)

| Продукт | Что объединяет | Чего нет vs Hub | TCO (50 users, год) |
|---------|----------------|-----------------|---------------------|
| **Atlassian Cloud** | Jira + Confluence + (Slack интеграция) | TMS, native calls, self-host | ~$8–17/user/mo × 50 ≈ **$5–10k/год** + TestRail отдельно |
| **GitLab Ultimate** | Issues + CI + MR + (calls beta) | TMS depth, messenger UX | ~$29/user/mo ≈ **$17k/год** |
| **Plane CE + Slack + TestRail** | Tracker (+ Slack sync) | Нет native TMS/calls в одном UI | Plane $7/seat + Slack ~$8 + TestRail ~$37 ≈ **$26k/год** |
| **Hub self-hosted** | Всё в одном | Automation UI, flaky ML, AI | **Инфра ~$200–500/mo** (VPS/K3s) |

**Вывод TCO:** Hub выигрывает на горизонте 2–3 лет для команд с требованием data residency и без лицензий TestOps ($30–39/user/mo).

---

## 2. Feature-by-feature: Hub vs лидеры

### 2.1 Task Tracker

| Фича | Jira | Linear | Plane | Hub |
|------|------|--------|-------|-----|
| Custom workflow | ⭐⭐⭐ deep | ⭐ basic | ⭐⭐ CE / ⭐⭐⭐ Commercial | ⭐⭐ CRUD + engine |
| **Workflow automation (no-code)** | ⭐⭐⭐ | ⭐ AI agents 2026 | ⭐ Commercial | ❌ **gap #1** (PROD-01) |
| Cycles / Sprints | Sprints + reports | Cycles auto-roll | Cycles | ✅ Sprints + burndown |
| Triage inbox | Backlog filters | ⭐ Triage view | Intake | ⚠️ Saved filters, нет triage UX |
| Roadmaps | Advanced Roadmaps ($) | ⭐ Initiatives | Modules | ⚠️ Epic hierarchy, нет timeline roadmap |
| JQL / query language | ⭐⭐⭐ JQL | Filters | Views | ⚠️ JQL partial, saved filters partial |
| Git auto-link PR | ⭐⭐⭐ | ⭐⭐⭐ auto status | ⭐⭐ sync | ✅ GitHub/GitLab webhooks |
| Keyboard / ⌘K | ⚠️ | ⭐⭐⭐ | ⭐⭐ | ✅ Command palette (2026) |
| Self-host | Data Center ($$$) | ❌ | ✅ AGPL CE | ✅ Docker/K3s |

### 2.2 Messenger

| Фича | Slack | Discord | Mattermost | Hub |
|------|-------|---------|------------|-----|
| Channels + threads | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ✅ |
| Search | ⭐⭐⭐ | ⭐⭐ | ⭐⭐ | ✅ global + channel |
| Bots / apps | ⭐⭐⭐ marketplace | ⭐⭐⭐ | ⭐⭐ plugins | ⚠️ Telegram bot only |
| **AI in chat** | Slack AI ($) | — | — | 📋 v3.0 MILESTONE-3 |
| Huddles / calls | ⭐⭐ | ⭐⭐⭐ voice | ⭐ calls plugin | ✅ LiveKit native |
| Issue ↔ thread sync | Integrations | — | — | ⚠️ Deep links, нет bidirectional sync |
| Self-host | ❌ | ❌ | ✅ | ✅ |

### 2.3 TMS

| Фича | TestRail | Allure TestOps | Hub |
|------|----------|----------------|-----|
| Manual + auto cases | ⭐⭐⭐ | ⭐⭐⭐ | ✅ |
| Allure CI ingest | Via API | ⭐⭐⭐ native | ✅ own parser |
| **Flaky detection ML** | ❌ | ⭐⭐⭐ | ❌ **gap #2** |
| Live test docs from code | ❌ | ⭐⭐⭐ | ⚠️ annotation link |
| Run comparison / trends | ⭐⭐ | ⭐⭐⭐ | ⚠️ basic trends |
| Traceability matrix | ⭐⭐ | ⭐⭐⭐ | ✅ |
| Self-host | ✅ | ✅ Docker/K8s | ✅ |

### 2.4 Calls & Dashboards

| Фича | Zoom/Meet | Hub |
|------|-----------|-----|
| Voice + screen share | ⭐⭐⭐ | ✅ |
| Recording | ⭐⭐⭐ | ✅ SeaweedFS |
| Scheduled / recurring | ⭐⭐⭐ | ✅ |
| Webcams | ⭐⭐⭐ | ❌ by design |
| Custom alerting dashboards | Grafana ⭐⭐⭐ | ❌ static widgets only |

### 2.5 Admin & Security

| Фича | Auth0 / Datadog | Hub |
|------|-----------------|-----|
| Audit log UI | ⭐⭐⭐ | ✅ |
| User ban + session revoke | ⭐⭐⭐ | ✅ |
| SSO / SCIM | ⭐⭐⭐ | ❌ out of scope |
| APM / traces | ⭐⭐⭐ | ⚠️ HTTP timing in audit |
| Secret rotation ops | — | 📋 PG-03 checklist |

---

## 3. Тренды 2025–2026

1. **AI-native workflows** — Linear AI triage, Jira Rovo, Plane AI actions, Slack AI summaries. Hub: v3.0 `@hub-ai` (RU-friendly providers).
2. **Consolidation fatigue** — команды устают от 5+ SaaS; спрос на self-hosted suites растёт (Plane 51k GitHub stars).
3. **Automation without code** — Jira Automation стал table stakes для PO; без этого Hub проигрывает в enterprise-adjacent сегменте.
4. **Quality engineering in CI** — flaky detection, DORA metrics, test orchestration — Allure TestOps и Datadog CI доминируют.
5. **Keyboard-first UX** — Linear задал планку; command palette обязателен (Hub: ✅).
6. **Compliance без SSO** — малые команды хотят audit + data control без Okta; Hub попадает в нишу.

---

## 4. Top 15 gaps (закрыть для победы над point tools)

| # | Gap | Конкурент-эталон | Hub milestone |
|---|-----|------------------|---------------|
| 1 | Workflow automation UI | Jira Automation | **v2.6 PROD-01** |
| 2 | Flaky test detection | Allure TestOps ML | v2.6 PROD-02 |
| 3 | Issue ↔ message bidirectional sync | Linear+Slack, Plane+Slack | v2.6 PROD-03 |
| 4 | Webhook delivery dashboard | Stripe, GitHub | v2.6 PROD-04 |
| 5 | Full JQL + saved/shared filters | Jira | v2.5 |
| 6 | Triage inbox для входящих багов | Linear Triage | v2.6 |
| 7 | Timeline roadmap / initiatives | Linear, Jira AR | v2.7 |
| 8 | RHF+Zod forms (validation UX) | Industry standard | v2.5 |
| 9 | E2E critical paths (project CRUD, messenger, TMS) | CI best practice | v2.5 |
| 10 | Lighthouse ≥90 + WCAG audit sign-off | NFR-16 | v2.5 |
| 11 | OpenTelemetry traces | Grafana Tempo | v2.6 |
| 12 | Notification digest (не только realtime) | Slack | v2.6 |
| 13 | Custom fields builder UI | Jira | v2.7 |
| 14 | CI workflow status in issue sidebar | GitHub Actions UI | v2.6 (US-C06) |
| 15 | AI chat in messenger | Slack AI | **v3.0** |

---

## 5. Top 10 differentiators (удвоить ставку)

1. **Единый self-hosted стек** — tracker + chat + TMS + calls + dashboards без 4–6 SaaS контрактов.
2. **Сквозная traceability** — задача → тест → баг → PR → уведомление в одном инстансе.
3. **Allure без TestOps license** — собственный парсер + TMS API (`hub-cli push-results`).
4. **Passwordless RU-friendly auth** — Magic Link + Telegram (без SSO complexity).
5. **LiveKit calls в контексте проекта** — созвон из канала/задачи, запись в проект.
6. **Optimistic locking + soft delete 30d** — enterprise data patterns без enterprise price.
7. **Deep module integration** — thread→issue promotion, failed case→bug, PR webhook→status.
8. **Command palette + user-scope search** — безопасный глобальный поиск (NF-03).
9. **Admin audit + ban + resource limits** — governance для single-team instance.
10. **v3.0 AI с GigaChat/YandexGPT/Ollama** — уникально для RU сегмента без VPN.

---

## 6. Стратегическая позиция

```
                    HIGH integration depth
                            │
         Allure TestOps     │     ★ Hub (target)
         (TMS only)         │     unified self-hosted
                            │
    LOW ────────────────────┼──────────────────── HIGH
    module count            │                    module count
                            │
         Linear             │     Atlassian Cloud
         (tracker only)     │     (expensive suite)
                            │
                    LOW integration depth
```

**Рекомендация:** не конкурировать с Jira в глубине workflow enterprise; побеждать **TCO + integration + self-host** для команд 10–50, затем закрыть automation + TMS analytics для удержания QA Lead.

---

## 7. Corporate super-apps (v4 benchmark) — NEW 2026-06-21

| Платформа | Фокус | Сильные стороны | Слабости vs Hub |
|-----------|-------|-----------------|-----------------|
| **Tinkoff Mighty** | Employee super-app | News, HR services, docs, mobile-first | Нет dev tracker/TMS; closed ecosystem |
| **Bitrix24** | SMB all-in-one | CRM, HR, chat, tasks, automation | Weak dev/TMS; heavy UI |
| **Lark (Feishu)** | Global teams | Chat, docs, Base no-code, AI | Cloud-only; China/US tension |
| **Microsoft Viva** | EX on Teams | Learning, topics, connections | Requires M365; no TMS |
| **Plane** | OSS tracker | Jira alt, Slack sync | No messenger/TMS/calls native |

### v4 gaps Hub should add (from corp apps)

| # | Feature | Benchmark |
|---|---------|-----------|
| 16 | Corp news feed + reactions | Mighty |
| 17 | Org directory + org chart | Viva Connections |
| 18 | Leave request workflow | Bitrix HR / T2 app |
| 19 | Unified calendar | Lark Calendar |
| 20 | CRM-lite pipeline | Bitrix CRM |
| 21 | **Multi-model AI picker (no VPN)** | OpenRouter pattern |
| 22 | 1C connector (not built-in accounting) | T-Bank API style |

**Hub unique v4 pitch:** Mighty experience **+** dev/TMS depth **+** self-host **+** AI hub.

---

*Следующий документ: [user_needs_analysis_2026.md](./user_needs_analysis_2026.md) · Vision: [PLATFORM-VISION-v4.md](../.planning/PLATFORM-VISION-v4.md)*
