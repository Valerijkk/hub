# MUTATION W5 slot 2 — TASK-144 (be2) · 2026-07-30

> Checker: **be2** (!= be1 fixer) · Env: localhost · Go 1.25.0  
> Criteria: `.planning/backlog/TASK-144-2026-07-29-a3-3-2-messenger-services.md` `[MUTATION CRITERIA]`

## Verdict

**6/6 PROVEN** · статус TASK-144 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W3-2-001 | **PROVEN** | `TestJoinChannel_IsMemberError_A3W3_2_001` | `isMember, _ := s.repo.IsMember(...)` (глотать err) | service |
| A3W3-2-002 | **PROVEN** | `TestCreateChannel_ExtraMembersFailRollsBack_A3W3_2_002` | убрать `DeleteChannel` при fail `AddMembers(extra)` | service |
| A3W3-2-003 | **PROVEN** | `TestAddBookmark_RejectsCrossChannel_A3W3_2_003` | убрать `msg.ChannelID != channelID` check | service |
| A3W3-2-004 | **PROVEN** | `TestToggleStarMuteUnmute_NonMemberForbidden_A3W3_2_004` | NotFound→`NewInternal` в ToggleStar/Mute/Unmute | service |
| A3W3-2-005 | **PROVEN** | `TestClosePoll_CreatorLeftForbiddenBeforeClose_A3W3_2_005` | убрать IsMember до `ClosePoll` | service |
| A3W3-2-006 | **PROVEN** | `TestAddMembers_DeleteChannel_WrapAppError_A3W3_2_006` | голый `return s.repo.AddMembers` / `DeleteChannel` | service |

Production restored (`git checkout --` после каждого BREAK). Mock-gap: нет (001–006 — mockRepository + service).

## Per-ID evidence

### A3W3-2-001 — JoinChannel IsMember err

```text
BREAK: service_channels.go — isMember, _ := IsMember (swallow err)
→ FAIL TestJoinChannel_IsMemberError_A3W3_2_001
   expected AppError 500, got <nil> (Join proceeded / AddMembers path)
RESTORE
→ PASS same test
```

### A3W3-2-002 — CreateChannel extra-members rollback

```text
BREAK: service_channels.go — drop DeleteChannel on AddMembers(extra) fail
→ FAIL TestCreateChannel_ExtraMembersFailRollsBack_A3W3_2_002
   DeletedAt == nil (orphan channel remains)
RESTORE
→ PASS same test
```

### A3W3-2-003 — AddBookmark channel-scope

```text
BREAK: service_bookmarks.go — skip msg.ChannelID != channelID
→ FAIL TestAddBookmark_RejectsCrossChannel_A3W3_2_003
   expected AppError 404, got <nil> (cross-channel bookmark created)
RESTORE
→ PASS same test
```

### A3W3-2-004 — ToggleStar/Mute/Unmute NotFound→403

```text
BREAK: service_channels.go — NotFound → NewInternal (ToggleStar + Mute + Unmute)
→ FAIL TestToggleStarMuteUnmute_NonMemberForbidden_A3W3_2_004
   expected 403, actual 500
RESTORE
→ PASS same test
```

### A3W3-2-005 — ClosePoll membership before close

```text
BREAK: service_polls.go — ClosePoll then GetPollResults (no IsMember pre-check)
→ FAIL TestClosePoll_CreatorLeftForbiddenBeforeClose_A3W3_2_005
   IsClosed=true при 403 creator-left
RESTORE
→ PASS same test
```

### A3W3-2-006 — AddMembers/DeleteChannel AppError wrap

```text
BREAK: service_channels.go — return s.repo.AddMembers / DeleteChannel raw
→ FAIL TestAddMembers_DeleteChannel_WrapAppError_A3W3_2_006
   expected AppError, got *errors.errorString: pg: connection reset
RESTORE
→ PASS same test
```

## Commands

```text
cd backend
# baseline + per-ID A3W3_2
go test -tags hubtest -count=1 -timeout=180s ./testframework/messenger/ -run A3W3_2
→ ok  ~1.8–2.9s (baseline + after each restore)

# TASK-144 related subset (invite/orphan/pin/promote/bot OOS smoke via name filter)
go test -tags hubtest -count=1 -timeout=300s ./testframework/messenger/ \
  -run "A3W3_2|Invite|orphan|TogglePin|promote|NotConfigured|CreateChannel|JoinChannel|AddBookmark|ClosePoll|AddMembers|DeleteChannel|ToggleStar|Mute"
→ ok  ~2.0s

# full package -race (observed parallel interference)
go test -tags hubtest -race -count=1 -timeout=300s ./testframework/messenger/
→ FAIL TestAddReaction_ConcurrentFirstReact_NoDuplicate_A2W3_3_004
   cause: OOS working-tree mutation in mongo_repository.go
   ("MUTATED: no $ne emoji guard" — parallel agent, not TASK-144)
   be2 did not touch / restore that file
```

## Assumptions / unresolved

- Full `./testframework/messenger/` `-race` FAIL — **не** регрессия TASK-144: чужая незакрытая мутация `mongo_repository.go` (TASK-089 reaction concurrent). A3W3_2 + related subset **PASS**.
- Коммит не создавался (инструкция слота).

## Handoff

TASK-144 **Закрыта** · evidence этот файл · README синхронизирован · be-ver → code-reviewer.
