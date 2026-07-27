# VERIFY W4–W20 — FINAL (post residual-close + 4×5 recheck)

**Branch:** `cursor-grok-4.5`  
**Date:** 2026-07-28  
**Status:** **FINAL** after residual-close wave + hubtest gate refresh  
**Method:** 3 fix waves × 5 agents → close WEAK/PARTIAL → 4 recheck waves × 5 agents → residual-close → hubtest re-run  
**Hubtest gate SHA:** [`1b83560`](https://github.com/) (`docs(audit): refresh hubtest gate artifact after residual close`)  
**Pre-artifact HEAD:** `3d313bf` (A10-007 notifications Query) · gate package run then artifact-committed as `1b83560`

Related: [`PROGRAM-W4-W20-SUMMARY.md`](./PROGRAM-W4-W20-SUMMARY.md) · [`FIX-W20.md`](./FIX-W20.md) · [`w20-gate-summary.txt`](./w20-gate-summary.txt)

---

## 1. Executive verdict — V1–V4

| Recheck wave | Waves covered | Overall |
|--------------|---------------|---------|
| **V1** | W4–W8 | **SOLID / PASS\*** |
| **V2** | W9–W13 | **PASS** (residuals closed in residual-close) |
| **V3** | W14–W18 | **PASS** (A8 indexes CLOSED `6031985`) |
| **V4** | W19–W20 + RT + scoreboard + gate | **PASS** |

| Dimension | Verdict |
|-----------|---------|
| **RT chains** | **12 CLOSED / 0 PARTIAL / 0 ALIVE** |
| **Schema soft-delete / indexes** | **CLOSED** (A8-005/006 + A8-001/002) — do not re-open as BLOCKED |
| **Hubtest gate** | **PASS** — **45 ok / 0 fail / 3 ?** @ commit **`1b83560`** (`-race`, exit 0) |
| **Zero debt claim** | **NO** — see §5 asterisks (ops MANUAL + polish leftovers) |

**No ALIVE RT. No PARTIAL RT.** Critical insider Broken Access Control surface from A19 is closed on this branch.

---

## 2. What residual-close closed (2026-07-28)

Schema / security / product gaps that recheck left honest-open, then closed in the residual-close wave:

| ID / theme | Close type | Evidence SHA(s) |
|------------|------------|-----------------|
| **A8-005** Channel soft-delete | Ent `deleted_at` + service filter + hubtest | `25028b5` / `293dee7` / `14c7929` |
| **A8-006** Comment soft-delete | Ent `deleted_at` + service filter + hubtest | `5c74170` / `9f4ce5b` / `c40c621` |
| **A8-001 / A8-002** indexes | `(project, deleted_at)` on Issue / TestCase | `6031985` (query rewrite earlier) |
| **A3-005** StartTimer race | Serialize open-timer create (tx / row-lock) | `bc06be9` / `54ac814` (+ prior `c502395`) |
| **RT-003** CRM List/Pipeline | Owner-scoped reads; admin/founder org-wide | `5e636f2` / `573c6fc` |
| **RT-005** AddIssues FK | Issues must belong to sprint project | `b4ab72d` (+ siblings `716c823`) |
| **BE Telegram login** | Dead `LoginWithTelegram*` service purge | `8ab8f76` |
| **A15-004** | REQUIREMENTS → Magic Link only (no TG login) | `d8e787b` |
| **A11-009…013** | SPA wire CRM / bases / marketplace / dashboard / admin | `68f9b0b` … `cbfae6c` |
| **A14-002 / A14-003** | Fail-closed SeaweedFS / NATS defaults in production | `f626397` |
| **A10-\*** class | notifications Query, Gantt virt, a11y, timer selectors, dead kanban | `3d313bf` / `108a6ca` / `a3e35cc` / `bbdb5fb` / `866fb7f` |
| **A2-007** flake | Await async Hub.Subscribe before broadcast | `5ffc26f` (unblocked hubtest green) |

Earlier WEAK→CLOSED evidence wave (pre residual-close): A2-007/010, A3-011, A1-009/013, A14-006/007, A4-002/006, A10-001, A13-014/018, A13-006, RT-005 siblings, W19 A6-006/A7-007/A7-008 — see [`PROGRAM-W4-W20-SUMMARY.md`](./PROGRAM-W4-W20-SUMMARY.md).

---

## 3. V1–V4 wave table (post residual-close)

### V1 — W4–W8

| Wave | Overall | Notes |
|------|---------|-------|
| **W4** | **SOLID** | A3-011 Postgres Ent twins CLOSED (`6342059`) |
| **W5** | **PASS\*** | Product IDOR CLOSED; honesty on mock vs prod SHAs |
| **W6** | **SOLID** | A2-007 Authorizer integration CLOSED (`cce9265`); flake fix `5ffc26f` |
| **W7** | **SOLID** | Founder / filters / dashboard / API key claims hold |
| **W8** | **CLOSED** | A2-010 Nak-on-Dispatch CLOSED (`612cf22`) |

### V2 — W9–W13

| Wave | Overall | Notes |
|------|---------|-------|
| **W9** | **PASS** | A3-005 sequential + serialize CLOSED |
| **W10** | **CLOSED** | A8-005/006 soft-delete CLOSED (\* trash basket nuance — §5) |
| **W11** | **CLOSED** | Dead FE + BE Telegram login purge; A11-009–013 SPA wired |
| **W12** | **PASS** | A10-001 WS→Query Vitest CLOSED (`575106d`) |
| **W13** | **PASS** | a11y / QueryState sample / TelegramButton absence |

### V3 — W14–W18

| Wave | Overall | Notes |
|------|---------|-------|
| **W14** | **FIXED** | A13-006 N+1 Counts CLOSED (`0eaea5f`) |
| **W15** | **CLOSED** | Virt thresholds + A8-001/002 indexes CLOSED |
| **W16** | **PASS** | A14-005 FIXED\*; **A14-009 OTEL still MANUAL** |
| **W17** | **PASS** | A1-009/013/A14-006 CLOSED; A14-007 mutation-proof (`586d3f5`) |
| **W18** | **CLOSED** | A4-002/A4-006 + StartSprint / RT-005 siblings CLOSED |

### V4 — W19–W20 + RT + gate

| Check | Overall | Notes |
|-------|---------|-------|
| **W19** | **CLOSED** | A6-006 / A7-007 / A7-008 hold |
| **W20 scoreboard** | **PASS** | **12 / 0 / 0** RT |
| **RT code** | **12 / 0 / 0** | RT-003 + RT-005 CLOSED; RT-001–002,004,006–012 CLOSED |
| **Hubtest gate** | **PASS** | **45 ok / 0 fail / 3 ?** @ **`1b83560`** |
| **Residual honesty** | **HIGH** | Not claiming zero debt — §5 |

---

## 4. True remaining (intentional — do not fake-close)

| ID / item | Status | Needs |
|-----------|--------|-------|
| **A14-009 OTEL** | FIXED\* / **MANUAL** | Set `OTEL_EXPORTER_OTLP_ENDPOINT` to enable spans (`68cbd14` wired; no-op without endpoint) |
| **PG-03 / FA-06** | **MANUAL** | Resend/Telegram secret rotation — **sole prod cutover blocker** |
| **A14-002 / A14-003** | FIXED\* / **MANUAL** | Rotate Seaweed S3 / set NATS creds + recreate if old defaults ever deployed (`f626397`) |
| **Trash basket 30d (Channel / Comment)** | **PARTIAL product\*** | Soft-delete + list filters CLOSED; full TZ **корзина 30 дней** (admin trash UI / restore / hard-purge job) for Channel/Comment **may still be incomplete** — orphans kept for future purge (`FIX-W10`) |
| **Orphan TG bot cemetery** | **Hygiene debt\*** | Login routes/UI/service purged; leftover legacy surfaces (`auth/telegram.go` VerifyTelegramAuth, `auth/telegram_bot.go` polling-disabled helpers, test export hooks) — not login paths, but dead cemetery until cleanup wave |

Closed in residual-close (removed from open): RT-003, A8-005/006 schema soft-delete, A8-001/002 indexes, A3-005 race, BE `LoginWithTelegram*`, A11-009…013 SPA wire, A15-004, AddIssues FK, stale-hubtest claim (refreshed).

---

## 5. Honest asterisks — not “100% zero debt”

Do **not** market this branch as debt-free. Residual-close + hubtest green close the **security/RT/schema P0–P1 program**; the following asterisks remain:

1. **Ops MANUAL:** A14-009 OTEL enable; PG-03/FA-06 secret rotation; A14-002/003 rotate-if-deployed.
2. **FIXED\* config:** A14-005 metrics CIDR — compose tightened; blanket RFC1918 in `.env` can re-open Docker peers.
3. **Trash / restore product gap:** Channel/Comment soft-delete ≠ full 30-day trash basket UX + scheduled hard-delete (TZ complete story).
4. **TG bot cemetery:** disabled/legacy auth-bot helpers still in tree after Magic-Link-only purge.
5. **Coverage floor:** COV **81.9%** (2026-07-27) — COV-01 ≥82% **MISS**; security 6-pack not remeasured 2026-07-27/28 — **no invented %**.
6. **Mild docs drift:** AGENTS / older STATE tables still quote **83.0%** baseline alongside gate **81.9%**.
7. **Block F / P2–P3 leftovers:** A4-005+, A3-013+, A5-012+, A7-010+, A14-008, deeper A10 InfiniteQuery polish — time-boxed out.
8. **Hubtest `?` packages:** 3 packages reported `?` (no test files / build skip pattern) — not failures; still listed for honesty.
9. **Uncommitted WIP:** working tree may hold unrelated FE API edits / planning-audit scripts — **not** part of this VERIFY commit.

**Post-FIX residual exploitability (honest):** ~**1.5/10** ops/schema-product nuance — **not** full named RT chains.

---

## 6. Hubtest gate (canonical)

| Field | Value |
|-------|-------|
| Artifact commit | **`1b83560`** |
| Pre-artifact HEAD | `3d313bf` |
| Command | `go test -tags hubtest -race -count=1 -timeout=900s ./testframework/...` |
| **ok** | **45** |
| **fail** | **0** |
| **?** | **3** |
| Exit | **0** |
| Logs | [`w20-gate-summary.txt`](./w20-gate-summary.txt) · [`w20-gate-hubtest.txt`](./w20-gate-hubtest.txt) |

Note: first full run after residual-close failed `TestReadPump_CallAuthorizer_ParticipantAllowed_A2_007` (async Subscribe race); fixed in `5ffc26f`, then gate green.

---

## 7. RT scoreboard (FINAL)

| CLOSED | PARTIAL | ALIVE |
|-------:|--------:|------:|
| **12** | **0** | **0** |

Full chain table → [`FIX-W20.md` §1](./FIX-W20.md).

---

## 8. Recommended next (human / ops)

1. Ops: **FA-06** secret rotation; enable **A14-009** OTEL where tracing is desired; rotate A14-002/003 if historical defaults were ever live.
2. Optional product: Channel/Comment → trash basket restore + 30d hard-purge; delete TG bot cemetery files/hooks.
3. Optional quality: coverage uplift remeasure to restore ≥82% (do not invent %).
4. User merge of `cursor-grok-4.5` when ready — **agents do not push/merge `master`**.

---

## 9. Key SHA index (residual-close + gate)

```
25028b5 / 293dee7 / 14c7929   A8-005 Channel soft-delete
5c74170 / 9f4ce5b / c40c621   A8-006 Comment soft-delete
6031985                       A8-001/002 indexes
bc06be9 / 54ac814             A3-005 serialize
5e636f2 / 573c6fc             RT-003 CRM list scope
b4ab72d                       AddIssues FK
8ab8f76 / d8e787b             Telegram login purge + REQUIREMENTS
68f9b0b … cbfae6c             A11-009…013 SPA wire
f626397                       A14-002/003 fail-closed
5ffc26f                       A2-007 hubtest flake
1b83560                       hubtest gate artifact 45/0/3
5a3d469                       residual-close scoreboard docs
```

---

*FINAL VERIFY W4–W20 · 2026-07-28 · residual-close closed · hubtest 45/0/3 @ `1b83560` · RT 12/0/0 · not zero-debt. No master push/merge.*
