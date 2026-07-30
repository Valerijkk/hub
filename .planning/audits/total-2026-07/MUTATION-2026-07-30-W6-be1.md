# MUTATION W6 slot 1 — TASK-145 (be1) · 2026-07-30

> Checker: **be1** · Env: SQLite in-memory (`newMessengerRepoEnv`; deployments-postgres/mongo up, не нужны для Ent-cap тестов) · Criteria: `.planning/backlog/TASK-145-2026-07-29-a3-3-3-messenger-repos.md` `[MUTATION CRITERIA]`
> Коммит не создавался (инструкция слота).

## Verdict

**Runnable PROVEN: 001a + 001b(GetChannel) + 002** · **N/A (не fake-PROVEN):** FindDM weak · 003 godoc-only  
Статус TASK-145 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W3-3-001a | **PROVEN** | `TestEntListUserChannels_HasHardCap_A3W3_3_001` | убрать `Limit(userChannelMembershipCap)` с `ListUserChannels` | repository |
| A3W3-3-001b | **PROVEN** (GetChannel) | `TestEntGetChannel_WithMembersCap_A3W3_3_001` | `WithMembers()` без `q.Limit(listMembersCap)` на GetChannel | repository |
| A3W3-3-001 FindDM | **N/A** weak | `TestEntFindDMChannel_WithMembersCap_A3W3_3_001` | снять Limit с FindDM → тест всё равно PASS (2 members ≤cap) | — |
| A3W3-3-002 | **PROVEN** | `TestEntListMembers_SkipsSoftDeletedChannel_A3W3_3_002` | убрать `HasChannelWith(channel.DeletedAtIsNil())` с `ListMembers` | repository |
| A3W3-3-003 | **N/A** godoc | — | runtime mutation N/A; godoc sync `presenceAwayThreshold (90s)` | docs/godoc |

Production restored (`git checkout --` после каждого BREAK). `git diff` на `ent_repository.go` / `presence.go` пуст.

## Per-ID evidence

### A3W3-3-001a — ListUserChannels hard cap

```text
BREAK: ent_repository.go ListUserChannels — drop Limit(userChannelMembershipCap)
→ FAIL TestEntListUserChannels_HasHardCap_A3W3_3_001
   (expected ≤500 / =500; got 501)
RESTORE
→ PASS same test (~3.8s)
```

### A3W3-3-001b — GetChannel WithMembers cap

```text
BREAK: GetChannel WithMembers() without Limit
→ FAIL TestEntGetChannel_WithMembersCap_A3W3_3_001
   (expected ≤500 / =500; got 501)
RESTORE
→ PASS same test (~3.5s)
```

### A3W3-3-001 FindDM — weak gap (N/A, not PROVEN)

```text
BREAK: FindDMChannel WithMembers() without Limit
→ PASS TestEntFindDMChannel_WithMembersCap_A3W3_3_001 (2 members always ≤cap)
RESTORE
→ не считается PROVEN (criteria Gaps)
```

### A3W3-3-002 — ListMembers soft-delete

```text
BREAK: ListMembers — drop HasChannelWith(channel.DeletedAtIsNil())
→ FAIL TestEntListMembers_SkipsSoftDeletedChannel_A3W3_3_002
   (after DeleteChannel: members not empty)
RESTORE
→ PASS same test
```

### A3W3-3-003 — godoc only (N/A runtime)

```text
presence.go:154-155:
  // GetOnlineUserIDs … (heartbeat менее presenceAwayThreshold (90s) назад). A3W3-3-003.
presenceAwayThreshold = 90 * time.Second (presence.go:21)
Нет runtime guard → mutation N/A; evidence = godoc sync (не TestGetOnlineUserIDs_*).
```

## Commands

```text
cd backend
# baseline / after restores
go test -tags hubtest -race -count=1 -timeout=300s ./testframework/messenger/ \
  -run "TestEntListUserChannels_HasHardCap_A3W3_3_001|TestEntGetChannel_WithMembersCap_A3W3_3_001|TestEntListMembers_SkipsSoftDeletedChannel_A3W3_3_002"
→ ok  ~3.3–3.8s

# full domain regression
go test -tags hubtest -race -count=1 -timeout=300s ./testframework/messenger/
→ ok  ~6.1s
```

## Assumptions / unresolved

- Коммит не создавался.
- FindDM WithMembers cap — **gap** (слабый тест); канон 001b = GetChannel.
- 003 — godoc FIXED в коде; runtime PROVEN = N/A.
- `layer_mismatch IDs: (none)`

## Handoff

TASK-145 **Закрыта** · evidence этот файл · README синхронизирован · be-ver → code-reviewer.
