# FIX-W8 — Block B reliability (Ack-on-error + delivery fail→success)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** be2

| ID | Status | SHA | Mutation note |
|----|--------|-----|---------------|
| A3-006 | CLOSED (W3) | `067a566` | `TestAutomationConsumer_NakOnBadJSON_A3_006` — always-Ack → `naks==0` FAIL (verified still green) |
| A1-008 | FIXED | `7cf6c7a` | `TestAuditConsumer_NakOnCreateError_A1_008` — Ack on CreateAuditLog err → FAIL |
| A2-010 | FIXED | `0253876` + WEAK `612cf22` | `TestNotificationConsumer_NakOnBadJSON_A2_010` — Ack on handler err → FAIL; **WEAK close 2026-07-28:** `TestNotificationConsumer_NakOnDispatchError_A2_010` — Nak when Dispatch fails |
| A1-007 | FIXED | `e8e3027` | `TestMagicLink_SendMagicLink_ResendErrorPropagates_A1_007` — `return code,nil` → FAIL |
| A7-004 | FIXED | `6f898c4` | `TestPerformBackup_PgDumpFailReturnsError_A7_004` — `return …, nil` → FAIL |

## Changes

- **audit consumer:** Ack/Nak inside `processEvent`; Nak on unmarshal / CreateAuditLog fail.
- **notification consumer:** handlers return `error`; Nak on fail, Ack on success.
- **auth magiclink:** Resend fail → wrapped error (HTTP 500 via handler).
- **backup:** dump/zip/S3 fail → `markFailed` + non-nil error (HTTP 500).
