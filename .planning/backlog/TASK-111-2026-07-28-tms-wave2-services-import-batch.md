# TASK-111 — 🐛 TMS W2 residual: tx/N+1/import/bases

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: На проверке  
> Источник: A2W2 residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008
> Закрыто: 2026-07-28 · be2 · ветка `fix/tms/task-111-w2-residual`

## Evidence (be2 · 2026-07-28)

| ID | Status | Evidence |
|----|--------|----------|
| A2W2-1-002 | FIXED | swagger: zip убран из WebhookAllure description |
| A2W2-1-003 | FIXED | StructValidator test PUT без version → 400; wave124 шлёт version |
| A2W2-1-005 | FIXED | raw body cap 5MB → 413; TestTask111_WebhookRawBodyCap |
| A2W2-2-002 | FIXED | UpdateTestCaseWithSnapshot Ent tx |
| A2W2-2-003 | FIXED | Rollback errors.As AppError (409 не → 500) |
| A2W2-2-004 | FIXED | CreateAutomaticImportRun Ent tx + abort cleanup |
| A2W2-2-005 | FIXED | maxEpochMs в ImportAllureToExistingRun |
| A2W2-2-007 | FIXED | publish EventTestRunCompleted |
| A2W2-2-008 | FIXED | UpdateTestResult(projectID, runID, resultID) |
| A2W2-2-009 | FIXED | mock+Ent Limit=1 pagination |
| A2W2-2-010 | FIXED | buildSnapshot возвращает err |
| A2W2-2-011 | FIXED | фиксированное «некорректный allure JSON» |
| A2W2-3-001 | FIXED | Mongo InsertMany/Find SetLimit cap |
| A2W2-3-002 | FIXED | List*/Batch Limit caps |
| A2W2-3-003 | FIXED | DeleteTestCase(projectID, id) |
| A2W2-4-001 | FIXED | ParseAllure/ParseJUnit depth+len |
| A2W2-4-002 | FIXED | PDF truncate() rune-safe |
| A2W2-4-003 | FIXED | matcher automation_id/case_key !IsNotFound |
| A2W2-4-004 | FIXED | ImportResult.Partial |
| A2W2-4-005 | FIXED | export paginate ListTestCases |
| A2W2-4-006 | FIXED | matcher mock projectID denial |
| A2W2-4-007 | FIXED | TestRail handler fixed 400 msg |
| A2W2-5-001 | FIXED | SoftDeleteBase cascade tx |
| A2W2-5-002 | FIXED | unique sqlite DSN в bases tests |

Mutation: убрать VersionEQ / depth cap / body cap / cascade → краснеют TestTask111_* / importer_limits / bases cascade.

## [OUTCOME]

Все open W2 вне TASK-091 закрыты (tx, N+1, parser caps, import limits, bases).

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/tms/**
  - backend/internal/bases/**
  - testframework/tms bases
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - ломать TASK-015/016 closures
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W2-1-002/003/005, A2W2-2-002..005/007..011, A2W2-3-*, A2W2-4-*, A2W2-5-* (24)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - чеклист 24 ID в отчёте
  - hubtest green
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
