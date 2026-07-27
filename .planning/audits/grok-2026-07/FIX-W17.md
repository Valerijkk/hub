# FIX-W17 — Block F handler polish

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** be1  
**Правила:** sample 3–5 critical handlers / domain · fix only CLEAR P1/P2 · 1 commit / finding · no schema · no push master

Домены: `auth`, `rbac`, `audit`, `platform`, `project`, `messenger`, `ws`.

Checklist: authz scoped · input validation · pagination limits · status 400/401/403/404/409 · `{error,request_id}` / AppError · 404 vs 403 foreign · no PII in logs · idempotency where dangerous.

---

## Итог fixes

| ID | Sev | Status | SHA | Notes |
|----|-----|--------|-----|-------|
| A1-009 | P1 | FIXED | `7002ce6` | Ban/Deactivate reject `SystemRoleFounder` → 403 |
| A1-013 | P2 | FIXED | `4884fb9` | `UpdateResourceLimit` 3-key allowlist (`limit.file.max_size_mb`, `limit.user.max_projects`, `limit.project.max_storage_mb`); reject `retention.*` / unknown → 400 |
| A2-009 | P2 | FIXED | `ed3b772` | `PublishTyping` requires `IsMember` → 403 |
| A14-007 | P2 | FIXED | `ec2e23f` | Fiber recover `EnableStackTrace` only if `DevMode` |
| A14-006 | P2 | FIXED | `a619a04` | Mask email in logs; never log OTP (even dev) |
| A1-010 | P2 | FIXED | `5a1858d` | Magic-link `Del` fail → verify error (one-time code) |

### WEAK close (2026-07-28 · qa-auto1)

| ID | Gap closed | Test SHA | Autotest |
|----|------------|----------|----------|
| A14-006 | OTP slog path unproven (only maskEmail) | `0cbc7d2` | `TestMagicLink_OTPNotLogged_A14_006` (slog spy; mutation: `"code", code`) |
| A14-007 | EnableStackTrace↔DevMode untested | `4ebd8c3` | `TestFiberRecoverEnableStackTrace_FollowsDevMode_A14_007` + wiring → helper |

Prod fixes remain `a619a04` / `ec2e23f`; helper already in tree (`c5613f6`); this wave closes WEAK with mutation-resistant tests.

---

## Coverage matrix (sampled handlers)

Legend: **PASS** = checklist item OK · **GAP** = residual (not clear P1/P2 or deferred) · **FIXED** = closed this wave.

### auth

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| `SendMagicLink` | n/a (public) | PASS email | n/a | PASS 400/429 | PASS | n/a | **FIXED** A14-006 | PASS rate NX |
| `VerifyMagicLink` | n/a | PASS rune len | n/a | PASS 400/401 | PASS | n/a | **FIXED** A14-006 | **FIXED** A1-010 |
| `RefreshToken` | PASS JWT | PASS body | n/a | PASS 401 | PASS | n/a | PASS | PASS rotate |
| `GetMe` / sessions | PASS user-scope | PASS | PASS | PASS | PASS | n/a | PASS | n/a |
| Admin pending/approve | PASS `IsSystemAdmin` | PASS UUID | PASS | PASS 403 | PASS | PASS | PASS | n/a |

Prior waves: A1-001 pending→JWT, A1-007 Resend fail-closed — already FIXED.

### rbac

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| `CreateRole` | PASS RBAC mw | PASS body | n/a | PASS 400 | PASS fiber.Error→EH | n/a | PASS | n/a |
| `UpdateRole` | PASS + project scope | PASS | n/a | PASS 403 system / **PASS** 404 mismatch | PASS | **PASS** (404 foreign role) | PASS | n/a |
| `DeleteRole` | PASS + project scope | PASS | n/a | PASS 403/404 | PASS | **PASS** | PASS | n/a |
| `AssignRole` | PASS + role∈project | PASS UUIDs | n/a | PASS 404 | PASS | **PASS** | PASS | n/a |

Prior: A1-002/003 IDOR, A1-005/006 founder bypass — FIXED earlier waves. No new P1/P2 this sample.

### audit

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| `UpdateLimit` | PASS admin | **FIXED** A1-013 | n/a | **FIXED** 400 unknown key | PASS | n/a | PASS | n/a |
| `DeactivateUser` | PASS admin | PASS UUID | n/a | **FIXED** 403 founder | PASS | n/a | PASS | n/a |
| `BanUser` | PASS admin | PASS reason | n/a | **FIXED** 403 founder | PASS | n/a | PASS | self-ban block |
| `ListTrash` | PASS admin | PASS type | **PASS** clamp ≤100 | PASS | PASS | n/a | PASS | n/a |
| `DeleteUserGDPR` | PASS admin | PASS | n/a | PASS 403 founder (A1-004) | PASS | n/a | PASS | n/a |

### platform

| Surface | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| ErrorHandler / RequestID | n/a | n/a | n/a | PASS | **PASS** AppError shape | n/a | PASS | n/a |
| RateLimit (API/auth/files/msg) | n/a | n/a | n/a | PASS 429 | PASS | n/a | PASS | n/a |
| Fiber recover | n/a | n/a | n/a | PASS | PASS | n/a | **FIXED** A14-007 | n/a |
| Logger access | n/a | n/a | n/a | n/a | n/a | n/a | PASS (no body dump) | n/a |
| SafeDialer (outbound) | n/a | PASS URL | n/a | n/a | n/a | n/a | PASS | n/a |

Residual GAP (not this Block F code fix): A14-008 access-log without `InfoContext`/trace — P2 obs, deferred.

### project

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| `CreateProject` | PASS auth | PASS body | n/a | PASS 400 | PASS | n/a | PASS | n/a |
| `ListProjects` | PASS user-scope | PASS | n/a (user-sized) | PASS | PASS | n/a | PASS | n/a |
| `GetProject` | PASS auth | PASS UUID | n/a | PASS 404 missing | PASS | **GAP** non-member still 200 metadata | PASS | n/a |
| `JoinByInvite` | PASS auth | PASS code | n/a | PASS 400/404 | PASS | PASS | PASS | invite consume |
| `AddMember` / remove | PASS (service) | PASS | n/a | PASS 404/409 | PASS | PASS | PASS | n/a |

`GetProject` non-member metadata leak = product-grey; not clear P1 without ADR — left **GAP**.

### messenger

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| `SendMessage` | PASS IsMember | PASS | n/a | PASS 403 | PASS AppError | PASS 403 | PASS | n/a |
| `ListMessages` | PASS IsMember | PASS | **PASS** mongo clamp 1..100 | PASS | PASS | PASS | PASS | cursor |
| `SearchMessages` | PASS member-scope (A2-001) | PASS | PASS limit | PASS 403 | PASS | PASS | PASS | n/a |
| `SendTyping` | **FIXED** A2-009 | PASS UUID | n/a | **FIXED** 403 | PASS | PASS | PASS | ephemeral |
| `GetPresence` | PASS auth | PASS user_ids | soft | PASS 400 | PASS | n/a | PASS | n/a |

Residual GAP P3: `queryInt` no handler clamp (A2-014) — mongo still clamps; skipped.

### ws

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Upgrade `/ws` | PASS JWT + active | PASS token | n/a | PASS 401/403 | PASS | n/a | **GAP** P3 token in query (A2-013) | n/a |
| Subscribe room | **PASS** Authorizer fail-closed | PASS room | n/a | deny unknown prefix | n/a | n/a | PASS | n/a |
| CheckOrigin | PASS CORS allowlist | PASS | n/a | reject | n/a | n/a | PASS | n/a |
| readPump size | n/a | PASS maxMessageSize | n/a | close | n/a | n/a | PASS | n/a |

---

## Tests run

```
go test -tags hubtest -race -count=1 ./testframework/audit/   # PASS (incl. A1_009, A1_013)
go test -tags hubtest -race -count=1 ./testframework/auth/    # PASS (incl. A14_006, A1_010)
go test -tags hubtest -race -count=1 ./testframework/messenger/ # PASS (incl. A2_009, SendTyping)
```

---

## Out of scope / residual

| Item | Why skipped |
|------|-------------|
| Project `GetProject` non-member 200 | Product-grey; needs ADR before 404-hide |
| A2-014 `queryInt` clamp | P3; mongo clamps |
| A2-013 WS JWT in query | P3; ticket exchange = larger design |
| A14-008 logging InfoContext | obs polish, not handler authz |
| Parallel dirty trees (imex/storage/search) | Other agents — not staged |

---

## Docs

- This report: `.planning/audits/grok-2026-07/FIX-W17.md`
- No API contract change beyond stricter 400/403 on existing admin/messenger paths
