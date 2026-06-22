# Hub — Конкурентный анализ 2026

> **Обновлено:** 2026-06-21  
> **Источники:** `docs/competitive_audit.md`, `docs/vision_and_scope.md`, `.planning/PLATFORM-VISION-v4.md`, web research (Linear, Jira, Plane, Allure TestOps, Slack, Mighty, Bitrix24, Lark).  
> **Целевой сегмент:** self-hosted команды **10–50 человек**, замена стека Jira + Slack + TestRail/Allure + Meet + Grafana **+** корпоративные сервисы (интранет, HR, CRM).

---

## 1. Резюме для руководства

Hub после milestone **v4 Super-app** (Phases 39–58) — единственный self-hosted продукт в нише, объединяющий **dev-suite + corporate super-app** с AI hub и data residency. Покрытие тестами backend **82.6%**, product score **93/100**, 286+ REST endpoints.

**Ключевое конкурентное преимущество:** TCO на горизонте 2–3 лет + сквозная traceability (задача → тест → PR → уведомление) + RU-friendly AI без VPN.

---

## 2. Матрица конкурентов по категориям

| Категория | Лидеры рынка | Open-source / self-hosted | Hub (2026-06-21) |
|-----------|--------------|---------------------------|------------------|
| **Task Tracker** | Jira, Linear, Asana, ClickUp | Plane, OpenProject, Taiga | ✅ Kanban+Scrum, workflow, JQL, automation lite |
| **Messenger** | Slack, Discord, Teams | Mattermost, Element | ✅ Каналы, треды, DM, reactions, @hub-ai |
| **TMS** | TestRail, Allure TestOps, Xray | Allure Report OSS | ✅ Кейсы, Allure CI, traceability, reports |
| **Calls** | Zoom, Meet, Discord | LiveKit (embed) | ✅ Voice + screen share + webcams (Phase 16) |
| **Dashboards** | Grafana, Datadog | — | ⚠️ Статичные виджеты; OBS stack ✅ |
| **Corp super-app** | Mighty, Bitrix24, Lark, Viva | — | ✅ Intranet, HR, CRM, calendar, bases, accounting |
| **Unified suite** | Atlassian Cloud, GitLab Ultimate | Plane (tracker only) | ✅ **Единый UX + deep links + self-host** |

### Прямые «заменители стека»

| Продукт | Что объединяет | Чего нет vs Hub | TCO (50 users, год) |
|---------|----------------|-----------------|---------------------|
| **Atlassian Cloud** | Jira + Confluence + Slack sync | TMS depth, native calls, self-host | ~$5–10k/год + TestRail отдельно |
| **GitLab Ultimate** | Issues + CI + MR | TMS depth, messenger UX, corp modules | ~$17k/год |
| **Plane CE + Slack + TestRail** | Tracker + chat sync | Нет TMS/calls/corp в одном UI | ~$26k/год |
| **Bitrix24** | CRM + HR + chat + tasks | Слабый dev/TMS; cloud-heavy | ~$3–8k/год (SaaS) |
| **Hub self-hosted** | Dev + corp + AI в одном | Enterprise JQL depth, flaky ML | **Инфра ~$200–500/mo** (VPS/K3s) |

---

## 3. Feature-by-feature: Hub vs лидеры

### 3.1 Task Tracker

| Фича | Jira | Linear | Plane | Hub |
|------|------|--------|-------|-----|
| Custom workflow | ⭐⭐⭐ | ⭐ basic | ⭐⭐ CE | ⭐⭐ engine + UI |
| Workflow automation (no-code) | ⭐⭐⭐ | ⭐ AI 2026 | ⭐ Commercial | ✅ **lite** (Phase 16) |
| Cycles / Sprints | Sprints | Cycles | Cycles | ✅ Sprints + burndown |
| Triage inbox | Backlog filters | ⭐ Triage | Intake | ⚠️ Phase 16 partial |
| Roadmaps / Gantt | Advanced ($) | Initiatives | Modules | ✅ Gantt (Phase 16) |
| JQL / query language | ⭐⭐⭐ JQL | Filters | Views | ✅ Full JQL (Phase 17) |
| Git auto-link PR | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ✅ GitHub/GitLab webhooks |
| Keyboard / ⌘K | ⚠️ | ⭐⭐⭐ | ⭐⭐ | ✅ Command palette |
| Self-host | Data Center ($$$) | ❌ | ✅ AGPL CE | ✅ Docker/K3s |

### 3.2 Messenger

| Фича | Slack | Discord | Mattermost | Hub |
|------|-------|---------|------------|-----|
| Channels + threads | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ✅ |
| Search | ⭐⭐⭐ | ⭐⭐ | ⭐⭐ | ✅ Meilisearch + Ctrl+K |
| Bots / apps | ⭐⭐⭐ marketplace | ⭐⭐⭐ | ⭐⭐ plugins | ⚠️ Telegram + @hub-ai |
| AI in chat | Slack AI ($) | — | — | ✅ v3.0 BYOK + Ollama |
| Huddles / calls | ⭐⭐ | ⭐⭐⭐ | ⭐ plugin | ✅ LiveKit native |
| Issue ↔ thread sync | Integrations | — | — | ⚠️ Deep links + partial sync |
| Self-host | ❌ | ❌ | ✅ | ✅ |

### 3.3 TMS

| Фича | TestRail | Allure TestOps | Hub |
|------|----------|----------------|-----|
| Manual + auto cases | ⭐⭐⭐ | ⭐⭐⭐ | ✅ |
| Allure CI ingest | Via API | ⭐⭐⭐ native | ✅ own parser + hub-cli |
| Flaky detection ML | ❌ | ⭐⭐⭐ | ⚠️ basic trends; ML — gap |
| Run comparison / trends | ⭐⭐ | ⭐⭐⭐ | ✅ reports + env compare |
| Traceability matrix | ⭐⭐ | ⭐⭐⭐ | ✅ |
| Self-host | ✅ | ✅ Docker/K8s | ✅ |

### 3.4 Calls, Dashboards, Corp

| Фича | Zoom/Meet | Hub |
|------|-----------|-----|
| Voice + screen share | ⭐⭐⭐ | ✅ |
| Webcams | ⭐⭐⭐ | ✅ Phase 16 |
| Recording | ⭐⭐⭐ | ✅ SeaweedFS |
| Custom alerting dashboards | Grafana ⭐⭐⭐ | ⚠️ static widgets; Prometheus/Grafana stack ✅ |

| Фича | Mighty | Bitrix24 | Lark | Hub v4 |
|------|--------|----------|------|--------|
| Corp news feed | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ✅ intranet |
| Org directory | ⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ✅ MOD-02 |
| Leave workflow | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ✅ HR-lite |
| Unified calendar | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ✅ + iCal |
| CRM pipeline | ❌ | ⭐⭐⭐ | ⭐⭐ | ✅ CRM-lite |
| No-code DB | ❌ | ⭐⭐ | ⭐⭐⭐ Base | ✅ `/bases` |
| Accounting | ❌ | ⭐⭐ | ❌ | ✅ MVP + invoices |
| Dev tracker + TMS | ❌ | ⚠️ weak | ❌ | ✅ **unique** |

### 3.5 Admin & Security

| Фича | Auth0 / Datadog | Hub |
|------|-----------------|-----|
| Audit log UI | ⭐⭐⭐ | ✅ + slow query filter |
| User ban + session revoke | ⭐⭐⭐ | ✅ |
| SSO / SCIM | ⭐⭐⭐ | ❌ out of scope (v5 ADR) |
| APM / traces | ⭐⭐⭐ | ✅ OTel + Prometheus + Loki |
| Secret rotation ops | — | ⚠️ PG-03 checklist (manual) |

---

## 4. Тренды 2025–2026 и ответ Hub

| Тренд | Ответ Hub |
|-------|-----------|
| AI-native workflows (Linear AI, Jira Rovo) | ✅ v3.0 AI gateway, @hub-ai, composer suggest |
| Consolidation fatigue (5+ SaaS) | ✅ v4 super-app — один инстанс |
| Automation without code | ✅ Phase 16 PROD-01 lite rules |
| Quality engineering in CI | ⚠️ Allure + hub-cli ✅; flaky ML — partial |
| Keyboard-first UX | ✅ Ctrl+K command palette |
| Compliance без SSO | ✅ Self-host + audit + Magic Link |
| Corp employee apps (Mighty model) | ✅ v4 intranet, HR, kudos, Random Coffee |

---

## 5. Top gaps (остаточные, 2026-06-21)

| # | Gap | Эталон | Статус Hub |
|---|-----|--------|------------|
| 1 | Flaky test ML-level detection | Allure TestOps | ⚠️ basic; PROD-02 partial |
| 2 | Issue ↔ message bidirectional sync | Linear+Slack | ⚠️ deep links ✅; full sync partial |
| 3 | Webhook delivery dashboard depth | Stripe, GitHub | ⚠️ API есть; admin UX polish |
| 4 | Timeline roadmap / initiatives | Linear, Jira AR | 📋 v4.5+ |
| 5 | Custom fields builder UI | Jira | 📋 schema ✅, UI нет |
| 6 | Notification digest | Slack | ⚠️ quiet hours ✅ |
| 7 | Lighthouse ≥90 CI gate | NFR-17 | 📋 not in CI |
| 8 | Full WCAG AA sign-off | NFR-16 | ⚠️ login PASS; full audit pending |
| 9 | Bot marketplace / plugins | Slack | ❌ by design (Telegram bind only) |
| 10 | SSO / SCIM | Enterprise | ❌ v5 ADR |

**Закрытые gaps (2026):** command palette, automation lite, full JQL, OTel, OBS stack, v4 corp modules, AI hub, Gantt, webcams, org directory, CRM, accounting MVP.

---

## 6. Top 10 differentiators (удвоить ставку)

1. **Единый self-hosted super-app** — dev + corp без 6–8 SaaS контрактов.
2. **Сквозная traceability** — задача → тест → баг → PR → invoice (CRM won).
3. **Allure без TestOps license** — парсер + TMS API + `hub-cli push-results`.
4. **Passwordless RU-friendly auth** — Magic Link + Telegram (без SSO complexity).
5. **LiveKit в контексте проекта** — созвон из канала/задачи, запись в storage.
6. **Optimistic locking + soft delete 30d** — enterprise patterns без enterprise price.
7. **Deep module integration** — thread→issue, failed case→bug, deal won→invoice.
8. **Command palette + user-scope search** — NF-03 compliant global search.
9. **Admin audit + ban + OBS stack** — governance + Prometheus/Loki/Alertmanager.
10. **Multi-model AI (GigaChat/YandexGPT/Ollama/BYOK)** — уникально для RU сегмента.

---

## 7. Стратегическая позиция

```
                    HIGH integration depth
                            │
         Allure TestOps     │     ★ Hub v4 (shipped)
         (TMS only)         │     dev + corp + AI
                            │
    LOW ────────────────────┼──────────────────── HIGH
    module count            │                    module count
                            │
         Linear             │     Atlassian Cloud
         (tracker only)     │     (expensive suite)
                            │
         Mighty              │     Bitrix24
         (no dev/TMS)       │     (weak dev)
                            │
                    LOW integration depth
```

**Рекомендация:** позиционировать Hub как **«Mighty для dev-команд»** — employee experience + полноценный dev/QA stack + self-host + AI hub. Не атаковать Jira enterprise head-on; закрывать automation depth и TMS analytics для удержания QA Lead.

---

## 8. Corporate super-apps — benchmark v4 (реализовано)

| Платформа | Фокус | Hub v4 ответ |
|-----------|-------|--------------|
| **Tinkoff Mighty** | Employee super-app | ✅ intranet, kudos, Random Coffee, directory |
| **Bitrix24** | SMB all-in-one | ✅ CRM-lite, HR leave, tasks; сильнее в dev/TMS |
| **Lark (Feishu)** | Global teams | ✅ bases (no-code), calendar; self-host vs cloud |
| **Microsoft Viva** | EX on Teams | ✅ directory, news; без M365 lock-in |
| **Plane** | OSS tracker | Hub шире: messenger + TMS + calls + corp |

**Уникальный pitch v4:** Mighty experience **+** dev/TMS depth **+** self-host **+** AI hub **+** 1C connector.

---

*Следующий документ: [user_needs_analysis_2026.md](./user_needs_analysis_2026.md) · Vision: [.planning/PLATFORM-VISION-v4.md](../.planning/PLATFORM-VISION-v4.md)*
