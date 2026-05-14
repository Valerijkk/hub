# Deferred Items — Phase 14 Architectural Hardening

Items discovered during execution that are outside the scope of the executing plan.

## From 14-01 (AuthService split)

### Pre-existing vet failure: unused import in apigateway webhook_test.go

- **File:** `backend/internal/apigateway/webhook_test.go:4`
- **Symptom:** `vet.exe: internal\apigateway\webhook_test.go:4:2: "crypto/hmac" imported and not used`
- **Discovered:** During 14-01 final `go vet ./...` over entire backend.
- **Origin:** Pre-existing, introduced in commit `253cc53` (LO-06 guard tests). Unrelated to AuthService.
- **Scope:** Out of scope for 14-01 (`files_modified` excludes this path; apigateway is not part of the auth split).
- **Recommendation:** One-line fix — remove the unused `crypto/hmac` import. Could be folded into a hygiene commit during 14-02..14-05 or assigned to a follow-up Phase 14 cleanup task.
- **Plan impact:** None. `go vet ./internal/auth/...` (the scope mandated by 14-01 verification) is clean.
