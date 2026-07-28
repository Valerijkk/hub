# W19-A4 — Red team (auth non-admin)

> 2026-07-28 · Агент **W19-A4** (security-auditor) · ветка `cursor-grok-4.5` · **no push master**  
> Задача: [TASK-008](../../backlog/TASK-008-total-audit-program.md) § Stage C / W19  
> Threat actor: **ordinary authenticated user** (JWT valid; **not** `system_role` admin/founder)  
> Метод: цепочки на **текущем** коде после Stage B; evidence = `path:line` + W18 FAIL reopen  
> Product code: **READ-ONLY** · этот файл — audit artifact  
> Re-verify pass: **2026-07-28 19:35 UTC+3** (post–Stage B HEAD на `cursor-grok-4.5`)

## Scope / assumptions

| In | Out |
|----|-----|
| Project roles owner / developer / tester / viewer | System admin / founder privilege |
| Org-level auth (`/api/teams`, magic-link session) | SSO (ADR pending), multi-org |
| Residual after Stage B / W18 FAIL | Closed W18 PASS as primary exploit (unless combo) |

Owner ≠ system admin — **in scope** (any user who created a project).

## Severity scale (this doc)

- **Likelihood:** Low / Med / High — prerequisites (role, UUID leak, race window, mis-wiring)
- **Impact:** Low / Med / High / Critical — confidentiality / integrity / compliance
- **OWASP:** A01 Access Control · A04 Insecure Design · A07 Auth · A10 SSRF · A08 Integrity

---

## Top 5 attack chains

### 1. Team attach → forced project membership (no team ACL)

| | |
|--|--|
| **Likelihood** | **Med** — needs `PermTeamManage` on **target project** (default: **owner**) + known `teamId` (UUID leak / prior membership / side channel; `ListTeams` already scoped to membership) |
| **Impact** | **High** — bulk-adds **all** team members as `developer` without consent / team-ACL; expands write surface (tasks, files, channels) |
| **Reopens** | **TASK-014** (P1-BE-08 stranger `AddTeamToProject`) |
| **OWASP** | A01 Broken Access Control |

**Chain**

1. Attacker is **project owner** (non-admin): `DefaultRolePermissions["owner"]` = all bits incl. `PermTeamManage` — `backend/internal/rbac/permissions.go:42-46`.
2. Learns foreign `teamId` (not current member; `ListTeams` only own teams — `project/service.go:521-531`).
3. `POST /api/teams/:teamId/projects/:projectId` — RBAC only checks project `team.manage` — `project/handler.go:58`, `688-701`.
4. Service **does not** call `requireTeamMember` (unlike `AddTeamMember` / `RemoveTeamMember`) — `project/service.go:545-579` vs `:534-543`.
5. Repo attaches team + loops members → `AddMember(..., ProjectRoleDeveloper)` — `project/repository.go:600-624`.

**PoC sketch:** owner of Project P + UUID of Team T (attacker not in T) → 200 → every T member is developer on P.

**Fix direction:** `requireTeamMember(ctx, teamID, actorID)` before `repo.AddTeamToProject`; hubtest stranger denial (W18-C2 FAIL criterion).

---

### 2. ConfirmUpload ghost object + weak entity bind

| | |
|--|--|
| **Likelihood** | **Med–High** — any member with `PermFileUpload` (developer/tester); HEAD path unproven by mutation |
| **Impact** | **Med** (→ High if combined with empty entity) — phantom metadata / bind to arbitrary message|channel UUID; NF-09-adjacent |
| **Reopens** | **TASK-045** (P2-BE-21 Confirm без object→404); combo W19-A3 **C-007** |
| **OWASP** | A01 · A04 |

**Chain**

1. Presign upload (project member) → pending `FileMetadata`.
2. `ConfirmUpload`: if `s3Client != nil`, `assertObjectWithinLimit` maps S3 404 → `ErrObjectNotFound` — `storage/service.go:154-157`, `:174-183`.
3. W18-C2b: mutating only `isS3NotFound` mapping **does not RED** Confirm suites; task test tautology — residual **unproven** Confirm→404.
4. `validateEntityScope`: `message`/`avatar`/`voice`/`channel` = non-empty string only — **no** membership / live-entity check — `storage/service.go:240-245`.
5. Empty `entityType`+`entityID` → **nil** (allow) — `:247-248`.

**PoC sketch:** Confirm pending file with `entity_type=message`, foreign/soft-deleted message UUID → confirmed attachment; or Confirm without object when HEAD guard regresses / `s3Client==nil` skips HEAD (`:154`).

**Fix direction:** Confirm integration: pending + Head 404 → `ErrObjectNotFound`; reject empty entity; membership for message/channel.

---

### 3. GDPR Mongo eraser silent success → HardDelete with live messages

| | |
|--|--|
| **Likelihood** | **Low–Med** — erase is **admin** endpoint; trigger = admin GDPR + nil/`coll==nil` eraser (mis-wire / `mongoDB` nil). Residual read by **any remaining non-admin** channel member |
| **Impact** | **Critical** (compliance / right-to-erasure) — user HardDeleted, Mongo messages remain; A3 **C-001** |
| **Reopens** | **TASK-079** (catch-up W18-C5-001/002); related TASK-034 cascade PASS but eraser fail-open |
| **OWASP** | A01 · A04 Insecure Design |

**Chain**

1. Admin calls `DELETE /api/admin/users/:id/gdpr` — `audit/handler.go` GDPR route.
2. Wiring always calls `SetMessageEraser(NewMongoMessageEraser(mongoDB))` — `cmd/server/wiring_messenger.go:66`.
3. `NewMongoMessageEraser(nil)` → struct with `coll==nil` (documented no-op) — `audit/gdpr_mongo.go:17-23`.
4. Cascade runs `msgEraser.DeleteMessagesBySender` — `audit/service.go:523-529`.
5. `DeleteMessagesBySender`: `e==nil || coll==nil` → **`(0, nil)`** — `gdpr_mongo.go:27-30`.
6. Cascade treats Mongo as OK → audit log → `HardDeleteUser` — `service.go:532-544`.
7. **Non-admin** members still read residual messages in Mongo channels (PII of «erased» user).

**PoC sketch:** wire `NewMongoMessageEraser(nil)` (or nil `mongoDB` at boot) → GDPR 200 → count messages by `sender_id` > 0.

**Fix direction:** nil coll → error; hubtest real `DeleteMany`; cascade abort (TASK-079 ACCEPTANCE). Also fail-closed if `msgEraser` unset (`service.go:523` skip path).

> Actor note: step 1 is admin; **impact surface for ordinary non-admin** = continued access to data that must be gone + false compliance. Included per W19 catch-up + C-001 CRITICAL.

---

### 4. User-driven SSRF (push / link-preview) + To4 mutation-blind residual

| | |
|--|--|
| **Likelihood** | **Low** — To4 present; Go 1.25 `Contains` may already block mapped IPv4; mutation drop-To4 stayed GREEN (W18-C2) |
| **Impact** | **High** — cloud metadata / RFC1918 via push endpoint or OG preview if blocklist bypass returns |
| **Reopens** | **TASK-019** (P1-BE-23 residual); Meili host SSRF tests missing → **TASK-044** partial |
| **OWASP** | A10 SSRF |

**Chain**

1. Non-admin registers WebPush endpoint — `ValidateOutboundURL` only (scheme/userinfo) — `notification/push.go:220-227`; dial-time block in `SafeDialContext`.
2. Or posts URL → `FetchOpenGraphMetadata` — `platform/http/opengraph.go` (ValidateOutboundURL + SafeHTTPClient).
3. Defense: `isBlockedIP` normalizes via `To4()` — `platform/http/safedialer.go:66-81`.
4. W18 FAIL: remove `To4()` → tests stay GREEN on Go 1.25 → **regression unwatched** for `::ffff:169.254.169.254`.
5. Parallel: Meili indexer has To4 link-local check — `search/meilisearch_indexer.go:70` — **no hubtest** SSRF (TASK-044).

**PoC sketch:** subscribe `http://[::ffff:169.254.169.254]/` after hypothetical To4 regression → SafeDial / metadata hit.

**Fix direction:** mutation-sensitive assert independent of Go Contains behavior; Meili `MEILI_HOST` private → nil indexer hubtest.

---

### 5. AI privacy / access fail-open + stream/Meili gaps

| | |
|--|--|
| **Likelihood** | **Low–Med** — prod wiring sets checker (`wiring_audit_ai.go`, `wiring_integrations.go`); fail-open if forgotten; user can set `allow_external_ai` |
| **Impact** | **High** — IDOR issue title/description / PR context via @hub-ai; external LLM data exfil if consent default regresses |
| **Reopens** | **TASK-044** (stream cancel + Meili SSRF hubtests missing); W18-C5-004 consumers; combo W19-A3 **C-006** |
| **OWASP** | A01 · A07 |

**Chain**

1. `TrackerIssueLookup`: `if access != nil` only — **skip** when nil — `ai/issue_lookup.go:56-65`.
2. `authorizeReview`: `if p.access == nil { return nil }` — `ai/code_review.go:66-69`.
3. Contrast: `RBACProjectAccess` fail-closed on `Svc==nil` — `ai/project_access.go:21-25`.
4. Preference default `AllowExternalAI: false` mutation-proven — `ai/service.go` GetOrCreatePreference — but stream cancel mid-read **no hubtest** (TASK-044).
5. Non-admin @mentions bot / code-review URL on foreign project → context leak if checker unset.

**PoC sketch:** unit without `SetAccessChecker` + foreign issue key → expect NotFound; today returns title/desc.

**Fix direction:** consumers fail-closed on nil access; stream cancel + Meili SSRF hubtests.

---

## Additional residual chains (not top-5, still mapped)

| # | Chain | L / I | TASK | Evidence |
|---|-------|-------|------|----------|
| 6 | Dual `StartSprintAtomic` → two active (SQLite masks) | Low–Med / Med | **041** | Guard present `sprint/repository.go:189-238`; W18-C3 Ent concurrent GREEN without conflict assert |
| 7 | `via_invite` payload assert handmade (invite join integrity signal) | Low / Low–Med | **042** PARTIAL | Production sets `"via_invite": true` — `messenger/service_channel_invites.go:194`; mutation false stays GREEN |
| 8 | Ent unique/CHECK source-grep only | Low / Med | **046** PARTIAL | No runtime duplicate Create / XOR debit-credit |
| 9 | FE DTO type-lie (`version` / guest cast) | Low / Low | **052** | Quality; weak optimistic-lock UX |
| 10 | Docs metrics drift | N/A exploit | **054** | No runtime attack; TASK-080 autotest |

---

## W18 FAIL → reopen map

| TASK | Reopens for red-team? | Why (current code) |
|------|----------------------|---------------------|
| **014** | **YES** | `AddTeamToProject` skips `requireTeamMember` — `project/service.go:571-579` |
| **019** | **YES** (residual) | To4 present but mutation-blind; push/OG user URLs |
| **041** | **YES** | Dual Start unproven under real concurrency |
| **042** | **PARTIAL** | via_invite static assert only; other ClassIDs mutation-PASS |
| **044** | **YES** | No Meili SSRF / stream-cancel hubtest; AI nil-access fail-open |
| **045** | **YES** | Confirm→404 unproven; weak non-issue entity bind |
| **046** | **PARTIAL** | Schema grep ≠ runtime unique/CHECK |
| **052** | **LOW** | FE types — no direct BE authz bypass |
| **054** | **NO** (exploit) | Docs drift only |
| **079** | **YES** | `gdpr_mongo.go:27-30` `(0,nil)` + cascade HardDelete; wiring `wiring_messenger.go:66` |

---

## Checklist (Hub invariants) — non-admin surface

| Check | Result |
|-------|--------|
| Fiber v3 RBAC order on `AddTeamToProject` | OK — handler then MW (`handler.go:58`) |
| Team ACL on attach | **FAIL** — missing `requireTeamMember` |
| Storage Confirm object proof | **Residual** — code path exists, mutation FAIL |
| SSRF SafeDial + ValidateOutboundURL | Present; To4 residual |
| AI access fail-closed | Adapter OK; **consumers** fail-open on nil |
| GDPR eraser fail-closed | **FAIL** — nil coll success |
| Search user-scope / WS | Not re-opened as top chain this pass (Stage B PASS assumed) |

---

## Verdict

- **0 new CRITICAL** proven as single-step exploit by non-admin **alone**.
- **1 CRITICAL combo** (TASK-079 + admin GDPR) leaves residual PII for ordinary members — treat as Stage C blocker with A3 C-001.
- Residual **Med/High** chains tied to W18 FAILs: **014, 045, 079, 019, 044** (then 041).

**Priority reopen-fix:** `014` → `079` → `045` → `044`/`019` → `041`.

```
verified: true — 2026-07-28 19:35 UTC+3
```

`verified: true` = audit complete with file:line evidence; **not** «product secure». Open HIGH/CRITICAL residuals above block «mutation-closed» for Stage C.
