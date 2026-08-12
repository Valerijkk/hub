# MUTATION W13 slot 3 refill — TASK-167 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-167-2026-07-29-a3-8-1-schemas.md` `[MUTATION CRITERIA]`  
> Findings: **A3W8-1-001**, **A3W8-1-002** (2 IDs)  
> Коммит не создавался.

## Verdict

**1/1 code PROVEN · 1/1 docs N/A** · TASK-167 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W8-1-001 | **PROVEN** | `TestSchema_TASK167_AIUserPreferenceDefaultProviderSetNull` |
| A3W8-1-002 | **N/A** (docs-only path B LWW ADR) | — (not fake-PROVEN; marker `TestSchema_TASK167_ChannelMemberLWW_NoVersionMixin` only cement exclusion) |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: A3W8-1-002`

## Per-ID evidence

### A3W8-1-001 — `default_provider` SetNull annotation

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/ent/task167/ \
    -run "TestSchema_TASK167_AIUserPreferenceDefaultProviderSetNull|TestSchema_TASK167_ChannelMemberLWW_NoVersionMixin"
  -> PASS (~0.2s)

BREAK:
  backend/ent/schema/aiuserpreference.go
  — remove Annotations(entsql.OnDelete(entsql.SetNull)) from edge.To("default_provider", …)
  — drop unused "entgo.io/ent/dialect/entsql" import
  — MUTATED comment must NOT contain string "entsql.OnDelete(entsql.SetNull)"
    (first attempt left that string in a comment → false GREEN; fixed kill)

  -> FAIL TestSchema_TASK167_AIUserPreferenceDefaultProviderSetNull
       does not contain "entsql.OnDelete(entsql.SetNull)"
       does not contain "\"entgo.io/ent/dialect/entsql\""

RESTORE annotation + entsql import (original Edges block)
  -> PASS (~0.3s)

git diff aiuserpreference.go → empty
MUTATED marker → отсутствует
```

### A3W8-1-002 — ChannelMember LWW (path B) · **N/A code mutation**

```text
Guard layer: ADR + PO Approval record (not VersionEQ / not EntRepository)

Evidence (docs-only, not PROVEN):
  - .planning/decisions/ADR-ENT-CHANNELMEMBER-LWW-2026.md
      title «LWW accepted» · A3W8-1-002 · «Owner lock / approve LWW path» · «Accepted 2026-07-29»
  - TASK-167 Approval record: ✅ Owner lock — LWW accepted 2026-07-29
  - schema channelmember.go: no VersionMixin{} (path A not chosen)

Kill recipe per criteria: N/A code mutate — evidence = ADR § + Owner lock.
Schema marker TestSchema_TASK167_ChannelMemberLWW_NoVersionMixin PASS on restore
  (cements exclusion; does NOT prove optimistic lock).

No fake-PROVEN.
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/ent/task167/ \
  -run "TestSchema_TASK167_AIUserPreferenceDefaultProviderSetNull|TestSchema_TASK167_ChannelMemberLWW_NoVersionMixin"
# BASELINE PASS · BREAK FAIL (001) · RESTORE PASS
```

## Gaps

- **A3W8-1-002** = docs-only N/A (path B LWW). Mute/star remain LWW until separate VersionMixin+codegen wave.
- Schema marker for 001 is static source scan (not runtime Ent cascade) — matches criteria Flag OK.

## Handoff

- Evidence: этот файл · TASK-167 **Закрыта** · README sync  
- be-ver → code-reviewer (002 path B: ADR+PO already locked)  
- Production restored; no leftover MUTATED; no commit
