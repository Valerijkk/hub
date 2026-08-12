# MUTATION W13 slot 4 refill — TASK-166 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-166-2026-07-29-a3-7-5-mp-status-dev.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**4/4 PROVEN** · **0 NOT_PROVEN** · TASK-166 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W7-5-001 | **PROVEN** | `TestListApproved_ReviewNoteHidden_A3W75001` | `resp.ReviewNote = l.ReviewNote` без `allowEmail` | DTO map |
| A3W7-5-002 | **PROVEN** | `TestSeedWorkflowTransitions_CreateFailPropagates_A3W75002` (+ links/channels suite) | slog+continue в `seedWorkflowTransitions` | seed helpers |
| A3W7-5-003 | **PROVEN** | `TestCurrencyNormalize_A3W75003` | убрать `normalizeCurrency` до Save Create/Update | service |
| A3W7-5-004 | **PROVEN** | `TestReviewNoteRuneCap_A3W75004` | убрать rune cap в `transitionReview` | service |

Production restored (`git checkout --`) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### A3W7-5-001 — public catalog hides review_note

```text
BASELINE → PASS (~4.6s)

BREAK: dto.go toListingResponseOpts — всегда копировать ReviewNote
→ FAIL: TestListApproved_ReviewNoteHidden_A3W75001
        Should be empty, but was internal: PII of seller phone +1-555
RESTORE → PASS (~4.3s)
```

### A3W7-5-002 — seed Create-fail propagates

```text
BASELINE → PASS (~5.3s)

BREAK: seed.go seedWorkflowTransitions — slog.Error + continue вместо return err
→ FAIL: TestSeedWorkflowTransitions_CreateFailPropagates_A3W75002
        An error is expected but got nil (Create FK fail must propagate)
(links/channels helpers untouched — suite still runs; kill target = one helper per criteria)
RESTORE → PASS (~5.9s)
```

### A3W7-5-003 — currency normalize before Save

```text
BASELINE → PASS (~4.6s)

BREAK: service.go CreateListing/UpdateListing — сырой req.Currency без normalize
→ FAIL: TestCurrencyNormalize_A3W75003
        create listing: ent: validator failed … currency … greater than the required length
        ("USD " → Ent MaxLen fail вместо stored USD)
RESTORE → PASS (~4.6s)
```

### A3W7-5-004 — ReviewNote rune cap

```text
BASELINE → PASS (~4.6s)

BREAK: service.go transitionReview — убрать maxReviewNoteRunes check
→ FAIL: TestReviewNoteRuneCap_A3W75004
        Expected error … invalid listing input … but got nil
RESTORE → PASS (~4.9s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=300s ./testframework/marketplace/ `
  -run "TestListApproved_ReviewNoteHidden_A3W75001|TestCurrencyNormalize_A3W75003|TestReviewNoteRuneCap_A3W75004"

go test -tags hubtest -count=1 -timeout=300s ./testframework/dev/ `
  -run "TestSeedWorkflowTransitions_CreateFailPropagates_A3W75002|TestSeedIssueLinks_CreateFailPropagates_A3W75002|TestSeedChannels_CreateFailPropagates_A3W75002"
```

Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task166.py` · JSON: `_mut-batch-w13-task166.json`.

## Gaps

- Docs-only / schema MaxLen `review_note`: N/A (out of scope; service cap достаточен).
- `layer_mismatch` / `mock≠Ent`: none.

## Handoff

- Evidence: этот файл · TASK-166 **Закрыта** · README sync  
- be-ver → code-reviewer (fix/tests already FIXED; mutation PROVEN)  
- Коммит: не создавался
