# Бэклог задач

> Единый вход для агентов и владельца. Одна папка, одна таблица, один формат.
> Обновлено: 2026-07-29 (A3: **все 195 → TASK-124…176** · 53 кластера · naming `TASK-NNN-YYYY-MM-DD-slug`)

Каждая задача — отдельный файл, написанный по
[мастер-промпту](../../prompt-engineering/agents/templates/_MASTER-PROMPT-TEMPLATE.md).
Правила ведения: [`.cursor/rules/57-backlog-tasks.mdc`](../../.cursor/rules/57-backlog-tasks.mdc)
(с 2026-07-28: **`TASK-<NNN>-<YYYY-MM-DD>-<функциональный-слаг>.md`**, emoji в заголовке/таблице).

Stage A3 dedup: [`.planning/audits/total-2026-07/A3-DEDUP-2026-07-29.md`](../audits/total-2026-07/A3-DEDUP-2026-07-29.md) —
**Owner lock: правим все 195** → TASK-124…176 (53 кластера, unmapped 0).

Stage A2 (brief): [A2-DEDUP-2026-07-28.md](../audits/total-2026-07/A2-DEDUP-2026-07-28.md) — **254 → TASK-081…123** (43 кластера, unmapped 0).

---

## Открытые

| ID | Задача | Приоритет | Статус | Исполнитель | Источник |
|----|--------|-----------|--------|-------------|----------|
| [TASK-008](TASK-008-total-audit-program.md) | Тотальный пофайловый аудит | P1 | В работе (A✅ · A2✅ · A3✅ 195→TASK-124…176 · B P0–P3✅ · не закрывать) | Grok + субагенты | Owner 2026-07-28 |
| [TASK-081](TASK-081-2026-07-28-tms-completerun-project-id-webhook.md) | ♻️🐛 TMS CompleteTestRun → webhook `project_id` | **P1** | Исправлена | be1 | A2W4-5-001 REGRESS |
| [TASK-082](TASK-082-2026-07-28-fe-sprint-backlog-without-sprint-filter.md) | 🐛 FE SprintBacklog без фильтра sprint | **P1** | Исправлена | fe1 | A2W10-A3-001 |
| [TASK-083](TASK-083-2026-07-28-fe-trash-hard-delete-api.md) | 🐛 FE Trash hard-delete без API | **P1** | Исправлена | fe2 | A2W10-A5-001 |
| [TASK-084](TASK-084-2026-07-28-storage-confirm-entity-scope.md) | 🔒 Storage Confirm non-issue scope | **P1** | Исправлена | be2 | W7-A4-002 |
| [TASK-085](TASK-085-2026-07-28-ai-task044-regress-bundle.md) | ♻️ AI TASK-044 false-close (RPM/JQL/stream) | P2 | Исправлена | be1 | A2W5 REGRESS |
| [TASK-086](TASK-086-2026-07-28-webhook-delivery-ctx-id.md) | ♻️ Webhook ctx + X-Hub-Delivery | P2 | Исправлена | be2 | A2W4-5-002/003 |
| [TASK-087](TASK-087-2026-07-28-outbox-claim-lease-flush.md) | 🐛 Outbox claim lease during flush | P2 | Исправлена | be1 | A2W4-2-001 NEW |
| [TASK-088](TASK-088-2026-07-28-tracker-assignee-team-tx-automation.md) | 🔒 Tracker assignee/team tx/automation Nak | P2 | Исправлена | be1 | A2W1 |
| [TASK-089](TASK-089-2026-07-28-messenger-crm-invoice-guards.md) | 🔒 Messenger reaction + CRM + invoice | P2 | Исправлена | be2 | A2W3/W7 |
| [TASK-090](TASK-090-2026-07-28-audit-purge-error-contract.md) | 🔒 Audit purge fail-closed + errors | P2 | Исправлена | be1 | A2W6 |
| [TASK-091](TASK-091-2026-07-28-tms-createbug-reports-batch.md) | 🐛 TMS createBug + reports N+1 | P2 | Исправлена | be2 | A2W2 |
| [TASK-092](TASK-092-2026-07-28-search-filter-status-safedial.md) | 🔒 Search 403 + Meili SafeDial | P2 | Исправлена | be2 | A2W5-A4 |
| [TASK-093](TASK-093-2026-07-28-fe-a2-error-empty-a11y-sweep.md) | 🧹 FE A2 error≠empty + a11y | P2 | Исправлена | fe1 | A2W9/W10/W11 |
| [TASK-094](TASK-094-2026-07-28-coffee-pairs-triple-edge.md) | ☕ Coffee Pairs[] triple-edges | P3 | Исправлена | be2 | A2W7-2-N01 |
| [TASK-095](TASK-095-2026-07-28-ent-version-mixin-adr-rollout.md) | 📐 Ent VersionMixin ADR + rollout | **P1** | Исправлена | arch→be1 | A2W8 · 19 rows |
| [TASK-096](TASK-096-2026-07-28-ent-ondelete-cascade-adr.md) | 📐 Ent OnDelete/cascade ADR | **P1** | Исправлена | arch→be1 | A2W8 · 4 |
| [TASK-097](TASK-097-2026-07-28-ent-soft-delete-schema-gaps.md) | 📐 Ent soft-delete schema gaps | **P1** | Исправлена | be2 | A2W8/W7 · 7 |
| [TASK-098](TASK-098-2026-07-28-intranet-author-delete-polls-adr.md) | 📐 Intranet author-delete + polls U03 | **P1** | Исправлена | product→be1 | W7-A2-001 |
| [TASK-099](TASK-099-2026-07-28-storage-nf09-size-zero-adr.md) | 📐 Storage NF-09 size==0 | **P1** | Исправлена | product→be2 | W7-A4-001 |
| [TASK-100](TASK-100-2026-07-28-backup-maintenance-gate-distributed-adr.md) | 📐 Backup MaintenanceGate U04 | **P1** | Исправлена | devops→be1 | W7-A4-003 |
| [TASK-101](TASK-101-2026-07-28-ai-ollama-localhost-consent-adr.md) | 📐 AI Ollama localhost consent | **P1** | Исправлена | product→be1 | A2W5-A2-005 |
| [TASK-102](TASK-102-2026-07-28-telegram-login-removal-adr-auth01.md) | 📐 Telegram login cleanup ADR-AUTH-01 | **P1** | Исправлена | be1 | A2W6 · 7 |
| [TASK-103](TASK-103-2026-07-28-jwt-ws-handshake-adr.md) | 📐 JWT/WS handshake (no ?token=) | P2 | Исправлена | arch→be+fe | A2W3/W11 · 3 |
| [TASK-104](TASK-104-2026-07-28-docs-cursor-pagination-adr.md) | 📐 Docs/API cursor vs offset | P2 | Исправлена | sa→be+docs | A2W11 · 2 |
| [TASK-105](TASK-105-2026-07-28-docs-vision-g07-wiki-kpi-adr.md) | 📐 Docs G-07 wiki KPI | **P1** | Исправлена | product→docs | A2W11-5-001 |
| [TASK-106](TASK-106-2026-07-28-intranet-coffee-dm-notify-adr.md) | 📐 Intranet FR-INTR-11 DM/notify U05 | P2 | Исправлена | product→be2 | W7-A2-008 |
| [TASK-107](TASK-107-2026-07-28-ent-indexes-composite-fk-polymorphic.md) | 🔒 Ent indexes/FK/polymorphic | P2 | Исправлена | be1 | A2W8 · 8 |
| [TASK-108](TASK-108-2026-07-28-messenger-invite-dm-unread-scope.md) | 🔒 Messenger invite/DM/unread/bot | P2 | Исправлена | be1 | A2W3 · 9 |
| [TASK-109](TASK-109-2026-07-28-ws-revoke-nats-publish.md) | 🔒 WS revoke → UnregisterBySession | P2 | Исправлена | be1 | A2W3-4-002 |
| [TASK-110](TASK-110-2026-07-28-calls-recording-room-guest-code.md) | 📞 Calls recording/room/guest | P2 | Исправлена | be2 | A2W3 · 3 |
| [TASK-111](TASK-111-2026-07-28-tms-wave2-services-import-batch.md) | 🐛 TMS W2 residual batch | P2 | Исправлена | be2 | A2W2 · 24 |
| [TASK-112](TASK-112-2026-07-28-platform-outbox-nats-aes-hsts.md) | 🐛 Platform NATS/HSTS/AES/RL | P2 | Исправлена | be1 | A2W4 · 9 |
| [TASK-113](TASK-113-2026-07-28-ai-residual-sanitize-tests-policy.md) | 🔒 AI residual sanitize/tests | P2 | Исправлена | be1 | A2W5 · 18 FIXED |
| [TASK-114](TASK-114-2026-07-28-audit-gdpr-limits-layering-tests.md) | 🔒 Audit/GDPR residual | P2 | Исправлена | be1 | A2W6 · 14 |
| [TASK-115](TASK-115-2026-07-28-rbac-permission-middleware-hygiene.md) | 🔒 RBAC permission hygiene | P2 | Исправлена | be2 | A2W6 · 5 |
| [TASK-116](TASK-116-2026-07-28-auth-avatar-xss-validation-tests.md) | 🔒 Auth AvatarURL + session assert | P2 | Исправлена | be1 | A2W6 · 2 |
| [TASK-117](TASK-117-2026-07-28-notification-mail-push-limits-tests.md) | 📧 Notification/mail/push limits | P2 | Исправлена | be2 | W7-A1 · 7 |
| [TASK-118](TASK-118-2026-07-28-v4-crm-hr-calendar-lists-trash.md) | 💼 v4 CRM/HR/calendar residual | P2 | Исправлена | be2 | W7-A2 · 5 |
| [TASK-119](TASK-119-2026-07-28-storage-backup-dashboard-residual.md) | 🔒 Storage/backup/dashboard residual | P2 | Исправлена | be1 | W7-A4 · 8 |
| [TASK-120](TASK-120-2026-07-28-integrations-secrets-imex-trash.md) | 🔌 Git*/1C/imex secrets+trash | P2 | Исправлена | be2 | W7-A3 · 4 |
| [TASK-121](TASK-121-2026-07-28-marketplace-status-dev-exposure.md) | 🛒 Marketplace/status/docs gates | P2 | Исправлена | be1 | W7-A5 · 6 |
| [TASK-122](TASK-122-2026-07-28-tracker-project-jql-cleanup-links.md) | 🔒 Tracker JQL + DeleteProject links | P2 | Исправлена | be1 | A2W1 · 2 |
| [TASK-123](TASK-123-2026-07-28-fe-lib-dead-export-docs-drift.md) | 🧹 FE dead export + docs drift | P2 | Исправлена | fe2+docs | A2W11 · 4 |
| [TASK-124](TASK-124-2026-07-29-a3-10-1-fe-routes-modules.md) | 🧹 A3 FE routes A1: accounting/marketplace/bases error≠empty + noopener | P3 | Открыта | fe1 | A3W10-1 · 5 |
| [TASK-125](TASK-125-2026-07-29-a3-10-2-fe-routes-modules.md) | 🔒🧹 A3 FE routes A2: github/gitlab sanitizeUrl + canMutate CI | P2 | На проверке | fe1 | A3W10-2 · 5 |
| [TASK-126](TASK-126-2026-07-29-a3-10-3-fe-sprint-backlog-regress.md) | ♻️🔒 A3 FE SprintBacklog filter REGRESS + sprint-store leaks | P1 | Открыта | fe1 | A3W10-3 · 4 |
| [TASK-127](TASK-127-2026-07-29-a3-10-4-fe-error-empty-settings.md) | 🧹 A3 FE routes A4: error≠empty messenger/project/TMS/settings | P2 | На проверке | fe1 | A3W10-4 · 6 |
| [TASK-128](TASK-128-2026-07-29-a3-10-5-fe-tms-onboarding-auth.md) | 🔒 A3 FE TMS canMutate + onboarding auth + login redirect | P2 | На проверке | fe1 | A3W10-5 · 5 |
| [TASK-129](TASK-129-2026-07-29-a3-11-1-fe-api-error-mapping.md) | ♻️🔒 A3 FE client.ts wire error→ApiError.message REGRESS | P1 | Открыта | fe1 | A3W11-1 · 1 |
| [TASK-130](TASK-130-2026-07-29-a3-11-2-fe-projects-sprint-version.md) | ♻️🔒 A3 FE projects/sprint API missing version (TASK-095) | P1 | На проверке | fe1 | A3W11-2 · 2 |
| [TASK-131](TASK-131-2026-07-29-a3-11-3-fe-stores-guards.md) | 🧹 A3 FE stores: loadLayout silent fail + reaction_toggled guard | P3 | На проверке | fe1 | A3W11-3 · 2 |
| [TASK-132](TASK-132-2026-07-29-a3-11-4-docs-inventory-drift.md) | 🧹 A3 docs README inventory + webcam/competitive HEAD drift | P2 | Открыта | docs-writer | A3W11-4 · 3 |
| [TASK-133](TASK-133-2026-07-29-a3-11-5-docs-metrics-spec-offset.md) | 🧹 A3 docs inventory/strategy Vitest + list-issues offset | P2 | На проверке | docs-writer | A3W11-5 · 3 |
| [TASK-134](TASK-134-2026-07-29-a3-1-2-tracker-services.md) | 🔒 A3 tracker services: labels scope + residual batch | P1 | Открыта | be1 | A3W1-2 · 6 |
| [TASK-135](TASK-135-2026-07-29-a3-1-3-tracker-repos.md) | 🐛 A3 tracker repos/automation residual | P2 | На проверке | be1 | A3W1-3 · 3 |
| [TASK-136](TASK-136-2026-07-29-a3-1-4-sprint.md) | 🐛 A3 sprint residual (CompleteSprint soft-delete etc) | P2 | Открыта | be1 | A3W1-4 · 3 |
| [TASK-137](TASK-137-2026-07-29-a3-1-5-project.md) | 🔒 A3 project archive/permanent-delete/resurrect residual | P2 | Открыта | be1 | A3W1-5 · 6 |
| [TASK-138](TASK-138-2026-07-29-a3-2-1-tms-handlers.md) | 🧹 A3 TMS handlers residual | P3 | На проверке | be2 | A3W2-1 · 1 |
| [TASK-139](TASK-139-2026-07-29-a3-2-2-tms-services.md) | 🔒 A3 TMS services: AssignTesters/FileIDs/ImportAllure scope | P2 | Открыта | be2 | A3W2-2 · 6 |
| [TASK-140](TASK-140-2026-07-29-a3-2-3-tms-repos.md) | 🐛 A3 TMS repos list caps residual | P2 | Открыта | be2 | A3W2-3 · 3 |
| [TASK-141](TASK-141-2026-07-29-a3-2-4-tms-import.md) | 🐛 A3 TMS import/export TestRail residual | P2 | Открыта | be2 | A3W2-4 · 5 |
| [TASK-142](TASK-142-2026-07-29-a3-2-5-bases.md) | 🧹 A3 bases residual | P3 | Открыта | be2 | A3W2-5 · 3 |
| [TASK-143](TASK-143-2026-07-29-a3-3-1-messenger-handlers.md) | 🧹 A3 messenger handlers poll test wrong URL | P2 | Открыта | be1 | A3W3-1 · 1 |
| [TASK-144](TASK-144-2026-07-29-a3-3-2-messenger-services.md) | 🔒 A3 messenger services Join/Create/Bookmark | P2 | Открыта | be1 | A3W3-2 · 6 |
| [TASK-145](TASK-145-2026-07-29-a3-3-3-messenger-repos.md) | 🐛 A3 messenger repos list caps/soft-delete | P2 | Открыта | be1 | A3W3-3 · 3 |
| [TASK-146](TASK-146-2026-07-29-a3-3-4-ws.md) | 🧹 A3 WS query+ticket mutation gap | P3 | Открыта | be1 | A3W3-4 · 1 |
| [TASK-147](TASK-147-2026-07-29-a3-3-5-calls.md) | 🐛 A3 calls LiveKit orphan + conflicts soft-delete | P2 | Открыта | be2 | A3W3-5 · 4 |
| [TASK-148](TASK-148-2026-07-29-a3-4-1-platform-http.md) | 🧹 A3 platform Request-ID mutation test | P3 | Открыта | be1 | A3W4-1 · 1 |
| [TASK-149](TASK-149-2026-07-29-a3-4-2-platform-nats-outbox.md) | 🔒 A3 outbox panic dead-letter + claim/Flush | P2 | Открыта | be1 | A3W4-2 · 3 |
| [TASK-150](TASK-150-2026-07-29-a3-4-3-platform-obs.md) | 🧹 A3 obs cancel→Error + UUID cardinality | P3 | Открыта | be1 | A3W4-3 · 2 |
| [TASK-151](TASK-151-2026-07-29-a3-4-4-platform-core.md) | 🐛 A3 Postgres Ping + cron timeout test | P2 | На проверке | be1 | A3W4-4 · 2 |
| [TASK-152](TASK-152-2026-07-29-a3-4-5-apigateway.md) | 🔒 A3 apigateway soft-delete webhooks + VersionEQ | P2 | Открыта | be2 | A3W4-5 · 4 |
| [TASK-153](TASK-153-2026-07-29-a3-5-1-ai-handlers.md) | 🐛 A3 AI stream ErrNoProvider mapping | P2 | На проверке | be1 | A3W5-1 · 3 |
| [TASK-154](TASK-154-2026-07-29-a3-5-2-ai-adapters.md) | 🐛 A3 OfflineAI stream Timeout residual | P2 | На проверке | be1 | A3W5-2 · 3 |
| [TASK-155](TASK-155-2026-07-29-a3-5-3-ai-service.md) | ♻️🔒 A3 AI service REGRESS + consent/version/IDOR | P2 | Открыта | be1 | A3W5-3 · 9 |
| [TASK-156](TASK-156-2026-07-29-a3-5-4-search-core.md) | ♻️🔒 A3 Meili operator dial REGRESS + JQL | P2 | На проверке | be2 | A3W5-4 · 3 |
| [TASK-157](TASK-157-2026-07-29-a3-5-5-search-tests.md) | 🧹 A3 search tests limit clamp | P3 | Открыта | be2 | A3W5-5 · 2 |
| [TASK-158](TASK-158-2026-07-29-a3-6-2-auth-tests.md) | 🔒 A3 auth tests deactivated JWT + founder race | P2 | Открыта | be1 | A3W6-2 · 4 |
| [TASK-159](TASK-159-2026-07-29-a3-6-3-rbac.md) | 🐛 A3 RBAC DeleteCustomRole cache-fail | P2 | Открыта | be2 | A3W6-3 · 3 |
| [TASK-160](TASK-160-2026-07-29-a3-6-4-audit-core.md) | 🔒 A3 GDPR Mongo eraser outside Ent tx | P2 | Открыта | be1 | A3W6-4 · 3 |
| [TASK-161](TASK-161-2026-07-29-a3-6-5-audit-tests.md) | 🧹 A3 audit tests CRM purge/Retention gaps | P2 | На проверке | be1 | A3W6-5 · 7 |
| [TASK-162](TASK-162-2026-07-29-a3-7-1-notification-mail.md) | 🐛 A3 mail DailyLimit distributed | P3 | Открыта | be2 | A3W7-1 · 2 |
| [TASK-163](TASK-163-2026-07-29-a3-7-2-v4-business.md) | ♻️ A3 coffee notify REGRESS + poll VersionEQ | P2 | Открыта | be2 | A3W7-2 · 3 |
| [TASK-164](TASK-164-2026-07-29-a3-7-3-integrations.md) | 🔒 A3 integrations trash filter + link-preview membership | P2 | Открыта | be2 | A3W7-3 · 3 |
| [TASK-165](TASK-165-2026-07-29-a3-7-4-storage-dash-backup.md) | 🐛 A3 storage/dashboard/backup residual | P2 | На проверке | be1 | A3W7-4 · 5 |
| [TASK-166](TASK-166-2026-07-29-a3-7-5-mp-status-dev.md) | 🔒 A3 marketplace review_note public catalog | P2 | На проверке | be1 | A3W7-5 · 4 |
| [TASK-167](TASK-167-2026-07-29-a3-8-1-schemas.md) | 🧹 A3 schemas SetNull annotation + ChannelMember | P3 | На проверке | be1 | A3W8-1 · 2 |
| [TASK-168](TASK-168-2026-07-29-a3-8-2-schemas.md) | 🐛 A3 schemas OnDelete Cascade batch (coffee/hub/integrations) | P2 | Открыта | be1 | A3W8-2 · 6 |
| [TASK-169](TASK-169-2026-07-29-a3-8-3-schemas.md) | 🐛 A3 HubTable/JournalEntry Cascade | P2 | Открыта | be1 | A3W8-3 · 2 |
| [TASK-170](TASK-170-2026-07-29-a3-8-4-schemas.md) | 🐛 A3 sprint active unique + SetNull issues | P2 | На проверке | be1 | A3W8-4 · 2 |
| [TASK-171](TASK-171-2026-07-29-a3-8-5-schemas.md) | 🔒 A3 Team/WorkflowStatus soft-delete + OnDelete Cascades | P1 | Открыта | be1 | A3W8-5 · 7 |
| [TASK-172](TASK-172-2026-07-29-a3-9-1-components.md) | 🧹 A3 FE admin error≠empty pending/limits/trash | P2 | Открыта | fe1 | A3W9-1 · 5 |
| [TASK-173](TASK-173-2026-07-29-a3-9-2-components.md) | 🧹 A3 FE bookmarks UUID + deadlines links | P2 | На проверке | fe1 | A3W9-2 · 5 |
| [TASK-174](TASK-174-2026-07-29-a3-9-3-components.md) | 🔒 A3 FE LinkPreview XSS whitelist | P2 | На проверке | fe1 | A3W9-3 · 4 |
| [TASK-175](TASK-175-2026-07-29-a3-9-4-components.md) | 🧹 A3 FE checklist/automation/comment error UI | P2 | Открыта | fe2 | A3W9-4 · 6 |
| [TASK-176](TASK-176-2026-07-29-a3-9-5-components.md) | 🧹 A3 FE avatar/popover/select hub tokens | P3 | Открыта | fe2 | A3W9-5 · 3 |
| [TASK-177](TASK-177-2026-07-29-c-w19-deactivate-ban-ws-kill.md) | 🔒 C-W19 Deactivate/Ban не рвёт живой WS | P2 | На проверке | be1 | C-W19-001 · 2 |
| [TASK-178](TASK-178-2026-07-29-c-w19-gdpr-mongo-erase-retry.md) | 🔒 C-W19 GDPR Mongo erase retry после Commit | P2 | На проверке | be1 | C-W19-002 · 1 |
| [TASK-179](TASK-179-2026-07-29-c-w19-fe-webhook-sprint-completed.md) | 🐛 C-W19 FE sprint.completed после BE trim | P2 | На проверке | fe1 | C-W19-FE-001 · 1 |
| [TASK-180](TASK-180-2026-07-29-c-w19-link-preview-modules-tc.md) | 🔒 C-W19 modules LinkPreview XSS + TC membership | P2 | Открыта | fe1+be2 | C-W19-FE-002 · 2 |
| [TASK-181](TASK-181-2026-07-29-c-w19-webhook-softdelete-deactivate.md) | 🔒 C-W19 soft-delete project гасит webhooks | P2 | На проверке | be2 | C-W19-FE-004 · 1 |
| [TASK-182](TASK-182-2026-07-29-c-w19-hr-pending-leave-error.md) | 🧹 C-W19 HR pending leave error≠empty | P3 | На проверке | fe2 | C-W19-FE-005 · 1 |
| [TASK-183](TASK-183-2026-07-29-c-w19-docs-inventory-drift.md) | 🧹 C-W19 docs inventory + AGENTS/spec drift | P2 | На проверке | docs-writer | C-W19-DOC-001 · 1 |
| [TASK-184](TASK-184-2026-07-29-c-w19-last-team-member.md) | 🐛 C-W19 RemoveTeamMember last-member orphan | P3 | Открыта | be1 | C-W19-TT-001 · 1 |
| [TASK-185](TASK-185-2026-07-29-c-w19-allure-existingrun-partial.md) | 🐛 C-W19 Allure ExistingRun silent partial | P2 | На проверке | be2 | C-W19-TT-002 · 1 |
| [TASK-186](TASK-186-2026-07-29-audit-fe-calls-error-empty.md) | 🧹 Audit FE calls calendar/dialog error≠empty | P2 | На проверке | fe2 | FE-MC-001…003 · 3 |
| [TASK-187](TASK-187-2026-07-29-audit-fe-pipeline-sanitize.md) | 🔒 Audit FE pipeline-status sanitizeUrl | P2 | На проверке | fe1 | FE-TA-001 · 1 |
| [TASK-188](TASK-188-2026-07-29-audit-be-push-telegram-rl.md) | 🔒 Audit BE PushSender layering + Telegram RL | P2 | На проверке | be2 | ASN-001…002 · 2 |
| [TASK-189](TASK-189-2026-07-29-audit-be-v4-calendar-poll.md) | 🔒 Audit BE calendar trash + draft poll IDOR | P2 | Открыта | be1 | BE-V4-001…003 · 3 |
| [TASK-190](TASK-190-2026-07-29-audit-be-tracker-messenger-softdelete.md) | 🔒 Audit BE tracker/messenger soft-delete + CreateChannel | P2 | Открыта | be1+be2 | AUDIT-TM-001…004 · 4 |
| [TASK-191](TASK-191-2026-07-29-audit-ent-automationrule-softdelete.md) | 🔒 Audit ent AutomationRule soft-delete | P2 | На проверке | be1 | EA-001 · 1 |
| [TASK-192](TASK-192-2026-07-29-audit-fe-settings-version-ol.md) | 🔒 Audit FE settings version OL (webhooks/AI) | P2 | На проверке | fe1 | FE-SMA-001…002 · 2 |
| [TASK-009](TASK-009-project-routes-rbac.md) | project routes без RBAC | **P0** | Исправлена | be1 | W1-A5-001 |
| [TASK-010](TASK-010-promote-to-issue-rbac.md) | promote-to-issue без PermTaskCreate | **P0** | Исправлена | be1 | W3-A2-001 |
| [TASK-011](TASK-011-webhook-event-match.md) | webhook UI events ≠ NATS subjects | **P0** | Исправлена | be1 | W4-A5-001 |
| [TASK-012](TASK-012-webhook-project-scope.md) | webhook dispatch без project-scope | **P0** | Исправлена | be1 | W4-A5-002 |
| [TASK-013](TASK-013-tracker-project-scope-idor.md) | Tracker IDOR workflow/fields/refs | P1 | Исправлена | be1 | W12 P1-BE-01..04 |
| [TASK-014](TASK-014-project-acl-sprint-tx.md) | Project ACL + sprint atomic | P1 | Закрыта | be2→rework · W1-be1 PASS | W12 P1-BE-05..09 · MUTATION-2026-07-29-W1-be1 |
| [TASK-015](TASK-015-tms-project-scope-idor-refs.md) | TMS suite/env IDOR + FK | P1 | Исправлена | be2 | W12 P1-BE-10..12 · W18-C2 PASS |
| [TASK-016](TASK-016-tms-import-limits-steps.md) | TMS import limits + steps | P1 | Исправлена | be2 | W12 P1-BE-13..14 · W18-C2 PASS |
| [TASK-017](TASK-017-messenger-membership-idor.md) | Messenger membership IDOR | P1 | Исправлена | be1 | W12 P1-BE-15..20 |
| [TASK-018](TASK-018-ws-calls-authz-race.md) | WS race + scheduled calls authz | P1 | Исправлена | be1 | W12 P1-BE-21..22 |
| [TASK-019](TASK-019-platform-ssrf-outbox-pool-config.md) | Platform SSRF/outbox/pool/config | P1 | Закрыта | be2→rework · W1-be1 PASS | W12 P1-BE-23..26 · MUTATION-2026-07-29-W1-be1 |
| [TASK-020](TASK-020-apigateway-webhook-create-contract.md) | Webhook secret once + allowlist | P1 | Исправлена | be2 | W12 P1-BE-27..28 · W18-C2 PASS |
| [TASK-021](TASK-021-ai-policy-bot-authz.md) | AI policy + bot authz | P1 | Исправлена | be1 | W12 P1-BE-29..31 |
| [TASK-022](TASK-022-search-nf03-softdelete.md) | Search soft-delete + private | P1 | Исправлена | be2 | W12 P1-BE-32 · W18-C2 PASS |
| [TASK-023](TASK-023-rbac-cache-gdpr.md) | RBAC cache + GDPR erase | P1 | Исправлена | be1 | W12 P1-BE-33..34 · W18-C3 PASS |
| [TASK-024](TASK-024-notif-integrations-storage-backup.md) | Notif/Git*/storage/backup/coffee | P1 | Исправлена | be2 | W12 P1-BE-35..39,43 · W18-C2 PASS |
| [TASK-025](TASK-025-ent-schema-constraints.md) | Ent unique/O2O constraints | P1 | Исправлена | be1 | W12 P1-BE-40..42 · W18-C3 PASS |
| [TASK-026](TASK-026-fe-api-client-auth-retry.md) | FE api client auth retry | P1 | Исправлена | fe1 | W12 FE-P1-API-CLIENT |
| [TASK-027](TASK-027-fe-callchat-ws-subscribe.md) | FE CallChat WS subscribe | P1 | Исправлена | fe1 | W12 FE-P1-CALLS-CHAT |
| [TASK-028](TASK-028-fe-calls-join-create-lifecycle.md) | FE calls lifecycle | P1 | Исправлена | fe1 | W12 FE-P1-CALLS-LIFECYCLE |
| [TASK-029](TASK-029-fe-messenger-reply-thread-unread.md) | FE messenger reply/unread | P1 | Исправлена | fe1 | W12 FE-P1-MESSENGER |
| [TASK-030](TASK-030-fe-board-dnd-single-path-rank.md) | FE board DnD + rank | P1 | Исправлена | fe1 | W12 FE-P1-BOARD-DND |
| [TASK-031](TASK-031-fe-sprint-loading-error-states.md) | FE sprint loading/error | P1 | Исправлена | fe1 | W12 FE-P1-SPRINT-UI |
| [TASK-032](TASK-032-fe-trash-list-and-webhooks-project.md) | FE trash + webhooks projectId | P1 | Исправлена · W18-C5 | fe2 | W12 FE-P1-TRASH-WEBHOOKS |
| [TASK-033](TASK-033-fe-error-empty-auth-race.md) | FE error≠empty + auth race | P1 | Исправлена · W18-C5 | fe2 | W12 FE-P1-ERROR-AS-EMPTY |
| [TASK-034](TASK-034-fe-a11y-nested-sidebar-notifications.md) | FE a11y nested interactive | P1 | Исправлена · W18-C5 | fe2 | W12 FE-P1-A11Y-CONTROLS |
| [TASK-035](TASK-035-fe-tms-compare-and-docs-testpath.md) | FE TMS compare + docs drift | P1 | Исправлена | fe1 | W12 FE-P1-TMS-DOCS |
| [TASK-036](TASK-036-be-p2-dto-error-contract.md) | BE DTO validator + error contract | P2 | Исправлена · W18-C2b | be2 | W12 P2-BE-01,07 |
| [TASK-037](TASK-037-be-p2-authz-ent-negative-tests.md) | BE Ent-negative authz tests | P2 | Исправлена · W18-C2b | be2 | W12 P2-BE-02 |
| [TASK-038](TASK-038-be-p2-ol-softdelete-tx-lists.md) | BE OL/soft-mutate/tx/lists | P2 | Исправлена | be1 | W12 P2-BE-03..06 · W18-C3 PASS (partial residual OK) |
| [TASK-039](TASK-039-be-p2-soft-delete-secrets.md) | BE soft-delete + secrets | P2 | Исправлена · W18-C2b | be2 | W12 P2-BE-08,09 |
| [TASK-040](TASK-040-be-p2-fail-closed-ratelimit.md) | BE fail-closed + rate-limit | P2 | Исправлена | be1 | W12 P2-BE-10 · W18-C3 PASS |
| [TASK-041](TASK-041-be-p2-sprint-poker-start.md) | BE sprint poker/start | P2 | Исправлена | be1→rework | W12 P2-BE-11 · second Start conflict |
| [TASK-042](TASK-042-be-p2-perm-messenger-ws-calls.md) | BE perm/messenger/ws/calls | P2 | Исправлена | be1→rework | W12 P2-BE-12..15 · NATS via_invite |
| [TASK-043](TASK-043-be-p2-platform-outbox-obs.md) | BE platform outbox/obs | P2 | Исправлена · W18-C2b | be2 | W12 P2-BE-16 |
| [TASK-044](TASK-044-be-p2-ai-search.md) | AI stream/privacy + search JQL/SSRF | P2 | Исправлена | be1→rework | W12 P2-BE-17/18 · Meili+cancel IT |
| [TASK-045](TASK-045-be-p2-rbac-gdpr-integrations-storage.md) | BE rbac/gdpr/imex/storage | P2 | Закрыта | be2→rework · W1-be1 PASS | W12 P2-BE-19..21 · MUTATION-2026-07-29-W1-be1 |
| [TASK-046](TASK-046-be-p2-ent-unique-fk-check.md) | BE Ent unique/FK/CHECK | P2 | Исправлена | be1→rework | W12 P2-BE-22 · enttest runtime |
| [TASK-047](TASK-047-fe-p2-error-vs-empty.md) | FE error≠empty sweep | P2 | Исправлена | fe1 | W12 P2-FE-01 |
| [TASK-048](TASK-048-fe-p2-mutation-feedback.md) | FE mutation feedback | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-02 |
| [TASK-049](TASK-049-fe-p2-a11y-names-dialogs.md) | FE a11y names/dialogs | P2 | Исправлена | fe1 | W12 P2-FE-03,04 |
| [TASK-050](TASK-050-fe-p2-tokens-dead-ui.md) | FE tokens/dead UI/authz | P2 | Исправлена | fe1 | W12 P2-FE-05,06,16 |
| [TASK-051](TASK-051-fe-p2-effects-hooks-race.md) | FE effects/hooks race | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-07 |
| [TASK-052](TASK-052-fe-p2-api-dto-gaps.md) | FE API DTO gaps | P2 | Исправлена | fe2→rework | W12 P2-FE-08 · tsc fixture |
| [TASK-053](TASK-053-fe-p2-auth-ws-upload.md) | FE auth/WS/upload | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-09..11 |
| [TASK-054](TASK-054-fe-p2-docs-metrics-drift.md) | Docs metrics drift | P2 | Исправлена | docs→rework | W12 P2-FE-12 · inventory script |
| [TASK-055](TASK-055-fe-p2-shell-tracker-ux.md) | FE shell + tracker UX | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-13,14 |
| [TASK-056](TASK-056-fe-p2-i18n-push-rooms.md) | FE i18n/push/rooms | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-15 |
| [TASK-057](TASK-057-be-p3-silent-error-swallow.md) | BE silent error swallow | P3 | Исправлена | be1 | W12 P3-BE-01 |
| [TASK-058](TASK-058-be-p3-assert-free-tests.md) | BE assert-free tests | P3 | Исправлена | be2 | W12 P3-BE-02 |
| [TASK-059](TASK-059-be-p3-mutation-proof-gaps.md) | BE mutation-proof gaps | P3 | Исправлена | be1 | W12 P3-BE-03 |
| [TASK-060](TASK-060-be-p3-utf8-rune-truncate.md) | BE UTF-8 rune truncate | P3 | Исправлена | be2 | W12 P3-BE-04 |
| [TASK-061](TASK-061-be-p3-repo-lists-defense.md) | BE repo/lists defense | P3 | Исправлена | be1 | W12 P3-BE-05,06 |
| [TASK-062](TASK-062-be-p3-dead-code-layering.md) | BE dead code/layering | P3 | Исправлена | be2 | W12 P3-BE-07 |
| [TASK-063](TASK-063-be-p3-errors-input-hygiene.md) | BE errors + input hygiene | P3 | Исправлена | be1 | W12 P3-BE-08,09 |
| [TASK-064](TASK-064-be-p3-ws-calls-lifecycle.md) | BE WS/calls lifecycle | P3 | Исправлена | be2 | W12 P3-BE-10 |
| [TASK-065](TASK-065-be-p3-platform-http-obs.md) | BE platform HTTP/obs | P3 | Исправлена | be1 | W12 P3-BE-11,12 |
| [TASK-066](TASK-066-be-p3-ai-auth-rbac-audit.md) | BE AI/auth/rbac/audit | P3 | Исправлена | be2 | W12 P3-BE-13,14 |
| [TASK-067](TASK-067-be-p3-v4-dev-storage.md) | BE v4/dev/storage | P3 | Исправлена | be1 | W12 P3-BE-15,16 |
| [TASK-068](TASK-068-be-p3-ent-hygiene-mojibake.md) | BE Ent hygiene/mojibake | P3 | Исправлена | be2 | W12 P3-BE-17,18 |
| [TASK-069](TASK-069-fe-p3-a11y-names.md) | FE a11y names residual | P3 | Исправлена | fe1 | W12 P3-FE-01 |
| [TASK-070](TASK-070-fe-p3-a11y-structure.md) | FE a11y structure | P3 | Исправлена | fe1 | W12 P3-FE-02 |
| [TASK-071](TASK-071-fe-p3-tokens-i18n.md) | FE tokens + i18n | P3 | Исправлена | fe1 | W12 P3-FE-03,04 |
| [TASK-072](TASK-072-fe-p3-timers-cleanup.md) | FE timers + NaN | P3 | Исправлена | fe2 | W12 P3-FE-05,16 |
| [TASK-073](TASK-073-fe-p3-silent-fail-empty.md) | FE silent fail residual | P3 | Исправлена | fe2 | W12 P3-FE-06 |
| [TASK-074](TASK-074-fe-p3-type-casts.md) | FE type casts | P3 | Исправлена | fe2 | W12 P3-FE-07 |
| [TASK-075](TASK-075-fe-p3-dead-ui-state.md) | FE dead UI/state | P3 | Исправлена | fe1 | W12 P3-FE-08 |
| [TASK-076](TASK-076-fe-p3-perf-stale.md) | FE perf stale | P3 | Исправлена | fe1 | W12 P3-FE-09 |
| [TASK-077](TASK-077-fe-p3-api-auth-naming.md) | FE api/auth/naming | P3 | Исправлена | fe1 | W12 P3-FE-10..12 |
| [TASK-078](TASK-078-fe-p3-docs-spa-ux.md) | FE docs/SPA/preview | P3 | Исправлена | fe1 | W12 P3-FE-13..15 |
| [TASK-079](TASK-079-gdpr-mongo-eraser-fail-closed.md) | GDPR Mongo eraser fail-closed + IT | P2 | Исправлена | be1 | W18-C5-001/002 |
| [TASK-080](TASK-080-docs-metrics-drift-autotest.md) | Docs metrics drift autotest | P2 | Исправлена | docs-writer | W18-C5 / closes 054 gate |

## Закрытые

| ID | Задача | Приоритет | Статус | Исполнитель | Источник |
|----|--------|-----------|--------|-------------|----------|
| [TASK-001](TASK-001-fr-intr-08-newspost-trash.md) | Новость невосстановима из корзины | **P0** | Закрыта | be1 | FR-INTR-08 |
| [TASK-002](TASK-002-fr-cal-05-series-trash.md) | У календарных серий нет корзины | **P0** | Закрыта | be1 | FR-CAL-05 |
| [TASK-004](TASK-004-state-goals-verified-reconcile.md) | STATE расходится с матрицей фаз | P1 | Закрыта | team-lead | Ревью W4–W20 |
| [TASK-005](TASK-005-august-phase-dependencies.md) | Связать августовские фазы зависимостями | P1 | Закрыта | team-lead | Аудит планирования |
| [TASK-003](TASK-003-a14-002-test-granularity.md) | Тест пропускает удаление одного guard'а | P2 | Закрыта | qa-auto1 | Ревью W4–W20 |
| [TASK-006](TASK-006-disable-cursor-coauthor-attribution.md) | Навсегда выключить Cursor Co-authored-by | **P0** | Закрыта | devops/docs | Owner 2026-07-28 |
| [TASK-007](TASK-007-utf8-mojibake-audit-fix.md) | Аудит и починка битой UTF-8 кодировки | P1 | Закрыта | docs+be1 | Owner 2026-07-28 |

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
