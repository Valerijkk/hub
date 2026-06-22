# Phase 110 — TMS + AI + Project Boundary Push

**Branch:** `feat/v2.5-trust`  
**Goal:** Push `tms`, `ai`, `project` toward 78%+ and lift COV toward 79.5%.

## Targets

| Package | Before | Target |
|---------|--------|--------|
| `tms` | 76.7% | ≥76.8% |
| `ai` | 77.0% | ≥78% |
| `project` | 77.6% | ≥78% |
| COV overall | 79.2% | ≥79.4% |

## Tasks

1. **AI** — OpenAI `ChatCompletion` success/error paths + admin handler validation (`adapter_wave110_test.go`, `admin_handler_wave110_test.go`)
2. **Project** — DeleteProject/Restore, invite/team error handlers (`handler_wave110_test.go`)
3. **TMS** — suite/component/run handler flows + invalid JSON (`handler_wave110_test.go`)

## Verification

```powershell
cd backend
go test -count=1 ./internal/tms/... ./internal/ai/... ./internal/project/...
go test -coverprofile=cov110_all ./internal/...
go tool cover -func=cov110_all | Select-String total
```
