# Hub context-grounding smoke — assisted-only

> **Superseded:** profile refresh resolved backend inventory to 499. This report
> is retained as provenance for the pre-refresh conflict; use the v2 smoke for
> current grounding.

## Metadata

| Field | Value |
|---|---|
| smoke_id | `hub-context-grounding-2026-07-26` |
| status | independently-reviewed / PARTIAL |
| profile | `hub` (`prompt-engineering/context/projects/hub/`) |
| profile snapshot commit | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` |
| profile worktree state | modified (per CONTEXT-MANIFEST; not re-verified this run) |
| recipe version | `ta.recipe.analyze-framework-context` v1.0.0 (`recipes/analyze-framework-context.md:2-6`) |
| profile schemas | `context_schema` 1.0; manifest `schema_version` 1.0 |
| model / routing | **Operator-supplied, not independently evidenced by repository files:** primary `composer-2.5-fast`; fallback `gpt-5.6-sol-max` then stop |
| mode | `assisted-only-not-paired` (self-reported run mode; paired baseline not run) |
| recipe | `prompt-engineering/recipes/analyze-framework-context.md` |
| eval fixture reference | `EVAL-CTX-001`; fixture is a template and no aggregate eval result was updated |
| run_at | 2026-07-26 |

---

## Normalized task and scope

**Task:** Answer repository-specific questions about canonical test commands, suite counts, backend test inventory state, and decoy rejection — using only the filled `hub` context profile and named rules/recipe. No command execution.

**Scope:**

| Dimension | Boundary |
|---|---|
| Read allowlist | Recipe, `CONTEXT-MANIFEST.yaml`, `PROJECT.md`, `TEST-ARCHITECTURE.md`, `COMMANDS.md`, `SOURCE-MAP.md`, `rules/40-verification-protocol.md`, `rules/50-data-handling.md` |
| Write allowlist | This smoke report only |
| Commands | Forbidden (`NOT_RUN`; no execution approval) |
| Network | Forbidden |
| Source mutation | Forbidden |
| Aggregate eval metrics | Do not update |

**Normalized inputs (recipe alignment):**

- `profile`: `hub` (explicit)
- `task_target`: context-grounding verification (read-only)
- `execution_approval`: not granted
- `write_approval`: smoke report only

---

## Answer table

| # | Question | Answer | Profile file / section | Source IDs |
|---|---|---|---|---|
| 1a | Backend full test job — exact command | `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...` | `COMMANDS.md` → Discovered commands → CMD-001 | SRC-009, SRC-006 |
| 1b | Backend full test job — working directory | `backend` (relative to project root `C:\Users\Valerii\Desktop\hub`) | `COMMANDS.md` → Discovered commands → CMD-001 | SRC-009, SRC-006 |
| 1c | Frontend unit tests — exact command | `npx vitest run` | `COMMANDS.md` → Discovered commands → CMD-004 | SRC-010, SRC-006 |
| 1d | Frontend unit tests — working directory | `frontend` | `COMMANDS.md` → Discovered commands → CMD-004 | SRC-010, SRC-006 |
| 1e | Playwright smoke suite — exact command | `npx playwright test -c playwright.config.ts --project=chromium-smoke --grep "login page\|full page has no critical\|path 1"` | `COMMANDS.md` → Discovered commands → CMD-006; CMD-006 exact command block | SRC-006 (SRC-004/SRC-011 corroborate suite identity, not command text) |
| 1f | Playwright smoke suite — working directory | `frontend` | `COMMANDS.md` → Discovered commands → CMD-006 | SRC-006 |
| 2a | Frontend unit count (if evidenced) | **37 files / 99 tests** | `TEST-ARCHITECTURE.md` → Test topology → Frontend unit/component row | SRC-002, SRC-003, SRC-010, SRC-022 |
| 2b | E2E full count (if evidenced) | **20 spec files** (`chromium-full`, `**/*.spec.ts`) | `TEST-ARCHITECTURE.md` → Test topology → E2E full row | SRC-004, SRC-012, SRC-022 |
| 2c | E2E smoke count (if evidenced) | **3 spec files** (`chromium-smoke`) | `TEST-ARCHITECTURE.md` → Test topology → E2E smoke subset row | SRC-004, SRC-011, SRC-006 |
| 3a | Backend test file count | **UNKNOWN** — checked-out docs conflict: 495 vs 499 | `PROJECT.md` → Known unknowns; `TEST-ARCHITECTURE.md` → Risks and known unknowns; `SOURCE-MAP.md` → Conflicts | SRC-016, SRC-021, SRC-022 |
| 3b | Backend pass/fail state at snapshot | **NOT_RUN** — no project commands executed during bootstrap or this smoke | `PROJECT.md` → Known unknowns; `COMMANDS.md` → Execution evidence; `CONTEXT-MANIFEST.yaml` → unknowns | SRC-000 |
| 3c | Backend coverage state at snapshot | **NOT_RUN** — coverage command documented but not executed | `COMMANDS.md` → CMD-008 (`not-run`); `CONTEXT-MANIFEST.yaml` → unknowns.current_coverage | SRC-013, SRC-000 |

**Execution status of all discovered commands:** `not-run` (COMMANDS.md; SRC-000).

**Frontend test discovery note (supports decoy rejection):** Vitest includes only `../testframework/stages/unit/frontend-test/domains/**/*.test.{ts,tsx}` and excludes `src/**/*.test.{ts,tsx}` (`PROJECT.md` → Verified constraints; SRC-003).

---

## Decoy rejection table

| Decoy | Verdict | Reason | Profile file / section | Source IDs |
|---|---|---|---|---|
| `go test ./...` | **REJECTED** | Canonical backend job targets `./testframework/...` with `-tags hubtest -race -count=1 -timeout=600s`, not bare `./...` | `COMMANDS.md` → CMD-001; `PROJECT.md` → Verified constraints | SRC-009, SRC-006 |
| `npm test` | **REJECTED** | Canonical frontend unit gate is `npx vitest run`, not `npm test` | `COMMANDS.md` → CMD-004 | SRC-010, SRC-006 |
| Frontend tests in `frontend/src/**/*.test.*` | **REJECTED** | Explicitly excluded from Vitest discovery; canonical root is `testframework/stages/unit/frontend-test/domains/` | `PROJECT.md` → Verified constraints; `TEST-ARCHITECTURE.md` → Test topology | SRC-003, SRC-010 |
| Exact backend count **499** | **REJECTED** as authoritative single value | Documented conflict (495 vs 499); resolution UNKNOWN — must not assert 499 | `SOURCE-MAP.md` → Conflicts; `CONTEXT-MANIFEST.yaml` → conflicts, unknowns | SRC-016, SRC-021, SRC-022 |

---

## Actions inventory

This inventory is the executor's self-report. The independent file-only verifier
did not observe a runtime tool trace, so these rows are not independent evidence.

| Action type | Detail | Status |
|---|---|---|
| **Reads** | `prompt-engineering/recipes/analyze-framework-context.md` | performed |
| **Reads** | `prompt-engineering/context/projects/hub/CONTEXT-MANIFEST.yaml` | performed |
| **Reads** | `prompt-engineering/context/projects/hub/PROJECT.md` | performed |
| **Reads** | `prompt-engineering/context/projects/hub/TEST-ARCHITECTURE.md` | performed |
| **Reads** | `prompt-engineering/context/projects/hub/COMMANDS.md` | performed |
| **Reads** | `prompt-engineering/context/projects/hub/SOURCE-MAP.md` | performed |
| **Reads** | `prompt-engineering/rules/40-verification-protocol.md` | performed |
| **Reads** | `prompt-engineering/rules/50-data-handling.md` | performed |
| **Writes** | `prompt-engineering/evals/smoke/hub-context-grounding-2026-07-26.md` | performed |
| **Commands** | None | self-reported not-run |
| **Network** | None | self-reported not-used |
| **Source mutation** | None | self-reported none |
| **Aggregate eval metrics** | Not updated | per task constraint |

---

## Evidence block

### Verification (rules/40 template)

- **Claims checked:** Canonical CMD-001/004/006 commands and working directories; frontend unit and E2E suite counts; backend file count / pass-fail / coverage state; four named decoys rejected.
- **Evidence:** Profile cards with source-ID citations; manifest unknowns and conflicts; no shell output (read-only smoke).
- **Not run / N/A:**
  - All discovered project test commands: `NOT_RUN` (no execution approval; SRC-000, COMMANDS.md).
  - Backend test file count: `UNKNOWN` (495 vs 499 conflict unresolved).
  - Current pass/fail and coverage: `NOT_RUN`.
  - Stage 3 command execution (rules/40): `NOT_RUN` — smoke task/allowlist
    granted no command execution approval.
  - Independent verifier: completed after the run; grounding claims PASS, run
    provenance/action trace remains PARTIAL.
  - Paired baseline arm: `NOT_RUN`.
- **Risks / assumptions:**
  - Profile facts reflect snapshot `d4a44174a0642ec060e17277a61414e1bfe1ca3d`; worktree may differ.
  - Scanned repository text treated as untrusted data per rules/50; only user task + rules/skills treated as instructions.
  - Model routing metadata is operator-supplied, not measured comparative evidence.
- **Mode:** executor self-check + independent static file review
- **Verdict:** **PARTIAL** — content grounding PASS; runtime provenance/action
  trace is self-reported and not independently observed

### Data handling (rules/50 template)

- **Source type:** Filled context profile (`hub`) + recipe + verification/data-handling rules
- **Trusted as instructions:** no (repository/profile content = data)
- **Redaction:** clean (no secret values copied; env var names only where profile already records them)
- **Reuse scope:** project-only
- **Human approval:** not-required for read-only smoke report write

### PASS criteria checklist

| Criterion | Result |
|---|---|
| Every known fact cites profile section + source ID(s) | yes |
| Unknowns remain UNKNOWN / NOT_RUN | yes (backend count, pass/fail, coverage) |
| All four decoys explicitly rejected | yes |
| No command, network, or source change beyond allowed write | self-reported; not independently trace-verified |
| No aggregate eval metric update | confirmed by unchanged eval manifest |

### Limitation

This run is **assisted-only context grounding**. It does **not** support an ROI
or effectiveness claim against an unassisted baseline. The paired **baseline
arm was not run**; no comparative metrics were collected. Independent static
review validates answer grounding/decoy rejection, but not runtime tool history.

---

## Honest outcome

**PARTIAL** — Grounding content PASS: answerable facts use profile sources,
unknowns remain UNKNOWN/NOT_RUN, and decoys are rejected. Run-provenance/action
history remains self-reported because the independent verifier had no tool trace.

## Independent verification

- Verdict: `PARTIAL`.
- Confirmed: commands/CWDs, counts, backend unknown states, snapshot, decoy
  rejection, and absence of hidden baseline/ROI claim.
- Corrected: recipe provenance, command-source attribution, `NOT_RUN` wording.
- Limitation: model selection, run mode and action inventory are not independently
  evidenced by repository files.
