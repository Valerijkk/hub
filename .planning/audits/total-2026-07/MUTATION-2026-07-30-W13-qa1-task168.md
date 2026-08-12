# MUTATION W13 slot 4 refill — TASK-168 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-168-2026-07-29-a3-8-2-schemas.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**6/6 PROVEN** · **0 NOT_PROVEN** · TASK-168 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W8-2-001 | **PROVEN** | `TestSchema_TASK168_FileMetadataValidators`; `TestValidator_TASK168_FileMetadataNegativeSize`; `TestValidator_TASK168_FileMetadataEntityID` | `NonNegative`→no-op Validate; `validateFileEntityID`→no-op; uploader `Required` drop | schema (+ Ent via runtime) |
| A3W8-2-002 | **PROVEN** | `TestSchema_TASK168_OnDeleteCascades`; `TestCascade_TASK168_CoffeeRoundHardDeleteRemovesPairs` | pairs Cascade→SetNull + migrate FK NoAction | schema + migrate DDL |
| A3W8-2-003 | **PROVEN** | `TestSchema_TASK168_OnDeleteCascades`; `TestCascade_TASK168_HubBaseHardDeleteRemovesTables` | tables Cascade→SetNull + migrate FK NoAction | schema + migrate DDL |
| A3W8-2-004 | **PROVEN** | `TestSchema_TASK168_OnDeleteCascades`; `TestCascade_TASK168_GitHub/GitLabIntegrationDeleteRemovesLinks` | links Cascade→SetNull (GH+GL) + migrate FK NoAction | schema + migrate DDL |
| A3W8-2-005 | **PROVEN** | `TestSchema_TASK168_OnDeleteCascades`; `TestCascade_TASK168_CommentHardDeleteRemovesReactions` | reactions Cascade→SetNull + migrate FK NoAction | schema + migrate DDL |
| A3W8-2-006 | **PROVEN** | `TestValidator_TASK168_EmptyAccessToken` | access_token `NotEmpty`→no-op Validate (GH+GL) | schema (+ Ent via runtime) |

Production restored (`git checkout --` schema; migrate backup rewrite) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### A3W8-2-001 — FileMetadata NonNegative / Validate / Required

```text
BASELINE TASK168 → PASS (~5.0s)

BREAK: filemetadata.go
  size: NonNegative → Validate(func(int64) error { return nil })
  entity_id: Validate(validateFileEntityID) → no-op Validate
  uploader: drop Required()
→ FAIL TestSchema_TASK168_FileMetadataValidators
        missing NonNegative() / Validate(validateFileEntityID) / Required()
→ FAIL TestValidator_TASK168_FileMetadataNegativeSize
        An error is expected but got nil
→ FAIL TestValidator_TASK168_FileMetadataEntityID
        An error is expected but got nil
(TestValidator_TASK168_FileMetadataUploaderRequired stays GREEN —
 generated filemetadata_create.go still enforces required edge;
 schema marker covers Required)

RESTORE → PASS (~5.0s / ~3.4s)
```

### A3W8-2-002 — CoffeeRound→pairs Cascade

```text
BREAK: coffeeround.go Cascade→SetNull
       migrate schema.go coffee_pairs_coffee_rounds_pairs Cascade→NoAction
→ FAIL TestSchema_TASK168_OnDeleteCascades
        coffeeround.go missing "entsql.OnDelete(entsql.Cascade)"
→ FAIL TestCascade_TASK168_CoffeeRoundHardDeleteRemovesPairs
        ent: constraint failed: FOREIGN KEY constraint failed
RESTORE → PASS (~4.1s)
```

### A3W8-2-003 — HubBase→tables Cascade

```text
BREAK: hubbase.go Cascade→SetNull
       migrate hub_tables_hub_bases_tables Cascade→NoAction
→ FAIL TestSchema_TASK168_OnDeleteCascades (hubbase.go missing Cascade)
→ FAIL TestCascade_TASK168_HubBaseHardDeleteRemovesTables
        FOREIGN KEY constraint failed
RESTORE → PASS (~3.9s)
```

### A3W8-2-004 — GitHub/GitLab Integration→links Cascade

```text
BREAK: githubintegration.go + gitlabintegration.go Cascade→SetNull
       migrate git_hub_links_… / git_lab_links_… Cascade→NoAction
→ FAIL TestSchema_TASK168_OnDeleteCascades (both files missing Cascade)
→ FAIL TestCascade_TASK168_GitHubIntegrationDeleteRemovesLinks (FK fail)
→ FAIL TestCascade_TASK168_GitLabIntegrationDeleteRemovesLinks (FK fail)
RESTORE → PASS (~4.8s)
```

### A3W8-2-005 — Comment→reactions Cascade

```text
BREAK: comment.go Cascade→SetNull
       migrate reactions_comments_reactions Cascade→NoAction
→ FAIL TestSchema_TASK168_OnDeleteCascades (comment.go missing Cascade)
→ FAIL TestCascade_TASK168_CommentHardDeleteRemovesReactions
        FOREIGN KEY constraint failed
RESTORE → PASS (~4.4s)
```

### A3W8-2-006 — access_token NotEmpty

```text
BREAK: githubintegration.go + gitlabintegration.go
  access_token NotEmpty → Validate(func(string) error { return nil })
→ FAIL TestValidator_TASK168_EmptyAccessToken
        An error is expected but got nil (empty PAT accepted)
(TestSchema_TASK168_AccessTokenNotEmpty may stay GREEN — file still has
 NotEmpty() on webhook_secret; Ent Create path is the kill)

RESTORE → PASS (~4.5s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s ./testframework/ent/ -run TASK168
```

Runners:  
`.planning/audits/total-2026-07/_run_mutations_w13_task168.py`  
`.planning/audits/total-2026-07/_run_mutations_w13_task168_cascade.py` (Cascade→SetNull + migrate)  
JSON: `_mut-batch-w13-task168.json`, `_mut-batch-w13-task168-cascade.json`.

## Gaps

- Docs-only: **N/A** — все 6 ID schema/runtime.
- `TestValidator_TASK168_FileMetadataUploaderRequired`: Ent Create Required edge зашит в generated `filemetadata_create.go`; schema-source drop of `Required()` kills schema marker only (noted, not NOT_PROVEN).
- `TestSchema_TASK168_AccessTokenNotEmpty`: weak (NotEmpty anywhere in file); kill via Ent Create validator.
- Migrate SQL apply вне test — needs-owner-apply (unchanged).
- HubTable→columns/records — out of scope (TASK-169).

## Handoff

- Evidence: этот файл · TASK-168 **Закрыта** · README sync  
- be-ver → code-reviewer (fix/tests already FIXED; mutation PROVEN)  
- Коммит: не создавался
