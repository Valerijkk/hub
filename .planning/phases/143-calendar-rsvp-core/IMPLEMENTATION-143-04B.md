# Implementation 143-04B — create events transactionally

> **Status:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS`
> **Source commit:** `75dafea889055126f768227a4ae62b28c91b7393`
> **Parent:** `a877848ef4d1ce80a30a4389f97ef0f863615159`
> **Tree:** `6b86570991e41e0c401934867a49a17849847cb3`
> **Subject:** `feat(calendar): create events transactionally`

| Path | Blob |
|---|---|
| `backend/internal/calendar/event_service_create.go` | `db34bbbc4b4c8a3ace8f768998060f82deecedbe` |
| `backend/tests/calendar/event_service_create_test.go` | `79225fb73481ffd34ee1f3b9a72965d255e0b14a` |

Tests first; Contract, Security and Lifecycle immutable exact-SHA reviews PASS C/I/M `0/0/0`; `dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Old 04B `PRECOMMIT_FAIL`, `NOT_ACCEPTED`, and current-next routes are `CONSUMED/HISTORICAL`. Static only: launcher `NOT_READY`, executable/generation/migration/runtime `NOT_RUN`; 03E remains blocked for Ent generation/migration. No next Calendar code packet or broader authority follows.
