# MUTATION W13 slot 4 refill — TASK-068 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/archive/legacy/TASK-068-be-p3-ent-hygiene-mojibake.md` `[MUTATION CRITERIA]`  
> Classes: **P3-BE-17**, **P3-BE-18** · Коммит: **не создавался** (инструкция слота).

## Verdict

**11/11 PROVEN** · **0 NOT_PROVEN** · **0 N/A** · TASK-068 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W8-A1-009 | **PROVEN** | `TestSchema_P3BE17_HygieneMarkers` — NotContains duplicate Unique |
| W8-A2-019 | **PROVEN** | `TestComment_MentionsUUID_P3BE17` + HygieneMarkers |
| W8-A3-017 | **PROVEN** | `TestInvite_CodeMaxLen_P3BE17` — 65-char Create `err==nil` |
| W8-A3-018 | **PROVEN** | `TestNewsPost_BodyMaxLen_P3BE17` + HygieneMarkers MaxLen(50000) |
| W8-A1-010 | **PROVEN** | `TestSchema_P3BE18_NoMojibake` — hits apikey/backuprecord/board |
| W8-A2-020 | **PROVEN** | `TestSchema_P3BE18_NoMojibake` — githubintegration/githublink/gitlablink |
| W8-A3-016 | **PROVEN** | `TestSchema_P3BE18_NoMojibake` — issue.go:ибра |
| W8-A4-011 | **PROVEN** | `TestSchema_P3BE18_NoMojibake` — project/projecttemplate |
| W8-A5-003 | **PROVEN** | `TestSchema_P3BE18_NoMojibake` — testcase.go:ибра |
| W8-A5-007 | **PROVEN** | `TestSchema_P3BE18_NoMojibake` — testrun allure_run_id Comment |
| W8-A5-011 | **PROVEN** | `TestSchema_P3BE18_NoMojibake` — 6 edges godoc files |

Production restored после каждого BREAK (`git checkout --` schema; generated `runtime.go` / `comment_create.go` — reverse-patch, gitignored). Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`.

## Baseline (pre-kill)

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"
$env:CGO_ENABLED="1"
go test -tags hubtest -count=1 -vet=off -timeout=180s ./testframework/ent/ `
  -run "TestSchema_P3BE18_NoMojibake|TestSchema_P3BE17_HygieneMarkers|TestInvite_CodeMaxLen_P3BE17|TestNewsPost_BodyMaxLen_P3BE17|TestComment_MentionsUUID_P3BE17" -v
```

```text
--- PASS: TestSchema_P3BE18_NoMojibake (0.01s)
--- PASS: TestSchema_P3BE17_HygieneMarkers (0.00s)
--- PASS: TestInvite_CodeMaxLen_P3BE17 (0.01s)
--- PASS: TestNewsPost_BodyMaxLen_P3BE17 (0.01s)
--- PASS: TestComment_MentionsUUID_P3BE17 (0.01s)
ok  github.com/valerii/hub/backend/testframework/ent  ~0.69s
```

## Per-ID evidence

### W8-A1-009 — duplicate Unique index hygiene · **PROVEN**

```text
BREAK: apikey/call/channelinvite Indexes — restore
  index.Fields("key_hash"|"room_name"|"invite_code"|"token").Unique()
→ FAIL TestSchema_P3BE17_HygieneMarkers (NotContains RED ×4)
RESTORE git checkout schema → PASS (~0.55s)
```

### W8-A2-019 — Comment mentions []uuid.UUID · **PROVEN**

```text
BREAK: schema comment.go []uuid.UUID{} → []string{}
      + comment_create.go SetMentions no-op (generated Mentions type
        stays UUID until ent generate; Exact RED = MentionsUUID)
→ FAIL HygieneMarkers (Contains UUID marker)
→ FAIL MentionsUUID (Len expected 1, got 0)
RESTORE schema + SetMentions → PASS (~0.57s)
```

### W8-A3-017 — Invite code MaxLen(64) · **PROVEN**

```text
BREAK: schema invite.go drop MaxLen(64)
      + runtime.go CodeValidator → validators[0] only (NotEmpty)
        (runtime reads schema.Invite{}.Fields(); hardcoding [0]+[1]
         panics after MaxLen drop — wire Must change with schema)
→ FAIL TestInvite_CodeMaxLen_P3BE17 (An error is expected but got nil)
RESTORE schema + runtime → PASS (~0.55s)
```

Note: schema `// MaxLen(64) removed` comment still matched HygieneMarkers Contains — Exact RED is Invite Create runtime.

### W8-A3-018 — NewsPost body MaxLen(50000) · **PROVEN**

```text
BREAK: schema newspost.go drop MaxLen(50000)
      + runtime.go BodyValidator → validators[0] only
→ FAIL HygieneMarkers (does not contain MaxLen(50000))
→ FAIL TestNewsPost_BodyMaxLen_P3BE17 (err==nil on 50001)
RESTORE → PASS (~0.57s)
```

### W8-A1-010 … W8-A5-011 — mojibake «ибра» · **PROVEN** (each)

Shared Exact RED: `TestSchema_P3BE18_NoMojibake`.

| ID | BREAK files | RED hits |
|----|-------------|----------|
| W8-A1-010 | apikey / backuprecord / board edges godoc | 3 files:ибра |
| W8-A2-020 | githubintegration / githublink / gitlablink | 3 files |
| W8-A3-016 | issue edges godoc | issue.go:ибра |
| W8-A4-011 | project / projecttemplate | 2 files |
| W8-A5-003 | testcase edges godoc | testcase.go:ибра |
| W8-A5-007 | testrun `allure_run_id` Comment | testrun.go:ибра |
| W8-A5-011 | testcase / testcaseversion / testrun / userblock / webhookdelivery / workflowstatus | 6 files |

Each: RESTORE `git checkout --` → PASS NoMojibake.

## Final GREEN (all restored)

```text
--- PASS: TestSchema_P3BE18_NoMojibake (0.01s)
--- PASS: TestSchema_P3BE17_HygieneMarkers (0.00s)
--- PASS: TestInvite_CodeMaxLen_P3BE17 (0.01s)
--- PASS: TestNewsPost_BodyMaxLen_P3BE17 (0.01s)
--- PASS: TestComment_MentionsUUID_P3BE17 (0.01s)
ok  ~0.76s
```

## Gaps

- Generated `backend/ent/{runtime,comment_create}.go` gitignored — MaxLen / SetMentions kills require reverse-patch restore (not `git checkout`).
- W8-A2-019: schema type change alone does not regenerate Mentions type; MentionsUUID killed via SetMentions no-op + schema marker.

## Handoff

- Статус TASK-068: **Закрыта** (11/11 PROVEN; 0 NOT_PROVEN)
- Evidence: этот файл
- Leftover MUTATED: нет
- Коммит: не создавался
