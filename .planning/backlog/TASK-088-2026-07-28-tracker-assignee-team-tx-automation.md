# TASK-088 — 🔒 Tracker: assignee membership + AddTeam tx + automation Nak

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Исправлена  
> Источник: A2W1 · Stage A2 dedup · TASK-008

## [OUTCOME]

Assignee вне проекта отклоняется; AddTeam+members атомарны; automation consumer Nak на transient/unknown subject.

## [ROLE]

- Role: be1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/tracker/service_issues.go
  - backend/internal/project/repository.go
  - backend/internal/tracker/automation_*.go
  - backend/testframework/tracker/** project/**
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - frontend board DnD
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W1-A2-002, A2W1-A5-002, A2W1-A3-001/002
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - Create/Update issue: assignee ∈ project members
  - AddTeamIDs+AddMember одна Ent tx
  - Engine error → Nak; unknown subject → Nak/metric
  - Мутации по каждому пункту
- Regression: затронутый `./testframework/...` или frontend unit
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Diff summary
- Test commands + PASS
- Mutation note
- Статус задачи обновлён

## [OPERATING RULES]

- Не закрывать соседние A2 ID «заодно» без evidence
- Docs sync только если контракт API изменился


## [EVIDENCE] (be1 2026-07-28)

- \IsProjectMember\ + check в Create/UpdateIssue
- \AddTeamToProject\ в одной Ent-tx + mid-fail inject rollback
- Engine \Handle*\ → \error\; consumer Nak на engine err / unknown subject + metric
- Tests: \ssignee_membership_task088_test.go\, \dd_team_tx_task088_test.go\, \utomation_nak_task088_test.go- Regression: \go test -tags hubtest -race ./testframework/tracker/... ./testframework/project/...\ PASS
