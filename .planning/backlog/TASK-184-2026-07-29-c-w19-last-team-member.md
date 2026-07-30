# 🐛 C-W19: RemoveTeamMember last-member orphan

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P3** · Статус: **Закрыта**  
> Источник: [`C-W19-tracker-tms.md`](../audits/total-2026-07/C-W19-tracker-tms.md) · A2W1-A5-003  
> Исполнитель: **qa-auto1** (W13 slot 4 refill: implement + mutate) · Finding ID: `C-W19-TT-001`  
> Mutation GAP: [`MUTATION-2026-07-30-W13-qa1-task184.md`](../audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task184.md)  
> Mutation PROVEN: [`MUTATION-2026-07-30-W13-qa1-task184-fix.md`](../audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task184-fix.md)

## [OUTCOME]

`RemoveTeamMember` не оставляет команду без участников (last-member → 409 Conflict); orphan team без `requireTeamMember` невозможен.

## [ROLE]

- Role: qa-auto1 (refill) · Handoff: closed after mutate PROVEN

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/internal/project/service.go` RemoveTeamMember; repo TeamMemberCount если нужно; `testframework/project/**`
- Do not touch: откат TASK-137 trash/hard-delete guards; dissolve-team API вне scope
- Commits EN/RU; no AI co-author

## [REFERENCES]

- `project/service.go` RemoveTeamMember: после `requireTeamMember` — `TeamMemberCount≤1` → Conflict
- `TeamMemberCount` в `Repository` / Ent
- TASK-137 явно исключал A2W1-A5-003
- Audit: C-W19-tracker-tms.md
- W13 GAP: NOT_PROVEN (guard отсутствовал на HEAD) → fix + mutate PROVEN 2026-07-30

## [ACCEPTANCE]

- Last remaining member remove → 409 Conflict (`cannot remove the last member*`); team остаётся ≥1 member — **DONE**
- ≥2 members: remove non-last OK; последующий last → 409 — **DONE**
- GREEN: `TestRemoveTeamMember_RejectsLastMember_TASK184`; `TestRemoveTeamMember_AllowsWhenMultiple_TASK184` — **DONE**
- Mutation: убрать guard → RED → restore GREEN — **PROVEN** (`MUTATION-…-task184-fix.md`)
- Stop: **Закрыта**

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 4 refill / qa-auto1 · 2026-07-30 · **1/1 PROVEN**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe | Verdict |
|----|------------------------------|------|----------------|-------------|--------|
| C-W19-TT-001 | `backend/internal/project/service.go` — `RemoveTeamMember`: после `requireTeamMember` — `TeamMemberCount≤1` → `apperrors.NewConflict("cannot remove the last member of a team")` (409); иначе `repo.RemoveTeamMember` | service | `TestRemoveTeamMember_RejectsLastMember_TASK184`; `TestRemoveTeamMember_AllowsWhenMultiple_TASK184` · `backend/testframework/project/remove_team_member_last_task184_test.go` | убрать `TeamMemberCount≤1` / Conflict guard → last-member remove `err==nil` + count=0 (orphan) → 409/count assert RED | **PROVEN** |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | ProjectService + Ent `TeamMemberCount` / AddTeamMember fixtures (`newProjectRepoEnv`) | **OK** — Ent (не mock-only count) |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Implement-guard GAP: **CLOSED** 2026-07-30 (qa-auto1 refill).
- Dissolve-team API вне scope (TASK); не подменять last-member guard dissolve-path.

## [OUTPUT SCHEMA]

Implement + GREEN TASK184 + mutate RED→GREEN + status **Закрыта** — см. `MUTATION-2026-07-30-W13-qa1-task184-fix.md`.

## [OPERATING RULES]

- Не закрывать без mutate PROVEN на живом guard.
- Не коммитить AI co-author.
