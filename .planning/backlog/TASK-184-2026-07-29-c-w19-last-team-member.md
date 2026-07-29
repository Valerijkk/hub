# 🐛 C-W19: RemoveTeamMember last-member orphan

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P3** · Статус: **На проверке**  
> Источник: [`C-W19-tracker-tms.md`](../audits/total-2026-07/C-W19-tracker-tms.md) · A2W1-A5-003  
> Исполнитель: **be1** · Finding ID: `C-W19-TT-001`

## [OUTCOME]

`RemoveTeamMember` не оставляет команду без участников (last-member → 400/Conflict); orphan team без `requireTeamMember` невозможен.

## [ROLE]

- Role: be1 · Handoff: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/internal/project/service.go` RemoveTeamMember; repo TeamMemberCount если нужно; `testframework/project/**`
- Do not touch: откат TASK-137 trash/hard-delete guards; dissolve-team API вне scope
- Commits EN/RU; no AI co-author

## [REFERENCES]

- `project/service.go` ~624–632: любой member удаляет любого; нет guard `count≤1`
- TASK-137 явно исключал A2W1-A5-003
- Audit: C-W19-tracker-tms.md

## [ACCEPTANCE]

- Last remaining member remove → 400/409; team остаётся ≥1 member (или явный dissolve path)
- Mutation: убрать guard → orphan/empty-team тест RED
- Stop: FIXED · «На проверке»
