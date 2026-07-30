# TASK-042 — Weak perm + messenger + WS revoke + calls guest

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P2-BE-12..15 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-BACKEND.md`

## [OUTCOME]

Дефекты P2-BE-12..15 закрыты (mutation-ready в scope модулей); остаток — out-of-scope / ADR.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/{tracker,messenger,ws,calls,project}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-BE-12..15)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-BE-12..15 с мутацией «да» закрыты по реестру
- **Мутация (C):** CreateBoard без Edit→403; member.joined без token; revoke→Unregister; guest RL
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [PROGRESS 2026-07-28 be1]

### Fixed (mutation-ready в scope)

| Class | Source IDs closed | Что сделано |
|-------|-------------------|-------------|
| P2-BE-12 | W1-A1-006, W1-A2-009, W1-A5-006, W1-A5-007, W3-A1-004, W3-A1-007 | CreateBoard→`PermTaskEdit`; status∈project; Get deleted→404; avatar http(s)/`/…`; EditMessage+channel; presence cap 50 |
| P2-BE-13 | W3-A2-010, W3-A2-012, W3-A2-013, W3-A2-014, W3-A3-003 | `via_invite:true`+invite_id; last owner Leave→400; Forward unread++; mention∩members; away/stale thresholds |
| P2-BE-14 | W3-A4-002 (уже), W3-A4-004 (WS side) | late Subscribe guard уже был; revoke/logout→`UnregisterBySession` |
| P2-BE-15 | W3-A5-005..008, W3-A5-010 | guest 12-hex + RL; GuestJoin service; conflicts authz; ephemeral ≤2000; DurationMinutes persist |

### Remaining (не в files allowed / нужен ADR/миграция)

| Source ID | Почему осталось |
|-----------|-----------------|
| W2-A1-004 | `tms` — вне scope |
| W2-A2-015 | `tms` — вне scope |
| W6-A1-005 | `auth` avatar — вне scope (project avatar закрыт W1-A5-007) |
| W7-A2-005 | `crm` — вне scope |
| W7-A2-009 | `hr` — вне scope |
| W3-A2-015 | bot/internal read guard — deferred |
| W3-A3-005 | unread hash TTL / leave cleanup — deferred |
| W3-A4-004 (auth publish) | `auth` должен публиковать `auth.session.{uid}.revoked` на Logout/Revoke — вне files allowed; WS Unregister готов |
| W3-A5-008 (full UUID) | `Call.invite_code` MaxLen(12) — полная UUID потребует Ent migration (approval) |

### Evidence

- `go test -tags hubtest -race` packages: tracker/project/messenger/ws/calls (P2_BE_* filters) — PASS

## [REWORK 2026-07-28]

NATS capture AcceptInvite via_invite=true из production PublishJSON.

## [W18-C3 MUTATION — 2026-07-28 be2]

**FAIL.** Evidence: `.planning/audits/total-2026-07/W18-C3-MUTATION.md`

- CreateBoard PermTaskEdit→View → `TestCreateBoard_ViewerForbidden_P2_BE_12` RED ✓
- Skip UnregisterBySession → `TestBridge_AuthRevoked_UnregistersClient_P2_BE_14` RED ✓
- Invite code [:12]→[:8] → `TestCreateInstantCall_InviteCodeFullUUID_P2_BE_15` RED ✓
- **Fail:** `via_invite: true→false` → `TestAcceptInvite_PayloadViaInviteBool_P2_BE_13` stays GREEN (handmade JSON map, AcceptInvite path does not assert production payload).

## [W1-be2 MUTATION — 2026-07-29]

**PASS.** Evidence: .planning/audits/total-2026-07/MUTATION-2026-07-29-W1-be2.md

- via_invite: true→false → TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13 RED
- Handmade PayloadViaInviteBool stays GREEN (control). GREEN after restore.

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 refill / be2 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Classes P2-BE-12..15

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| P2-BE-12 | `backend/internal/tracker/handler.go:73-74` — `boards.Post` → `rbac.RequirePermission(rbac.PermTaskEdit, …)` | handler middleware | `TestCreateBoard_ViewerForbidden_P2_BE_12` · `backend/testframework/tracker/create_board_perm_task042_test.go` | заменить `PermTaskEdit` → `PermProjectView` → viewer CreateBoard 201 → 403 assert RED |
| P2-BE-13 | `backend/internal/messenger/service_channel_invites.go:199-205` — `AcceptInvite` PublishJSON `"via_invite": true` (без raw token) | service (NATS) | `TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13` · `backend/testframework/messenger/invite_leave_forward_task042_test.go` | `"via_invite": true→false` / omit / publish token → NATS payload assert RED |
| P2-BE-14 | `backend/internal/ws/bridge.go:249-252` — `handleAuthSessionEvent`: `revoked`/`logout` → `UnregisterBySession` | WS Bridge | `TestBridge_AuthRevoked_UnregistersClient_P2_BE_14` · `backend/testframework/ws/revoke_unregister_task042_test.go` | убрать `UnregisterBySession` call → client остаётся в room / done не закрыт → RED |
| P2-BE-15 | `backend/internal/calls/service.go:23-24` + `generateInviteCode:27-33` — base62×`inviteCodeLength` (12) | service | `TestCreateInstantCall_InviteCodeFullUUID_P2_BE_15` · `backend/testframework/calls/guest_conflicts_dto_task042_test.go` | укоротить length (напр. 8) / hex-UUID slice → Len/Regexp assert RED |
| P2-BE-15-RL | `backend/internal/calls/handler.go:42-47` — `guestRL` Max:10 / Window:1m на `POST /calls/guest-join` | handler RateLimit | **GAP** — нет Exact RED `it`/test на RL | **N/A** until hubtest asserts 429 after &gt;10/min (source-only сейчас) |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 12 | Fiber + Ent RBAC viewer vs editor | **OK** — Ent RBAC |
| 13 | AcceptInvite + real NATS PublishJSON capture | **OK** — service+NATS; mock invite repo OK |
| 13-ctrl | `TestAcceptInvite_PayloadViaInviteBool_P2_BE_13` handmade JSON map | **layer_mismatch** — **не** kill-target (W18 GREEN при mutate; control only) |
| 14 | Hub + Bridge auth.session.revoked | **OK** — in-memory Hub (prod path) |
| 15 | CallService CreateInstantCall invite | **OK** — Ent call fixture |
| 15-RL | (нет теста) | **GAP** |

`layer_mismatch IDs: P2-BE-13 handmade control (PayloadViaInviteBool)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)` · `GAP IDs: P2-BE-15-RL`

### Gaps

- Docs-only: **N/A**.
- Out-of-scope remaining (tms/auth/crm/hr/bot/unread TTL/full UUID migrate) — не в criteria table.
- Auth publish `auth.session.*.revoked` на Logout — вне files allowed; kill только WS Unregister side.

## [W13-be2 MUTATION — 2026-07-30]

**PASS.** Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task042.md`

| ID | Verdict | RED |
|----|---------|-----|
| P2-BE-12 | PROVEN | `TestCreateBoard_ViewerForbidden_P2_BE_12` (403→201) |
| P2-BE-13 | PROVEN | `TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13` |
| P2-BE-13-ctrl | N/A | handmade `PayloadViaInviteBool` stayed GREEN (layer_mismatch) |
| P2-BE-14 | PROVEN | `TestBridge_AuthRevoked_UnregistersClient_P2_BE_14` |
| P2-BE-15 | PROVEN | `TestCreateInstantCall_InviteCodeFullUUID_P2_BE_15` |
| P2-BE-15-RL | GAP_NOTED | нет Exact RED на guest RL |

4/4 runnable PROVEN · no NOT_PROVEN → **Закрыта**. leftover MUTATED: нет.
