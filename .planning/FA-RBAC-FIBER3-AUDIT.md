# FA-RBAC-FIBER3 — Fiber v3 RBAC Middleware Order Audit

**Date:** 2026-06-22  
**Severity:** CRITICAL (OWASP A01 — Broken Access Control)  
**Status:** **FIXED** (73 routes across 5 modules + TMS 50 routes prior)

---

## Root Cause

Fiber v3 signature: `Post(path string, handler Handler, middleware ...Handler)`.

The router builds the chain as `[middleware..., handler]`. Variadic middleware runs **before** the main handler.

| Pattern | Chain | RBAC |
|---------|-------|------|
| ❌ `Post(path, rbacMW, h.Handler)` | `[h.Handler, rbacMW]` | **Bypassed** |
| ✅ `Post(path, h.Handler, rbacMW)` | `[rbacMW, h.Handler]` | **Enforced** |

Hub historically used the wrong order on tracker, sprint, github, gitlab, rbac (and TMS before FA-02-02).

---

## Fixed Modules

| File | Routes fixed |
|------|-------------:|
| `internal/tms/handler.go` | 50 (FA-02-02) |
| `internal/tracker/handler.go` | 46 |
| `internal/sprint/handler.go` | 11 |
| `internal/rbac/handler.go` | 6 |
| `internal/github/handler.go` | 5 |
| `internal/gitlab/handler.go` | 5 |
| **Total** | **123** |

---

## Not Affected

Modules using `Group().Use(authMw)` + handler-only routes, or in-handler `authorizeProject()` (storage, apigateway post-APIGW-01).

---

## Regression Tests

- `testframework/tms/handler_webhook_rbac_test.go`
- Add tracker/sprint non-member 403 tests (recommended)

---

## Related

- FA-02-02 TMS webhook IDOR
- APIGW-01 webhook CRUD RBAC
- `.planning/FA-01-AUDIT.md`
