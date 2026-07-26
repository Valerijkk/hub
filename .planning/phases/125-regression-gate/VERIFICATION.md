# Phase 125 Verification — COV / Security / Regression Gate

**Date:** 2026-07-27 (retro)  
**Auditor:** Grok 4.5 · PW2 planning-audit  
**Verdict:** PARTIAL

> PLAN Status COMPLETE (**2026-06-21**) records phase-exit metrics below. Live Docker race + cover gate **not** re-run on 2026-07-27. Later published overall is **83.0%** (**2026-06-22** Docker baseline in `COVERAGE.md`) — distinct dated run, not a substitute PASS stamp for this file without re-execution.

## Goal

| Metric | Target |
|--------|--------|
| COV overall | ≥82% |
| auth, ws, rbac, audit | ≥80% each |
| notification | ≥82% |
| Full regression (`-tags hubtest -race`) | pass |

## PLAN-recorded results (phase exit, 2026-06-21)

| Metric | Target | Recorded result |
|--------|--------|-----------------|
| COV overall | ≥82% | **82.6%** |
| auth | ≥80% | **82.4%** |
| ws | ≥80% | **87.7%** |
| rbac | ≥80% | **84.7%** |
| audit | ≥80% | **80.5%** |
| notification | ≥82% | **82.5%** (testframework) |
| Full regression | pass | ✅ claimed in PLAN |

## Later dated baseline (context)

| Metric | Value | Date |
|--------|-------|------|
| Overall `internal/` | **83.0%** | **2026-06-22** Docker (`COVERAGE.md`, post FA-11/12) |
| Security 6/6 ≥80% | met | **2026-06-22** Docker |

These support that the **program** later held the floor; they are **not** a 2026-07-27 re-verification of Phase 125 commands.

## Live re-measure

```text
Docker compose test stack + go test -tags hubtest -race / -cover
→ NOT EXECUTED in this PW2 pass
```

## Criteria

| Criterion | Target | Actual (dated) | Status |
|-----------|--------|----------------|--------|
| COV overall | ≥82% | **82.6%** (2026-06-21 PLAN) | recorded PASS · **not re-verified** |
| Security packages | ≥80% | PLAN table all ≥80% | recorded PASS · **not re-verified** |
| notification | ≥82% | **82.5%** (PLAN) | recorded PASS · **not re-verified** |
| Race regression | pass | PLAN claim only | **not re-verified** |

## Why not PASS

Avoids inventing PASS to match STATE 100% without a current Docker gate. Historical PLAN numbers are dated and internally consistent with later 83.0% baseline, but verification artifact requires live evidence for an unqualified PASS.

## Remaining

Run PLAN Verify block under Docker; stamp new date; then upgrade to PASS if thresholds hold.
