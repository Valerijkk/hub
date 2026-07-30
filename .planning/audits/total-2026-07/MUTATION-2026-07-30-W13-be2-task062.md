# MUTATION W13 slot 3 refill — TASK-062 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-062-be-p3-dead-code-layering.md` `[MUTATION CRITERIA]`  
> Findings: **14** IDs (P3-BE-07) · **7 PROVEN** · **6 GAP_NOTED** · **1 N/A**  
> Коммит не создавался. leftover MUTATED: **нет**.

## Verdict

**7/14 PROVEN** · **6 GAP_NOTED** · **1 N/A** · **0 NOT_PROVEN** · TASK-062 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| W1-A1-008 | **GAP_NOTED** | `TestHandler_SearchIssuesJQL_Success` (layering bypass stays GREEN) |
| W1-A2-014 | **PROVEN** | `TestDeleteIssue_LogsDeletedActivity_Task062` |
| W1-A2-015 | **PROVEN** | `TestListWorkflowTransitions_SkipsNilEdges_Task062` |
| W2-A2-018 | **GAP_NOTED** | `TestUpdateTestResult_Success` (dead `_ = result` stays GREEN) |
| W2-A4-010 | **GAP_NOTED** | `TestParseTestRailCSV` (unused `csv.NewReader` stays GREEN) |
| W2-A4-011 | **GAP_NOTED** | `TestTrimStepNumbering` (MustCompile-per-call stays GREEN) |
| W2-A5-007 | **PROVEN** | `TestCreateTable_DefaultNameColumnInTx_Task062` |
| W3-A2-018 | **PROVEN** | `TestForwardMessage_SourceChannelMismatch_Task062` |
| W3-A3-009 | **N/A** | docs-only godoc `presenceAwayThreshold` |
| W4-A3-007 | **PROVEN** | `TestMetricsEndpoint_BlocksNonMatchingCIDR` |
| W4-A4-010 | **GAP_NOTED** | `TestApplyPoolLimits_MaxOpenConns_P1_BE_25` (restored `OpenRawDB` no callers) |
| W4-A5-009 | **GAP_NOTED** | `TestHashKey` (unused `stopCh` field stays GREEN) |
| W6-A1-009 | **PROVEN** | compile-guard `TestEntRepository_OrganizationPendingFlow` |
| W7-A4-017 | **PROVEN** | `TestRequireSystemAdmin_Layering_Task062` |

`layer_mismatch IDs: W1-A1-008` · `mock≠Ent IDs: W1-A2-014, W1-A2-015, W3-A2-018` · `docs-only IDs: W3-A3-009` · `GAP IDs: W1-A1-008, W2-A2-018, W2-A4-010, W2-A4-011, W4-A4-010, W4-A5-009`

## Per-ID evidence

### W1-A1-008 — handler JQL layering (GAP)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/tracker/ \
    -run TestHandler_SearchIssuesJQL_Success$
  -> PASS (~6.5s)

BREAK (handler — hardcoded total=1, skip SearchIssuesJQL):
  backend/internal/tracker/handler_issues.go SearchIssuesJQL
  — MUTATED W1-A1-008

  -> PASS (~7.5s)  ← adjacent Success only checks total==1; no layering RED

RESTORE
  -> PASS (~5.8s)
```

**GAP_NOTED (layer_mismatch):** нет Exact RED на handler→Ent bypass.

### W1-A2-014 — DeleteIssue LogActivity «deleted» (mock≠Ent)

```text
BASELINE PASS (~6.5s)
BREAK: убрать LogActivity из DeleteIssue
  -> FAIL (~7.2s) "[]" should have 1 item(s), but has 0
RESTORE PASS (~6.1s)
```

Layer: service + mockRepo (**mock≠Ent** flagged; Exact RED на service-слое).

### W1-A2-015 — nil From/ToStatus skip (mock≠Ent)

```text
BASELINE PASS (~5.7s)
BREAK: убрать nil-guard
  -> FAIL (~7.3s) panic: nil pointer dereference
RESTORE PASS (~5.5s)
```

### W2-A2-018 — dead `_ = result` (GAP)

```text
BASELINE TestUpdateTestResult_Success PASS (~8.0s)
BREAK: restore `result, err := Update…` + `_ = result`
  -> PASS (~10.0s)
RESTORE PASS (~7.0s)
```

**GAP_NOTED:** нет assert отсутствия dead assign.

### W2-A4-010 — unused csv.NewReader (GAP)

```text
BASELINE TestParseTestRailCSV PASS (~6.9s)
BREAK: `_ = csv.NewReader(bytes.NewReader(data))`
  -> PASS (~8.6s)
RESTORE PASS (~6.5s)
```

### W2-A4-011 — package-level regexp (GAP / perf)

```text
BASELINE TestTrimStepNumbering PASS (~6.0s)
BREAK: MustCompile внутри trim на каждый вызов
  -> PASS (~8.8s)
RESTORE PASS (~6.5s)
```

### W2-A5-007 — CreateTable default Name column in tx (Ent)

```text
BASELINE PASS (~4.7s)
BREAK (CreateTable-unique anchor — не CreateBase):
  skip HubColumn.Create Name в CreateTable tx
  — MUTATED W2-A5-007

  -> FAIL (~6.0s) "[]" should have 1 item(s), but has 0
RESTORE PASS (~4.9s)
```

Layer OK: Ent CreateTable tx.

### W3-A2-018 — ForwardMessage channel scope (mock≠Ent)

```text
BASELINE PASS (~6.3s)
BREAK: убрать source.ChannelID != channelID
  -> FAIL (~7.9s) An error is expected but got nil
RESTORE PASS (~6.3s)
```

### W3-A3-009 — presence godoc (N/A)

```text
docs-only: godoc synced to 90s presenceAwayThreshold; runtime unchanged
→ N/A
```

### W4-A3-007 — MetricsEndpoint authorizeMetricsIP

```text
BASELINE PASS (~3.6s)
BREAK: skip authorizeMetricsIP
  -> FAIL (~3.6s) expected: 403
RESTORE PASS (~3.1s)
```

### W4-A4-010 — OpenRawDB deleted (GAP)

```text
BASELINE TestApplyPoolLimits_MaxOpenConns_P1_BE_25 PASS (~6.8s)
BREAK: restore OpenRawDB() без callers
  -> PASS (~6.7s)
RESTORE PASS (~4.7s)
```

**GAP_NOTED:** нет RED на dead API.

### W4-A5-009 — unused stopCh (GAP)

```text
BASELINE TestHashKey PASS (~7.6s)
BREAK: restore unused stopCh field
  -> PASS (~8.9s)
RESTORE PASS (~6.9s)
```

### W6-A1-009 — CreateOrganizationWithFounder signature (compile-guard)

```text
BASELINE compile ./testframework/auth/ PASS (~7.0s)
BREAK: inviteToken param restored on interface+impl
  -> FAIL (~1.1s) not enough arguments in call to CreateOrganizationWithFounder
RESTORE compile PASS (~6.7s)
```

### W7-A4-017 — RequireSystemAdmin layering (Ent)

```text
BASELINE PASS (~5.3s)
BREAK: non-admin → return nil
  -> FAIL (~6.6s) An error is expected but got nil
RESTORE PASS (~5.2s)
```

## Cleanup

- `git checkout --` на всех touched prod-файлах после каждой пары BREAK/RESTORE
- `rg MUTATED` по touched → пусто
- `git diff` touched TASK-062 prod → empty
- Коммит не создавался
- Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task062.py`
- Re-prove W2-A5-007: `.planning/audits/total-2026-07/_reprove_w2a5007_task062.py`
- Raw JSON: `.planning/audits/total-2026-07/_mut-batch-w13-task062.json`
