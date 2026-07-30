# MUTATION W13 slot 3 refill — TASK-042 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-042-be-p2-perm-messenger-ws-calls.md` `[MUTATION CRITERIA]`  
> Findings: **P2-BE-12**, **P2-BE-13**, **P2-BE-14**, **P2-BE-15** (+ **P2-BE-15-RL** GAP; handmade via_invite control)  
> Коммит не создавался. Handler/service/WS kills; leftover MUTATED: **нет**.

## Verdict

**4/4 runnable PROVEN** · **P2-BE-15-RL GAP_NOTED** · handmade control **N/A** · TASK-042 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| P2-BE-12 | **PROVEN** | `TestCreateBoard_ViewerForbidden_P2_BE_12` |
| P2-BE-13 | **PROVEN** | `TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13` |
| P2-BE-13-ctrl | **N/A** | `TestAcceptInvite_PayloadViaInviteBool_P2_BE_13` (layer_mismatch control; stayed GREEN under kill) |
| P2-BE-14 | **PROVEN** | `TestBridge_AuthRevoked_UnregistersClient_P2_BE_14` |
| P2-BE-15 | **PROVEN** | `TestCreateInstantCall_InviteCodeFullUUID_P2_BE_15` |
| P2-BE-15-RL | **GAP_NOTED** | нет Exact RED на guest RL 10/min |

`layer_mismatch IDs: P2-BE-13 handmade control (PayloadViaInviteBool)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)` · `GAP IDs: P2-BE-15-RL`

## Per-ID evidence

### P2-BE-12 — CreateBoard `PermTaskEdit` (не View)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/tracker/ \
    -run TestCreateBoard_ViewerForbidden_P2_BE_12$
  -> PASS (~1.0s)

BREAK (handler middleware correct layer):
  backend/internal/tracker/handler.go boards.Post
  — PermTaskEdit → PermProjectView
  — MUTATED P2-BE-12

  -> FAIL TestCreateBoard_ViewerForbidden_P2_BE_12 (~0.9s)
       expected: 403
       actual  : 201

RESTORE PermTaskEdit
  -> PASS (~1.0s)

git diff handler.go → empty
MUTATED marker → отсутствует
```

Layer OK: Fiber + Ent RBAC viewer vs editor.

### P2-BE-13 — AcceptInvite PublishJSON `"via_invite": true`

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/messenger/ \
    -run "TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13$|TestAcceptInvite_PayloadViaInviteBool_P2_BE_13$"
  -> PASS (~0.9s)

BREAK (service NATS correct layer):
  backend/internal/messenger/service_channel_invites.go AcceptInvite
  — "via_invite": true → false
  — MUTATED P2-BE-13

  -> FAIL TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13 (~1.0s)
       does not contain "\"via_invite\":true"
       expected: true / actual: false

  -> PASS TestAcceptInvite_PayloadViaInviteBool_P2_BE_13 (control; handmade JSON)

RESTORE via_invite: true
  -> PASS (~1.1s)

git diff service_channel_invites.go → empty
MUTATED marker → отсутствует
```

Layer OK: AcceptInvite + real NATS PublishJSON capture.  
Control **N/A**: handmade map stays GREEN under prod mutate (W18 layer_mismatch).

### P2-BE-14 — revoke/logout → `UnregisterBySession`

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/ws/ \
    -run TestBridge_AuthRevoked_UnregistersClient_P2_BE_14$
  -> PASS (~0.8s)

BREAK (WS Bridge correct layer):
  backend/internal/ws/bridge.go handleAuthSessionEvent
  — убрать UnregisterBySession call
  — MUTATED P2-BE-14

  -> FAIL TestBridge_AuthRevoked_UnregistersClient_P2_BE_14 (~1.4s)
       Condition never satisfied
       client.done must close after revoke

RESTORE UnregisterBySession
  -> PASS (~1.1s)

git diff bridge.go → empty
MUTATED marker → отсутствует
```

Layer OK: in-memory Hub + Bridge auth.session.revoked (prod path).

### P2-BE-15 — invite_code length 12 base62

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/calls/ \
    -run TestCreateInstantCall_InviteCodeFullUUID_P2_BE_15$
  -> PASS (~1.1s)

BREAK (service correct layer):
  backend/internal/calls/service.go inviteCodeLength
  — 12 → 8
  — MUTATED P2-BE-15

  -> FAIL TestCreateInstantCall_InviteCodeFullUUID_P2_BE_15 (~1.1s)
       Expect "…" to match "^[0-9A-Za-z]{12}$"

RESTORE inviteCodeLength = 12
  -> PASS (~1.1s)

git diff service.go → empty
MUTATED marker → отсутствует
```

Layer OK: CallService CreateInstantCall invite (Ent call fixture).

### P2-BE-15-RL — guest-join RateLimit Max:10 / Window:1m

```text
backend/internal/calls/handler.go:42-47 guestRL на POST /calls/guest-join
  Exact RED hubtest на 429 after >10/min — отсутствует
  → GAP_NOTED (source-only; N/A until dedicated RL assert)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s ./testframework/tracker/ -run TestCreateBoard_ViewerForbidden_P2_BE_12$
go test -tags hubtest -count=1 -timeout=180s ./testframework/messenger/ -run "TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13$|TestAcceptInvite_PayloadViaInviteBool_P2_BE_13$"
go test -tags hubtest -count=1 -timeout=180s ./testframework/ws/ -run TestBridge_AuthRevoked_UnregistersClient_P2_BE_14$
go test -tags hubtest -count=1 -timeout=180s ./testframework/calls/ -run TestCreateInstantCall_InviteCodeFullUUID_P2_BE_15$
# BASELINE PASS · FINAL PASS (tracker~1.0s · messenger~1.0s · ws~0.9s · calls~1.1s)
```

## Gaps

- **P2-BE-15-RL** — **GAP_NOTED** (нет Exact RED на guest RL).
- **P2-BE-13-ctrl** — **N/A** layer_mismatch (handmade JSON; не kill-target).
- Out-of-scope remaining (tms/auth/crm/hr/bot/unread TTL/full UUID migrate) — вне criteria table.
- Auth publish `auth.session.*.revoked` — вне files allowed; kill только WS Unregister side (PROVEN).

## Handoff

- Evidence: этот файл · TASK-042 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git diff` на 4 prod files — empty · no leftover MUTATED  
- Коммит не создавался
