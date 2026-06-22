# Phase 27 — Architecture: auth (ARCH-01)

> **Branch:** `feat/v2.5-trust` · **Depends on:** Phase 26 rbac Repository

## Exit criteria

`auth` on Repository · middleware без `*ent.Client` · **17/17 ARCH-01** · tests green

---

## Wave 1 — auth Repository

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W1-01 `auth.Repository` + `EntRepository` + tx helpers | be | arch-ver | ✓ |
| W1-02 Sessions/Login/Profile/Reg/Org → repo | be | be-ver | ✓ |
| W1-03 middleware + tests + main.go | qa | be-ver | ✓ |

---
