# SCORECARD — Repo Scan 2026-07-23

> **Non-authoritative aggregate** · readonly · код не менялся в сессии fallback
> **Provenance:** [`PROVENANCE.md`](./PROVENANCE.md) — rollup **одной** fallback-сессии поверх 19 pre-existing W*; **не** 40 independent runs
> **Счётчик файлов: 40 / 40** (39 domain slots + этот SCORECARD) · все W*: `current_verified: false`

## ⚠️ Ограничения (обязательно прочитать)

| Claim | Status |
|-------|--------|
| «40× Grok 4.5 independent runs» | **FALSE** — 19 PRE_EXISTING + 20 FALLBACK + 1 AGGREGATE; см. PROVENANCE |
| Per-file `verified: true` | **FALSE** — все W* сейчас `current_verified: false`; PASS = historical_self_check |
| Totals ниже | **Approximate**, deduplicated вручную, **с эхом**; **не verified severity** |
| «Overall repo verified» | **FALSE** — `current_verified: false` для репо |
| FE/test CRITICAL | Часто **echo** BE — не independent findings |
| AI-expansion Must PASS | **Qualified** — structural/package PASS; runtime/eval gates remain; см. `07-ACCEPTANCE.md` |

## Missing list

| Expected slot | Status |
|---------------|--------|
| — | **Пусто** — все 39 domain-отчётов + SCORECARD на диске |

## Сводка CRITICAL / HIGH по файлам

> Severity из W*.md **без independent code recheck**. Использовать для **triage**, не для sign-off. Provenance/confidence — [`PROVENANCE.md` registry](./PROVENANCE.md#per-file-registry-39-domain-reports).

| File | CRITICAL | HIGH | Headline | Provenance | Confidence |
|------|----------|------|----------|------------|------------|
| W1-auth | 0 | 1 | pending→session | PRE_EXISTING | MIXED |
| W1-rbac-ws | 0 | 3 | custom role IDOR | PRE_EXISTING | MIXED |
| W1-tracker-sprint | 0 | 7 | JQL layering + issue/sprint IDOR + version | PRE_EXISTING | MIXED |
| W1-messenger | 3 | 7 | search NF-03; delete; polls | PRE_EXISTING | MIXED |
| W1-security-crosscut | 0 | 0 | PASS (historical self-check) | PRE_EXISTING | MIXED |
| W2-tms | 4 | 4 | case/run/suite/env/import IDOR | PRE_EXISTING | MIXED |
| W2-calls | 1 | 7 | CreateCall; chat/recording/scheduled | PRE_EXISTING | MIXED |
| W2-notification-search | 0 | 0 | MEDIUM only | PRE_EXISTING | MIXED |
| W2-storage-apigateway | 0 | 0 | MEDIUM only | PRE_EXISTING | MIXED |
| W2-project-dashboard | 6 | 4 | project+dashboard IDOR cluster | PRE_EXISTING | MIXED |
| W3-v4-bases-crm-hr | 4 | 3 | **crm** mutate IDOR | PRE_EXISTING | MIXED |
| W3-platform | 0 | 0 | PASS (historical self-check) | FALLBACK_SUMMARY | PRELIMINARY |
| W3-ai | 0 | 0 | SSRF PASS claim; layering MEDIUM | FALLBACK_SUMMARY | PRELIMINARY |
| W3-v4-intranet-calendar | 0 | 1 | leave org-wide | FALLBACK_SUMMARY | PRELIMINARY |
| W3-audit-backup-imex | 1 | 1 | **imex** auth-only | FALLBACK_SUMMARY | PRELIMINARY |
| W4-fe-tracker | 0 | 1* | *echo BE tracker | FALLBACK_SUMMARY | PRELIMINARY |
| W4-fe-messenger | 1* | 1* | *echo BE + vitest gap | FALLBACK_SUMMARY | PRELIMINARY |
| W4-fe-tms | 1* | 1* | *echo BE | FALLBACK_SUMMARY | PRELIMINARY |
| W4-fe-calls | 1* | 2* | *echo BE + no vitest | FALLBACK_SUMMARY | PRELIMINARY |
| W4-fe-auth-shell | 0 | 1* | *echo pending | FALLBACK_SUMMARY | PRELIMINARY |
| W5-fe-v4 | 1* | 1* | *echo crm + smoke | FALLBACK_SUMMARY | PRELIMINARY |
| W5-fe-admin | 0 | 0 | MEDIUM UX (historical self-check) | FALLBACK_SUMMARY | PRELIMINARY |
| W5-fe-design-system | 0 | 0 | `!important` ui/=0 PASS (self-check) | FALLBACK_SUMMARY | PRELIMINARY |
| W5-fe-a11y | 0 | 1 | NFR-16 full open | FALLBACK_SUMMARY | PRELIMINARY |
| W5-fe-stores-api | 0 | 0 | docs 11→13 (historical self-check) | FALLBACK_SUMMARY | PRELIMINARY |
| W6-tests-coverage | 0 | 4 | 6×80% CI; inventory; waitForTimeout | PRE_EXISTING | MIXED |
| W6-tests-vitest | 1 | 3 | TC↔Vitest; calls/messenger holes | PRE_EXISTING | MIXED |
| W6-tests-hubtest-core | 0 | 2 | inventory/CI echo | FALLBACK_SUMMARY | PRELIMINARY |
| W6-tests-hubtest-domains | 3* | 3* | *missing denial suites | FALLBACK_SUMMARY | PRELIMINARY |
| W6-tests-e2e | 0 | 1 | waitForTimeout ai-hub | FALLBACK_SUMMARY | PRELIMINARY |
| W7-ci | 0 | 0 | MEDIUM minutes | PRE_EXISTING | MIXED |
| W7-deploy-compose | 0 | 0 | (см. файл) | PRE_EXISTING | MIXED |
| W7-docs-drift | 0 | 0 | (см. файл) | PRE_EXISTING | MIXED |
| W7-tz-grep | 0 | 0 | PASS (grep scope, self-check) | PRE_EXISTING | MIXED |
| W7-fa06-secrets | 1 | 2 | **Phase 0 only — prod blocker** | FALLBACK_SUMMARY | PRELIMINARY |
| W8-gsd-ru | 0 | 0 | PASS (historical self-check) | PRE_EXISTING | MIXED |
| W8-pro-qa-ai | 0 | 1 | corporate-delivery link / structural gaps | PRE_EXISTING | MIXED |
| W8-cursor-agents | 0 | 0 | 32/32 GSD PASS (self-check) | FALLBACK_SUMMARY | PRELIMINARY |
| W8-ai-expansion | 0 | 2 | M6/M7 PARTIAL; acceptance qualified | FALLBACK_SUMMARY | PRELIMINARY |

\*FE/test «CRITICAL» часто **эхо** backend-находок или process-gap — **не** independent findings.

### Totals (approximate, deduplicated, non-authoritative)

| Severity | ~Count | Notes |
|----------|--------|-------|
| CRITICAL (unique product+ops, manual dedup) | **≈20** | messenger 3 + tms 4 + calls 1 + project 6 + crm 4 + imex 1 + FA-06 1 — **UNVERIFIED** |
| HIGH (unique, без полного FE-echo) | **≈45+** | approximate |
| Process CRITICAL (vitest TC matrix) | 1 | test process, not prod exploit |

**Не использовать** эти числа в KPI, COVERAGE или «verified» claims без spot-check в коде. **19 PRE_EXISTING + 20 FALLBACK** — не 39 независимых прогонов.

## Top-10 fix order (triage recommendation — UNVERIFIED)

| # | Priority | Finding | Primary report | Owner |
|---|----------|---------|----------------|-------|
| 1 | **P0 ops** | FA-06 Phase 1–3 secret rotation (Resend/TG) | W7-fa06-secrets | DevOps / user |
| 2 | **P0 sec** | Project API: RBAC + membership на CRUD/members/invites | W2-project-dashboard | be1/be2 |
| 3 | **P0 sec** | Messenger `SearchMessages` NF-03 org-wide | W1-messenger | be |
| 4 | **P0 sec** | TMS Get/Update/Delete case/run/suite/env project-scope | W2-tms | be |
| 5 | **P0 sec** | CRM Update/Delete contact+deal authorize | W3-v4-bases-crm-hr | be |
| 6 | **P0 sec** | Calls `CreateCall` membership | W2-calls | be |
| 7 | **P0 sec** | Messenger delete IDOR + GetPollResults | W1-messenger | be |
| 8 | **P0 sec** | Imex import/export — RequirePermission + membership | W3-audit-backup-imex | be |
| 9 | **P0 sec** | Dashboard `GetProjectStats` membership | W2-project-dashboard | be |
| 10 | **P1 sec** | Auth pending→JWT (ORG-06) + rbac role IDOR×3 | W1-auth · W1-rbac-ws | be |

**Сразу после P0:** denial-suites (W6-hubtest-domains), `waitForTimeout` remove
(W6-e2e), CI 6×80% (W6-coverage), FE pending gate, a11y NFR-16, PE paired
eval/runtime provenance gates (см. 07-ACCEPTANCE), M6/M7 polish.

## Вердикт скана (qualified)

| Gate | Status | Notes |
|------|--------|-------|
| TZ grep (W7) | historical_self_check PASS | grep scope only — not full repo; `current_verified: false` |
| Platform SSRF / AI SSRF | UNVERIFIED | FALLBACK_SUMMARY — recheck |
| Design `!important` in ui/ | historical_self_check PASS | FALLBACK file scope only |
| Agents GSD 32/32 | historical_self_check PASS | FALLBACK file scope only |
| AI-expansion Must | **QUALIFIED PASS** | structural/package PASS; paired eval NOT_RUN |
| Security IDOR cluster | **LIKELY FAIL** | UNVERIFIED until code recheck |
| FA-06 prod | **FAIL** open | ops |
| Overall merge-ready | **NO** | top-10 + recheck fallback findings |
| **Repo verified** | **NO** | `current_verified: false` on all W* |

`current_verified: false` на каждом W*-файле и для репозитория как целого — до закрытия P0 **и** выборочной reverification fallback-отчётов. Embedded PASS = historical self-check, не independent acceptance.
