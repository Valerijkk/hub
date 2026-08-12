# Stage C residual — 50 file-audit slots (2026-07-29)

> TASK-008 · READ-ONLY product · оркестратор Grok · параллельный запуск
> Выход каждого: `.planning/audits/total-2026-07/C50-<NN>-<slug>.md`
> Не коммитить. TASK-008 не закрывать.

| # | Slug | Scope |
|---|------|-------|
| 01 | be-auth | `backend/internal/auth/**/*.go` |
| 02 | be-rbac | `backend/internal/rbac/**/*.go` |
| 03 | be-audit | `backend/internal/audit/**/*.go` |
| 04 | be-ws | `backend/internal/ws/**/*.go` |
| 05 | be-search | `backend/internal/search/**/*.go` |
| 06 | be-ai-core | `backend/internal/ai/*handler*`, `service*`, `policy*` |
| 07 | be-ai-adapters | `backend/internal/ai/adapter*.go`, `http_stream*` |
| 08 | be-tracker-h | `backend/internal/tracker/handler*.go` |
| 09 | be-tracker-s | `backend/internal/tracker/service*.go` |
| 10 | be-tracker-r | `backend/internal/tracker/repo*.go`, `ent_repository*` |
| 11 | be-sprint | `backend/internal/sprint/**/*.go` |
| 12 | be-project | `backend/internal/project/**/*.go` |
| 13 | be-messenger-h | `backend/internal/messenger/handler*.go` |
| 14 | be-messenger-s | `backend/internal/messenger/service*.go`, `repo*` |
| 15 | be-calls | `backend/internal/calls/**/*.go` |
| 16 | be-tms-h | `backend/internal/tms/handler*.go` |
| 17 | be-tms-s | `backend/internal/tms/service*.go` |
| 18 | be-notification | `backend/internal/notification/**/*.go` |
| 19 | be-mail | `backend/internal/mail/**/*.go` |
| 20 | be-storage | `backend/internal/storage/**/*.go` |
| 21 | be-apigateway | `backend/internal/apigateway/**/*.go` |
| 22 | be-platform-http | `backend/internal/platform/http/**`, `middleware/**` |
| 23 | be-platform-nats | `backend/internal/platform/nats/**`, `outbox/**` |
| 24 | be-github | `backend/internal/github/**/*.go` |
| 25 | be-gitlab | `backend/internal/gitlab/**/*.go` |
| 26 | be-dashboard | `backend/internal/dashboard/**/*.go` |
| 27 | be-backup-imex | `backend/internal/backup/**`, `imex/**` |
| 28 | be-bases | `backend/internal/bases/**/*.go` |
| 29 | be-calendar | `backend/internal/calendar/**/*.go` |
| 30 | be-crm | `backend/internal/crm/**/*.go` |
| 31 | be-hr | `backend/internal/hr/**/*.go` |
| 32 | be-intranet | `backend/internal/intranet/**/*.go` |
| 33 | be-accounting | `backend/internal/accounting/**/*.go` |
| 34 | be-marketplace | `backend/internal/marketplace/**/*.go` |
| 35 | be-status-dev | `backend/internal/status/**`, `dev/**`, `integrations/**` |
| 36 | ent-a | `backend/ent/schema/` A–M |
| 37 | ent-b | `backend/ent/schema/` N–Z |
| 38 | fe-messenger | `frontend/src/modules/messenger/**` |
| 39 | fe-calls | `frontend/src/modules/calls/**` |
| 40 | fe-tracker | `frontend/src/modules/tracker/**` |
| 41 | fe-tms | `frontend/src/modules/tms/**` |
| 42 | fe-settings-admin | `frontend/src/modules/settings/**`, `admin/**` |
| 43 | fe-v4-people | `frontend/src/modules/{intranet,crm,hr}/**` |
| 44 | fe-v4-ops | `frontend/src/modules/{bases,calendar,marketplace,accounting}/**` |
| 45 | fe-api | `frontend/src/api/**` |
| 46 | fe-stores-hooks | `frontend/src/stores/**`, `hooks/**` |
| 47 | fe-components | `frontend/src/components/**` (не ui Radix behavior) |
| 48 | fe-routes | `frontend/src/routes/**` |
| 49 | infra-ci | `.github/workflows/**`, `deployments/**/*.yml` |
| 50 | docs-security-tz | `docs/security_guidelines.md`, `tz_constraints.md`, `architecture.md` (drift only) |
