# Phase 22 — Architecture: storage (ARCH-01)

> **Branch:** `feat/v2.5-trust` · **Depends on:** Phase 21 backup Repository

## Exit criteria

`storage` module on Repository · orphan cleanup via repo · tests green

---

## Wave 1 — storage Repository

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W1-01 `storage.Repository` + `EntRepository` | be | arch-ver | ✓ |
| W1-02 `StorageService` + `OrphanCleanupService` → repo | be | be-ver | ✓ |
| W1-03 Tests + main.go wiring | qa | be-ver | ✓ |

**Next:** apigateway (APIKey + Webhook), imex, dashboard, rbac

---
