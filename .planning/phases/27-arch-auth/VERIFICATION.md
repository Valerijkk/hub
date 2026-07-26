# Phase 27 Verification — Architecture: auth (ARCH-01)

**Date:** 2026-07-27  
**Auditor:** Grok 4.5 (PW2 retro-verify)  
**Mode:** documentation-only · measured against current HEAD  
**Verdict:** PASS

## Goal-backward

| Criterion | Expected | Actual (2026-07-27) | Status |
|-----------|----------|---------------------|--------|
| `auth.Repository` + `EntRepository` | present | `repository.go` — interface + `EntRepository` + tx helpers | PASS |
| Sessions/Login/Profile/Reg/Org → repo | services on Repository | `SessionService`, `LoginService`, `ProfileService`, `RegistrationModeService`, `OrganizationService`, `AuthService` all hold `repo Repository` | PASS |
| Middleware without `*ent.Client` | repo / service only | `RequireAuth(svc *AuthService)`; `userIsDeactivated(..., repo Repository, ...)`; 0 non-comment `*ent.Client` outside `repository.go` (handler comment ARCH-H1 confirms intent) | PASS |
| Wiring | DI | `wiring_auth.go` — `auth.NewAuthService(auth.NewEntRepository(d.db), ...)` | PASS |
| ARCH-01 | **17/17** | Former holdouts from ARCHITECTURE.md §6 all have Repository interfaces: **7/7** (`auth,dashboard,backup,imex,storage,apigateway,rbac`) | PASS |
| Tests green | package hubtest | `go test -tags hubtest -count=1 ./testframework/auth/` → **ok** (3.308s) | PASS |

## Commands run

```text
# leak scan auth (exclude repository + comments): 0
Select-String backend/internal/auth/middleware.go 'RequireAuth|Repository'
cd backend && go test -tags hubtest -count=1 -timeout=300s ./testframework/auth/
```

## Residual

None for exit criteria. Milestone close formalized in phase 28.
