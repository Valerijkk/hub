# Бэклог задач

> Единый вход для агентов и владельца. Одна папка, одна таблица, один формат.
> Обновлено: 2026-07-28 (W18-C5 mutation + catch-up TASK-079/080)

Каждая задача — отдельный файл, написанный по
[мастер-промпту](../../prompt-engineering/agents/templates/_MASTER-PROMPT-TEMPLATE.md).
Правила ведения: [`.cursor/rules/57-backlog-tasks.mdc`](../../.cursor/rules/57-backlog-tasks.mdc).

---

## Открытые

| ID | Задача | Приоритет | Статус | Исполнитель | Источник |
|----|--------|-----------|--------|-------------|----------|
| [TASK-008](TASK-008-total-audit-program.md) | Тотальный пофайловый аудит | P1 | В работе (B: P0/P1/P2✅ → P3 TASK-057…078) | Grok + субагенты | Owner 2026-07-28 |
| [TASK-009](TASK-009-project-routes-rbac.md) | project routes без RBAC | **P0** | Исправлена | be1 | W1-A5-001 |
| [TASK-010](TASK-010-promote-to-issue-rbac.md) | promote-to-issue без PermTaskCreate | **P0** | Исправлена | be1 | W3-A2-001 |
| [TASK-011](TASK-011-webhook-event-match.md) | webhook UI events ≠ NATS subjects | **P0** | Исправлена | be1 | W4-A5-001 |
| [TASK-012](TASK-012-webhook-project-scope.md) | webhook dispatch без project-scope | **P0** | Исправлена | be1 | W4-A5-002 |
| [TASK-013](TASK-013-tracker-project-scope-idor.md) | Tracker IDOR workflow/fields/refs | P1 | Исправлена | be1 | W12 P1-BE-01..04 |
| [TASK-014](TASK-014-project-acl-sprint-tx.md) | Project ACL + sprint atomic | P1 | Открыта | be2 | W12 P1-BE-05..09 · W18-C2 FAIL |
| [TASK-015](TASK-015-tms-project-scope-idor-refs.md) | TMS suite/env IDOR + FK | P1 | Исправлена | be2 | W12 P1-BE-10..12 · W18-C2 PASS |
| [TASK-016](TASK-016-tms-import-limits-steps.md) | TMS import limits + steps | P1 | Исправлена | be2 | W12 P1-BE-13..14 · W18-C2 PASS |
| [TASK-017](TASK-017-messenger-membership-idor.md) | Messenger membership IDOR | P1 | Исправлена | be1 | W12 P1-BE-15..20 |
| [TASK-018](TASK-018-ws-calls-authz-race.md) | WS race + scheduled calls authz | P1 | Исправлена | be1 | W12 P1-BE-21..22 |
| [TASK-019](TASK-019-platform-ssrf-outbox-pool-config.md) | Platform SSRF/outbox/pool/config | P1 | Открыта | be2 | W12 P1-BE-23..26 · W18-C2 FAIL |
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
| [TASK-041](TASK-041-be-p2-sprint-poker-start.md) | BE sprint poker/start | P2 | Открыта | be1 | W12 P2-BE-11 · W18-C3 FAIL |
| [TASK-042](TASK-042-be-p2-perm-messenger-ws-calls.md) | BE perm/messenger/ws/calls | P2 | Открыта | be1 | W12 P2-BE-12..15 · W18-C3 FAIL |
| [TASK-043](TASK-043-be-p2-platform-outbox-obs.md) | BE platform outbox/obs | P2 | Исправлена · W18-C2b | be2 | W12 P2-BE-16 |
| [TASK-044](TASK-044-be-p2-ai-search.md) | BE AI privacy + search | P2 | Открыта | be1 | W12 P2-BE-17,18 · W18-C3 FAIL |
| [TASK-045](TASK-045-be-p2-rbac-gdpr-integrations-storage.md) | BE rbac/gdpr/imex/storage | P2 | Открыта · W18-C2b FAIL | be2 | W12 P2-BE-19..21 |
| [TASK-046](TASK-046-be-p2-ent-unique-fk-check.md) | BE Ent unique/FK/CHECK | P2 | Открыта | be1 | W12 P2-BE-22 · W18-C3 FAIL |
| [TASK-047](TASK-047-fe-p2-error-vs-empty.md) | FE error≠empty sweep | P2 | Исправлена | fe1 | W12 P2-FE-01 |
| [TASK-048](TASK-048-fe-p2-mutation-feedback.md) | FE mutation feedback | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-02 |
| [TASK-049](TASK-049-fe-p2-a11y-names-dialogs.md) | FE a11y names/dialogs | P2 | Исправлена | fe1 | W12 P2-FE-03,04 |
| [TASK-050](TASK-050-fe-p2-tokens-dead-ui.md) | FE tokens/dead UI/authz | P2 | Исправлена | fe1 | W12 P2-FE-05,06,16 |
| [TASK-051](TASK-051-fe-p2-effects-hooks-race.md) | FE effects/hooks race | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-07 |
| [TASK-052](TASK-052-fe-p2-api-dto-gaps.md) | FE API DTO gaps | P2 | Открыта · W18-C5 FAIL | fe2 | W12 P2-FE-08 |
| [TASK-053](TASK-053-fe-p2-auth-ws-upload.md) | FE auth/WS/upload | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-09..11 |
| [TASK-054](TASK-054-fe-p2-docs-metrics-drift.md) | Docs metrics/wiring drift | P2 | Открыта · W18-C5 FAIL | docs-writer | W12 P2-FE-12 |
| [TASK-055](TASK-055-fe-p2-shell-tracker-ux.md) | FE shell + tracker UX | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-13,14 |
| [TASK-056](TASK-056-fe-p2-i18n-push-rooms.md) | FE i18n/push/rooms | P2 | Исправлена · W18-C5 | fe2 | W12 P2-FE-15 |
| [TASK-057](TASK-057-be-p3-silent-error-swallow.md) | BE silent error swallow | P3 | Открыта | be1 | W12 P3-BE-01 |
| [TASK-058](TASK-058-be-p3-assert-free-tests.md) | BE assert-free tests | P3 | Открыта | be2 | W12 P3-BE-02 |
| [TASK-059](TASK-059-be-p3-mutation-proof-gaps.md) | BE mutation-proof gaps | P3 | Открыта | be1 | W12 P3-BE-03 |
| [TASK-060](TASK-060-be-p3-utf8-rune-truncate.md) | BE UTF-8 rune truncate | P3 | Открыта | be2 | W12 P3-BE-04 |
| [TASK-061](TASK-061-be-p3-repo-lists-defense.md) | BE repo/lists defense | P3 | Открыта | be1 | W12 P3-BE-05,06 |
| [TASK-062](TASK-062-be-p3-dead-code-layering.md) | BE dead code/layering | P3 | Открыта | be2 | W12 P3-BE-07 |
| [TASK-063](TASK-063-be-p3-errors-input-hygiene.md) | BE errors + input hygiene | P3 | Открыта | be1 | W12 P3-BE-08,09 |
| [TASK-064](TASK-064-be-p3-ws-calls-lifecycle.md) | BE WS/calls lifecycle | P3 | Открыта | be2 | W12 P3-BE-10 |
| [TASK-065](TASK-065-be-p3-platform-http-obs.md) | BE platform HTTP/obs | P3 | Открыта | be1 | W12 P3-BE-11,12 |
| [TASK-066](TASK-066-be-p3-ai-auth-rbac-audit.md) | BE AI/auth/rbac/audit | P3 | Открыта | be2 | W12 P3-BE-13,14 |
| [TASK-067](TASK-067-be-p3-v4-dev-storage.md) | BE v4/dev/storage | P3 | Открыта | be1 | W12 P3-BE-15,16 |
| [TASK-068](TASK-068-be-p3-ent-hygiene-mojibake.md) | BE Ent hygiene/mojibake | P3 | Открыта | be2 | W12 P3-BE-17,18 |
| [TASK-069](TASK-069-fe-p3-a11y-names.md) | FE a11y names residual | P3 | Открыта | fe1 | W12 P3-FE-01 |
| [TASK-070](TASK-070-fe-p3-a11y-structure.md) | FE a11y structure | P3 | Открыта | fe2 | W12 P3-FE-02 |
| [TASK-071](TASK-071-fe-p3-tokens-i18n.md) | FE tokens + i18n | P3 | Открыта | fe1 | W12 P3-FE-03,04 |
| [TASK-072](TASK-072-fe-p3-timers-cleanup.md) | FE timers + NaN | P3 | Открыта | fe2 | W12 P3-FE-05,16 |
| [TASK-073](TASK-073-fe-p3-silent-fail-empty.md) | FE silent fail residual | P3 | Открыта | fe1 | W12 P3-FE-06 |
| [TASK-074](TASK-074-fe-p3-type-casts.md) | FE type casts | P3 | Открыта | fe2 | W12 P3-FE-07 |
| [TASK-075](TASK-075-fe-p3-dead-ui-state.md) | FE dead UI/state | P3 | Открыта | fe1 | W12 P3-FE-08 |
| [TASK-076](TASK-076-fe-p3-perf-stale.md) | FE perf stale | P3 | Открыта | fe2 | W12 P3-FE-09 |
| [TASK-077](TASK-077-fe-p3-api-auth-naming.md) | FE api/auth/naming | P3 | Открыта | fe1 | W12 P3-FE-10..12 |
| [TASK-078](TASK-078-fe-p3-docs-spa-ux.md) | FE docs/SPA/preview | P3 | Открыта | docs-writer | W12 P3-FE-13..15 |
| [TASK-079](TASK-079-gdpr-mongo-eraser-fail-closed.md) | GDPR Mongo eraser fail-closed + IT | P2 | Открыта | be1 | W18-C5-001/002 |
| [TASK-080](TASK-080-docs-metrics-drift-autotest.md) | Docs metrics drift autotest | P2 | Открыта | docs-writer | W18-C5 TASK-054 FAIL |

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
