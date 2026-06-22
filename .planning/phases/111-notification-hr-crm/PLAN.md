# Phase 111 — Notification + HR + CRM Push

**Branch:** `feat/v2.5-trust`  
**Goal:** Triple push `notification (77.8%) + hr (77.9%) + crm (78%)` → COV **≥79.5%**.

## Targets

| Package | Before | Target |
|---------|--------|--------|
| `notification` | 77.8% | ≥78% |
| `hr` | 77.9% | ≥80% |
| `crm` | 78.0% | ≥80% |
| COV overall | 79.4% | ≥79.5% |

## Tasks

1. **HR** — RejectLeave/ApproveLeave version-zero + forbidden service paths; handler JSON/403/400 errors (`service_wave111_test.go`, `handler_wave111_test.go`)
2. **CRM** — UpdateContact linked user, validateDealRefs on deals, handler invalid contact/JSON (`service_wave111_test.go`, `handler_wave111_test.go`)
3. **Notification** — Dispatch multi-channel (telegram+push), publishToWS with project/actor, telegram status linked, push subscribe 503 (`service_wave111_test.go`)

## Verification

```powershell
cd backend
go test -count=1 ./internal/notification/... ./internal/hr/... ./internal/crm/...
go test -coverprofile=cov111_all ./internal/...
go tool cover -func=cov111_all | Select-String total
go test -race ./internal/notification/... ./internal/hr/... ./internal/crm/...
```
