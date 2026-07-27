# VERIFY W4–W20 — post-closure agent recheck

**Branch:** `cursor-grok-4.5`  
**Date:** 2026-07-28  
**Method:** 3 fix waves × 5 agents → close WEAK/PARTIAL → 4 recheck waves × 5 agents → residual-close wave  
**HEAD at residual-close scoreboard:** includes A11 SPA wire (`cbfae6c`) + A14 fail-closed (`f626397`) + RT-003 (`5e636f2`)

Related: [`PROGRAM-W4-W20-SUMMARY.md`](./PROGRAM-W4-W20-SUMMARY.md) · [`FIX-W20.md`](./FIX-W20.md)

---

## Executive verdict

| Recheck wave | Waves | Overall |
|--------------|-------|---------|
| **V1** | W4–W8 | **SOLID / PASS\*** |
| **V2** | W9–W13 | **PASS** (residuals later closed in residual-close wave) |
| **V3** | W14–W18 | **PASS** (A8 indexes closed `6031985`) |
| **V4** | W19–W20 + RT + scoreboard + gate | **PASS** (hubtest artifact **stale**; RT **12 / 0 / 0**) |

**No ALIVE RT chains. No PARTIAL RT.** RT-003 + RT-005 CLOSED. Schema soft-delete/indexes CLOSED in residual-close wave (do not re-open as BLOCKED).

---

## Fix waves executed (before recheck)

### Fix Wave 1
| SHA | Close |
|-----|--------|
| `c502395` | A3-005 StartTimer reject open |
| `8e0c9f2` | StartSprint → GetSprintInProject |
| `cce9265` / `612cf22` | A2-007 Authorizer + A2-010 Nak Dispatch |
| `6342059` | A3-011 Postgres Ent twins |
| `957472e` / `9764850` | A1-009/013 HTTP |

### Fix Wave 2
| SHA | Close |
|-----|--------|
| `0cbc7d2` / `4ebd8c3` / `586d3f5` | A14-006 OTP slog + A14-007 wiring |
| `c3bc077` / `1b19d13` | A4-002 runtime + ConfirmUpload 409 |
| `575106d` | A10-001 WS→setQueryData Vitest |
| `3536e0e` / `53d72c8` | A13-014/018 virt thresholds |
| `c5613f6` / `76080f8` | A14-005 loopback CIDR + honesty docs |

### Fix Wave 3
| SHA | Close |
|-----|--------|
| `716c823` | RT-005 Get/Update/Retro/Burndown/AddIssues |
| `cfc4959` / `fbc20a1` / `ab580a3` | W19 A6-006 / A7-007 / A7-008 |
| `35b13c1` | RT-003 ADR-blocked honesty (superseded — RT-003 CLOSED) |
| `0eaea5f` | A13-006 batch ListSprints counts |
| `54e6cb4` | PROGRAM scoreboard update |

### Residual-close wave (2026-07-28)
| SHA | Close |
|-----|--------|
| `25028b5` / `293dee7` / `14c7929` | A8-005 Channel soft-delete |
| `5c74170` / `9f4ce5b` / `c40c621` | A8-006 Comment soft-delete |
| `6031985` | A8-001/002 project+deleted_at indexes |
| `bc06be9` / `54ac814` | A3-005 StartTimer serialize (TOCTOU) |
| `5e636f2` / `573c6fc` | RT-003 List/Pipeline owner scope |
| `b4ab72d` | AddIssues issue↔project FK filter |
| `8ab8f76` | BE Telegram login surface purge |
| `d8e787b` | A15-004 REQUIREMENTS Magic Link only |
| `68f9b0b`…`cbfae6c` | A11-009…013 SPA wire |
| A10-* / A14-002/003 | UX/a11y + fail-closed Seaweed/NATS (`f626397`) |

---

## V1 — W4–W8 (5 agents)

| Wave | Overall | Notes |
|------|---------|-------|
| **W4** | **SOLID** | A3-011 Postgres twins CLOSED (`6342059`) |
| **W5** | **PASS\*** | Product IDOR CLOSED; honesty fixed (`7937060`=mock; prod `2ffd60f`); residual under-disclose `c2dd881` handler smuggle |
| **W6** | **SOLID** | A2-007 Authorizer integration CLOSED (`cce9265`) |
| **W7** | **SOLID** | Six claims hold; `w67-after-hubtest.txt` still MISSING (artifact only) |
| **W8** | **CLOSED** | A2-010 Nak-on-Dispatch CLOSED (`612cf22`) |

---

## V2 — W9–W13 (5 agents)

| Wave | Overall | Notes |
|------|---------|-------|
| **W9** | **PASS** | A3-005 sequential + serialize CLOSED (`c502395` / `bc06be9`) |
| **W10** | **CLOSED** | A8-005/006 soft-delete CLOSED (residual-close) |
| **W11** | **CLOSED** | Dead FE + BE Telegram purge; A11-009–013 SPA wired |
| **W12** | **PASS** | A10-001 WS Vitest CLOSED (`575106d`) |
| **W13** | **PASS** | A10-008 FIXED via A13-014; a11y/QueryState sample closed |

---

## V3 — W14–W18 (5 agents)

| Wave | Overall | Notes |
|------|---------|-------|
| **W14** | **FIXED** | A13-006 N+1 Counts CLOSED (`0eaea5f`) |
| **W15** | **CLOSED** | Virt tests + A8-001/002 indexes CLOSED (`6031985`) |
| **W16** | **PASS** | A14-005 FIXED\* loopback (`c5613f6`); **A14-009 OTEL still MANUAL** |
| **W17** | **PASS** | A1-009/013/A14-006 CLOSED; A14-007 hardened (`586d3f5`) |
| **W18** | **CLOSED** | A4-002/A4-006/StartSprint/RT-005 siblings CLOSED |

---

## V4 — W19–W20 + RT + gate (5 agents)

| Check | Overall | Notes |
|-------|---------|-------|
| **W19** | **CLOSED** | A6-006 / A7-007 / A7-008 hold |
| **W20 scoreboard** | **PASS** | **12 CLOSED / 0 PARTIAL / 0 ALIVE** |
| **RT code** | **12 / 0 / 0** | RT-003 CLOSED (`5e636f2`); RT-005 CLOSED; RT-001–002,004,006–012 CLOSED |
| **Hubtest gate** | **WEAK (stale)** | `w20-gate-*.txt` @ `831da02` — not a gate for HEAD; re-run pending |
| **Residual honesty** | **HIGH** | Schema/Telegram/RT-003 not falsely left open; only ops MANUAL residuals |

---

## Still open (intentional — do not fake-close)

| ID | Status | Needs |
|----|--------|-------|
| **A14-009** | FIXED\* / **MANUAL** | Set `OTEL_EXPORTER_OTLP_ENDPOINT` in env to enable spans (`68cbd14` wired; no-op without endpoint) |
| **A14-002 / A14-003** | FIXED\* / **MANUAL** | Rotate Seaweed S3 if defaults ever deployed; set `NATS_USER`/`NATS_PASSWORD` + recreate on existing prod (`f626397` fail-closed defaults) |
| **PG-03 / FA-06** | MANUAL | Resend/Telegram secret rotation — sole prod cutover blocker |
| **Hubtest gate** | STALE | Re-run full `go test -tags hubtest ./testframework/...` on HEAD; replace `w20-gate-*.txt` |

Closed in residual-close (removed from open): RT-003, A8-005/006, A8-001/002 indexes, A3-005 race, BE `LoginWithTelegram*`, A11-009…013 SPA wire, A15-004, AddIssues FK (`b4ab72d`).

---

## RT scoreboard (post residual-close)

| CLOSED | PARTIAL | ALIVE |
|-------:|--------:|------:|
| **12** | **0** | **0** |

---

## Recommended next (engineering)

1. Re-run hubtest gate on HEAD; replace `w20-gate-*.txt` with SHA-pinned summary (`ok/fail/?`, `-race`, exit).
2. Ops: enable OTEL endpoint where tracing is desired (A14-009); FA-06 secret rotation.
3. Optional: coverage uplift to restore ≥82% (do not invent % — remeasure only).

---

*Updated 2026-07-28 residual-close scoreboard. No master push/merge.*
