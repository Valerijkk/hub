# W19-A4 — Red team (auth non-admin)

> 2026-07-28 · TASK-008 W19 · After Stage B · Product READ-ONLY  
> Threat actor: ordinary authenticated user (not system admin/founder)  
> Evidence: current code + W18 FAILs 014,019,041,042,044,045,046,052,054 + TASK-079

## Top 5 chains

1. **AddTeamToProject without team ACL** (Med×High) — TASK-014  
   Owner/`PermTeamManage` attaches any org team → bulk `developer` without consent.  
   `project/service.go` (no `requireTeamMember`) · `repository.go` bulk AddMember

2. **ConfirmUpload ghost / weak entity bind** (Med–High×Med) — TASK-045  
   Confirm without proven S3 object / weak message|channel ownership → phantom attachments.  
   `storage/service.go` ConfirmUpload

3. **GDPR Mongo eraser `(0,nil)` → HardDelete** (Low–Med×Critical) — TASK-079  
   Nil coll returns success → user HardDelete while messages remain.  
   `audit/gdpr_mongo.go:27-30` · `audit/service.go` cascade

4. **Push/OG SSRF + To4 residual** (Low×High) — TASK-019  
   User-controlled outbound URL; To4 mutation blind on Go 1.25.  
   `safedialer.go` · `notification/push.go` · `opengraph.go`

5. **AI privacy / nil-access fail-open** (Low–Med×High) — TASK-044  
   Nil access checker skips authz; Meili/stream hubtests missing.  
   `ai/issue_lookup.go` · `ai/code_review.go`

## Fix priority

`014` → `079` → `045` → `044`/`019` → `041`/`046`
