# Phase 64 — Auth Coverage Uplift (74.7% → 80%)

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-25

## Goal

Close COV-01 security-critical gap: `auth` package **≥80%** statement coverage.

## Deliverables

### Backend tests
- `repository_organization_test.go` — org, pending users, approve/reject/purge
- `service_organization_test.go` — OrganizationService + ValidateInviteToken
- `handler_organization_test.go` — ListPending, Approve, Reject, GetOrgInvite
- `service_facade_test.go` — AuthService Telegram facade delegates

## Results

| Metric | Before | After |
|--------|--------|-------|
| `auth/` coverage | 74.7% | **81.6%** |

## Verification

```bash
cd backend && go test -count=1 -coverprofile=cov_auth.txt ./internal/auth/...
go tool cover -func cov_auth.txt
```

## Success criteria

- [x] auth ≥80%
- [x] all auth tests pass
- [x] security-critical 6/6 ≥80% restored
