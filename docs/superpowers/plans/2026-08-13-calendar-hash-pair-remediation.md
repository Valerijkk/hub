# Calendar Canonical Hash Pair Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `superpowers:subagent-driven-development` to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce at the Ent schema layer that a canonical calendar payload hash version
and its 32-byte hash are either both null or both present.

**Architecture:** Add one named portable SQL CHECK annotation to `CalendarEvent`, then one
to `CalendarEventException`, as two sequential atomic packets. Each packet has one external
schema-metadata unit oracle, changes no field/edge/index/migration/generated file, and can be
reviewed or reverted independently.

**Tech Stack:** Go 1.25, Ent v0.14.1 schema annotations, standard testing package.

## Global Constraints

- Frozen source decision:
  `.planning/phases/143-calendar-rsvp-core/REMEDIATION-143-03-HASH-PAIR.md`.
- Worker runtime for each code packet: exact `gpt-5.6-terra` / `xhigh`, actual telemetry
  required, fallback none. Reviewer: exact `gpt-5.6-terra` / `high` with the same rule.
- Current execution state: `BLOCKED_MODEL_PROVENANCE`; no product write begins until actual
  model/effort is visible.
- Packet A implementation ownership: `backend/ent/schema/calendarevent.go`; test ownership:
  `backend/tests/calendar/calendar_event_schema_test.go`.
- Packet C implementation ownership:
  `backend/ent/schema/calendareventexception.go`; test ownership:
  `backend/tests/calendar/calendar_event_exception_schema_test.go`.
- Do not change fields, validators, edges, indexes, enums, comments unrelated to the new
  `Annotations` method, generated Ent output, migration SQL or dependencies.
- Do not read or change protected `testframework/**`, `backend/testframework/**`,
  `.planning/audits/test-rebase/**`, phases 128–141/TASK-193–234, `pro-qa-ai/**`,
  `.env.local`, or `.git/index.lock`.
- Host tests/build/lint/validators/generation are forbidden. Direct Docker/Compose is not a
  fallback. Until the accepted launcher exists, every executable status remains `NOT_RUN`.
- Exported Go comments are Russian. Commits contain no AI attribution.

---

## File Structure

- `backend/ent/schema/calendarevent.go` — canonical master schema and its named pair CHECK.
- `backend/tests/calendar/calendar_event_schema_test.go` — external metadata oracle for the
  master CHECK name and predicate.
- `backend/ent/schema/calendareventexception.go` — occurrence override schema and its named
  pair CHECK.
- `backend/tests/calendar/calendar_event_exception_schema_test.go` — external metadata
  oracle for the exception CHECK name and predicate.

---

### Task 1: CalendarEvent payload hash pair

**Files:**

- Create: `backend/tests/calendar/calendar_event_schema_test.go`
- Modify: `backend/ent/schema/calendarevent.go`

**Interfaces:**

- Consumes: `entsql.Checks(map[string]string)` and the existing nullable
  `payload_hash_version` / `payload_hash` fields.
- Produces: named check `calendar_events_payload_hash_pair` with exact predicate
  `(payload_hash_version IS NULL) = (payload_hash IS NULL)`.

- [ ] **Step 1: Prove the packet entry boundary before reading/writing**

The worker response must be able to record:

```text
branch=codex/mail-module-20260813
worktree=clean
requested_model=gpt-5.6-terra/xhigh
actual_model=gpt-5.6-terra/xhigh
allowed_product_file=backend/ent/schema/calendarevent.go
allowed_test_file=backend/tests/calendar/calendar_event_schema_test.go
runtime=NOT_RUN
```

Unavailable actual telemetry, a dirty overlap or an unexpected branch returns `BLOCKED`
before repository reads or writes.

- [ ] **Step 2: Write the external failing annotation oracle**

Create exactly:

```go
package calendar_test

import (
	"testing"

	"entgo.io/ent/dialect/entsql"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestCalendarEvent_PayloadHashColumnsArePaired(t *testing.T) {
	const (
		checkName = "calendar_events_payload_hash_pair"
		predicate = "(payload_hash_version IS NULL) = (payload_hash IS NULL)"
	)

	for _, annotation := range (entschema.CalendarEvent{}).Annotations() {
		sqlAnnotation, ok := annotation.(*entsql.Annotation)
		if !ok {
			continue
		}
		if got, found := sqlAnnotation.Checks[checkName]; found {
			if got != predicate {
				t.Fatalf("%s = %q, want %q", checkName, got, predicate)
			}
			return
		}
	}

	t.Fatalf("missing named CHECK %s", checkName)
}
```

The test inspects Ent metadata only; it opens no DB and performs no I/O.

- [ ] **Step 3: Record the expected RED without host execution**

Before the schema change, `CalendarEvent` has no `Annotations` method, so the external
oracle is expected not to compile. Record:

```text
TDD_RED=STATIC_EXPECTED
RUNTIME_RED=NOT_RUN
reason=task-owned Docker launcher NOT_READY; host/direct-Docker fallback forbidden
```

- [ ] **Step 4: Add only the two annotation imports**

Add to the existing import group in `calendarevent.go`:

```go
"entgo.io/ent/dialect/entsql"
"entgo.io/ent/schema"
```

Do not reorder or modify unrelated imports except formatting required by Go syntax.

- [ ] **Step 5: Add the named portable check**

After `Mixin()` and before `Edges()`, add exactly:

```go
// Annotations задаёт ограничения целостности канонического календарного события.
func (CalendarEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_events_payload_hash_pair": "(payload_hash_version IS NULL) = (payload_hash IS NULL)",
		}),
	}
}
```

No timed/all-day or end-order SQL predicate belongs in this remediation.

- [ ] **Step 6: Perform static GREEN inspection**

```powershell
rg -n "entgo.io/ent/dialect/entsql|entgo.io/ent/schema|calendar_events_payload_hash_pair|payload_hash_version IS NULL" backend/ent/schema/calendarevent.go backend/tests/calendar/calendar_event_schema_test.go
git diff --check -- backend/ent/schema/calendarevent.go backend/tests/calendar/calendar_event_schema_test.go
git diff --stat -- backend/ent/schema/calendarevent.go backend/tests/calendar/calendar_event_schema_test.go
```

Expected: one production predicate, one independently literal test predicate, zero whitespace
errors, and exactly two packet paths.

- [ ] **Step 7: Return the uncommitted Terra handoff**

Report exact starting SHA, requested/actual model+effort, two changed paths, source-content
hashes, static command output, and `generation/compile/unit/runtime=NOT_RUN`. Do not commit.

- [ ] **Step 8: Root reviews and commits Packet A**

Root rejects any field/edge/index/generated/migration change, validates the test is external
and literal rather than derived from production, stages only the schema/test plus
`IMPLEMENTATION-143-03A.md`, `REMEDIATION-143-03-HASH-PAIR.md`, and `.planning/STATE.md`,
runs staged `diff --check`, then commits:

```text
fix(calendar): pair event payload hash fields
```

The evidence status is `STATIC_CONFORMANT`, never runtime PASS.

---

### Task 2: CalendarEventException payload hash pair

**Files:**

- Create: `backend/tests/calendar/calendar_event_exception_schema_test.go`
- Modify: `backend/ent/schema/calendareventexception.go`

**Interfaces:**

- Consumes: accepted Task 1 commit and the existing exception hash fields.
- Produces: named check `calendar_event_exceptions_payload_hash_pair` with the same exact
  portable null-pair predicate.

- [ ] **Step 1: Prove the sequential packet boundary**

Require clean branch, Task 1 as an ancestor, exact Terra/xhigh actual telemetry, and only the
exception schema/test paths. Any mismatch returns `BLOCKED` before reads or writes.

- [ ] **Step 2: Write the external failing exception oracle**

Create exactly:

```go
package calendar_test

import (
	"testing"

	"entgo.io/ent/dialect/entsql"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestCalendarEventException_PayloadHashColumnsArePaired(t *testing.T) {
	const (
		checkName = "calendar_event_exceptions_payload_hash_pair"
		predicate = "(payload_hash_version IS NULL) = (payload_hash IS NULL)"
	)

	for _, annotation := range (entschema.CalendarEventException{}).Annotations() {
		sqlAnnotation, ok := annotation.(*entsql.Annotation)
		if !ok {
			continue
		}
		if got, found := sqlAnnotation.Checks[checkName]; found {
			if got != predicate {
				t.Fatalf("%s = %q, want %q", checkName, got, predicate)
			}
			return
		}
	}

	t.Fatalf("missing named CHECK %s", checkName)
}
```

- [ ] **Step 3: Record expected RED without execution**

The current exception schema has no `Annotations` method. Record `STATIC_EXPECTED` and every
runtime field as `NOT_RUN`; do not invoke Go or Docker tooling.

- [ ] **Step 4: Add the exact annotation imports**

Add only:

```go
"entgo.io/ent/dialect/entsql"
"entgo.io/ent/schema"
```

- [ ] **Step 5: Add the exception check**

After `Fields()` and before `Edges()`, add exactly:

```go
// Annotations задаёт ограничения целостности переопределения экземпляра события.
func (CalendarEventException) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_event_exceptions_payload_hash_pair": "(payload_hash_version IS NULL) = (payload_hash IS NULL)",
		}),
	}
}
```

- [ ] **Step 6: Perform static GREEN inspection**

```powershell
rg -n "entgo.io/ent/dialect/entsql|entgo.io/ent/schema|calendar_event_exceptions_payload_hash_pair|payload_hash_version IS NULL" backend/ent/schema/calendareventexception.go backend/tests/calendar/calendar_event_exception_schema_test.go
git diff --check -- backend/ent/schema/calendareventexception.go backend/tests/calendar/calendar_event_exception_schema_test.go
git diff --stat -- backend/ent/schema/calendareventexception.go backend/tests/calendar/calendar_event_exception_schema_test.go
```

Expected: exactly the exception schema/test paths and no whitespace error.

- [ ] **Step 7: Return uncommitted handoff and independent review**

Return exact provenance/SHA/diff evidence with all executable statuses `NOT_RUN`. Root then
dispatches a fresh exact Terra/high read-only reviewer for Task 2 and rechecks Task 1 remains
unchanged. Unavailable telemetry is `BLOCKED`, not PASS.

- [ ] **Step 8: Root docs sync and atomic commit**

After static content acceptance, stage only the two Task 2 paths plus
`IMPLEMENTATION-143-03C.md`, `REMEDIATION-143-03-HASH-PAIR.md`, and `.planning/STATE.md`,
inspect staged names/check, then commit:

```text
fix(calendar): pair exception payload hash fields
```

Keep Ent generation, compile, migration, unit and runtime `NOT_RUN`.

---

## Plan Self-Review

- Spec coverage: both known half-null persistence gaps have one exact named CHECK and one
  independent literal metadata oracle.
- Atomicity: event and exception remain sequential packets with disjoint product/test files
  and separate commits/reviews.
- Scope: no temporal union, end-order, field, edge, index, generator or migration change is
  included.
- Type/API consistency: both tests use Ent v0.14.1's actual `*entsql.Annotation.Checks`
  representation and import the repository schema externally.
- Placeholder scan: no guessed launcher/generator command or incomplete implementation step
  exists.
- Current state: `BLOCKED_BEFORE_WRITE` until actual Terra telemetry is available.
