# Grok FIX Program — W4…W20 summary

**Branch:** `cursor-grok-4.5`  
**Finale:** 2026-07-27 · [`FIX-W20.md`](./FIX-W20.md)  
**Remediation recheck:** 2026-07-28 (WEAK / evidence closure wave)  
**Residual-close scoreboard:** 2026-07-28 (A8 schema / RT-003 / A11 SPA / Telegram purge)  
**Audit root:** `.planning/audits/grok-2026-07/` (zones `00`–`19` + FIX-W*)

---

## Remediation recheck (2026-07-28)

Post-W20 wave closed **WEAK** evidence gaps (mutation-resistant / HTTP / Postgres / Vitest threshold tests) and one residual product fix (`StartSprint` project-scope). **No new coverage % claimed** — gate remains **81.9%** (2026-07-27).

### Newly closed / WEAK→CLOSED

| ID | Wave home | Close type | Evidence SHA(s) |
|----|-----------|------------|-----------------|
| A3-005 | W9 | Fix + test — reject `StartTimer` if open timer | `c502395` |
| A3-005 race | W9 | Serialize open-timer create (tx / row-lock) | `bc06be9` / `54ac814` |
| A3-016 / RT-005 sibling | W18 | Fix — `StartSprint` → `GetSprintInProject` | `8e0c9f2` |
| A2-007 | W6 | WEAK — Authorizer deny integration | `cce9265` |
| A2-010 | W8 | WEAK — Nak on Dispatch failure mutation | `612cf22` |
| A3-011-link/react/target | W4 | WEAK — Postgres Ent twins (not sqlite) | `6342059` |
| A1-009 | W17 | WEAK — HTTP 403 founder ban/deactivate | `957472e` |
| A1-013 | W17 | WEAK — HTTP 400 retention allowlist | `9764850` |
| A14-006 | W17 | WEAK — OTP not logged (slog spy) | `0cbc7d2` |
| A14-007 | W17 | WEAK — `EnableStackTrace`↔DevMode (mutation-proof) | `4ebd8c3` / `586d3f5` |
| A13-006 | W14 | residual N+1 Counts → batch `GetSprintIssueCounts` | `0eaea5f` |
| RT-005 siblings | W20 | Get/Update/Retro/Burndown/AddIssues scoped | `716c823` |
| AddIssues FK | W20 | Issues must belong to sprint project | `b4ab72d` |
| A4-002 | W18 | WEAK — runtime soft-delete exclusion | `c3bc077` |
| A4-006 | W18 | WEAK — ConfirmUpload HTTP 409 | `1b19d13` |
| A10-001 | W12 | WEAK — WS→Query handler path Vitest | `575106d` |
| A13-014 | W15 | WEAK — board column virtualizes at threshold | `3536e0e` |
| A13-018 | W15 | WEAK — channel list virtualizes at threshold | `53d72c8` |
| A14-005 | W16 | Tighten metrics CIDR default (FIXED*) | `c5613f6` |
| **A8-005** | W10 | Channel soft-delete schema + service | `25028b5` / `293dee7` / `14c7929` |
| **A8-006** | W10 | Comment soft-delete schema + service | `5c74170` / `9f4ce5b` / `c40c621` |
| **A8-001/002 indexes** | W15 | `(project, deleted_at)` on Issue/TestCase | `6031985` |
| **RT-003** | W20 | List/Pipeline owner-scoped; admin org-wide | `5e636f2` / `573c6fc` |
| BE Telegram login | W11 | Dead `LoginWithTelegram*` service purge | `8ab8f76` |
| A15-004 | docs | REQUIREMENTS Magic Link only (no TG login) | `d8e787b` |
| A11-009…013 | W11 | SPA wire CRM/bases/marketplace/dashboard/admin | `68f9b0b`…`cbfae6c` |

Honesty pass (no fake CLOSED): [`76080f8`](./FIX-W5.md) — W5/W10/W13/W16 residuals (later superseded by residual-close where code landed).

### Schema blocked — CLEARED

| ID | Status | Notes |
|----|--------|-------|
| **A8-006** | **CLOSED** | Comment `deleted_at` + soft-delete (`5c74170` / `9f4ce5b` / `c40c621`) |
| **A8-001** indexes | **CLOSED** | Query rewrite + composite index (`0ce8cd7` / `6031985`) |
| **A8-002** indexes | **CLOSED** | Query rewrite + composite index (`443a569` / `6031985`) |
| **A8-005** | **CLOSED** | Channel soft-delete (`25028b5` / `293dee7` / `14c7929`) |

### RT residual (honest)

- **RT-003** **CLOSED** — writes + List/Pipeline owner scope; admin/founder org-wide (`5e636f2` / `573c6fc`).
- **RT-005** **CLOSED** — Complete/Start/Get/Update/Retro/Burndown/AddIssues all `GetSprintInProject`; AddIssues FK (`1ac8669` / `8e0c9f2` / `716c823` / `b4ab72d`).
- **Scoreboard:** **12 CLOSED / 0 PARTIAL / 0 ALIVE**.

### True residuals (ops / deferred only)

| ID | Status | Notes |
|----|--------|-------|
| **A14-009** | FIXED\* / **MANUAL** | OTEL wired (`68cbd14`); enable endpoint in env to get spans |
| **A14-002 / A14-003** | FIXED\* / **MANUAL** | Fail-closed defaults (`f626397`); rotate/recreate if old secrets deployed |
| **PG-03 / FA-06** | MANUAL | Sole prod cutover blocker |
| **Hubtest gate artifact** | STALE | Re-run on HEAD |

---

## Wave status

| Wave | Focus | Status | Report |
|------|-------|--------|--------|
| W0 | Ent generate / A12-001 local panic | **DONE** | [FIX-W0.md](./FIX-W0.md) |
| W1 | Security P0 (pending JWT, messenger search, RBAC roles, join, CSV) | **DONE** | [FIX-W1.md](./FIX-W1.md) |
| W2 | Remaining P0 (sprint complete, CRM write, invoice tx, TMS FormData, GDPR) | **DONE** | [FIX-W2.md](./FIX-W2.md) |
| W3 | Tracker/TMS P1 IDOR + locks + consumers | **DONE** | [FIX-W3.md](./FIX-W3.md) |
| W4 | Tracker/TMS residual IDOR / links / Ent gaps | **DONE** | [FIX-W4.md](./FIX-W4.md) |
| W5 | GitHub/GitLab/imex IDOR + PDF scope | **DONE** | [FIX-W5.md](./FIX-W5.md) |
| W6 | Calls + push IDOR (A2-002…007) | **DONE** | [FIX-W6.md](./FIX-W6.md) |
| W7 | Founder admin, saved filters, dashboard stats, API key auth | **DONE** | [FIX-W7.md](./FIX-W7.md) |
| W8 | JetStream Nak / Resend fail-closed / backup error | **DONE** | [FIX-W8.md](./FIX-W8.md) |
| W9 | Optimistic lock / coffee / invoice dup | **DONE** | [FIX-W9.md](./FIX-W9.md) |
| W10 | Channel tx delete, Mongo disconnect, Listen cancel, restore maintenance | **DONE** | [FIX-W10.md](./FIX-W10.md) |
| W11 | FE↔BE contract dead clients | **DONE** | [FIX-W11.md](./FIX-W11.md) |
| W12 | FE Query/WS/cursor/presence | **DONE** | [FIX-W12.md](./FIX-W12.md) |
| W13 | FE a11y / TelegramButton / QueryStateView sample | **DONE** | [FIX-W13.md](./FIX-W13.md) |
| W14 | Perf caps / aggregates / N+1 (A13-001…006) | **DONE** | [FIX-W14.md](./FIX-W14.md) |
| W15 | Lazy/virtual/waterfalls + A8 query rewrite + indexes | **DONE** | [FIX-W15.md](./FIX-W15.md) |
| W16 | Obs/health/secrets/metrics/OTEL (A14-*) | **DONE** (A14-009 MANUAL) | [FIX-W16.md](./FIX-W16.md) |
| W17 | Block F sample — auth/rbac/audit/platform/messenger/ws | **DONE** | [FIX-W17.md](./FIX-W17.md) |
| W18 | Block F sample — tms/tracker/sprint/calls/search/storage/ai | **DONE** | [FIX-W18.md](./FIX-W18.md) |
| W19 | Block F sample — crm/hr/calendar/imex/mail/dashboard/backup… | **DONE** | [FIX-W19.md](./FIX-W19.md) |
| **W20** | **Finale: RT re-verify + planning/COV hygiene + scoreboard** | **DONE** (docs) | [FIX-W20.md](./FIX-W20.md) |

W4–W20 = remediation program after readonly zones `01`–`19`. Earlier W0–W3 included for continuity.

---

## RT outcome (W20 + residual-close 2026-07-28)

| CLOSED | PARTIAL | ALIVE |
|-------:|--------:|------:|
| **12** | **0** | **0** |

Full table → [FIX-W20.md §1](./FIX-W20.md).

---

## Headline metrics

| Metric | Value |
|--------|-------|
| Unique findings (`00-SUMMARY`) | **262** |
| Approx closed W0–W19 + residual-close | **~120+** |
| Coverage (2026-07-27) | **81.9%** (−1.1 pp vs 83.0) |
| COV-01 ≥82% | **MISS** |
| Prod blocker | **PG-03/FA-06** (manual) |
| Schema blocked | **none** (A8-005/006 + A8-001/002 indexes CLOSED) |
| RT scoreboard | **12 / 0 / 0** |

---

## Stop / handoff

- Feature code freeze for July Rebuild planning remains product context; this branch holds security/quality fixes pending user merge.
- Do **not** push/merge `master` from agents.
- Next engineering priority: hubtest re-run on HEAD → A14-009 OTEL enable (ops) → FA-06 secret rotation → optional COV uplift (remeasure only).
