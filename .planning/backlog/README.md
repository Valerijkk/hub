# Р‘СЌРєР»РѕРі Р·Р°РґР°С‡

> Р•РґРёРЅС‹Р№ РІС…РѕРґ РґР»СЏ Р°РіРµРЅС‚РѕРІ Рё РІР»Р°РґРµР»СЊС†Р°. РћРґРЅР° РїР°РїРєР°, РѕРґРЅР° С‚Р°Р±Р»РёС†Р°, РѕРґРёРЅ С„РѕСЂРјР°С‚.
> РћР±РЅРѕРІР»РµРЅРѕ: 2026-07-28 (W18-C5 mutation + catch-up TASK-079/080)

РљР°Р¶РґР°СЏ Р·Р°РґР°С‡Р° вЂ” РѕС‚РґРµР»СЊРЅС‹Р№ С„Р°Р№Р», РЅР°РїРёСЃР°РЅРЅС‹Р№ РїРѕ
[РјР°СЃС‚РµСЂ-РїСЂРѕРјРїС‚Сѓ](../../prompt-engineering/agents/templates/_MASTER-PROMPT-TEMPLATE.md).
РџСЂР°РІРёР»Р° РІРµРґРµРЅРёСЏ: [`.cursor/rules/57-backlog-tasks.mdc`](../../.cursor/rules/57-backlog-tasks.mdc).

---

## РћС‚РєСЂС‹С‚С‹Рµ

| ID | Р—Р°РґР°С‡Р° | РџСЂРёРѕСЂРёС‚РµС‚ | РЎС‚Р°С‚СѓСЃ | РСЃРїРѕР»РЅРёС‚РµР»СЊ | РСЃС‚РѕС‡РЅРёРє |
|----|--------|-----------|--------|-------------|----------|
| [TASK-008](TASK-008-total-audit-program.md) | РўРѕС‚Р°Р»СЊРЅС‹Р№ РїРѕС„Р°Р№Р»РѕРІС‹Р№ Р°СѓРґРёС‚ | P1 | Р’ СЂР°Р±РѕС‚Рµ (B: P0/P1/P2вњ… в†’ P3 TASK-057вЂ¦078) | Grok + СЃСѓР±Р°РіРµРЅС‚С‹ | Owner 2026-07-28 |
| [TASK-009](TASK-009-project-routes-rbac.md) | project routes Р±РµР· RBAC | **P0** | РСЃРїСЂР°РІР»РµРЅР° | be1 | W1-A5-001 |
| [TASK-010](TASK-010-promote-to-issue-rbac.md) | promote-to-issue Р±РµР· PermTaskCreate | **P0** | РСЃРїСЂР°РІР»РµРЅР° | be1 | W3-A2-001 |
| [TASK-011](TASK-011-webhook-event-match.md) | webhook UI events в‰  NATS subjects | **P0** | РСЃРїСЂР°РІР»РµРЅР° | be1 | W4-A5-001 |
| [TASK-012](TASK-012-webhook-project-scope.md) | webhook dispatch Р±РµР· project-scope | **P0** | РСЃРїСЂР°РІР»РµРЅР° | be1 | W4-A5-002 |
| [TASK-013](TASK-013-tracker-project-scope-idor.md) | Tracker IDOR workflow/fields/refs | P1 | РСЃРїСЂР°РІР»РµРЅР° | be1 | W12 P1-BE-01..04 |
| [TASK-014](TASK-014-project-acl-sprint-tx.md) | Project ACL + sprint atomic | P1 | РСЃРїСЂР°РІР»РµРЅР° | be2в†’rework | W12 P1-BE-05..09 В· ACL+Ent mid-fail |
| [TASK-015](TASK-015-tms-project-scope-idor-refs.md) | TMS suite/env IDOR + FK | P1 | РСЃРїСЂР°РІР»РµРЅР° | be2 | W12 P1-BE-10..12 В· W18-C2 PASS |
| [TASK-016](TASK-016-tms-import-limits-steps.md) | TMS import limits + steps | P1 | РСЃРїСЂР°РІР»РµРЅР° | be2 | W12 P1-BE-13..14 В· W18-C2 PASS |
| [TASK-017](TASK-017-messenger-membership-idor.md) | Messenger membership IDOR | P1 | РСЃРїСЂР°РІР»РµРЅР° | be1 | W12 P1-BE-15..20 |
| [TASK-018](TASK-018-ws-calls-authz-race.md) | WS race + scheduled calls authz | P1 | РСЃРїСЂР°РІР»РµРЅР° | be1 | W12 P1-BE-21..22 |
| [TASK-019](TASK-019-platform-ssrf-outbox-pool-config.md) | Platform SSRF/outbox/pool/config | P1 | РСЃРїСЂР°РІР»РµРЅР° | be2в†’rework | W12 P1-BE-23..26 В· To4 normalize proof |
| [TASK-020](TASK-020-apigateway-webhook-create-contract.md) | Webhook secret once + allowlist | P1 | РСЃРїСЂР°РІР»РµРЅР° | be2 | W12 P1-BE-27..28 В· W18-C2 PASS |
| [TASK-021](TASK-021-ai-policy-bot-authz.md) | AI policy + bot authz | P1 | РСЃРїСЂР°РІР»РµРЅР° | be1 | W12 P1-BE-29..31 |
| [TASK-022](TASK-022-search-nf03-softdelete.md) | Search soft-delete + private | P1 | РСЃРїСЂР°РІР»РµРЅР° | be2 | W12 P1-BE-32 В· W18-C2 PASS |
| [TASK-023](TASK-023-rbac-cache-gdpr.md) | RBAC cache + GDPR erase | P1 | РСЃРїСЂР°РІР»РµРЅР° | be1 | W12 P1-BE-33..34 В· W18-C3 PASS |
| [TASK-024](TASK-024-notif-integrations-storage-backup.md) | Notif/Git*/storage/backup/coffee | P1 | РСЃРїСЂР°РІР»РµРЅР° | be2 | W12 P1-BE-35..39,43 В· W18-C2 PASS |
| [TASK-025](TASK-025-ent-schema-constraints.md) | Ent unique/O2O constraints | P1 | РСЃРїСЂР°РІР»РµРЅР° | be1 | W12 P1-BE-40..42 В· W18-C3 PASS |
| [TASK-026](TASK-026-fe-api-client-auth-retry.md) | FE api client auth retry | P1 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 FE-P1-API-CLIENT |
| [TASK-027](TASK-027-fe-callchat-ws-subscribe.md) | FE CallChat WS subscribe | P1 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 FE-P1-CALLS-CHAT |
| [TASK-028](TASK-028-fe-calls-join-create-lifecycle.md) | FE calls lifecycle | P1 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 FE-P1-CALLS-LIFECYCLE |
| [TASK-029](TASK-029-fe-messenger-reply-thread-unread.md) | FE messenger reply/unread | P1 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 FE-P1-MESSENGER |
| [TASK-030](TASK-030-fe-board-dnd-single-path-rank.md) | FE board DnD + rank | P1 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 FE-P1-BOARD-DND |
| [TASK-031](TASK-031-fe-sprint-loading-error-states.md) | FE sprint loading/error | P1 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 FE-P1-SPRINT-UI |
| [TASK-032](TASK-032-fe-trash-list-and-webhooks-project.md) | FE trash + webhooks projectId | P1 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C5 | fe2 | W12 FE-P1-TRASH-WEBHOOKS |
| [TASK-033](TASK-033-fe-error-empty-auth-race.md) | FE errorв‰ empty + auth race | P1 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C5 | fe2 | W12 FE-P1-ERROR-AS-EMPTY |
| [TASK-034](TASK-034-fe-a11y-nested-sidebar-notifications.md) | FE a11y nested interactive | P1 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C5 | fe2 | W12 FE-P1-A11Y-CONTROLS |
| [TASK-035](TASK-035-fe-tms-compare-and-docs-testpath.md) | FE TMS compare + docs drift | P1 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 FE-P1-TMS-DOCS |
| [TASK-036](TASK-036-be-p2-dto-error-contract.md) | BE DTO validator + error contract | P2 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C2b | be2 | W12 P2-BE-01,07 |
| [TASK-037](TASK-037-be-p2-authz-ent-negative-tests.md) | BE Ent-negative authz tests | P2 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C2b | be2 | W12 P2-BE-02 |
| [TASK-038](TASK-038-be-p2-ol-softdelete-tx-lists.md) | BE OL/soft-mutate/tx/lists | P2 | РСЃРїСЂР°РІР»РµРЅР° | be1 | W12 P2-BE-03..06 В· W18-C3 PASS (partial residual OK) |
| [TASK-039](TASK-039-be-p2-soft-delete-secrets.md) | BE soft-delete + secrets | P2 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C2b | be2 | W12 P2-BE-08,09 |
| [TASK-040](TASK-040-be-p2-fail-closed-ratelimit.md) | BE fail-closed + rate-limit | P2 | РСЃРїСЂР°РІР»РµРЅР° | be1 | W12 P2-BE-10 В· W18-C3 PASS |
| [TASK-041](TASK-041-be-p2-sprint-poker-start.md) | BE sprint poker/start | P2 | РћС‚РєСЂС‹С‚Р° | be1 | W12 P2-BE-11 В· W18-C3 FAIL |
| [TASK-042](TASK-042-be-p2-perm-messenger-ws-calls.md) | BE perm/messenger/ws/calls | P2 | РћС‚РєСЂС‹С‚Р° | be1 | W12 P2-BE-12..15 В· W18-C3 FAIL |
| [TASK-043](TASK-043-be-p2-platform-outbox-obs.md) | BE platform outbox/obs | P2 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C2b | be2 | W12 P2-BE-16 |
| [TASK-044](TASK-044-be-p2-ai-search.md) | AI stream/privacy + search JQL/SSRF | P2 | Исправлена | be1→rework | W12 P2-BE-17/18 · Meili+cancel IT |
| [TASK-045](TASK-045-be-p2-rbac-gdpr-integrations-storage.md) | BE rbac/gdpr/imex/storage | P2 | РСЃРїСЂР°РІР»РµРЅР° | be2в†’rework | W12 P2-BE-19..21 В· Confirm ghost IT |
| [TASK-046](TASK-046-be-p2-ent-unique-fk-check.md) | BE Ent unique/FK/CHECK | P2 | РћС‚РєСЂС‹С‚Р° | be1 | W12 P2-BE-22 В· W18-C3 FAIL |
| [TASK-047](TASK-047-fe-p2-error-vs-empty.md) | FE errorв‰ empty sweep | P2 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 P2-FE-01 |
| [TASK-048](TASK-048-fe-p2-mutation-feedback.md) | FE mutation feedback | P2 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C5 | fe2 | W12 P2-FE-02 |
| [TASK-049](TASK-049-fe-p2-a11y-names-dialogs.md) | FE a11y names/dialogs | P2 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 P2-FE-03,04 |
| [TASK-050](TASK-050-fe-p2-tokens-dead-ui.md) | FE tokens/dead UI/authz | P2 | РСЃРїСЂР°РІР»РµРЅР° | fe1 | W12 P2-FE-05,06,16 |
| [TASK-051](TASK-051-fe-p2-effects-hooks-race.md) | FE effects/hooks race | P2 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C5 | fe2 | W12 P2-FE-07 |
| [TASK-052](TASK-052-fe-p2-api-dto-gaps.md) | FE API DTO gaps | P2 | РћС‚РєСЂС‹С‚Р° В· W18-C5 FAIL | fe2 | W12 P2-FE-08 |
| [TASK-053](TASK-053-fe-p2-auth-ws-upload.md) | FE auth/WS/upload | P2 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C5 | fe2 | W12 P2-FE-09..11 |
| [TASK-054](TASK-054-fe-p2-docs-metrics-drift.md) | Docs metrics/wiring drift | P2 | РћС‚РєСЂС‹С‚Р° В· W18-C5 FAIL | docs-writer | W12 P2-FE-12 |
| [TASK-055](TASK-055-fe-p2-shell-tracker-ux.md) | FE shell + tracker UX | P2 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C5 | fe2 | W12 P2-FE-13,14 |
| [TASK-056](TASK-056-fe-p2-i18n-push-rooms.md) | FE i18n/push/rooms | P2 | РСЃРїСЂР°РІР»РµРЅР° В· W18-C5 | fe2 | W12 P2-FE-15 |
| [TASK-057](TASK-057-be-p3-silent-error-swallow.md) | BE silent error swallow | P3 | РћС‚РєСЂС‹С‚Р° | be1 | W12 P3-BE-01 |
| [TASK-058](TASK-058-be-p3-assert-free-tests.md) | BE assert-free tests | P3 | РћС‚РєСЂС‹С‚Р° | be2 | W12 P3-BE-02 |
| [TASK-059](TASK-059-be-p3-mutation-proof-gaps.md) | BE mutation-proof gaps | P3 | РћС‚РєСЂС‹С‚Р° | be1 | W12 P3-BE-03 |
| [TASK-060](TASK-060-be-p3-utf8-rune-truncate.md) | BE UTF-8 rune truncate | P3 | РћС‚РєСЂС‹С‚Р° | be2 | W12 P3-BE-04 |
| [TASK-061](TASK-061-be-p3-repo-lists-defense.md) | BE repo/lists defense | P3 | РћС‚РєСЂС‹С‚Р° | be1 | W12 P3-BE-05,06 |
| [TASK-062](TASK-062-be-p3-dead-code-layering.md) | BE dead code/layering | P3 | РћС‚РєСЂС‹С‚Р° | be2 | W12 P3-BE-07 |
| [TASK-063](TASK-063-be-p3-errors-input-hygiene.md) | BE errors + input hygiene | P3 | РћС‚РєСЂС‹С‚Р° | be1 | W12 P3-BE-08,09 |
| [TASK-064](TASK-064-be-p3-ws-calls-lifecycle.md) | BE WS/calls lifecycle | P3 | РћС‚РєСЂС‹С‚Р° | be2 | W12 P3-BE-10 |
| [TASK-065](TASK-065-be-p3-platform-http-obs.md) | BE platform HTTP/obs | P3 | РћС‚РєСЂС‹С‚Р° | be1 | W12 P3-BE-11,12 |
| [TASK-066](TASK-066-be-p3-ai-auth-rbac-audit.md) | BE AI/auth/rbac/audit | P3 | РћС‚РєСЂС‹С‚Р° | be2 | W12 P3-BE-13,14 |
| [TASK-067](TASK-067-be-p3-v4-dev-storage.md) | BE v4/dev/storage | P3 | РћС‚РєСЂС‹С‚Р° | be1 | W12 P3-BE-15,16 |
| [TASK-068](TASK-068-be-p3-ent-hygiene-mojibake.md) | BE Ent hygiene/mojibake | P3 | РћС‚РєСЂС‹С‚Р° | be2 | W12 P3-BE-17,18 |
| [TASK-069](TASK-069-fe-p3-a11y-names.md) | FE a11y names residual | P3 | РћС‚РєСЂС‹С‚Р° | fe1 | W12 P3-FE-01 |
| [TASK-070](TASK-070-fe-p3-a11y-structure.md) | FE a11y structure | P3 | РћС‚РєСЂС‹С‚Р° | fe2 | W12 P3-FE-02 |
| [TASK-071](TASK-071-fe-p3-tokens-i18n.md) | FE tokens + i18n | P3 | РћС‚РєСЂС‹С‚Р° | fe1 | W12 P3-FE-03,04 |
| [TASK-072](TASK-072-fe-p3-timers-cleanup.md) | FE timers + NaN | P3 | РћС‚РєСЂС‹С‚Р° | fe2 | W12 P3-FE-05,16 |
| [TASK-073](TASK-073-fe-p3-silent-fail-empty.md) | FE silent fail residual | P3 | РћС‚РєСЂС‹С‚Р° | fe1 | W12 P3-FE-06 |
| [TASK-074](TASK-074-fe-p3-type-casts.md) | FE type casts | P3 | РћС‚РєСЂС‹С‚Р° | fe2 | W12 P3-FE-07 |
| [TASK-075](TASK-075-fe-p3-dead-ui-state.md) | FE dead UI/state | P3 | РћС‚РєСЂС‹С‚Р° | fe1 | W12 P3-FE-08 |
| [TASK-076](TASK-076-fe-p3-perf-stale.md) | FE perf stale | P3 | РћС‚РєСЂС‹С‚Р° | fe2 | W12 P3-FE-09 |
| [TASK-077](TASK-077-fe-p3-api-auth-naming.md) | FE api/auth/naming | P3 | РћС‚РєСЂС‹С‚Р° | fe1 | W12 P3-FE-10..12 |
| [TASK-078](TASK-078-fe-p3-docs-spa-ux.md) | FE docs/SPA/preview | P3 | РћС‚РєСЂС‹С‚Р° | docs-writer | W12 P3-FE-13..15 |
| [TASK-079](TASK-079-gdpr-mongo-eraser-fail-closed.md) | GDPR Mongo eraser fail-closed + IT | P2 | РСЃРїСЂР°РІР»РµРЅР° | be1 | W18-C5-001/002 |
| [TASK-080](TASK-080-docs-metrics-drift-autotest.md) | Docs metrics drift autotest | P2 | РћС‚РєСЂС‹С‚Р° | docs-writer | W18-C5 TASK-054 FAIL |

## Р—Р°РєСЂС‹С‚С‹Рµ

| ID | Р—Р°РґР°С‡Р° | РџСЂРёРѕСЂРёС‚РµС‚ | РЎС‚Р°С‚СѓСЃ | РСЃРїРѕР»РЅРёС‚РµР»СЊ | РСЃС‚РѕС‡РЅРёРє |
|----|--------|-----------|--------|-------------|----------|
| [TASK-001](TASK-001-fr-intr-08-newspost-trash.md) | РќРѕРІРѕСЃС‚СЊ РЅРµРІРѕСЃСЃС‚Р°РЅРѕРІРёРјР° РёР· РєРѕСЂР·РёРЅС‹ | **P0** | Р—Р°РєСЂС‹С‚Р° | be1 | FR-INTR-08 |
| [TASK-002](TASK-002-fr-cal-05-series-trash.md) | РЈ РєР°Р»РµРЅРґР°СЂРЅС‹С… СЃРµСЂРёР№ РЅРµС‚ РєРѕСЂР·РёРЅС‹ | **P0** | Р—Р°РєСЂС‹С‚Р° | be1 | FR-CAL-05 |
| [TASK-004](TASK-004-state-goals-verified-reconcile.md) | STATE СЂР°СЃС…РѕРґРёС‚СЃСЏ СЃ РјР°С‚СЂРёС†РµР№ С„Р°Р· | P1 | Р—Р°РєСЂС‹С‚Р° | team-lead | Р РµРІСЊСЋ W4вЂ“W20 |
| [TASK-005](TASK-005-august-phase-dependencies.md) | РЎРІСЏР·Р°С‚СЊ Р°РІРіСѓСЃС‚РѕРІСЃРєРёРµ С„Р°Р·С‹ Р·Р°РІРёСЃРёРјРѕСЃС‚СЏРјРё | P1 | Р—Р°РєСЂС‹С‚Р° | team-lead | РђСѓРґРёС‚ РїР»Р°РЅРёСЂРѕРІР°РЅРёСЏ |
| [TASK-003](TASK-003-a14-002-test-granularity.md) | РўРµСЃС‚ РїСЂРѕРїСѓСЃРєР°РµС‚ СѓРґР°Р»РµРЅРёРµ РѕРґРЅРѕРіРѕ guard'Р° | P2 | Р—Р°РєСЂС‹С‚Р° | qa-auto1 | Р РµРІСЊСЋ W4вЂ“W20 |
| [TASK-006](TASK-006-disable-cursor-coauthor-attribution.md) | РќР°РІСЃРµРіРґР° РІС‹РєР»СЋС‡РёС‚СЊ Cursor Co-authored-by | **P0** | Р—Р°РєСЂС‹С‚Р° | devops/docs | Owner 2026-07-28 |
| [TASK-007](TASK-007-utf8-mojibake-audit-fix.md) | РђСѓРґРёС‚ Рё РїРѕС‡РёРЅРєР° Р±РёС‚РѕР№ UTF-8 РєРѕРґРёСЂРѕРІРєРё | P1 | Р—Р°РєСЂС‹С‚Р° | docs+be1 | Owner 2026-07-28 |

---

## РџРѕСЂСЏРґРѕРє РІС‹РїРѕР»РЅРµРЅРёСЏ (TASK-008)

```
P0:  TASK-009 в†’ 010 в†’ 011+012 (webhooks РІРјРµСЃС‚Рµ)
P1:  TASK-013вЂ¦025 (BE IDOR/authz СЂР°РЅСЊС€Рµ schema)
     TASK-026вЂ¦035 (FE)
P2/P3: РїРѕСЃР»Рµ P1; СЂРµРµСЃС‚СЂ СЃС‹СЂС‹С… вЂ” audits/total-2026-07/
Blocked ADR: W12-REGISTRY-P1-BACKEND В§В«С‚СЂРµР±СѓРµС‚ СѓС‚РѕС‡РЅРµРЅРёСЏВ» (U01вЂ“U06)
```

---

## Р›РµРіРµРЅРґР°

**РџСЂРёРѕСЂРёС‚РµС‚:** P0 Р±Р»РѕРєРµСЂ В· P1 РІР°Р¶РЅРѕ В· P2 РјРѕР¶РЅРѕ РѕС‚Р»РѕР¶РёС‚СЊ В· P3 РєРѕСЃРјРµС‚РёРєР°

**РЎС‚Р°С‚СѓСЃ:** `РћС‚РєСЂС‹С‚Р°` в†’ `Р’ СЂР°Р±РѕС‚Рµ` в†’ `РќР° РїСЂРѕРІРµСЂРєРµ` в†’ `Р—Р°РєСЂС‹С‚Р°` / `РћС‚РјРµРЅРµРЅР°`

РЎС‚Р°С‚СѓСЃ РјРµРЅСЏРµС‚СЃСЏ **РІ С‚РѕРј Р¶Рµ РєРѕРјРјРёС‚Рµ**, С‡С‚Рѕ Рё СЂР°Р±РѕС‚Р° РїРѕ Р·Р°РґР°С‡Рµ.

---

## Р§С‚Рѕ РЅРµ СЃСЋРґР°

| Р§С‚Рѕ | Р“РґРµ |
|-----|-----|
| РљСЂСѓРїРЅС‹Р№ Р±Р»РѕРє СЂР°Р±РѕС‚ | `.planning/phases/NN-*/PLAN.md` (GSD) |
| РРґРµСЏ Р±РµР· СЃСЂРѕРєР° | `ROADMAP.md`, СЃРµРєС†РёСЏ `## Backlog`, РЅСѓРјРµСЂР°С†РёСЏ `999.x` |
| Р РµС€РµРЅРёРµ, Р¶РґСѓС‰РµРµ РІР»Р°РґРµР»СЊС†Р° | [`../audits/planning-audit-2026-07/DECISIONS-PENDING.md`](../audits/planning-audit-2026-07/DECISIONS-PENDING.md) |
| РќР°С…РѕРґРєР° Р°СѓРґРёС‚Р° Р±РµР· РїРѕСЃС‚Р°РЅРѕРІРєРё | `.planning/audits/<Р°СѓРґРёС‚>/` |
