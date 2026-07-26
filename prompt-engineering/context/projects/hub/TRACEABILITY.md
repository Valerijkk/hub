---
context_schema: "1.0"
profile_slug: "hub"
status: filled
generated_at: "2026-07-26T02:25:00+03:00"
snapshot_commit: d4a44174a0642ec060e17277a61414e1bfe1ca3d
---

# Traceability

This map records observed links. It does not claim coverage when a requirement, test case, automated test, or CI job cannot be evidenced.

## Identifier conventions

| Kind | Observed format | Source | Confidence |
|---|---|---|---|
| Requirement | `US-*`, `FR-*`, `NFR-*`; security requirements/findings also use `SEC-*` and `NF-*` | SRC-020 | high |
| Test case | `TC-<MODULE>-NN` | SRC-020, SRC-022 | high |
| Automated test | Go `Test<Function>_<Scenario>`; frontend `*.test.ts(x)`; E2E `*.spec.ts` | SRC-017, SRC-003, SRC-004 | high |
| Issue/change | `ADR-*`, `FA-*`, `PG-*`, `TF-*`, `UX-*`, and other catalogued prefixes | SRC-020 | high |

## Requirement-to-test map

| Requirement/source | Test case | Automated test path and lines | CI job | Evidence |
|---|---|---|---|---|
| `SEC-01` API-key middleware ordering | UNKNOWN | `backend/testframework/apigateway/apikey_auth_test.go:146-185` | `backend-test` | SRC-016, SRC-018, SRC-043, SRC-009 |
| `SEC-02` mandatory LiveKit webhook body hash | UNKNOWN | `backend/testframework/calls/recording_webhook_test.go:156-271` | `backend-test` | SRC-016, SRC-018, SRC-044, SRC-009 |
| `SEC-05` no send-on-closed-channel panic | UNKNOWN | `backend/testframework/ws/hub_test.go:80-163` | `backend-test` | SRC-016, SRC-018, SRC-045, SRC-009 |
| `docs/TESTING.md:212` | `TC-AUTH-05` | `backend/testframework/auth/service_test.go:142-157` (`TestJWTExpired`) | `backend-test` | SRC-022, SRC-027, SRC-009 |
| `docs/test_plan.md:147-150` | UNKNOWN | `testframework/stages/e2e/domains/auth/login.spec.ts:8-40` | `e2e-smoke`, `e2e-full` | SRC-016, SRC-042, SRC-011, SRC-012 |

## Unlinked tests

| Test path and lines | Observed purpose | Missing link | Evidence |
|---|---|---|---|
| `testframework/stages/unit/frontend-test/domains/tracker/issue-form.test.tsx:41-80` | Form validation and successful create submission | Inline `TC-*` / `US-*` / `FR-*` / `NFR-*` identifier: UNKNOWN | SRC-034 |
| `testframework/stages/unit/frontend-test/domains/ai/ai-privacy.test.tsx:10-33` | External-AI consent load/toggle | Inline `TC-*` / requirement identifier: UNKNOWN | SRC-035 |
| `testframework/stages/unit/frontend-test/domains/auth/auth-store.test.ts:14-65` | Auth store token, user, and logout behavior | Inline `TC-*` / requirement identifier: UNKNOWN | SRC-033 |

## Uncovered or unverifiable requirements

| Requirement/source | Gap | Why unresolved | Evidence |
|---|---|---|---|
| `TC-AUTH-08` | Exact automated test path UNKNOWN | Checked-out test matrix labels it E2E/manual without an exact path | SRC-022 |
| `TC-AUTH-09` | Exact automated test path UNKNOWN | Checked-out test matrix labels it E2E/manual without an exact path | SRC-022 |
| `TC-AUTH-10` | Exact automated test path UNKNOWN | Checked-out matrix says manual/middleware test but does not identify a function or file | SRC-022 |
| Complete `US → FR/NFR → TC → automated test → CI` coverage | UNKNOWN | Sampled sources evidence selected links, not an exhaustive resolvable chain for every requirement | SRC-020, SRC-022, SRC-046 |

## Traceability rules

| Rule | Enforcement location | Evidence |
|---|---|---|
| New test automation should follow `US → FR/NFR → TC → automated test` traceability | `docs/GLOSSARY.md` | SRC-020 |
| Edge/corner automation should add a TC first, then record a TC-ID comment and the automated-test column; Allure may use `AS_ID` | `testframework/guides/06-edge-corner-coverage.md` | SRC-046 |
| `docs/TESTING.md` maps TC rows to automated test names or explicitly marks manual/E2E gaps | `docs/TESTING.md` | SRC-022 |
| CI suite-to-job mapping is recorded in checked-out job manifests | `testframework/pipeline/jobs/*.yaml` | SRC-009, SRC-010, SRC-011, SRC-012 |

## Manual mappings

Refresh preserves this human-owned block.

<!-- BEGIN MANUAL -->
UNKNOWN
<!-- END MANUAL -->
