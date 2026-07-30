# MUTATION W13 slot 4 refill — TASK-046 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/archive/legacy/TASK-046-be-p2-ent-unique-fk-check.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**2/4 PROVEN** (Ent-runtime) · **2 N/A** (schema `layer_mismatch`, source-assert only — **не** fake-PROVEN) · **0 NOT_PROVEN** · TASK-046 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W8-A2-006 | **PROVEN** | `TestComponent_UniqueNamePerProject_P2BE22` — duplicate Create `err==nil` after `Unique: true→false` |
| W8-A3-008 | **PROVEN** | `TestJournalLine_DebitXorCredit_P2BE22` — debit+credit Create `err==nil` after XOR Checks drop |
| W8-A5-004 | **N/A** | `TestSchema_P2BE22_UniqueFKCheck` RED on Unique string drop — schema Contains only (**layer_mismatch**) |
| W8-A1-007 | **N/A** | `TestSchema_P2BE22_UniqueFKCheck` RED on channel edge drop — schema Contains only (**layer_mismatch**) |

Production restored after each BREAK. Leftover MUTATED: **нет** (`git diff` ent/schema + migrate/schema.go empty; `MUTATED` grep empty).

`layer_mismatch IDs: W8-A5-004, W8-A1-007` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`.

## Baseline (pre-kill)

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"
$env:CGO_ENABLED="1"
go test -tags hubtest -count=1 -timeout=120s -vet=off ./testframework/ent/ `
  -run "TestComponent_UniqueNamePerProject_P2BE22|TestJournalLine_DebitXorCredit_P2BE22|TestSchema_P2BE22_UniqueFKCheck" -v
```

```text
--- PASS: TestComponent_UniqueNamePerProject_P2BE22 (0.02s)
--- PASS: TestJournalLine_DebitXorCredit_P2BE22 (0.01s)
--- PASS: TestSchema_P2BE22_UniqueFKCheck (0.00s)
ok  github.com/valerii/hub/backend/testframework/ent  ~0.87s
```

## Per-ID evidence

### W8-A2-006 — Component unique `(project,name)` · **PROVEN**

Layer: Ent-runtime DDL (`ent/migrate/schema.go` — enttest sqlite Create).

```text
BREAK: component_name_project_components Unique true → false
→ FAIL TestComponent_UniqueNamePerProject_P2BE22
     An error is expected but got nil  (duplicate Create accepted)
RESTORE Unique true
→ PASS TestComponent_UniqueNamePerProject_P2BE22 (~0.01s)
```

### W8-A3-008 — JournalLine XOR debit/credit CHECK · **PROVEN**

Layer: Ent-runtime DDL Checks map.

```text
BREAK: JournalLinesTable.Annotation.Checks — drop journalline_debit_xor_credit
→ FAIL TestJournalLine_DebitXorCredit_P2BE22
     XOR CHECK must reject debit>0 AND credit>0  (err==nil)
RESTORE Checks map
→ PASS TestJournalLine_DebitXorCredit_P2BE22 (~0.01s)
```

### W8-A5-004 — TestCaseVersion unique index · **N/A** (layer_mismatch)

Layer: schema source `Contains` only — **не** Ent-runtime duplicate Create.

```text
BREAK: testcaseversion.go drop .Unique() on index.Edges("test_case").Fields("version_number")
→ FAIL TestSchema_P2BE22_UniqueFKCheck
     does not contain `index.Edges("test_case").Fields("version_number").Unique()`
RESTORE .Unique()
→ PASS
```

**Не PROVEN:** kill доказывает только text-assert; runtime unique Create для `(test_case, version_number)` в TASK-046 **нет** (criteria GAP_NOTED).

### W8-A1-007 — Bookmark→Channel FK edge · **N/A** (layer_mismatch)

Layer: schema source `Contains` only — **не** runtime FK insert.

```text
BREAK: bookmark.go remove edge.From("channel", Channel.Type).Field("channel_id")
→ FAIL TestSchema_P2BE22_UniqueFKCheck
     does not contain `edge.From("channel", Channel.Type)`
RESTORE channel edge
→ PASS
```

**Не PROVEN:** schema grep ≠ runtime FK enforcement.

## Final GREEN (all restored)

```text
--- PASS: TestComponent_UniqueNamePerProject_P2BE22 (0.01s)
--- PASS: TestJournalLine_DebitXorCredit_P2BE22 (0.01s)
--- PASS: TestSchema_P2BE22_UniqueFKCheck (0.00s)
ok  ~0.56s
```

## Gaps (criteria, not NOT_PROVEN)

- W8-A5-004 / W8-A1-007: schema-only — future Ent-runtime tests would upgrade N/A→PROVEN.
- FIXED без dedicated RED в TASK-046: W8-A2-007 (environment), W8-A4-005 (role) — вне kill table этой волны.
- U02/U03 VersionMixin/OnDelete — после ADR.

## Handoff

- Статус TASK-046: **Закрыта** (2 PROVEN + 2 N/A; 0 NOT_PROVEN)
- Evidence: этот файл
- Leftover MUTATED: нет
