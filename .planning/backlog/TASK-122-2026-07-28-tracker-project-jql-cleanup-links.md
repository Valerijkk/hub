# TASK-122 — 🔒 Tracker/project: JQL via repo + DeleteProject links

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Источник: A2W1 residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008
> Mutation: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W12-be1.md` · 2/2 PROVEN (A5 Ent twin)

## [OUTCOME]

JQL не через GetDB; DeleteProject link cleanup fail-closed/tx.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/tracker/**
  - backend/internal/project/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - дублировать TASK-088
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W1-A2-001, A2W1-A5-001 (2)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 2 ID + tests
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

## [EVIDENCE] (be1 2026-07-29)

### Checklist ID

| ID | Статус | Evidence |
|----|--------|----------|
| A2W1-A2-001 | FIXED | `repo.SearchIssuesJQL`; `GetDB` убран из `Repository`; тест `TestEntRepository_SearchIssuesJQL_ProjectScoped` |
| A2W1-A5-001 | FIXED | `SoftDeleteWithCrossLinkCleanup` (Ent tx); fail-closed unit: `TestDeleteProject_CleanupListError_DoesNotSoftDelete`, `TestDeleteProject_CleanupDeleteError_DoesNotSoftDelete` |

### Mutation

- A2W1-A2-001: вернуть `s.repo.GetDB().Issue.Query()` в service → compile fail (нет `GetDB`) или обход layering краснеет на ProjectScoped.
- A2W1-A5-001: вернуть log+continue в cleanup → `Cleanup*Error_DoesNotSoftDelete` краснеют (проект soft-delete'ится при ошибке).

### Tests (PASS)

```
go build -tags release ./internal/tracker/... ./internal/project/...
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/tracker/   # ok ~7s
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/project/   # ok ~7s
```

## [MUTATION CRITERIA] — verifier-ready

> W7 slot 3 / qa-auto1 · criteria only (no mutate) · 2026-07-30

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A2W1-A2-001 | `tracker/service_issues.go:400` — обойти `repo.SearchIssuesJQL` (вернуть GetDB/прямой Ent); **или** `repo_issues.go:489-495` — убрать `issue.ProjectID`/`DeletedAtIsNil` scope | `TestEntRepository_SearchIssuesJQL_ProjectScoped` | repository (+ service layering) |
| A2W1-A5-001 | `project/service.go:281-288` — ignore err от `SoftDeleteWithCrossLinkCleanup` / soft-delete при cleanup fail; **или** Ent `repository.go:636+` — soft-delete вне tx / continue после DeleteOne fail | `TestDeleteProject_CleanupListError_DoesNotSoftDelete`; `TestDeleteProject_CleanupDeleteError_DoesNotSoftDelete` | service (mock) |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| A2-001 | EntRepository SearchIssuesJQL | Project scope — **Ent OK**; GetDB removal = compile |
| A5-001 | mock `crossLinkRepo` SoftDeleteWithCrossLinkCleanup **+** Ent twin mid-fail | Service fail-closed — OK на mock; **Ent tx** — `TestSoftDeleteWithCrossLinkCleanup_TxRollbackOnMidFail_A2W1_A5_001` |

`layer_mismatch IDs: (none)` — A2W1-A5-001 закрыт Ent twin (W12 be1)

### Gaps

- **A2W1-A5-001:** ~~нет Ent IT~~ → закрыто W12: `TestSoftDeleteWithCrossLinkCleanup_TxRollbackOnMidFail_A2W1_A5_001` + `ExportSetAfterCrossLinkDeleteHook`; mutation without Ent tx → RED (link not rolled back).
- GetDB compile-gate — N/A runtime mutation (ожидаемо).

## [MUTATION W12 be1] — 2026-07-30

| ID | Verdict | Evidence |
|----|---------|----------|
| A2W1-A2-001 | **PROVEN** | drop `issueProjectIDEQ` → `TestEntRepository_SearchIssuesJQL_ProjectScoped` RED (len=2) → restore GREEN |
| A2W1-A5-001 | **PROVEN** (service + Ent twin) | service: soft-delete on cleanup fail → Cleanup* RED; Ent: no-tx SoftDelete → twin RED (link not found) |

Отчёт: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W12-be1.md`
