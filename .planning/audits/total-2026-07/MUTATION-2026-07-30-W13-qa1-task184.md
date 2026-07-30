# MUTATION W13 slot 4 refill — TASK-184 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-184-2026-07-29-c-w19-last-team-member.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**0/1 PROVEN** · **1 NOT_PROVEN (implement-guard GAP)** · TASK-184 → **Открыта**

| ID | Verdict | RED |
|----|---------|-----|
| C-W19-TT-001 | **NOT_PROVEN** — GAP: `TeamMemberCount≤1` / Conflict guard **отсутствует** в `RemoveTeamMember` | baseline уже FAIL (не mutate) — kill recipe неприменим |

**Не fake-PROVEN:** mutate→RED невозможен, пока production guard не реализован. Leftover MUTATED: **нет** (мутация не применялась).

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## HEAD verify (pre-kill)

`backend/internal/project/service.go:624-632` на HEAD:

```go
func (s *ProjectService) RemoveTeamMember(ctx context.Context, teamID, actorID, userID uuid.UUID) error {
	if err := s.requireTeamMember(ctx, teamID, actorID); err != nil {
		return err
	}
	if err := s.repo.RemoveTeamMember(ctx, teamID, userID); err != nil {
		return apperrors.NewInternal("failed to remove team member")
	}
	return nil
}
```

- Нет вызова `TeamMemberCount`.
- Нет `apperrors.NewConflict("cannot remove the last member")`.
- Repo метод `TeamMemberCount` **есть** (`repository.go` / Ent) — guard можно добавить в service без нового API.

Criteria note («GAP: guard still missing») **подтверждён**. Session note be1 «FIXED» / STATE 2026-07-29 **не соответствует** текущему HEAD (false-close / не смержено / откат).

## Baseline evidence (без mutate)

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s -vet=off ./testframework/project/ `
  -run "TestRemoveTeamMember_(RejectsLastMember|AllowsWhenMultiple)_TASK184" -v
```

```text
=== RUN   TestRemoveTeamMember_RejectsLastMember_TASK184
    remove_team_member_last_task184_test.go:27:
        	Error: An error is expected but got nil.
--- FAIL: TestRemoveTeamMember_RejectsLastMember_TASK184 (0.01s)
=== RUN   TestRemoveTeamMember_AllowsWhenMultiple_TASK184
    remove_team_member_last_task184_test.go:59:
        	Error: An error is expected but got nil.
--- FAIL: TestRemoveTeamMember_AllowsWhenMultiple_TASK184 (0.08s)
FAIL  github.com/valerii/hub/backend/testframework/project  ~1.0s
```

Интерпретация: last-member remove возвращает `err==nil` → orphan team возможен; RED-тесты описывают **целевой** контракт, а не текущий.

## Why not mutate

Kill recipe criteria: «убрать `TeamMemberCount≤1` / Conflict guard → last-member remove `err==nil`».  
На HEAD **нечего убирать** — поведение уже = broken target. Mutate бы только подтвердил уже красный baseline → **не PROVEN**.

## Implement-guard GAP (для be1)

В `RemoveTeamMember` после `requireTeamMember`, **до** `repo.RemoveTeamMember`:

1. `count, err := s.repo.TeamMemberCount(ctx, teamID)`
2. если `count <= 1` → `return apperrors.NewConflict("cannot remove the last member of a team")` (или message, совпадающий с assert `cannot remove the last member`)
3. иначе — существующий `repo.RemoveTeamMember`

Acceptance после фикса:

- `TestRemoveTeamMember_RejectsLastMember_TASK184` GREEN
- `TestRemoveTeamMember_AllowsWhenMultiple_TASK184` GREEN
- Mutation: убрать guard → оба RED → restore → GREEN → тогда «Закрыта»

Dissolve-team API вне scope (не подменять).

## Handoff

- Статус TASK-184: **Открыта** (implement-guard GAP; не «На проверке»)
- Исполнитель: **be1** (повторная реализация guard + mutate self-check)
- После фикса: qa-auto1 re-slot mutate → RED → restore → GREEN
- Docs: `docs/wiki/API.md` DELETE last-member 409 — сверить при реализации (не трогали в этом слоте)
