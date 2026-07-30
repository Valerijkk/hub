# MUTATION W13 slot 3 refill — TASK-164 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-164-2026-07-29-a3-7-3-integrations.md` `[MUTATION CRITERIA]`  
> Findings: **A3W7-3-001**, **A3W7-3-002**, **A3W7-3-003** (3 IDs)  
> Коммит не создавался.

## Verdict

**3/3 PROVEN** · TASK-164 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W7-3-001 | **PROVEN** | `TestListSprintIssues_ExcludesTrash_A3W73_001` |
| A3W7-3-002 | **PROVEN** | `TestFindIssueByKey_ExcludesTrash_A3W73_002` (github + gitlab) |
| A3W7-3-003 | **PROVEN** | `TestGetLinkPreview_IssueMembership_A3W73_003` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`

## Per-ID evidence

### A3W7-3-001 — `ListSprintIssues` + `DeletedAtIsNil`

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/imex/ \
    -run TestListSprintIssues_ExcludesTrash_A3W73_001
  -> PASS (~0.95s)

BREAK:
  imex/repository.go ListSprintIssues — убран Where(issue.DeletedAtIsNil())
  -> FAIL TestListSprintIssues_ExcludesTrash_A3W73_001
       should have 1 item(s), but has 2 (trash ST-2 в списке)

RESTORE DeletedAtIsNil
  -> PASS (~0.95s)

git diff imex/repository.go → empty
MUTATED marker → отсутствует
```

### A3W7-3-002 — `FindIssueByKey*` + `DeletedAtIsNil` (GH + GL)

```text
BREAK:
  github/repository.go — FindIssueByKey + FindIssueByKeyInProject без DeletedAtIsNil
  gitlab/repository.go — то же
  -> FAIL TestFindIssueByKey_ExcludesTrash_A3W73_002 (github)
       trashed issue must be NotFound via FindIssueByKey, got <nil>
  -> FAIL TestFindIssueByKey_ExcludesTrash_A3W73_002 (gitlab)
       trashed issue must be NotFound via FindIssueByKey, got <nil>

RESTORE DeletedAtIsNil на всех 4 методах
  -> PASS github (~0.9s) · gitlab (~1.0s)

git diff github/repository.go gitlab/repository.go → empty
MUTATED marker → отсутствует
```

### A3W7-3-003 — `resolveIssuePreview` + `IsProjectMember`

```text
BREAK:
  github/preview.go resolveIssuePreview — skip IsProjectMember
  -> FAIL TestGetLinkPreview_IssueMembership_A3W73_003
       Type=issue (want external); Title/Status/Assignee утекают (SECRET-FOREIGN-*)

RESTORE IsProjectMember → denial nil
  -> PASS (~10.8s suite filter)

git diff github/preview.go → empty
MUTATED marker → отсутствует
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s ./testframework/imex/ \
  -run TestListSprintIssues_ExcludesTrash_A3W73_001
go test -tags hubtest -count=1 -timeout=180s ./testframework/github/ \
  -run "TestFindIssueByKey_ExcludesTrash_A3W73_002|TestGetLinkPreview_IssueMembership_A3W73_003"
go test -tags hubtest -count=1 -timeout=180s ./testframework/gitlab/ \
  -run TestFindIssueByKey_ExcludesTrash_A3W73_002
# BASELINE PASS · BREAK FAIL (per ID) · RESTORE PASS
```

## Gaps

- none — criteria IDs PROVEN  
- Docs-only / N/A: none  
- 003: mock membership intentional (criteria layer notes OK)

## Handoff

- Evidence: этот файл · TASK-164 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored; no leftover MUTATED; no commit
