# Бэклог задач

> Единый вход для агентов и владельца. Одна папка, одна таблица, один формат.
> Обновлено: 2026-07-28 (TASK-008 стадия B · W12)

Каждая задача — отдельный файл, написанный по
[мастер-промпту](../../prompt-engineering/agents/templates/_MASTER-PROMPT-TEMPLATE.md).
Правила ведения: [`.cursor/rules/57-backlog-tasks.mdc`](../../.cursor/rules/57-backlog-tasks.mdc).

---

## Открытые

| ID | Задача | Приоритет | Статус | Исполнитель | Источник |
|----|--------|-----------|--------|-------------|----------|
| [TASK-008](TASK-008-total-audit-program.md) | Тотальный пофайловый аудит | P1 | В работе (B: W13 P0✅ → W14 P1) | Grok + субагенты | Owner 2026-07-28 |
| [TASK-009](TASK-009-project-routes-rbac.md) | project routes без RBAC | **P0** | На проверке | be1 | W1-A5-001 |
| [TASK-010](TASK-010-promote-to-issue-rbac.md) | promote-to-issue без PermTaskCreate | **P0** | На проверке | be1 | W3-A2-001 |
| [TASK-011](TASK-011-webhook-event-match.md) | webhook UI events ≠ NATS subjects | **P0** | На проверке | be1 | W4-A5-001 |
| [TASK-012](TASK-012-webhook-project-scope.md) | webhook dispatch без project-scope | **P0** | На проверке | be1 | W4-A5-002 |
| [TASK-013](TASK-013-tracker-project-scope-idor.md) | Tracker IDOR workflow/fields/refs | P1 | На проверке | be1 | W12 P1-BE-01..04 |
| [TASK-014](TASK-014-project-acl-sprint-tx.md) | Project ACL + sprint atomic | P1 | На проверке | be2 | W12 P1-BE-05..09 |
| [TASK-015](TASK-015-tms-project-scope-idor-refs.md) | TMS suite/env IDOR + FK | P1 | На проверке | be2 | W12 P1-BE-10..12 |
| [TASK-016](TASK-016-tms-import-limits-steps.md) | TMS import limits + steps | P1 | На проверке | be2 | W12 P1-BE-13..14 |
| [TASK-017](TASK-017-messenger-membership-idor.md) | Messenger membership IDOR | P1 | На проверке | be1 | W12 P1-BE-15..20 |
| [TASK-018](TASK-018-ws-calls-authz-race.md) | WS race + scheduled calls authz | P1 | На проверке | be1 | W12 P1-BE-21..22 |
| [TASK-019](TASK-019-platform-ssrf-outbox-pool-config.md) | Platform SSRF/outbox/pool/config | P1 | На проверке | be1 | W12 P1-BE-23..26 |
| [TASK-020](TASK-020-apigateway-webhook-create-contract.md) | Webhook secret once + allowlist | P1 | На проверке | be2 | W12 P1-BE-27..28 |
| [TASK-021](TASK-021-ai-policy-bot-authz.md) | AI policy + bot authz | P1 | На проверке | be1 | W12 P1-BE-29..31 |
| [TASK-022](TASK-022-search-nf03-softdelete.md) | Search soft-delete + private | P1 | На проверке | be2 | W12 P1-BE-32 |
| [TASK-023](TASK-023-rbac-cache-gdpr.md) | RBAC cache + GDPR erase | P1 | На проверке | be1 | W12 P1-BE-33..34 |
| [TASK-024](TASK-024-notif-integrations-storage-backup.md) | Notif/Git*/storage/backup/coffee | P1 | На проверке | be2 | W12 P1-BE-35..39,43 |
| [TASK-025](TASK-025-ent-schema-constraints.md) | Ent unique/O2O constraints | P1 | На проверке | be1 | W12 P1-BE-40..42 |
| [TASK-026](TASK-026-fe-api-client-auth-retry.md) | FE api client auth retry | P1 | На проверке | fe1 | W12 FE-P1-API-CLIENT |
| [TASK-027](TASK-027-fe-callchat-ws-subscribe.md) | FE CallChat WS subscribe | P1 | На проверке | fe1 | W12 FE-P1-CALLS-CHAT |
| [TASK-028](TASK-028-fe-calls-join-create-lifecycle.md) | FE calls lifecycle | P1 | На проверке | fe1 | W12 FE-P1-CALLS-LIFECYCLE |
| [TASK-029](TASK-029-fe-messenger-reply-thread-unread.md) | FE messenger reply/unread | P1 | На проверке | fe1 | W12 FE-P1-MESSENGER |
| [TASK-030](TASK-030-fe-board-dnd-single-path-rank.md) | FE board DnD + rank | P1 | На проверке | fe1 | W12 FE-P1-BOARD-DND |
| [TASK-031](TASK-031-fe-sprint-loading-error-states.md) | FE sprint loading/error | P1 | На проверке | fe1 | W12 FE-P1-SPRINT-UI |
| [TASK-032](TASK-032-fe-trash-list-and-webhooks-project.md) | FE trash + webhooks projectId | P1 | На проверке | fe2 | W12 FE-P1-TRASH-WEBHOOKS |
| [TASK-033](TASK-033-fe-error-empty-auth-race.md) | FE error≠empty + auth race | P1 | На проверке | fe2 | W12 FE-P1-ERROR-AS-EMPTY |
| [TASK-034](TASK-034-fe-a11y-nested-sidebar-notifications.md) | FE a11y nested interactive | P1 | На проверке | fe2 | W12 FE-P1-A11Y-CONTROLS |
| [TASK-035](TASK-035-fe-tms-compare-and-docs-testpath.md) | FE TMS compare + docs drift | P1 | На проверке | fe1 | W12 FE-P1-TMS-DOCS |

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
