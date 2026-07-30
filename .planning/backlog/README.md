# Бэклог задач

> Единый вход для агентов и владельца. Одна таблица, один формат.
> Обновлено: 2026-07-30 · трек 1 августа заведён — 9 задач по фазам 128–136

## Состояние

| | |
|---|---|
| Активных задач | **9** — TASK-193…201, трек 1 августа |
| Закрыто | **192** |
| Архив | `archive/legacy/` 80 · `archive/2026-07-28/` 43 · `archive/2026-07-29/` 69 |

Новая задача заводится **плоско в корне** `backlog/`, по закрытии переезжает
в `archive/<дата заведения>/`. Подробности — в правиле 57.

`archive/legacy/` — задачи TASK-001…080, заведённые до введения даты в имени.
Не переименовывать: их пути стоят в телах коммитов.

Каждая задача — отдельный файл, написанный по
[мастер-промпту](../../prompt-engineering/agents/templates/_MASTER-PROMPT-TEMPLATE.md).
Правила ведения: [`.cursor/rules/57-backlog-tasks.mdc`](../../.cursor/rules/57-backlog-tasks.mdc)
(с 2026-07-28: **`TASK-<NNN>-<YYYY-MM-DD>-<функциональный-слаг>.md`**, emoji в заголовке/таблице).

Stage A3 dedup: [`.planning/audits/total-2026-07/A3-DEDUP-2026-07-29.md`](../audits/total-2026-07/A3-DEDUP-2026-07-29.md) —
**Owner lock: правим все 195** → TASK-124…176 (53 кластера, unmapped 0).

Stage A2 (brief): [A2-DEDUP-2026-07-28.md](../audits/total-2026-07/A2-DEDUP-2026-07-28.md) — **254 → TASK-081…123** (43 кластера, unmapped 0).

---

## Открытые — трек 1 августа, тестовый фреймворк

Девять задач по фазам 128–136. Порядок строгий: каждая следующая
опирается на предыдущую. Подробности —
[`../active/august-2026/04-REVISED-PRIORITIES.md`](../active/august-2026/04-REVISED-PRIORITIES.md).

| ID | Задача | Приоритет | Статус | Исполнитель | Источник |
|----|--------|-----------|--------|-------------|----------|
| [TASK-193](TASK-193-2026-07-30-tf-python-scaffold.md) | 🧪 Каркас Python-проекта /tests | P1 | Открыта | qa-auto1 | Трек 1 · фаза 128 |
| [TASK-194](TASK-194-2026-07-30-tf-e2e-shared-fixtures.md) | 🧪 Общие фикстуры E2E: авторизация и данные | P1 | Открыта | qa-auto1 | Трек 1 · фаза 129 |
| [TASK-195](TASK-195-2026-07-30-tf-e2e-mechanical.md) | 🧪 Перенос E2E: механическая часть | P1 | Открыта | qa-auto2 | Трек 1 · фаза 130 |
| [TASK-196](TASK-196-2026-07-30-tf-e2e-rewrite.md) | 🧪 Перенос E2E: сценарии с переработкой | P2 | Открыта | qa-auto2 | Трек 1 · фаза 131 |
| [TASK-197](TASK-197-2026-07-30-tf-backend-units-colocated.md) | 🧪 Go-юниты возвращаются к коду | P1 | Открыта | be1 + be2 | Трек 1 · фаза 132 |
| [TASK-198](TASK-198-2026-07-30-tf-frontend-units-colocated.md) | 🧪 Фронт-юниты: с нуля рядом с кодом | P1 | Открыта | fe1 + fe2 | Трек 1 · фаза 133 |
| [TASK-199](TASK-199-2026-07-30-tf-integration-python.md) | 🧪 Интеграционные тесты на Python | P2 | Открыта | qa-auto1 | Трек 1 · фаза 134 |
| [TASK-200](TASK-200-2026-07-30-tf-locust-load.md) | 🧪 Нагрузочные сценарии на Locust | P2 | Открыта | qa-auto2 | Трек 1 · фаза 135 |
| [TASK-201](TASK-201-2026-07-30-tf-ci-cleanup.md) | 🧹 CI и удаление старого фреймворка | P1 | Открыта | devops | Трек 1 · фаза 136 |

**Раскладка после миграции:**

```
backend/internal/<домен>/unit/   юниты на Go, рядом с кодом
frontend/src/<модуль>/unit/      юниты на TS, рядом с кодом
/tests/                          всё остальное — Python
    integration/  e2e/  api/  load/  full/
```

Юниты пишут разработчики на языке своего слоя, остальное —
тестировщики на Python.

---

## Закрытые
| [TASK-008](archive/legacy/TASK-008-total-audit-program.md) | Тотальный пофайловый аудит | P1 | Закрыта (W20 FINAL вердикт A · hubtest PASS 50 ok · 2026-07-30) | Grok + субагенты | Owner 2026-07-28 |
| [TASK-081](archive/2026-07-28/TASK-081-2026-07-28-tms-completerun-project-id-webhook.md) | ♻️🐛 TMS CompleteTestRun → webhook `project_id` | **P1** | Закрыта · MUTATION-2026-07-29-W3 | be1 | A2W4-5-001 REGRESS |
| [TASK-082](archive/2026-07-28/TASK-082-2026-07-28-fe-sprint-backlog-without-sprint-filter.md) | 🐛 FE SprintBacklog без фильтра sprint | **P1** | Закрыта · MUTATION-2026-07-29-W3 | fe1 | A2W10-A3-001 |
| [TASK-083](archive/2026-07-28/TASK-083-2026-07-28-fe-trash-hard-delete-api.md) | 🐛 FE Trash hard-delete без API | **P1** | Закрыта · MUTATION-2026-07-29-W3 | fe2 | A2W10-A5-001 |
| [TASK-084](archive/2026-07-28/TASK-084-2026-07-28-storage-confirm-entity-scope.md) | 🔒 Storage Confirm non-issue scope | **P1** | Закрыта · MUTATION-2026-07-29-W3 | be2 | W7-A4-002 |
| [TASK-085](archive/2026-07-28/TASK-085-2026-07-28-ai-task044-regress-bundle.md) | ♻️ AI TASK-044 false-close (RPM/JQL/stream) | P2 | Закрыта · MUTATION-2026-07-29-W4 | be1 | A2W5 REGRESS |
| [TASK-086](archive/2026-07-28/TASK-086-2026-07-28-webhook-delivery-ctx-id.md) | ♻️ Webhook ctx + X-Hub-Delivery | P2 | Закрыта · MUTATION-2026-07-29-W3 | be2 | A2W4-5-002/003 |
| [TASK-087](archive/2026-07-28/TASK-087-2026-07-28-outbox-claim-lease-flush.md) | 🐛 Outbox claim lease during flush | P2 | Закрыта · MUTATION-2026-07-30-W8-be2 | be1 | A2W4-2-001 NEW |
| [TASK-088](archive/2026-07-28/TASK-088-2026-07-28-tracker-assignee-team-tx-automation.md) | 🔒 Tracker assignee/team tx/automation Nak | P2 | Закрыта · MUTATION-2026-07-29-W4 | be1 | A2W1 |
| [TASK-089](archive/2026-07-28/TASK-089-2026-07-28-messenger-crm-invoice-guards.md) | 🔒 Messenger reaction + CRM + invoice | P2 | Закрыта | be2→qa-auto2 | A2W3/W7 · MUTATION-W8 |
| [TASK-090](archive/2026-07-28/TASK-090-2026-07-28-audit-purge-error-contract.md) | 🔒 Audit purge fail-closed + errors | P2 | Закрыта · MUTATION-2026-07-30-W9-be1 | be1 | A2W6 |
| [TASK-091](archive/2026-07-28/TASK-091-2026-07-28-tms-createbug-reports-batch.md) | 🐛 TMS createBug + reports N+1 | P2 | Закрыта · MUTATION-2026-07-30-W4-be1 | be1 | A2W2 |
| [TASK-092](archive/2026-07-28/TASK-092-2026-07-28-search-filter-status-safedial.md) | 🔒 Search 403 + Meili SafeDial | P2 | Закрыта · MUTATION-2026-07-29-W4 | be2 | A2W5-A4 |
| [TASK-093](archive/2026-07-28/TASK-093-2026-07-28-fe-a2-error-empty-a11y-sweep.md) | 🧹 FE A2 error≠empty + a11y | P2 | Закрыта · MUTATION-W13-fe1 + 008-retry (14 PROVEN) | fe1 | A2W9/W10/W11 |
| [TASK-094](archive/2026-07-28/TASK-094-2026-07-28-coffee-pairs-triple-edge.md) | ☕ Coffee Pairs[] triple-edges | P3 | Закрыта · MUTATION-2026-07-29-W4 | be2 | A2W7-2-N01 |
| [TASK-095](archive/2026-07-28/TASK-095-2026-07-28-ent-version-mixin-adr-rollout.md) | 📐 Ent VersionMixin ADR + rollout | **P1** | Закрыта · MUTATION-2026-07-29-W5 | arch→be1 | A2W8 · 19 rows |
| [TASK-096](archive/2026-07-28/TASK-096-2026-07-28-ent-ondelete-cascade-adr.md) | 📐 Ent OnDelete/cascade ADR | **P1** | Закрыта · MUTATION-2026-07-29-W6 | arch→be1 | A2W8 · 4 |
| [TASK-097](archive/2026-07-28/TASK-097-2026-07-28-ent-soft-delete-schema-gaps.md) | 📐 Ent soft-delete schema gaps | **P1** | Закрыта · MUTATION-2026-07-29-W6 | be2 | A2W8/W7 · 7 |
| [TASK-098](archive/2026-07-28/TASK-098-2026-07-28-intranet-author-delete-polls-adr.md) | 📐 Intranet author-delete + polls U03 | **P1** | Закрыта · MUTATION-2026-07-29-W6 | product→be1 | W7-A2-001 |
| [TASK-099](archive/2026-07-28/TASK-099-2026-07-28-storage-nf09-size-zero-adr.md) | 📐 Storage NF-09 size==0 | **P1** | Закрыта · MUTATION-2026-07-29-W3 | product→be2 | W7-A4-001 |
| [TASK-100](archive/2026-07-28/TASK-100-2026-07-28-backup-maintenance-gate-distributed-adr.md) | 📐 Backup MaintenanceGate U04 | **P1** | Закрыта · MUTATION-2026-07-29-W5 | devops→be1 | W7-A4-003 |
| [TASK-101](archive/2026-07-28/TASK-101-2026-07-28-ai-ollama-localhost-consent-adr.md) | 📐 AI Ollama localhost consent | **P1** | Закрыта · MUTATION-2026-07-29-W5 | product→be1 | A2W5-A2-005 |
| [TASK-102](archive/2026-07-28/TASK-102-2026-07-28-telegram-login-removal-adr-auth01.md) | 📐 Telegram login cleanup ADR-AUTH-01 | **P1** | Закрыта · MUTATION-2026-07-29-W4 | be1 | A2W6 · 7 |
| [TASK-103](archive/2026-07-28/TASK-103-2026-07-28-jwt-ws-handshake-adr.md) | 📐 JWT/WS handshake (no ?token=) | P2 | Закрыта · MUTATION-2026-07-29-W6 | arch→be+fe | A2W3/W11 · 3 |
| [TASK-104](archive/2026-07-28/TASK-104-2026-07-28-docs-cursor-pagination-adr.md) | 📐 Docs/API cursor vs offset | P2 | Закрыта · mutation N/A | sa→be+docs | A2W11 · 2 |
| [TASK-105](archive/2026-07-28/TASK-105-2026-07-28-docs-vision-g07-wiki-kpi-adr.md) | 📐 Docs G-07 wiki KPI | **P1** | Закрыта · mutation N/A | product→docs | A2W11-5-001 |
| [TASK-106](archive/2026-07-28/TASK-106-2026-07-28-intranet-coffee-dm-notify-adr.md) | 📐 Intranet FR-INTR-11 DM/notify U05 | P2 | Закрыта · MUTATION-2026-07-29-W5 | product→be2 | W7-A2-008 |
| [TASK-107](archive/2026-07-28/TASK-107-2026-07-28-ent-indexes-composite-fk-polymorphic.md) | 🔒 Ent indexes/FK/polymorphic | P2 | Закрыта | be1 · mut qa-auto1 | A2W8 · 8 · MUTATION-2026-07-30-W11-qa-auto1 |
| [TASK-108](archive/2026-07-28/TASK-108-2026-07-28-messenger-invite-dm-unread-scope.md) | 🔒 Messenger invite/DM/unread/bot | P2 | Закрыта · MUTATION-2026-07-29-W5 | be1 | A2W3 · 9 |
| [TASK-109](archive/2026-07-28/TASK-109-2026-07-28-ws-revoke-nats-publish.md) | 🔒 WS revoke → UnregisterBySession | P2 | Закрыта · MUTATION-2026-07-29-W3 | be1 | A2W3-4-002 |
| [TASK-110](archive/2026-07-28/TASK-110-2026-07-28-calls-recording-room-guest-code.md) | 📞 Calls recording/room/guest | P2 | Закрыта · MUTATION-2026-07-29-W5 | be2 | A2W3 · 3 |
| [TASK-111](archive/2026-07-28/TASK-111-2026-07-28-tms-wave2-services-import-batch.md) | 🐛 TMS W2 residual batch | P2 | Закрыта · MUTATION-2026-07-30-W13-be1-task111 | be2→be1 mut | A2W2 · 12/24 PROVEN · gaps documented |
| [TASK-112](archive/2026-07-28/TASK-112-2026-07-28-platform-outbox-nats-aes-hsts.md) | 🐛 Platform NATS/HSTS/AES/RL | P2 | Закрыта | be2 (W10 mut) | A2W4 · 9 · MUTATION-W10-be2 |
| [TASK-113](archive/2026-07-28/TASK-113-2026-07-28-ai-residual-sanitize-tests-policy.md) | 🔒 AI residual sanitize/tests | P2 | Закрыта · MUTATION-2026-07-29-W6 | be1 | A2W5 · 18 FIXED |
| [TASK-114](archive/2026-07-28/TASK-114-2026-07-28-audit-gdpr-limits-layering-tests.md) | 🔒 Audit/GDPR residual | P2 | Закрыта | be2 | A2W6 · 14 · MUTATION-2026-07-30-W11-be2 |
| [TASK-115](archive/2026-07-28/TASK-115-2026-07-28-rbac-permission-middleware-hygiene.md) | 🔒 RBAC permission hygiene | P2 | Закрыта · MUTATION-2026-07-29-W6 | be2 | A2W6 · 5 |
| [TASK-116](archive/2026-07-28/TASK-116-2026-07-28-auth-avatar-xss-validation-tests.md) | 🔒 Auth AvatarURL + session assert | P2 | Закрыта · MUTATION-2026-07-30-W8-be1 | be1 | A2W6 · 2 |
| [TASK-117](archive/2026-07-28/TASK-117-2026-07-28-notification-mail-push-limits-tests.md) | 📧 Notification/mail/push limits | P2 | Закрыта | be2 | W7-A1 · MUTATION-W9-be2 · 6/7 PROVEN · 005 N/A |
| [TASK-118](archive/2026-07-28/TASK-118-2026-07-28-v4-crm-hr-calendar-lists-trash.md) | 💼 v4 CRM/HR/calendar residual | P2 | Закрыта · MUTATION-2026-07-30-W10-be1 | be2→be1 mut | W7-A2 · 5/5 PROVEN |
| [TASK-119](archive/2026-07-28/TASK-119-2026-07-28-storage-backup-dashboard-residual.md) | 🔒 Storage/backup/dashboard residual | P2 | Закрыта · MUTATION-2026-07-30-W11-be1 | be1 | W7-A4 · 7/8 PROVEN · 008 N/A |
| [TASK-120](archive/2026-07-28/TASK-120-2026-07-28-integrations-secrets-imex-trash.md) | 🔌 Git*/1C/imex secrets+trash | P2 | Закрыта | be2 → qa-auto2 mut | W7-A3 · 4/4 PROVEN W10 |
| [TASK-121](archive/2026-07-28/TASK-121-2026-07-28-marketplace-status-dev-exposure.md) | 🛒 Marketplace/status/docs gates | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-task121 | be1→be2 mut | W7-A5 · 5/6 PROVEN · 001 N/A |
| [TASK-122](archive/2026-07-28/TASK-122-2026-07-28-tracker-project-jql-cleanup-links.md) | 🔒 Tracker JQL + DeleteProject links | P2 | Закрыта | be1 | A2W1 · 2 · MUTATION-W12-be1 |
| [TASK-123](archive/2026-07-28/TASK-123-2026-07-28-fe-lib-dead-export-docs-drift.md) | 🧹 FE dead export + docs drift | P2 | Закрыта | fe2+docs · mut fe1 W8 | A2W11 · 4 |
| [TASK-124](archive/2026-07-29/TASK-124-2026-07-29-a3-10-1-fe-routes-modules.md) | 🧹 A3 FE routes A1: accounting/marketplace/bases error≠empty + noopener | P3 | Закрыта | fe1 | A3W10-1 · 5 |
| [TASK-125](archive/2026-07-29/TASK-125-2026-07-29-a3-10-2-fe-routes-modules.md) | 🔒🧹 A3 FE routes A2: github/gitlab sanitizeUrl + canMutate CI | P2 | Закрыта · MUTATION-2026-07-30-W5-fe2 | fe1→fe2 mutate | A3W10-2 · 5 |
| [TASK-126](archive/2026-07-29/TASK-126-2026-07-29-a3-10-3-fe-sprint-backlog-regress.md) | ♻️🔒 A3 FE SprintBacklog filter REGRESS + sprint-store leaks | P1 | Закрыта · MUTATION-2026-07-29-W2-fe-p1 | fe1 | A3W10-3 · 4 |
| [TASK-127](archive/2026-07-29/TASK-127-2026-07-29-a3-10-4-fe-error-empty-settings.md) | 🧹 A3 FE routes A4: error≠empty messenger/project/TMS/settings | P2 | Закрыта | fe1→fe2 mut | A3W10-4 · 6 · MUTATION-2026-07-30-W6-fe2 |
| [TASK-128](archive/2026-07-29/TASK-128-2026-07-29-a3-10-5-fe-tms-onboarding-auth.md) | 🔒 A3 FE TMS canMutate + onboarding auth + login redirect | P2 | Закрыта · MUTATION-2026-07-30-W6-qa-auto1 | fe1→qa-auto1 | A3W10-5 · 5 |
| [TASK-129](archive/2026-07-29/TASK-129-2026-07-29-a3-11-1-fe-api-error-mapping.md) | ♻️🔒 A3 FE client.ts wire error→ApiError.message REGRESS | P1 | Закрыта · MUTATION-2026-07-29-W2-fe-p1 | fe1 | A3W11-1 · 1 |
| [TASK-130](archive/2026-07-29/TASK-130-2026-07-29-a3-11-2-fe-projects-sprint-version.md) | ♻️🔒 A3 FE projects/sprint API missing version (TASK-095) | P1 | Закрыта · MUTATION-2026-07-30-W3-fe2 | fe1 | A3W11-2 · 2 |
| [TASK-131](archive/2026-07-29/TASK-131-2026-07-29-a3-11-3-fe-stores-guards.md) | 🧹 A3 FE stores: loadLayout silent fail + reaction_toggled guard | P3 | Закрыта | fe1 | A3W11-3 · 2 |
| [TASK-132](archive/2026-07-29/TASK-132-2026-07-29-a3-11-4-docs-inventory-drift.md) | 🧹 A3 docs README inventory + webcam/competitive HEAD drift | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-docs-132-133 (N/A) | docs-writer | A3W11-4 · 3 |
| [TASK-133](archive/2026-07-29/TASK-133-2026-07-29-a3-11-5-docs-metrics-spec-offset.md) | 🧹 A3 docs inventory/strategy Vitest + list-issues offset | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-docs-132-133 (N/A) | docs-writer | A3W11-5 · 3 |
| [TASK-134](archive/2026-07-29/TASK-134-2026-07-29-a3-1-2-tracker-services.md) | 🔒 A3 tracker services: labels scope + residual batch | P1 | Закрыта · MUTATION-2026-07-30-W3-be2 | be1→be2 verify | A3W1-2 · 6 |
| [TASK-135](archive/2026-07-29/TASK-135-2026-07-29-a3-1-3-tracker-repos.md) | 🐛 A3 tracker repos/automation residual | P2 | Закрыта · MUTATION-2026-07-30-W5-be1 | be1 | A3W1-3 · 3 |
| [TASK-136](archive/2026-07-29/TASK-136-2026-07-29-a3-1-4-sprint.md) | 🐛 A3 sprint residual (CompleteSprint soft-delete etc) | P2 | Закрыта · MUTATION-2026-07-30-W4-be2 | be1→be2 verify | A3W1-4 · 3 |
| [TASK-137](archive/2026-07-29/TASK-137-2026-07-29-a3-1-5-project.md) | 🔒 A3 project archive/permanent-delete/resurrect residual | P2 | Закрыта · MUTATION-2026-07-30-W3-qa-auto1 | be2→qa-auto1 | A3W1-5 · 6 |
| [TASK-138](archive/2026-07-29/TASK-138-2026-07-29-a3-2-1-tms-handlers.md) | 🧹 A3 TMS handlers residual | P3 | Закрыта | be2 | A3W2-1 · 1 |
| [TASK-139](archive/2026-07-29/TASK-139-2026-07-29-a3-2-2-tms-services.md) | 🔒 A3 TMS services: AssignTesters/FileIDs/ImportAllure scope | P2 | Закрыта | be2 | A3W2-2 · 6 |
| [TASK-140](archive/2026-07-29/TASK-140-2026-07-29-a3-2-3-tms-repos.md) | 🐛 A3 TMS repos list caps residual | P2 | Закрыта · MUTATION-2026-07-30-W9-qa-auto1 | be2→qa-auto1 | A3W2-3 · 3 |
| [TASK-141](archive/2026-07-29/TASK-141-2026-07-29-a3-2-4-tms-import.md) | 🐛 A3 TMS import/export TestRail residual | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-task141 | be2 | A3W2-4 · 5 |
| [TASK-142](archive/2026-07-29/TASK-142-2026-07-29-a3-2-5-bases.md) | 🧹 A3 bases residual | P3 | Закрыта · MUTATION-2026-07-30-W13-be1-task142 | be2→be1 mutate | A3W2-5 · 3 · MUT 3/3 PROVEN |
| [TASK-143](archive/2026-07-29/TASK-143-2026-07-29-a3-3-1-messenger-handlers.md) | 🧹 A3 messenger handlers poll test wrong URL | P2 | Закрыта | be1 | A3W3-1 · 1 |
| [TASK-144](archive/2026-07-29/TASK-144-2026-07-29-a3-3-2-messenger-services.md) | 🔒 A3 messenger services Join/Create/Bookmark | P2 | Закрыта · MUTATION-2026-07-30-W5-be2 | be1→be2 verify | A3W3-2 · 6 |
| [TASK-145](archive/2026-07-29/TASK-145-2026-07-29-a3-3-3-messenger-repos.md) | 🐛 A3 messenger repos list caps/soft-delete | P2 | Закрыта · MUTATION-2026-07-30-W6-be1 | be1 | A3W3-3 · 3 |
| [TASK-146](archive/2026-07-29/TASK-146-2026-07-29-a3-3-4-ws.md) | 🧹 A3 WS query+ticket mutation gap | P3 | Закрыта · MUTATION-2026-07-30-W13-be2-task146 | be1→be2 | A3W3-4 · 1 · MUT 1/1 PROVEN |
| [TASK-147](archive/2026-07-29/TASK-147-2026-07-29-a3-3-5-calls.md) | 🐛 A3 calls LiveKit orphan + conflicts soft-delete | P2 | Закрыта | be2 | A3W3-5 · 4 · MUT 4/4 PROVEN |
| [TASK-148](archive/2026-07-29/TASK-148-2026-07-29-a3-4-1-platform-http.md) | 🧹 A3 platform Request-ID mutation test | P3 | Закрыта · MUTATION-2026-07-30-W13-be2-task148 | be1→be2 | A3W4-1 · 1 · MUT 1/1 PROVEN |
| [TASK-149](archive/2026-07-29/TASK-149-2026-07-29-a3-4-2-platform-nats-outbox.md) | 🔒 A3 outbox panic dead-letter + claim/Flush | P2 | Закрыта · MUTATION-2026-07-30-W5-qa-auto1 | be1→qa-auto1 | A3W4-2 · 3 |
| [TASK-150](archive/2026-07-29/TASK-150-2026-07-29-a3-4-3-platform-obs.md) | 🧹 A3 obs cancel→Error + UUID cardinality | P3 | Закрыта | be1 | A3W4-3 · 2 |
| [TASK-151](archive/2026-07-29/TASK-151-2026-07-29-a3-4-4-platform-core.md) | 🐛 A3 Postgres Ping + cron timeout test | P2 | Закрыта · MUTATION-2026-07-30-W13-qa1-task151 | be1→qa-auto1 | A3W4-4 · 2 |
| [TASK-152](archive/2026-07-29/TASK-152-2026-07-29-a3-4-5-apigateway.md) | 🔒 A3 apigateway soft-delete webhooks + VersionEQ | P2 | Закрыта · MUTATION-2026-07-30-W4-qa-auto1 | be2→qa-auto1 | A3W4-5 · 4 |
| [TASK-153](archive/2026-07-29/TASK-153-2026-07-29-a3-5-1-ai-handlers.md) | 🐛 A3 AI stream ErrNoProvider mapping | P2 | Закрыта · MUTATION-2026-07-30-W13-qa1-task153 | be1→qa-auto1 | A3W5-1 · 3 |
| [TASK-154](archive/2026-07-29/TASK-154-2026-07-29-a3-5-2-ai-adapters.md) | 🐛 A3 OfflineAI stream Timeout residual | P2 | Закрыта · MUTATION-2026-07-30-W13-qa1-task154 | be1→qa-auto1 | A3W5-2 · 3 |
| [TASK-155](archive/2026-07-29/TASK-155-2026-07-29-a3-5-3-ai-service.md) | ♻️🔒 A3 AI service REGRESS + consent/version/IDOR | P2 | Закрыта · MUTATION-2026-07-30-W13-qa1-task155 | be1→qa-auto1 | A3W5-3 · 9 |
| [TASK-156](archive/2026-07-29/TASK-156-2026-07-29-a3-5-4-search-core.md) | ♻️🔒 A3 Meili operator dial REGRESS + JQL | P2 | Закрыта · MUTATION-2026-07-30-W13-qa1-task156 | be2→qa-auto1 | A3W5-4 · 3 |
| [TASK-157](archive/2026-07-29/TASK-157-2026-07-29-a3-5-5-search-tests.md) | 🧹 A3 search tests limit clamp | P3 | Закрыта · MUTATION-2026-07-30-W13-be2-task157 | be2 | A3W5-5 · 001 PROVEN · 002 N/A |
| [TASK-158](archive/2026-07-29/TASK-158-2026-07-29-a3-6-2-auth-tests.md) | 🔒 A3 auth tests deactivated JWT + founder race | P2 | Закрыта | be1 | A3W6-2 · 4 |
| [TASK-159](archive/2026-07-29/TASK-159-2026-07-29-a3-6-3-rbac.md) | 🐛 A3 RBAC DeleteCustomRole cache-fail | P2 | Закрыта · MUTATION-2026-07-30-W13-qa1-task159 | be2→qa-auto1 | A3W6-3 · 3 |
| [TASK-160](archive/2026-07-29/TASK-160-2026-07-29-a3-6-4-audit-core.md) | 🔒 A3 GDPR Mongo eraser outside Ent tx | P2 | Закрыта · MUTATION-2026-07-30-W13-be1-task160 | be1 | A3W6-4 · 3 |
| [TASK-161](archive/2026-07-29/TASK-161-2026-07-29-a3-6-5-audit-tests.md) | 🧹 A3 audit tests CRM purge/Retention gaps | P2 | Закрыта | be1 | A3W6-5 · 7 |
| [TASK-162](archive/2026-07-29/TASK-162-2026-07-29-a3-7-1-notification-mail.md) | 🐛 A3 mail DailyLimit distributed | P3 | Закрыта · MUTATION-2026-07-30-W13-be2-task162 | be2 | A3W7-1 · 2 |
| [TASK-163](archive/2026-07-29/TASK-163-2026-07-29-a3-7-2-v4-business.md) | ♻️ A3 coffee notify REGRESS + poll VersionEQ | P2 | Закрыта | be2 | A3W7-2 · 3 |
| [TASK-164](archive/2026-07-29/TASK-164-2026-07-29-a3-7-3-integrations.md) | 🔒 A3 integrations trash filter + link-preview membership | P2 | Закрыта | be2 | A3W7-3 · 3 |
| [TASK-165](archive/2026-07-29/TASK-165-2026-07-29-a3-7-4-storage-dash-backup.md) | 🐛 A3 storage/dashboard/backup residual | P2 | Закрыта | be1 | A3W7-4 · 5 |
| [TASK-166](archive/2026-07-29/TASK-166-2026-07-29-a3-7-5-mp-status-dev.md) | 🔒 A3 marketplace review_note public catalog | P2 | Закрыта | be1 | A3W7-5 · 4 · mutation PROVEN qa1 |
| [TASK-167](archive/2026-07-29/TASK-167-2026-07-29-a3-8-1-schemas.md) | 🧹 A3 schemas SetNull annotation + ChannelMember | P3 | Закрыта · MUTATION-2026-07-30-W13-be2-task167 | be1→be2 mutate | A3W8-1 · 001 PROVEN · 002 N/A |
| [TASK-168](archive/2026-07-29/TASK-168-2026-07-29-a3-8-2-schemas.md) | 🐛 A3 schemas OnDelete Cascade batch (coffee/hub/integrations) | P2 | Закрыта | be1 | A3W8-2 · 6 |
| [TASK-169](archive/2026-07-29/TASK-169-2026-07-29-a3-8-3-schemas.md) | 🐛 A3 HubTable/JournalEntry Cascade | P2 | Закрыта | be1 | A3W8-3 · 2 |
| [TASK-170](archive/2026-07-29/TASK-170-2026-07-29-a3-8-4-schemas.md) | 🐛 A3 sprint active unique + SetNull issues | P2 | Закрыта · MUTATION-2026-07-30-W13-be1-task170 | be1 | A3W8-4 · 2/2 PROVEN |
| [TASK-171](archive/2026-07-29/TASK-171-2026-07-29-a3-8-5-schemas.md) | 🔒 A3 Team/WorkflowStatus soft-delete + OnDelete Cascades | P1 | Закрыта · MUTATION-2026-07-30-W3-be1 | be1 verify | A3W8-5 · 7 |
| [TASK-172](archive/2026-07-29/TASK-172-2026-07-29-a3-9-1-components.md) | 🧹 A3 FE admin error≠empty pending/limits/trash | P2 | Закрыта · MUTATION-2026-07-30-W9-fe2 | fe1→fe2 mutate | A3W9-1 · 5 |
| [TASK-173](archive/2026-07-29/TASK-173-2026-07-29-a3-9-2-components.md) | 🧹 A3 FE bookmarks UUID + deadlines links | P2 | Закрыта | fe1→fe2 | A3W9-2 · 5 · W11 PROVEN FE |
| [TASK-174](archive/2026-07-29/TASK-174-2026-07-29-a3-9-3-components.md) | 🔒 A3 FE LinkPreview XSS whitelist | P2 | Закрыта | fe1 | A3W9-3 · 4 · mut 001–003 PROVEN · 004 N/A |
| [TASK-175](archive/2026-07-29/TASK-175-2026-07-29-a3-9-4-components.md) | 🧹 A3 FE checklist/automation/comment error UI | P2 | Закрыта · MUTATION-2026-07-30-W10-fe1 | fe2→fe1 mutate | A3W9-4 · 6/6 PROVEN |
| [TASK-176](archive/2026-07-29/TASK-176-2026-07-29-a3-9-5-components.md) | 🧹 A3 FE avatar/popover/select hub tokens | P3 | Закрыта · MUTATION-2026-07-30-W13-fe1-task176 | fe2→fe1 mutate | A3W9-5 · 3/3 PROVEN |
| [TASK-177](archive/2026-07-29/TASK-177-2026-07-29-c-w19-deactivate-ban-ws-kill.md) | 🔒 C-W19 Deactivate/Ban не рвёт живой WS | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-task177 | be1→be2 mutate | C-W19-001 · 2/2 PROVEN |
| [TASK-178](archive/2026-07-29/TASK-178-2026-07-29-c-w19-gdpr-mongo-erase-retry.md) | 🔒 C-W19 GDPR Mongo erase retry после Commit | P2 | Закрыта · MUTATION-2026-07-30-W13-be1-task178 | be1 | C-W19-002 · 1/1 PROVEN |
| [TASK-179](archive/2026-07-29/TASK-179-2026-07-29-c-w19-fe-webhook-sprint-completed.md) | 🐛 C-W19 FE sprint.completed после BE trim | P2 | Закрыта | fe1 · qa-auto2 mut | C-W19-FE-001 · 1 |
| [TASK-180](archive/2026-07-29/TASK-180-2026-07-29-c-w19-link-preview-modules-tc.md) | 🔒 C-W19 modules LinkPreview XSS + TC membership | P2 | Закрыта | fe1+be2 | C-W19-FE-002 · MUTATION-W12-fe2 · FE PROVEN · BE N/A |
| [TASK-181](archive/2026-07-29/TASK-181-2026-07-29-c-w19-webhook-softdelete-deactivate.md) | 🔒 C-W19 soft-delete project гасит webhooks | P2 | Закрыта · MUTATION-2026-07-30-W13-be1-task181 | be2→be1 mutate | C-W19-FE-004 · 2/2 PROVEN |
| [TASK-182](archive/2026-07-29/TASK-182-2026-07-29-c-w19-hr-pending-leave-error.md) | 🧹 C-W19 HR pending leave error≠empty | P3 | Закрыта | fe2 · qa-auto1 mut | C-W19-FE-005 · 1 |
| [TASK-183](archive/2026-07-29/TASK-183-2026-07-29-c-w19-docs-inventory-drift.md) | 🧹 C-W19 docs inventory + AGENTS/spec drift | P2 | Закрыта · MUTATION-2026-07-30-W13-be1-docs-183 (N/A) | docs-writer→be1 | C-W19-DOC-001 · N/A docs-only |
| [TASK-184](archive/2026-07-29/TASK-184-2026-07-29-c-w19-last-team-member.md) | 🐛 C-W19 RemoveTeamMember last-member orphan | P3 | Закрыта · MUTATION-2026-07-30-W13-qa1-task184-fix | qa-auto1 | C-W19-TT-001 · 1/1 PROVEN |
| [TASK-185](archive/2026-07-29/TASK-185-2026-07-29-c-w19-allure-existingrun-partial.md) | 🐛 C-W19 Allure ExistingRun silent partial | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-task185 | be2 | C-W19-TT-002 · 2/2 PROVEN |
| [TASK-186](archive/2026-07-29/TASK-186-2026-07-29-audit-fe-calls-error-empty.md) | 🧹 Audit FE calls calendar/dialog error≠empty | P2 | Закрыта | fe1 (mut W9) / fe2 | FE-MC-001…006 · 5 PROVEN + 005 N/A |
| [TASK-187](archive/2026-07-29/TASK-187-2026-07-29-audit-fe-pipeline-sanitize.md) | 🔒 Audit FE pipeline-status sanitizeUrl | P2 | Закрыта | fe1→fe2 mut W13 | FE-TA-001…005 · 5/5 PROVEN · MUTATION-W13-fe2-task187 |
| [TASK-188](archive/2026-07-29/TASK-188-2026-07-29-audit-be-push-telegram-rl.md) | 🔒 Audit BE PushSender layering + Telegram RL | P2 | Закрыта | be2 | ASN-001…002 · 2/2 PROVEN · W13 be1 |
| [TASK-189](archive/2026-07-29/TASK-189-2026-07-29-audit-be-v4-calendar-poll.md) | 🔒 Audit BE calendar trash + draft poll IDOR | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-task189 | be1→be2 mut | BE-V4-001…003 · 3/3 PROVEN |
| [TASK-190](archive/2026-07-29/TASK-190-2026-07-29-audit-be-tracker-messenger-softdelete.md) | 🔒 Audit BE tracker/messenger soft-delete + CreateChannel | P2 | Закрыта · MUTATION-2026-07-30-W13-be1-task190 | be1 | AUDIT-TM-001…004 · 4/4 PROVEN (+004 GAP_NOTED) |
| [TASK-191](archive/2026-07-29/TASK-191-2026-07-29-audit-ent-automationrule-softdelete.md) | 🔒 Audit ent AutomationRule soft-delete | P2 | Закрыта | be1 | EA-001 · 1 |
| [TASK-192](archive/2026-07-29/TASK-192-2026-07-29-audit-fe-settings-version-ol.md) | 🔒 Audit FE settings version OL (webhooks/AI) | P2 | Закрыта | fe1→fe2 mut | FE-SMA-001…003 · 3/3 PROVEN W10 |
| [TASK-009](archive/legacy/TASK-009-project-routes-rbac.md) | project routes без RBAC | **P0** | Закрыта | be1 | W1-A5-001 |
| [TASK-010](archive/legacy/TASK-010-promote-to-issue-rbac.md) | promote-to-issue без PermTaskCreate | **P0** | Закрыта | be1 | W3-A2-001 |
| [TASK-011](archive/legacy/TASK-011-webhook-event-match.md) | webhook UI events ≠ NATS subjects | **P0** | Закрыта | be1 | W4-A5-001 |
| [TASK-012](archive/legacy/TASK-012-webhook-project-scope.md) | webhook dispatch без project-scope | **P0** | Закрыта | be1 | W4-A5-002 |
| [TASK-013](archive/legacy/TASK-013-tracker-project-scope-idor.md) | Tracker IDOR workflow/fields/refs | P1 | Закрыта | be1 | W12 P1-BE-01..04 |
| [TASK-014](archive/legacy/TASK-014-project-acl-sprint-tx.md) | Project ACL + sprint atomic | P1 | Закрыта · MUTATION-2026-07-29-W1-be1 | be2→rework · W1-be1 PASS | W12 P1-BE-05..09 · MUTATION-2026-07-29-W1-be1 |
| [TASK-015](archive/legacy/TASK-015-tms-project-scope-idor-refs.md) | TMS suite/env IDOR + FK | P1 | Закрыта | be2 | W12 P1-BE-10..12 · W18-C2 PASS |
| [TASK-016](archive/legacy/TASK-016-tms-import-limits-steps.md) | TMS import limits + steps | P1 | Закрыта | be2 | W12 P1-BE-13..14 · W18-C2 PASS |
| [TASK-017](archive/legacy/TASK-017-messenger-membership-idor.md) | Messenger membership IDOR | P1 | Закрыта | be1 | W12 P1-BE-15..20 |
| [TASK-018](archive/legacy/TASK-018-ws-calls-authz-race.md) | WS race + scheduled calls authz | P1 | Закрыта | be1 | W12 P1-BE-21..22 |
| [TASK-019](archive/legacy/TASK-019-platform-ssrf-outbox-pool-config.md) | Platform SSRF/outbox/pool/config | P1 | Закрыта · MUTATION-2026-07-29-W1-be1 | be2→rework · W1-be1 PASS | W12 P1-BE-23..26 · MUTATION-2026-07-29-W1-be1 |
| [TASK-020](archive/legacy/TASK-020-apigateway-webhook-create-contract.md) | Webhook secret once + allowlist | P1 | Закрыта | be2 | W12 P1-BE-27..28 · W18-C2 PASS |
| [TASK-021](archive/legacy/TASK-021-ai-policy-bot-authz.md) | AI policy + bot authz | P1 | Закрыта | be1 | W12 P1-BE-29..31 |
| [TASK-022](archive/legacy/TASK-022-search-nf03-softdelete.md) | Search soft-delete + private | P1 | Закрыта | be2 | W12 P1-BE-32 · W18-C2 PASS |
| [TASK-023](archive/legacy/TASK-023-rbac-cache-gdpr.md) | RBAC cache + GDPR erase | P1 | Закрыта | be1 | W12 P1-BE-33..34 · W18-C3 PASS |
| [TASK-024](archive/legacy/TASK-024-notif-integrations-storage-backup.md) | Notif/Git*/storage/backup/coffee | P1 | Закрыта | be2 | W12 P1-BE-35..39,43 · W18-C2 PASS |
| [TASK-025](archive/legacy/TASK-025-ent-schema-constraints.md) | Ent unique/O2O constraints | P1 | Закрыта | be1 | W12 P1-BE-40..42 · W18-C3 PASS |
| [TASK-026](archive/legacy/TASK-026-fe-api-client-auth-retry.md) | FE api client auth retry | P1 | Закрыта | fe1 | W12 FE-P1-API-CLIENT |
| [TASK-027](archive/legacy/TASK-027-fe-callchat-ws-subscribe.md) | FE CallChat WS subscribe | P1 | Закрыта | fe1 | W12 FE-P1-CALLS-CHAT |
| [TASK-028](archive/legacy/TASK-028-fe-calls-join-create-lifecycle.md) | FE calls lifecycle | P1 | Закрыта | fe1 | W12 FE-P1-CALLS-LIFECYCLE |
| [TASK-029](archive/legacy/TASK-029-fe-messenger-reply-thread-unread.md) | FE messenger reply/unread | P1 | Закрыта | fe1 | W12 FE-P1-MESSENGER |
| [TASK-030](archive/legacy/TASK-030-fe-board-dnd-single-path-rank.md) | FE board DnD + rank | P1 | Закрыта | fe1 | W12 FE-P1-BOARD-DND |
| [TASK-031](archive/legacy/TASK-031-fe-sprint-loading-error-states.md) | FE sprint loading/error | P1 | Закрыта | fe1 | W12 FE-P1-SPRINT-UI |
| [TASK-032](archive/legacy/TASK-032-fe-trash-list-and-webhooks-project.md) | FE trash + webhooks projectId | P1 | Закрыта | fe2 | W12 FE-P1-TRASH-WEBHOOKS |
| [TASK-033](archive/legacy/TASK-033-fe-error-empty-auth-race.md) | FE error≠empty + auth race | P1 | Закрыта | fe2 | W12 FE-P1-ERROR-AS-EMPTY |
| [TASK-034](archive/legacy/TASK-034-fe-a11y-nested-sidebar-notifications.md) | FE a11y nested interactive | P1 | Закрыта | fe2 | W12 FE-P1-A11Y-CONTROLS |
| [TASK-035](archive/legacy/TASK-035-fe-tms-compare-and-docs-testpath.md) | FE TMS compare + docs drift | P1 | Закрыта | fe1 | W12 FE-P1-TMS-DOCS |
| [TASK-036](archive/legacy/TASK-036-be-p2-dto-error-contract.md) | BE DTO validator + error contract | P2 | Закрыта | be2 | W12 P2-BE-01,07 |
| [TASK-037](archive/legacy/TASK-037-be-p2-authz-ent-negative-tests.md) | BE Ent-negative authz tests | P2 | Закрыта | be2 | W12 P2-BE-02 |
| [TASK-038](archive/legacy/TASK-038-be-p2-ol-softdelete-tx-lists.md) | BE OL/soft-mutate/tx/lists | P2 | Закрыта | be1 | W12 P2-BE-03..06 · W18-C3 PASS (partial residual OK) |
| [TASK-039](archive/legacy/TASK-039-be-p2-soft-delete-secrets.md) | BE soft-delete + secrets | P2 | Закрыта | be2 | W12 P2-BE-08,09 |
| [TASK-040](archive/legacy/TASK-040-be-p2-fail-closed-ratelimit.md) | BE fail-closed + rate-limit | P2 | Закрыта | be1 | W12 P2-BE-10 · W18-C3 PASS |
| [TASK-041](archive/legacy/TASK-041-be-p2-sprint-poker-start.md) | BE sprint poker/start | P2 | Закрыта · MUTATION-2026-07-30-W13-be1-task041 | be1 | W12 P2-BE-11 · 3 PROVEN + 007 GAP_NOTED |
| [TASK-042](archive/legacy/TASK-042-be-p2-perm-messenger-ws-calls.md) | BE perm/messenger/ws/calls | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-task042 | be2 | W12 P2-BE-12..15 · 4 PROVEN + RL GAP_NOTED |
| [TASK-043](archive/legacy/TASK-043-be-p2-platform-outbox-obs.md) | BE platform outbox/obs | P2 | Закрыта | be2 | W12 P2-BE-16 |
| [TASK-044](archive/legacy/TASK-044-be-p2-ai-search.md) | AI stream/privacy + search JQL/SSRF | P2 | Закрыта · MUTATION-2026-07-29-W2-044 | be1→rework · W2-044 PASS | W12 P2-BE-17/18 · Meili+cancel IT |
| [TASK-045](archive/legacy/TASK-045-be-p2-rbac-gdpr-integrations-storage.md) | BE rbac/gdpr/imex/storage | P2 | Закрыта · MUTATION-2026-07-29-W1-be1 | be2→rework · W1-be1 PASS | W12 P2-BE-19..21 · MUTATION-2026-07-29-W1-be1 |
| [TASK-046](archive/legacy/TASK-046-be-p2-ent-unique-fk-check.md) | BE Ent unique/FK/CHECK | P2 | Закрыта · MUTATION-2026-07-30-W13-qa1-task046 | be1→qa-auto1 | W12 P2-BE-22 · 2 PROVEN + 2 N/A |
| [TASK-047](archive/legacy/TASK-047-fe-p2-error-vs-empty.md) | FE error≠empty sweep | P2 | Закрыта | fe1 | W12 P2-FE-01 |
| [TASK-048](archive/legacy/TASK-048-fe-p2-mutation-feedback.md) | FE mutation feedback | P2 | Закрыта | fe2 | W12 P2-FE-02 |
| [TASK-049](archive/legacy/TASK-049-fe-p2-a11y-names-dialogs.md) | FE a11y names/dialogs | P2 | Закрыта | fe1 | W12 P2-FE-03,04 |
| [TASK-050](archive/legacy/TASK-050-fe-p2-tokens-dead-ui.md) | FE tokens/dead UI/authz | P2 | Закрыта | fe1 | W12 P2-FE-05,06,16 |
| [TASK-051](archive/legacy/TASK-051-fe-p2-effects-hooks-race.md) | FE effects/hooks race | P2 | Закрыта | fe2 | W12 P2-FE-07 |
| [TASK-052](archive/legacy/TASK-052-fe-p2-api-dto-gaps.md) | FE API DTO gaps | P2 | Закрыта · MUTATION-2026-07-30-W13-fe2 | fe2 | W12 P2-FE-08 · 10/10 PROVEN |
| [TASK-053](archive/legacy/TASK-053-fe-p2-auth-ws-upload.md) | FE auth/WS/upload | P2 | Закрыта | fe2 | W12 P2-FE-09..11 |
| [TASK-054](archive/legacy/TASK-054-fe-p2-docs-metrics-drift.md) | Docs metrics drift | P2 | Закрыта · MUTATION-2026-07-30-W13-fe1-task054 (N/A docs-only) | docs-writer | W12 P2-FE-12 · TASK-080 follow-up |
| [TASK-055](archive/legacy/TASK-055-fe-p2-shell-tracker-ux.md) | FE shell + tracker UX | P2 | Закрыта | fe2 | W12 P2-FE-13,14 |
| [TASK-056](archive/legacy/TASK-056-fe-p2-i18n-push-rooms.md) | FE i18n/push/rooms | P2 | Закрыта | fe2 | W12 P2-FE-15 |
| [TASK-057](archive/legacy/TASK-057-be-p3-silent-error-swallow.md) | BE silent error swallow | P3 | Закрыта | be1 | W12 P3-BE-01 |
| [TASK-058](archive/legacy/TASK-058-be-p3-assert-free-tests.md) | BE assert-free tests | P3 | Закрыта · MUTATION-2026-07-30-W13-be1-task058 | be1 | W12 P3-BE-02 · 17/17 PROVEN |
| [TASK-059](archive/legacy/TASK-059-be-p3-mutation-proof-gaps.md) | BE mutation-proof gaps | P3 | Закрыта · MUTATION-2026-07-30-W13-be2-task059 | be2 | W12 P3-BE-03 · 9/10 PROVEN · W1-A4-015 GAP_NOTED |
| [TASK-060](archive/legacy/TASK-060-be-p3-utf8-rune-truncate.md) | BE UTF-8 rune truncate | P3 | Закрыта | be2 | W12 P3-BE-04 |
| [TASK-061](archive/legacy/TASK-061-be-p3-repo-lists-defense.md) | BE repo/lists defense | P3 | Закрыта · MUTATION-2026-07-30-W13-be2-task061 | be2 | W12 P3-BE-05,06 · 11/11 PROVEN |
| [TASK-062](archive/legacy/TASK-062-be-p3-dead-code-layering.md) | BE dead code/layering | P3 | Закрыта · MUTATION-2026-07-30-W13-be2-task062 · 7 PROVEN · 6 GAP · 1 N/A | be2 | W12 P3-BE-07 |
| [TASK-063](archive/legacy/TASK-063-be-p3-errors-input-hygiene.md) | BE errors + input hygiene | P3 | Закрыта | be1 | W12 P3-BE-08,09 |
| [TASK-064](archive/legacy/TASK-064-be-p3-ws-calls-lifecycle.md) | BE WS/calls lifecycle | P3 | Закрыта · MUTATION-2026-07-30-W13-qa1-task064 | be2→qa-auto1 | W12 P3-BE-10 · 5 PROVEN + 1 N/A |
| [TASK-065](archive/legacy/TASK-065-be-p3-platform-http-obs.md) | BE platform HTTP/obs | P3 | Закрыта | be1 | W12 P3-BE-11,12 |
| [TASK-066](archive/legacy/TASK-066-be-p3-ai-auth-rbac-audit.md) | BE AI/auth/rbac/audit | P3 | Закрыта · MUTATION-2026-07-30-W13-be1-task066-retry (10/10 PROVEN) | be1 (mutation) | W12 P3-BE-13,14 · W6-A4-010 PROVEN |
| [TASK-067](archive/legacy/TASK-067-be-p3-v4-dev-storage.md) | BE v4/dev/storage | P3 | Закрыта | be1 | W12 P3-BE-15,16 |
| [TASK-068](archive/legacy/TASK-068-be-p3-ent-hygiene-mojibake.md) | BE Ent hygiene/mojibake | P3 | Закрыта · MUTATION-2026-07-30-W13-qa1-task068 (11/11 PROVEN) | be2→qa-auto1 | W12 P3-BE-17,18 |
| [TASK-069](archive/legacy/TASK-069-fe-p3-a11y-names.md) | FE a11y names residual | P3 | Закрыта · C: 6/6 PROVEN (W13 fe1 retry) | fe1 | W12 P3-FE-01 |
| [TASK-070](archive/legacy/TASK-070-fe-p3-a11y-structure.md) | FE a11y structure | P3 | Закрыта · C: 7/7 PROVEN (W13 fe2 retry) | fe2 | W12 P3-FE-02 |
| [TASK-071](archive/legacy/TASK-071-fe-p3-tokens-i18n.md) | FE tokens + i18n | P3 | Закрыта · MUTATION-2026-07-30-W13-fe1-task071 (9/9 PROVEN) | fe1 | W12 P3-FE-03,04 |
| [TASK-072](archive/legacy/TASK-072-fe-p3-timers-cleanup.md) | FE timers + NaN | P3 | Закрыта | fe1 | W12 P3-FE-05,16 · C: 4/4 PROVEN (W13 fe1 retry) |
| [TASK-073](archive/legacy/TASK-073-fe-p3-silent-fail-empty.md) | FE silent fail residual | P3 | Закрыта · C: 5/5 PROVEN (W13 fe2) | fe2 | W12 P3-FE-06 |
| [TASK-074](archive/legacy/TASK-074-fe-p3-type-casts.md) | FE type casts | P3 | Закрыта · C: 8/8 PROVEN (W13 fe2 retry) | fe2 | W12 P3-FE-07 |
| [TASK-075](archive/legacy/TASK-075-fe-p3-dead-ui-state.md) | FE dead UI/state | P3 | Закрыта · C: 11/11 PROVEN (W13 fe1 retry) | fe1 | W12 P3-FE-08 |
| [TASK-076](archive/legacy/TASK-076-fe-p3-perf-stale.md) | FE perf stale | P3 | Закрыта | fe2 | W12 P3-FE-09 · C W13 3/3 PROVEN |
| [TASK-077](archive/legacy/TASK-077-fe-p3-api-auth-naming.md) | FE api/auth/naming | P3 | Закрыта · MUTATION-2026-07-30-W13-fe2-task077 | fe1→fe2 | W12 P3-FE-10..12 · 6/6 PROVEN |
| [TASK-078](archive/legacy/TASK-078-fe-p3-docs-spa-ux.md) | FE docs/SPA/preview | P3 | Закрыта · C: 4/4 PROVEN + 3 N/A (W13 fe1) | fe1 | W12 P3-FE-13..15 |
| [TASK-079](archive/legacy/TASK-079-gdpr-mongo-eraser-fail-closed.md) | GDPR Mongo eraser fail-closed + IT | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-task079 | be2 | W18-C5-001/002 |
| [TASK-080](archive/legacy/TASK-080-docs-metrics-drift-autotest.md) | Docs metrics drift autotest | P2 | Закрыта · MUTATION-2026-07-30-W13-be2-task079-080 (N/A docs/script) | docs-writer | W18-C5 / closes 054 gate |

## Закрытые

| ID | Задача | Приоритет | Статус | Исполнитель | Источник |
|----|--------|-----------|--------|-------------|----------|
| [TASK-001](archive/legacy/TASK-001-fr-intr-08-newspost-trash.md) | Новость невосстановима из корзины | **P0** | Закрыта | be1 | FR-INTR-08 |
| [TASK-002](archive/legacy/TASK-002-fr-cal-05-series-trash.md) | У календарных серий нет корзины | **P0** | Закрыта | be1 | FR-CAL-05 |
| [TASK-004](archive/legacy/TASK-004-state-goals-verified-reconcile.md) | STATE расходится с матрицей фаз | P1 | Закрыта | team-lead | Ревью W4–W20 |
| [TASK-005](archive/legacy/TASK-005-august-phase-dependencies.md) | Связать августовские фазы зависимостями | P1 | Закрыта | team-lead | Аудит планирования |
| [TASK-003](archive/legacy/TASK-003-a14-002-test-granularity.md) | Тест пропускает удаление одного guard'а | P2 | Закрыта | qa-auto1 | Ревью W4–W20 |
| [TASK-006](archive/legacy/TASK-006-disable-cursor-coauthor-attribution.md) | Навсегда выключить Cursor Co-authored-by | **P0** | Закрыта | devops/docs | Owner 2026-07-28 |
| [TASK-007](archive/legacy/TASK-007-utf8-mojibake-audit-fix.md) | Аудит и починка битой UTF-8 кодировки | P1 | Закрыта | docs+be1 | Owner 2026-07-28 |

---

## Порядок выполнения (TASK-008)

```
P0:  TASK-009 → 010 → 011+012 (webhooks вместе)
P1:  TASK-013…025 (BE IDOR/authz раньше schema)
     TASK-026…035 (FE)
P2/P3: после P1; реестр сырых — audits/total-2026-07/
Blocked ADR: W12-REGISTRY-P1-BACKEND §«требует уточнения» (U01–U06)
```

---

## Легенда

**Приоритет:** P0 блокер · P1 важно · P2 можно отложить · P3 косметика

**Статус:** `Открыта` → `В работе` → `На проверке` → `Закрыта` / `Отменена`

Статус меняется **в том же коммите**, что и работа по задаче.

---

## Что не сюда

| Что | Где |
|-----|-----|
| Крупный блок работ | `.planning/phases/NN-*/PLAN.md` (GSD) |
| Идея без срока | `ROADMAP.md`, секция `## Backlog`, нумерация `999.x` |
| Решение, ждущее владельца | [`../audits/planning-audit-2026-07/DECISIONS-PENDING.md`](../audits/planning-audit-2026-07/DECISIONS-PENDING.md) |
| Находка аудита без постановки | `.planning/audits/<аудит>/` |
