# MUTATION W8 slot 3 — TASK-116 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost (SQLite/Ent in-memory auth tests) · Criteria: `.planning/backlog/TASK-116-2026-07-28-auth-avatar-xss-validation-tests.md` `[MUTATION CRITERIA]`
> Коммит не создавался (инструкция слота).

## Verdict

**Runnable PROVEN: 2/2** (A2W6-A1-003, A2W6-A1-004)  
Статус TASK-116 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A2W6-A1-003 | **PROVEN** | `TestProfileService_UpdateProfile_AvatarURL_SchemeWhitelist_A2W6_A1_003` | убрать вызов `validateProfileAvatarURL` из `UpdateProfile` | service |
| A2W6-A1-004 | **PROVEN** | `TestHandler_ListSessions_MarksCurrent` | `IsCurrent: s.ID != currentSessionID` (инверсия; `false` без использования sid → compile fail unused) | handler |

Production restored (`git checkout --` после каждого BREAK). `git diff` на `service_profile.go` / `handler.go` пуст.

## Per-ID evidence

### A2W6-A1-003 — AvatarURL scheme whitelist

```text
BASELINE → PASS (~4.0s)

BREAK: service_profile.go UpdateProfile — drop validateProfileAvatarURL(req.AvatarURL)
→ FAIL TestProfileService_UpdateProfile_AvatarURL_SchemeWhitelist_A2W6_A1_003
   (avatar "javascript:alert(1)" must be rejected — error expected but got nil)
RESTORE (git checkout -- backend/internal/auth/service_profile.go)
→ PASS same test (~3.4s)
```

### A2W6-A1-004 — ListSessions is_current

```text
BREAK attempt: IsCurrent: false → build failed (currentSessionID declared and not used)
BREAK: handler.go:350 — IsCurrent: s.ID != currentSessionID (criteria alt)
→ FAIL TestHandler_ListSessions_MarksCurrent
   (ровно одна сессия должна быть is_current — expected 1, actual 0)
RESTORE (git checkout -- backend/internal/auth/handler.go)
→ PASS both named tests (~3.3s)
```

## Commands

```text
cd backend
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/auth/ \
  -run "TestProfileService_UpdateProfile_AvatarURL_SchemeWhitelist_A2W6_A1_003|TestHandler_ListSessions_MarksCurrent"
→ ok  ~3.3–4.0s (baseline + final restore)
```

## Assumptions / unresolved

- Коммит не создавался.
- `IsCurrent: false` как единственная мутация не компилируется (unused `currentSessionID`); criteria допускает `s.ID != currentSessionID` — использован этот якорь.
- `layer_mismatch IDs: (none)`

## Handoff

TASK-116 **Закрыта** · evidence этот файл · README синхронизирован · be-ver → code-reviewer.
