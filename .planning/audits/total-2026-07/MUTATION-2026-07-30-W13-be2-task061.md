# MUTATION W13 slot 3 refill — TASK-061 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-061-be-p3-repo-lists-defense.md` `[MUTATION CRITERIA]`  
> Findings: **11** runnable IDs (P3-BE-05, P3-BE-06) · **15** sub-kill runs  
> Коммит не создавался. leftover MUTATED: **нет**.

## Verdict

**11/11 PROVEN** · **0 NOT_PROVEN** · **0 GAP_NOTED** · TASK-061 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| W1-A3-009 | **PROVEN** | `TestEntRepository_UpdateAutomationRule_RequiresProjectID_P3BE05` |
| W1-A4-013 | **PROVEN** | `TestEntRepository_ClearIssuesSprint_RequiresProject_P3BE05` |
| W2-A3-009 | **PROVEN** | `…_DeleteTestCase_IdempotentNotFound_P3BE05`; `…_MoveSuite_RequiresProject_P3BE05`; `…_AddFilesToResult_RequiresRun_P3BE05` |
| W2-A3-010 | **PROVEN** | `TestEntRepository_GetProjectIssuesWithTestCases_SkipsSoftDeleted_P3BE05` |
| W3-A3-008 | **PROVEN** | `TestEntRepository_GetChannelMemberIDs_SkipsSoftDeleted_P3BE05` |
| W3-A5-014 | **PROVEN** | `TestRepository_GetChannelMemberRole_SkipsSoftDeletedChannel_P3BE05` |
| W7-A4-014 | **PROVEN** | `TestEntRepository_SoftDelete_IdempotentNotFound_P3BE05` |
| W7-A5-011 | **PROVEN** | `TestEntRepository_SoftDeleteListing_RequiresVersion_P3BE05` |
| W1-A3-010 | **PROVEN** | `…_ListActivities_Caps_P3BE06`; `…_ListComments_Caps_P3BE06`; `…_ListTimeLogs_Caps_P3BE06` |
| W2-A5-012 | **PROVEN** | `TestEntRepository_ListBases_Caps_P3BE06` |
| W7-A2-013 | **PROVEN** | `TestEntRepository_ListContactsAndDeals_Caps_P3BE06` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)` · `GAP IDs: (none)`

## Per-ID evidence

### W1-A3-009 — `UpdateAutomationRule` ProjectIDEQ

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/tracker/ \
    -run TestEntRepository_UpdateAutomationRule_RequiresProjectID_P3BE05$
  -> PASS (~5.7s)

BREAK (EntRepository correct layer):
  backend/internal/tracker/repo_automation.go UpdateAutomationRule
  — убрать ProjectIDEQ
  — MUTATED W1-A3-009

  -> FAIL (~7.1s)
       expected: "Scoped" (foreign project hijack)

RESTORE ProjectIDEQ
  -> PASS (~6.6s)

git diff repo_automation.go → empty
```

### W1-A4-013 — `ClearIssuesSprint` HasProjectWith

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/sprint/ \
    -run TestEntRepository_ClearIssuesSprint_RequiresProject_P3BE05$
  -> PASS (~5.2s)

BREAK:
  backend/internal/sprint/repository.go ClearIssuesSprint
  — убрать HasProjectWith из Count + Update
  — MUTATED W1-A4-013

  -> FAIL (~6.2s)
       Expected error "issue not in project" but got nil

RESTORE HasProjectWith
  -> PASS (~5.0s)
```

### W2-A3-009 — DeleteTestCase / MoveSuite / AddFilesToResult (3 guards)

```text
DeleteTestCase — снять DeletedAtIsNil
  BASELINE PASS (~5.9s) → BREAK FAIL (~8.3s) An error is expected but got nil
  → RESTORE PASS (~6.3s)

MoveSuite — снять HasProjectWith
  BASELINE PASS (~6.3s) → BREAK FAIL (~7.7s) An error is expected but got nil
  → RESTORE PASS (~6.5s)

AddFilesToResult — снять HasRunWith
  BASELINE PASS (~5.7s) → BREAK FAIL (~7.7s) An error is expected but got nil
  → RESTORE PASS (~6.6s)
```

### W2-A3-010 — `GetProjectIssuesWithTestCases` DeletedAtIsNil

```text
BASELINE PASS (~9.1s)
BREAK: снять entissue.DeletedAtIsNil
  -> FAIL (~9.2s) Should be empty, but was [Issue(... deleted_at=...)]
RESTORE PASS (~7.3s)
```

### W3-A3-008 — `GetChannelMemberIDs` HasChannelWith(DeletedAtIsNil)

```text
BASELINE PASS (~6.3s)
BREAK: снять HasChannelWith(DeletedAtIsNil)
  -> FAIL (~7.5s) Should be empty, but was [uuid uuid]
RESTORE PASS (~8.0s)
```

### W3-A5-014 — `GetChannelMemberRole` HasChannelWith(DeletedAtIsNil)

```text
BASELINE PASS (~6.6s)
BREAK: снять HasChannelWith(DeletedAtIsNil)
  -> FAIL (~8.1s) expected: "" (role still returned)
RESTORE PASS (~6.1s)
```

### W7-A4-014 — storage SoftDelete DeletedAtIsNil

```text
BASELINE PASS (~5.0s)
BREAK: снять DeletedAtIsNil
  -> FAIL (~6.7s) An error is expected but got nil (2nd SoftDelete)
RESTORE PASS (~5.0s)
```

### W7-A5-011 — SoftDeleteListing VersionEQ+DeletedAtIsNil

```text
BASELINE PASS (~4.2s)
BREAK: UpdateOneID без Where VersionEQ/DeletedAtIsNil
  -> FAIL (~4.8s) An error is expected but got nil (stale version)
RESTORE PASS (~3.8s)
```

### W1-A3-010 — ListActivities/Comments/TimeLogs Limit caps (3)

```text
ListActivities — снять Limit(MaxIssueActivities)
  BASELINE PASS (~7.9s) → BREAK FAIL (~8.2s) expected: 200
  → RESTORE PASS (~6.8s)

ListComments — снять Limit(MaxIssueComments)
  BASELINE PASS (~6.0s) → BREAK FAIL (~8.0s) expected: 200
  → RESTORE PASS (~6.4s)

ListTimeLogs — снять Limit(MaxIssueTimeLogs)
  BASELINE PASS (~5.6s) → BREAK FAIL (~7.3s) expected: 200
  → RESTORE PASS (~5.5s)
```

### W2-A5-012 — ListBases Limit(MaxOwnerBases)

```text
BASELINE PASS (~4.1s)
BREAK: снять Limit(MaxOwnerBases)
  -> FAIL (~5.1s) expected: 100
RESTORE PASS (~4.9s)
```

### W7-A2-013 — ListContacts + ListDeals Limit(MaxCRMList)

```text
BASELINE PASS (~5.5s)
BREAK: снять Limit(MaxCRMList) на обоих списках
  -> FAIL (~5.3s) expected: 100 (contacts + deals)
RESTORE PASS (~4.4s)
```

## Cleanup

- `git checkout --` на всех touched prod-файлах после каждой пары BREAK/RESTORE
- `rg MUTATED` по touched → пусто
- `git diff` touched → empty
- Коммит не создавался
- Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task061.py`
- Raw JSON: `.planning/audits/total-2026-07/_mut-batch-w13-task061.json`
