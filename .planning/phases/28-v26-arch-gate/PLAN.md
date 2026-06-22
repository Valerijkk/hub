# Phase 28 — v2.6 Architecture Milestone Gate

> **Branch:** `feat/v2.5-trust` · **Depends on:** Phase 27 auth Repository (ARCH-01 17/17)

## Exit criteria

ARCH-01 verified · no `*ent.Client` in rbac middleware · full `go test ./...` green · milestone v2.6 Architecture closed

---

## Wave 1 — ARCH cleanup + gate

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W1-01 `rbac.RequireAdmin` → Repository | be | arch-ver | ✓ |
| W1-02 Full backend regression | qa | be-ver | ✓ |
| W1-03 STATE/ROADMAP/COVERAGE milestone close | sa | arch-ver | ✓ |

**Next:** Phase 29 OBS-01 / Wave 3 Vitest

---
