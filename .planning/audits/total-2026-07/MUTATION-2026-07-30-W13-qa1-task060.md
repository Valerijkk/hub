# MUTATION W13 slot 4 refill — TASK-060 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-060-be-p3-utf8-rune-truncate.md` `[MUTATION CRITERIA]`  
> Class: **P3-BE-04** · Коммит: **не создавался** (инструкция слота).

## Verdict

**5/5 PROVEN** · **0 NOT_PROVEN** · TASK-060 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| W5-A4-010 | **PROVEN** | `TestTruncateRunes_Cyrillic_NotMidRune`; `TestTruncateRunes_Emoji_KeepsWholeRunes` | `[]rune` → `s[:N]` bytes (+ `"..."`) | helper / search |
| W3-A2-017 | **PROVEN** | `TestTruncateRunes_PromoteTitle_Cyrillic`; `TestTruncateRunes_PromoteTitle_Emoji` | `[]rune` → `s[:N]` bytes | helper / messenger |
| W7-A3-012 | **PROVEN** | `TestTruncateCell_Cyrillic_NotMidRune`; `TestTruncateCell_Emoji_KeepsWholeRunes` | `[]rune` → `s[:keep]` bytes | helper / imex |
| W7-A5-009 | **PROVEN** | `TestValidateListingInput_CyrillicTitle_RuneLimit` | `utf8.RuneCountInString` → `len()` | service / marketplace |
| W2-A2-019 | **PROVEN** | `TestNormalizeErrorPattern_Cyrillic_NotMidRune` | `[]rune` → `msg[:100]` bytes | converter / tms |

Production restored (`git checkout --` touched guards) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### W5-A4-010 — search `truncateRunes`

```text
BASELINE TruncateRunes_ → PASS (~6.3s)

BREAK: backend/internal/search/truncate.go
  runes := []rune(s); string(runes[:maxRunes])+"..."
  → s[:maxRunes]+"..." // MUTATED W5-A4-010
→ FAIL TestTruncateRunes_Cyrillic_NotMidRune
        ValidString=false; RuneCount expected 100
→ FAIL TestTruncateRunes_Emoji_KeepsWholeRunes
        ValidString=false; expected whole 🔥 prefix
RESTORE → PASS (~6.3s)
```

### W3-A2-017 — messenger `truncateRunes` (promote title)

```text
BASELINE TruncateRunes_PromoteTitle_ → PASS (~9.3s)

BREAK: backend/internal/messenger/truncate.go
  string(runes[:maxRunes]) → s[:maxRunes] // MUTATED W3-A2-017
→ FAIL TestTruncateRunes_PromoteTitle_Cyrillic
        ValidString=false; RuneCount expected 500
→ FAIL TestTruncateRunes_PromoteTitle_Emoji
        ValidString=false; RuneCount expected 500
RESTORE → PASS (~7.0s)
```

### W7-A3-012 — imex `truncateCell`

```text
BASELINE TruncateCell_ → PASS (~7.4s)

BREAK: backend/internal/imex/truncate.go
  string(runes[:keep])+"..." → s[:keep]+"..." // MUTATED W7-A3-012
→ FAIL TestTruncateCell_Cyrillic_NotMidRune
        ValidString=false; expected 47×«я»+"..."
→ FAIL TestTruncateCell_Emoji_KeepsWholeRunes
        ValidString=false; expected whole 📦 cells
RESTORE → PASS (~7.9s)
```

### W7-A5-009 — marketplace `validateListingInput`

```text
BASELINE ValidateListingInput_ → PASS (~5.3s)

BREAK: backend/internal/marketplace/service.go:376-378
  utf8.RuneCountInString(title/description)
  → len(title)/len(description) // MUTATED W7-A5-009
→ FAIL TestValidateListingInput_CyrillicTitle_RuneLimit
        150×«я» (300 bytes) falsely rejected as >200
(TestValidateListingInput_CyrillicTitle_OverRuneLimit /
 TestValidateListingInput_EmojiDescription_RuneLimit stay GREEN —
 over-limit still rejected by byte len; anti-regression companions)
RESTORE → PASS (~6.2s)
```

### W2-A2-019 — tms `normalizeErrorPattern`

```text
BASELINE NormalizeErrorPattern_ → PASS (~10.5s)

BREAK: backend/internal/tms/service_converters.go:396-403
  string(runes[:100]) → msg[:100] // MUTATED W2-A2-019
→ FAIL TestNormalizeErrorPattern_Cyrillic_NotMidRune
        RuneCount expected 100 got 50; string ≠ 100×«я»
(msg[:100] на чётной границе «я» остаётся ValidString —
 kill = RuneCount/equality, как в criteria)
RESTORE → PASS (~7.3s)
```

```text
FINAL all 5 patterns → PASS (~5.6–8.2s)
leftover_mutated=False
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=120s ./testframework/search/ -run TruncateRunes_
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/messenger/ -run TruncateRunes_PromoteTitle_
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/imex/ -run TruncateCell_
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/marketplace/ -run ValidateListingInput_
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/tms/ -run NormalizeErrorPattern_
```

Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task060.py`  
JSON: `_mut-batch-w13-task060.json`.

## Gaps

- Short-string unchanged tests — anti-regression, не primary kill — **GAP_NOTED**, не NOT_PROVEN.
- Marketplace OverRuneLimit / EmojiDescription — companions; primary kill = CyrillicTitle_RuneLimit (150 рун / 300 байт).
- TMS `msg[:100]` на чётной байтовой границе «я» может остаться ValidString; kill = RuneCount — **GAP_NOTED**.
- Посторонний leftover MUTATED в `messenger/unread.go` (W3-A3-010) блокировал baseline compile — восстановлен `git checkout` до прогона; не часть TASK-060.

## Handoff

- Evidence: этот файл · TASK-060 **Закрыта** · README sync  
- be-ver → code-reviewer (fix/tests already FIXED; mutation PROVEN)  
- Коммит: не создавался
