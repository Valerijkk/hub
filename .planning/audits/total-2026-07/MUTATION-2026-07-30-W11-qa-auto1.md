# MUTATION W11 slot 3 — TASK-107 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost · Go 1.25 · Postgres env URL (не поднимали)  
> Criteria: `.planning/backlog/archive/2026-07-28/TASK-107-2026-07-28-ent-indexes-composite-fk-polymorphic.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).  
> Batch: `_mut-batch-w11-task107.json` · runner: `_run_mutations_w11_task107.py`

## Verdict

**8/8 PROVEN** · `red_class=assert` на каждом ID · статус TASK-107 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A2W8-1-004 | **PROVEN** | `TestSchema_TASK107_Markers` | drop `index.Edges("project").Fields("deleted_at")` (fields remain) | schema source |
| A2W8-1-005 | **PROVEN** | `TestSchema_TASK107_Markers` | split `edge.To("default_provider",…)` across lines | schema source |
| A2W8-2-005 | **PROVEN** | `TestSchema_TASK107_Markers` | drop email + owner+deleted_at indexes | schema source |
| A2W8-2-007 | **PROVEN** | `TestSchema_TASK107_Markers` | inject `Optional()` comment in webhook_secret block | schema source |
| A2W8-2-008 | **PROVEN** | `TestSchema_TASK107_Markers` | drop `.Unique()` on GH+GL link indexes | schema source |
| A2W8-2-009 | **PROVEN** | `TestSchema_TASK107_Markers` | drop `index.Edges("table").Fields("name").Unique()` | schema source |
| A2W8-3-002 | **PROVEN** | `TestSchema_TASK107_Markers` | split `field.Int("version")` / `field.Time("revoked_at")` tokens | schema source |
| A2W8-3-006 | **PROVEN** | `TestSchema_TASK107_Markers` | split `field.UUID("entity_id"…)` / `field.Int("version")` tokens | schema source |

Production restored (`git checkout --` после каждого ID). Leftover MUTATED: **нет**.  
`final_markers=true` · `final_runtime=true` · wall **511.5s**.

## Technique note (reformulated)

Naive field/validator removal → `MUTATION_FAILED_TO_BUILD` (unused import) или `ent/runtime.go` panic (desync без `go generate`).  
Reformulation: **source-assert Markers only** — index drops / token splits / NotContains poison comment; no `go generate`, tree stays buildable.

Previous aborted run incorrectly labeled build/panic as PROVEN — **superseded** by this assert-class prove.

## Per-ID evidence

### A2W8-1-004 — board project+deleted_at index

```text
BASELINE → PASS Markers
BREAK: board.go Indexes — remove index.Edges("project").Fields("deleted_at")
→ FAIL assert.Contains … "index.Edges(\"project\").Fields(\"deleted_at\")"  (red_class=assert, 5.9s)
RESTORE → PASS (7.0s)
```

### A2W8-1-005 — AIUserPreference default_provider edge

```text
BREAK: edge.To(\n\t"default_provider", …) — contiguous marker gone
→ FAIL does not contain edge.To("default_provider", AIProvider.Type)  (assert, 67.8s)
RESTORE → PASS
```

### A2W8-2-005 — CRM email / owner+deleted_at indexes

```text
BREAK: remove index.Fields("email"); remove index.Edges("owner").Fields("deleted_at")
→ FAIL Contains email / owner+deleted_at  (assert, 59.4s)
RESTORE → PASS
```

### A2W8-2-007 — webhook_secret NotEmpty / not Optional

```text
BREAK: NotEmpty(). // MUTATED: was Optional() — poison NotContains
→ FAIL should not contain "Optional()"  (assert, 7.6s)
RESTORE → PASS
```

### A2W8-2-008 — GH/GL link unique

```text
BREAK: drop .Unique() on both githublink.go + gitlablink.go
→ FAIL does not contain …Fields("external_id", "link_type").Unique()  (assert, 59.2s)
RESTORE → PASS
```

### A2W8-2-009 — HubColumn unique (table,name)

```text
BREAK: remove index.Edges("table").Fields("name").Unique()
→ FAIL Contains unique (table,name)  (assert, 70.2s)
RESTORE → PASS
```

### A2W8-3-002 — Invite version / revoked_at

```text
BREAK: field.Int(\n\t"version"); field.Time(\n\t"revoked_at")
→ FAIL does not contain field.Time("revoked_at")  (assert, 69.5s)
RESTORE → PASS
```

### A2W8-3-006 — Notification entity_id UUID + version

```text
BREAK: field.UUID(\n\t"entity_id"…); field.Int(\n\t"version")
→ FAIL does not contain field.Int("version")  (assert, 68.7s)
RESTORE → PASS
```

## Commands

```bash
# Baseline + per-ID Break→RED→restore→GREEN (script)
$env:POSTGRES_TEST_URL = "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
python .planning/audits/total-2026-07/_run_mutations_w11_task107.py
# → DONE 8/8 PROVEN · wall=511.5s · exit 0

# Manual equivalent
go test -tags hubtest -race -count=1 ./testframework/ent/ -run '^TestSchema_TASK107_Markers$'
go test -tags hubtest -race -count=1 ./testframework/ent/ -run '^TestSchema_TASK107_RuntimeConstraints$'
```

## Gaps (from criteria, unchanged)

- **A2W8-2-005:** Markers only — Runtime не ловит снятие index на sqlite.
- Markers = source-guard; не доказывает migrate SQL drift без assert на `20260728_task107_*.sql` (он есть в Markers — не мутировали).
