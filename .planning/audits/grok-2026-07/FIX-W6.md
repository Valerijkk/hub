# FIX-W6 — Calls/push IDOR A2-002…A2-007

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Предшественник:** FIX-W5 (`2fccf15`)

Docker: postgres/mongo/redis healthy; `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` set.

## Цель
CreateCall / chat / recording / RaiseHand без membership; push hijack; instant-call WS deny.

## Baseline / after
| | ok | fail | `?` |
|--|----|------|-----|
| **ДО** | **42** | **0** | **3** (= W5 after) |
| **ПОСЛЕ** | _после прогона_ | | |

## Закрыто

| ID | SHA | Фикс |
|----|-----|------|
| A2-002 | `c4b1dc2` | CreateCall → IsChannelMember |
| A2-003 | `591fa0d` | chat send/list → participant/channel |
| A2-004 | `ef72607` | recording start/stop → membership |
| A2-005 | `a90fb5a` | push Update/Delete Endpoint+UserID |
| A2-006 | `cb80edd` | RaiseHand → active participant |
| A2-007 | `2b78dd3` + WEAK `cce9265` | instant WS → IsCallParticipant; **WEAK close 2026-07-28:** `ws_authorizer_integration_a2007_test.go` (prod Authorizer + stranger deny) |

Каждый: regression + mutation (снять guard → red → restore → green).

*Next: W7 authz remainder.*
