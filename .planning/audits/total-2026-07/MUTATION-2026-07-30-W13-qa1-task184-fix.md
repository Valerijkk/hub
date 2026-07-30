# MUTATION W13 slot 4 refill FIX — TASK-184 (qa-auto1) · 2026-07-30

> Checker/implementer: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-184-2026-07-29-c-w19-last-team-member.md` `[MUTATION CRITERIA]`  
> Prior GAP: [`MUTATION-2026-07-30-W13-qa1-task184.md`](MUTATION-2026-07-30-W13-qa1-task184.md)  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**1/1 PROVEN** · **0 NOT_PROVEN** · TASK-184 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| C-W19-TT-001 | **PROVEN** | `TestRemoveTeamMember_RejectsLastMember_TASK184`; `TestRemoveTeamMember_AllowsWhenMultiple_TASK184` — убрать `TeamMemberCount≤1` / Conflict → last-member `err==nil` (orphan) |

Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Implement summary

В `ProjectService.RemoveTeamMember` после `requireTeamMember`, до `repo.RemoveTeamMember`:

1. `count, err := s.repo.TeamMemberCount(ctx, teamID)`
2. `count <= 1` → `apperrors.NewConflict("cannot remove the last member of a team")` (HTTP **409**)
3. иначе — `repo.RemoveTeamMember`

Регрессия wave122: `TestService_Wave122_RemoveTeamMember` переведён с orphan-OK на 409 Conflict (старый контракт противоречил TASK-184).

Docs: `docs/wiki/API.md` уже имел DELETE last-member → 409; `docs/TESTING.md` — TC-PROJ-11/12.

## Evidence chain

### 1. GREEN (post-implement, no mutate)

```text
go test -tags hubtest -count=1 -timeout=180s -vet=off ./testframework/project/ `
  -run "TestRemoveTeamMember_(RejectsLastMember|AllowsWhenMultiple)_TASK184|TestService_Wave122_RemoveTeamMember|TestService_Wave124_RemoveTeamMember|TestRemoveTeamMember_Success" -v

=== RUN   TestRemoveTeamMember_RejectsLastMember_TASK184
--- PASS: TestRemoveTeamMember_RejectsLastMember_TASK184 (0.01s)
=== RUN   TestRemoveTeamMember_AllowsWhenMultiple_TASK184
--- PASS: TestRemoveTeamMember_AllowsWhenMultiple_TASK184 (0.01s)
=== RUN   TestRemoveTeamMember_Success
--- PASS: TestRemoveTeamMember_Success (0.00s)
=== RUN   TestService_Wave122_RemoveTeamMember
--- PASS: TestService_Wave122_RemoveTeamMember (0.01s)
=== RUN   TestService_Wave124_RemoveTeamMember
--- PASS: TestService_Wave124_RemoveTeamMember (0.01s)
PASS  ok  github.com/valerii/hub/backend/testframework/project  ~0.9s
```

### 2. MUTATE → RED

Kill: removed `TeamMemberCount` + `NewConflict` block from `RemoveTeamMember` (`// MUTATED` comment).

```text
=== RUN   TestRemoveTeamMember_RejectsLastMember_TASK184
    remove_team_member_last_task184_test.go:27: An error is expected but got nil.
--- FAIL: TestRemoveTeamMember_RejectsLastMember_TASK184 (0.01s)
=== RUN   TestRemoveTeamMember_AllowsWhenMultiple_TASK184
    remove_team_member_last_task184_test.go:59: An error is expected but got nil.
--- FAIL: TestRemoveTeamMember_AllowsWhenMultiple_TASK184 (0.01s)
FAIL
```

### 3. RESTORE → GREEN

Guard restored. No `MUTATED` leftover in `service.go`.

```text
=== RUN   TestRemoveTeamMember_RejectsLastMember_TASK184
--- PASS
=== RUN   TestRemoveTeamMember_AllowsWhenMultiple_TASK184
--- PASS
=== RUN   TestService_Wave122_RemoveTeamMember
--- PASS
=== RUN   TestService_Wave124_RemoveTeamMember
--- PASS
PASS  ok  ~1.0s
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s -vet=off ./testframework/project/ `
  -run "TestRemoveTeamMember_(RejectsLastMember|AllowsWhenMultiple)_TASK184" -v
```

## Handoff

- Статус TASK-184: **Закрыта**
- Исполнитель слота: qa-auto1 (implement + mutate; originally be1 false-close не на HEAD)
- Dissolve-team API вне scope — не трогали
- Коммит не создавался
