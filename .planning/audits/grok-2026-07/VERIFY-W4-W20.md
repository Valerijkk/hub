# VERIFY W4–W20 — post-closure agent recheck

**Branch:** `cursor-grok-4.5`  
**Date:** 2026-07-28  
**Method:** 3 fix waves × 5 agents → close WEAK/PARTIAL → 4 recheck waves × 5 agents  
**HEAD at recheck close:** includes `586d3f5` (A14-007 mutation-proof)

Related: [`PROGRAM-W4-W20-SUMMARY.md`](./PROGRAM-W4-W20-SUMMARY.md) · [`FIX-W20.md`](./FIX-W20.md)

---

## Executive verdict

| Recheck wave | Waves | Overall |
|--------------|-------|---------|
| **V1** | W4–W8 | **SOLID / PASS\*** |
| **V2** | W9–W13 | **PASS** (W10 incomplete by design; W11 product-scaffold WEAK honest) |
| **V3** | W14–W18 | **PASS** (A8 indexes still NEEDS_APPROVAL) |
| **V4** | W19–W20 + RT + scoreboard + gate | **PASS** (hubtest artifact **stale**) |

**No ALIVE RT chains.** RT-005 CLOSED. RT-003 ADR-blocked PARTIAL. Schema soft-delete/indexes not falsely CLOSED.

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
| `35b13c1` | RT-003 ADR-blocked honesty |
| `0eaea5f` | A13-006 batch ListSprints counts |
| `54e6cb4` | PROGRAM scoreboard update |

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
| **W9** | **PASS\*** | A3-005 sequential reject CLOSED; **TOCTOU race residual** (no tx/unique) — do not claim “race CLOSED” |
| **W10** | **INCOMPLETE** | A8-007/008/A7-003 FIXED; **A8-005 PARTIAL**; **A8-006 BLOCKED** — honest |
| **W11** | **WEAK/PARTIAL** | Dead FE clients CLOSED; A11-009–013 = API stubs (no UI) — honest FIXED-as-wrapper; BE `LoginWithTelegram*` DEFERRED |
| **W12** | **PASS** | A10-001 WS Vitest CLOSED (`575106d`) |
| **W13** | **PASS** | A10-008 FIXED via A13-014 (not DEFERRED); deferred UX debt remains |

---

## V3 — W14–W18 (5 agents)

| Wave | Overall | Notes |
|------|---------|-------|
| **W14** | **FIXED** | A13-006 N+1 Counts CLOSED (`0eaea5f`); A8 indexes out-of-scope NEEDS_APPROVAL |
| **W15** | **DONE\*** | Virt tests CLOSED; **A8-001/002 indexes still MISSING** |
| **W16** | **PASS** | A14-005 FIXED\* loopback (`c5613f6`); MANUAL ops asterisks hold |
| **W17** | **PASS** | A1-009/013/A14-006 CLOSED; A14-007 hardened (`586d3f5`) after soft first test |
| **W18** | **CLOSED** | A4-002/A4-006/StartSprint/RT-005 siblings CLOSED |

---

## V4 — W19–W20 + RT + gate (5 agents)

| Check | Overall | Notes |
|-------|---------|-------|
| **W19** | **CLOSED** | A6-006 / A7-007 / A7-008 hold |
| **W20 scoreboard** | **PASS\*** | 11 CLOSED / 1 PARTIAL / 0 ALIVE; fuzzy ~120+ close count admitted |
| **RT code** | **11 / 1 / 0** | RT-005 CLOSED; RT-003 ADR-blocked PARTIAL; RT-001–002,004,006–012 CLOSED |
| **Hubtest gate** | **WEAK (stale)** | `w20-gate-*.txt` @ `831da02`: 45 ok / 3 ?; mislabeled “42”; **26 commits** since (17 backend) — **not a gate for HEAD** |
| **Residual honesty** | **HIGH** | No false CLOSED on A8-006 / indexes / A8-005 / A14 MANUAL / Telegram BE / RT-003 |

---

## Still open (intentional — do not fake-close)

| ID | Status | Needs |
|----|--------|-------|
| **RT-003** | PARTIAL | Product ADR (CRM list scope / `PermCrm*`) |
| **A8-006** | BLOCKED | Ent `deleted_at` on Comment + approval |
| **A8-001/002 indexes** | NEEDS_APPROVAL | Atlas composite `(project, deleted_at)` |
| **A8-005** | PARTIAL | Channel soft-delete schema (cascade already done) |
| **A14-002/003/009** | FIXED\* | MANUAL rotate / NATS recreate / OTEL enable |
| **A3-005 race** | residual WEAK | Tx/lock or partial unique + concurrent test |
| **BE LoginWithTelegram\*** | DEFERRED | Service/DTO purge (routes already gone) |
| **A11 product UI** | residual | Wire CRM/bases/marketplace/admin wrappers in SPA |
| **Hubtest gate** | STALE | Re-run full `go test -tags hubtest ./testframework/...` on HEAD |

---

## RT scoreboard (post-fix)

| CLOSED | PARTIAL | ALIVE |
|-------:|--------:|------:|
| **11** | **1** (RT-003) | **0** |

---

## Recommended next (engineering)

1. Re-run hubtest gate on HEAD; replace `w20-gate-*.txt` with SHA-pinned summary (`ok/fail/?`, `-race`, exit).
2. ADR: CRM List/Pipeline scope (closes RT-003).
3. Schema approval: A8-006 + A8-001/002 indexes (+ optional A8-005 soft-delete).
4. Optional: A3-005 Exist→Create under tx; AddIssues issue-project FK filter; purge `LoginWithTelegram*`.

---

*Generated by 4×5 agent recheck after 3×5 fix waves. No master push/merge.*
