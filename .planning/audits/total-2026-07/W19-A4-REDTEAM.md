# W19-A4 — Red team (auth non-admin)

> 2026-07-28 · TASK-008 W19 · After Stage B · Product READ-ONLY  
> Threat actor: ordinary authenticated user (not system admin/founder)  
> Evidence: current code + W18 FAILs 014,019,041,042,044,045,046,052,054

## Scope / assumptions

- JWT session valid; no admin/founder bypass
- Default project roles: owner / developer / tester / viewer (`rbac/permissions.go`)
- Owner ≠ system admin (in-scope)

## Top 5 chains

### 1. Team attach → forced project membership
**Likelihood: Med · Impact: High · Reopens: TASK-014**

Project owner (or role with `team.manage`) attaches any org team via `POST /api/teams/{teamId}/projects/{projectId}` without team membership/consent. Service lacks `requireTeamMember`; repo bulk-adds members as developer.

### 2. ConfirmUpload ghost / weak entity bind
**Likelihood: Med–High · Impact: Med · Reopens: TASK-045**

Member with `PermFileUpload` can Confirm without proving S3 object / strong entity ownership for message/avatar/voice/channel → phantom attachments.

### 3. Concurrent StartSprint → dual active sprints
**Likelihood: Low–Med · Impact: Med · Reopens: TASK-041**

Race on StartSprint; Ent concurrent test stayed GREEN under SQLite masking — residual under real PG concurrency.

### 4. User-driven SSRF (push / link-preview) + To4 residual
**Likelihood: Low · Impact: High · Reopens: TASK-019**

Push subscribe / link-preview user URLs; To4() normalize present but drop-To4 mutation blind on Go 1.25 Contains → regression unwatched for IPv4-mapped metadata SSRF.

### 5. AI privacy / access fail-open + stream cancel gap
**Likelihood: Low–Med · Impact: High · Reopens: TASK-044**

External AI consent + nil access-checker fail-open paths; stream cancel / Meili SSRF hubtests missing.

## W18 FAIL → reopen map

| TASK | Reopens? | Why |
|------|----------|-----|
| 014 | YES | AddTeamToProject no team ACL; CompleteSprint mid-fail unproven |
| 019 | YES (residual) | To4 mutation blind on Go 1.25 |
| 041 | YES | Dual StartSprint Ent race unproven |
| 042 | PARTIAL | via_invite static assert |
| 044 | YES | No Meili/stream-cancel hubtest |
| 045 | YES | Confirm→404 unproven |
| 046 | PARTIAL | Runtime unique/CHECK unproven |
| 052 | LOW | FE type quality |
| 054 | NO (exploit) | Docs drift |

## Verdict

0 new CRITICAL proven open for ordinary non-admin; residual Med/High chains tied to W18 FAILs. Blockers for “mutation-closed”: **014, 041, 045** first.
