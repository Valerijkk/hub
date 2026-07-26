# Hub context-grounding smoke — assisted-only (v2, refreshed profile)

## Status

**PARTIAL** — by design for assisted-only smoke: content grounding may self-check PASS; runtime provenance remains self-reported and is not independent benchmark evidence.

| Dimension | Result |
|---|---|
| `content_grounding` | **PASS** (self-check) |
| `runtime_provenance` | **PARTIAL** (self-reported action trace; no independent tool trace) |
| `paired_baseline` | **NOT_RUN** |
| `effectiveness` | **NOT_RUN** |
| `ROI` | **NOT_RUN** |

---

## Metadata

| Field | Value |
|---|---|
| smoke_id | `hub-context-grounding-2026-07-26-v2` |
| supersedes | `hub-context-grounding-2026-07-26` (pre-refresh conflict state) |
| status | **PARTIAL** (assisted-only; see dimensions above) |
| profile | `hub` (`prompt-engineering/context/projects/hub/`) |
| profile snapshot commit | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` |
| profile refreshed_at | `2026-07-26T11:40:48.3583334Z` (CONTEXT-MANIFEST) |
| profile worktree state | modified (per CONTEXT-MANIFEST; not re-verified this run) |
| recipe id / version | `ta.recipe.analyze-framework-context` v1.0.0 |
| recipe version provenance | `recipes/analyze-framework-context.md:2-6` |
| profile schemas | `context_schema` 1.0; manifest `schema_version` 1.0 |
| model / routing | **Self-reported operator metadata — not independent benchmark evidence:** primary `composer-2.5-fast`; fallback `gpt-5.6-sol-max` then stop |
| mode | `assisted-only-not-paired` |
| eval fixture reference | `EVAL-TBD-ANALYZE-FRAMEWORK-CONTEXT` (recipe placeholder); aggregate eval metrics **not updated** |
| run_at | 2026-07-26 |

---

## Normalized task and scope

**Task:** Answer repository-specific questions about canonical test commands, suite counts, backend test inventory state, pass/fail and coverage state, and decoy rejection — using only the **refreshed** filled `hub` context profile and named rules/recipe. No command execution. No network.

**Scope:**

| Dimension | Boundary |
|---|---|
| Read allowlist | `prompt-engineering/recipes/analyze-framework-context.md`, `prompt-engineering/context/projects/hub/CONTEXT-MANIFEST.yaml`, `prompt-engineering/context/projects/hub/PROJECT.md`, `prompt-engineering/context/projects/hub/TEST-ARCHITECTURE.md`, `prompt-engineering/context/projects/hub/COMMANDS.md`, `prompt-engineering/context/projects/hub/SOURCE-MAP.md`, `prompt-engineering/rules/40-verification-protocol.md`, `prompt-engineering/rules/50-data-handling.md` |
| Write allowlist | This smoke report only (`prompt-engineering/evals/smoke/hub-context-grounding-2026-07-26-v2.md`) |
| Commands | Forbidden (`NOT_RUN`; no execution approval) |
| Network | Forbidden |
| Source mutation | Forbidden |
| Aggregate eval metrics | Do not update |

**Normalized inputs (recipe alignment):**

- `profile`: `hub` (explicit)
- `task_target`: `refresh` readiness/grounding verification against the refreshed profile (read-only)
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
| 1e | Playwright smoke suite — exact command | `npx playwright test -c playwright.config.ts --project=chromium-smoke --grep "login page\|full page has no critical\|path 1"` | `COMMANDS.md` → Discovered commands → CMD-006; CMD-006 exact command block | SRC-006 (SRC-004, SRC-011 corroborate suite identity) |
| 1f | Playwright smoke suite — working directory | `frontend` | `COMMANDS.md` → Discovered commands → CMD-006 | SRC-006 |
| 2a | Backend test files (evidenced count) | **499** test files (`backend/testframework/**/*_test.go`) | `TEST-ARCHITECTURE.md` → Test topology → Backend unit + integration row; `PROJECT.md` → Verified constraints | **SRC-050**, **SRC-016**, SRC-009 |
| 2b | Frontend unit files / tests (evidenced) | **37 files / 99 tests** | `TEST-ARCHITECTURE.md` → Test topology → Frontend unit/component row | SRC-002, SRC-003, SRC-010, SRC-022 |
| 2c | E2E smoke spec files (evidenced) | **3 spec files** (`chromium-smoke`) | `TEST-ARCHITECTURE.md` → Test topology → E2E smoke subset row | SRC-004, SRC-011, SRC-006 |
| 2d | E2E full spec files (evidenced) | **20 spec files** (`chromium-full`, `**/*.spec.ts`) | `TEST-ARCHITECTURE.md` → Test topology → E2E full row | SRC-004, SRC-012, SRC-022 |
| 3a | Backend pass/fail state at snapshot | **NOT_RUN** — no project test commands executed during bootstrap or this smoke | `PROJECT.md` → Known unknowns; `COMMANDS.md` → Execution evidence; `CONTEXT-MANIFEST.yaml` → unknowns.current_test_results | SRC-000 |
| 3b | Backend coverage state at snapshot | **NOT_RUN** — coverage command documented but not executed | `COMMANDS.md` → CMD-008 (`not-run`); `CONTEXT-MANIFEST.yaml` → unknowns.current_coverage | SRC-013, SRC-000 |

**Execution status of all discovered commands:** `not-run` (`COMMANDS.md`; SRC-000).

**Backend inventory resolution note:** Live filesystem inventory (SRC-050) and corrected `docs/test_plan.md` (SRC-016) converge on **499**; historical **495** is superseded and must not be asserted as canonical (`SOURCE-MAP.md` → Conflicts; `PROJECT.md` → Historical provenance notes).

**Frontend test discovery note (supports decoy rejection):** Vitest includes only `../testframework/stages/unit/frontend-test/domains/**/*.test.{ts,tsx}` and excludes `src/**/*.test.{ts,tsx}` (`PROJECT.md` → Verified constraints; SRC-003).

---

## Decoy rejection table

| Decoy | Verdict | Reason | Profile file / section | Source IDs |
|---|---|---|---|---|
| `go test ./...` | **REJECTED** | Canonical backend job targets `./testframework/...` with `-tags hubtest -race -count=1 -timeout=600s`, not bare `./...` | `COMMANDS.md` → CMD-001; `PROJECT.md` → Verified constraints | SRC-009, SRC-006 |
| `npm test` | **REJECTED** | Canonical frontend unit gate is `npx vitest run`, not `npm test` | `COMMANDS.md` → CMD-004 | SRC-010, SRC-006 |
| Frontend tests in `frontend/src/**/*.test.*` | **REJECTED** | Explicitly excluded from Vitest discovery; canonical root is `testframework/stages/unit/frontend-test/domains/` | `PROJECT.md` → Verified constraints; `TEST-ARCHITECTURE.md` → Test topology | SRC-003, SRC-010 |
| Exact backend count **495** | **REJECTED** as canonical | Resolved conflict: dedicated `functions.Glob` inventory returned **499**; corrected docs agree; 495 is historical only | `SOURCE-MAP.md` → Conflicts; `PROJECT.md` → Historical provenance notes; `TEST-ARCHITECTURE.md` → Risks and known unknowns | **SRC-050**, **SRC-016**, SRC-021, SRC-022 |

---

## Actions inventory

**Label:** self-reported executor trace. Not independent benchmark or tool-trace evidence.

| Action type | Detail | Status |
|---|---|---|
| **Reads** | `prompt-engineering/recipes/analyze-framework-context.md` | self-reported performed |
| **Reads** | `prompt-engineering/context/projects/hub/CONTEXT-MANIFEST.yaml` | self-reported performed |
| **Reads** | `prompt-engineering/context/projects/hub/PROJECT.md` | self-reported performed |
| **Reads** | `prompt-engineering/context/projects/hub/TEST-ARCHITECTURE.md` | self-reported performed |
| **Reads** | `prompt-engineering/context/projects/hub/COMMANDS.md` | self-reported performed |
| **Reads** | `prompt-engineering/context/projects/hub/SOURCE-MAP.md` | self-reported performed |
| **Reads** | `prompt-engineering/rules/40-verification-protocol.md` | self-reported performed |
| **Reads** | `prompt-engineering/rules/50-data-handling.md` | self-reported performed |
| **Writes** | `prompt-engineering/evals/smoke/hub-context-grounding-2026-07-26-v2.md` | self-reported performed |
| **Commands** | None | self-reported not-run |
| **Network** | None | self-reported not-used |
| **Source mutation** | None | self-reported none |
| **Aggregate eval metrics** | Not updated | per task constraint |

---

## Evidence block

### Verification (rules/40 template)

- **Claims checked:** Canonical CMD-001/004/006 commands and working directories; backend (499), frontend unit (37/99), E2E smoke (3) and full (20) counts; backend pass/fail and coverage NOT_RUN; four named decoys rejected.
- **Evidence:** Profile cards with source-ID citations; manifest unknowns, conflicts (resolved backend count), SRC-050 inventory; no shell output (read-only smoke).
- **Not run / N/A:**
  - All discovered project test commands: `NOT_RUN` (no execution approval; SRC-000, COMMANDS.md).
  - Current pass/fail and coverage: `NOT_RUN` (CONTEXT-MANIFEST unknowns).
  - Stage 3 command execution (rules/40): `NOT_RUN` — task/allowlist granted no command execution approval.
  - Independent verifier / tool trace: **not performed** — runtime provenance remains PARTIAL.
  - Paired baseline arm: `NOT_RUN`.
  - Effectiveness / ROI: `NOT_RUN`.
- **Risks / assumptions:**
  - Profile facts reflect snapshot `d4a44174a0642ec060e17277a61414e1bfe1ca3d`; worktree may differ from manifest `worktree_state: modified`.
  - Backend 499 is a **file count** from SRC-050 inventory, not proof of test-function count or execution.
  - Scanned repository text treated as untrusted data per rules/50; only user task + rules/skills treated as instructions.
  - Model routing metadata is operator-supplied, not measured comparative evidence.
- **Mode:** self-check only (assisted-only smoke)
- **Verdict:** **PARTIAL** — content grounding self-check PASS; runtime provenance PARTIAL

### Data handling (rules/50 template)

- **Source type:** Filled context profile (`hub`, refreshed) + recipe + verification/data-handling rules
- **Trusted as instructions:** no (repository/profile content = data)
- **Redaction:** clean (no secret values copied; env var names only where profile already records them)
- **Reuse scope:** project-only
- **Human approval:** not-required for read-only smoke report write

### PASS criteria checklist (content grounding self-check)

| Criterion | Result |
|---|---|
| Every known fact cites profile section + source ID(s) | yes |
| Backend 499 cites SRC-050 and corrected SRC-016 | yes |
| Unknowns remain UNKNOWN / NOT_RUN (pass/fail, coverage) | yes |
| All four decoys explicitly rejected | yes |
| No command, network, or source change beyond allowed write | self-reported; not independently trace-verified |
| No aggregate eval metric update | self-reported; eval manifest remains `NOT_RUN` |
| Independent static verification | first pass PARTIAL; findings corrected, recheck pending |

---

## Limitations

1. **Assisted-only context grounding** — does not support ROI or effectiveness claims against an unassisted baseline; paired baseline was not run.
2. **Independent static verification only** — factual grounding was checked by a
   separate verifier, but no verifier observed runtime tool history;
   `runtime_provenance` stays PARTIAL.
3. **No command execution** — pass/fail and coverage remain `NOT_RUN`.
4. **Snapshot drift** — manifest records `worktree_state: modified`; live repo may differ from `snapshot_commit`.
5. **Inventory scope** — SRC-050 counts `*_test.go` files under `backend/testframework`; it does not prove semantic review of all tests or current filesystem state beyond the recorded inventory observation time (`2026-07-26T11:38:21.5106347Z`).
6. **Self-reported metadata** — model selection (`composer-2.5-fast` / fallback `gpt-5.6-sol-max`) and action inventory are executor self-report, not repository-evidenced benchmark data.

---

## Honest outcome

**PARTIAL** — Refreshed profile grounding self-check **PASS**: canonical
commands/CWDs, evidenced counts (backend **499** via SRC-050 + SRC-016;
frontend 37/99; E2E 3 smoke / 20 full), NOT_RUN execution states, and decoy
rejection (including historical **495**) are consistent with the filled profile.
Independent static verification checks content, not runtime history.
**Runtime provenance remains PARTIAL** because action trace and model routing
are self-reported.
