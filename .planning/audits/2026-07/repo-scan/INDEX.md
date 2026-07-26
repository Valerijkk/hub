# Repo Scan 2026-07-23 — 40 report slots

> **Счётчик: 40 / 40** файлов на диске · **не** 40 независимых Grok-прогонов  
> **Provenance:** [`PROVENANCE.md`](./PROVENANCE.md) — [per-file registry](./PROVENANCE.md#per-file-registry-39-domain-reports): **19 PRE_EXISTING** + **20 FALLBACK_SUMMARY** + **1 SCORECARD (AGGREGATE)**  
> **Все W*:** `current_verified: false` · embedded PASS = historical self-check only  
> Папка: `.planning/audits/2026-07/repo-scan/`

## Происхождение (кратко)

| Компонент | Count | Class | Confidence | current_verified |
|-----------|-------|-------|------------|------------------|
| W*.md до fallback | 19 | PRE_EXISTING | MIXED | false (all) |
| W*.md fallback-сессия | 20 | FALLBACK_SUMMARY | PRELIMINARY | false (all) |
| SCORECARD | 1 | AGGREGATE | non-authoritative | false (rollup) |

Один **fallback generalPurpose**-агент (оркестратор) создал оставшиеся 20 отчётов и SCORECARD после того, как на диске уже было 19 `W*.md`. Канон per-file — [`PROVENANCE.md` § registry](./PROVENANCE.md#per-file-registry-39-domain-reports).

## Волны

| Wave | Status | On disk |
|------|--------|---------|
| 1 Backend core | ✅ 5/5 | auth, rbac-ws, tracker-sprint, messenger, security-crosscut |
| 2 Backend domains | ✅ 5/5 | tms, calls, notif-search, storage-apigateway, project-dashboard |
| 3 v4+platform+ai | ✅ 5/5 | v4-bases-crm-hr · platform · ai · intranet-calendar · audit-backup-imex |
| 4 FE core | ✅ 5/5 | tracker · messenger · tms · calls · auth-shell |
| 5 FE rest | ✅ 5/5 | v4 · admin · design-system · a11y · stores-api |
| 6 Tests | ✅ 5/5 | coverage · vitest · hubtest-core · hubtest-domains · e2e |
| 7 CI/deploy/docs | ✅ 5/5 | ci · deploy-compose · docs-drift · tz-grep · **fa06-secrets** |
| 8 AI ops + SCORECARD | ✅ 5/5 | gsd-ru · prompt-engineering · cursor-agents · ai-expansion · **SCORECARD** |

## Готовые (кратко — headlines; severity **не verified**)

| File | Headline | Class | Confidence |
|------|----------|-------|------------|
| W1-* | auth HIGH pending; rbac 3×HIGH IDOR; tracker JQL/IDOR; messenger 3 CRITICAL; security cross-cut PASS (self-check) | PRE_EXISTING | MIXED |
| W2-* | tms/project CRITICAL IDOR; calls CreateCall; notif PASS core; storage MEDIUM | PRE_EXISTING | MIXED |
| W3-* | crm CRITICAL (pre-existing); imex/calendar/platform/ai — fallback summaries | mixed | MIXED + PRELIMINARY |
| W4-fe-* | FE echo BE IDOR; calls Vitest=0; auth-shell pending | FALLBACK_SUMMARY | PRELIMINARY (echo) |
| W5-fe-* | design `!important` ui/=0 PASS (self-check); a11y NFR-16 open; stores 13 | FALLBACK_SUMMARY | PRELIMINARY |
| W6-* | 499 tests; waitForTimeout ai-hub; denial suites missing; Vitest TC gap | mixed | MIXED + PRELIMINARY |
| W7-* | tz-grep PASS (grep scope, self-check); **FA-06 Phase 0 only — prod blocker** | mixed | MIXED + PRELIMINARY |
| W8-* | gsd-ru PASS (self-check); PE link gap; agents 32/32; Must PASS **qualified** | mixed | MIXED + PRELIMINARY |
| **SCORECARD** | ≈20 unique CRITICAL (approx) · top-10 triage · **non-authoritative aggregate** | AGGREGATE | rollup only |

## Next

1. Читать [`PROVENANCE.md`](./PROVENANCE.md) (per-file registry) перед triage.
2. [`SCORECARD.md`](./SCORECARD.md) — approximate dedup only; **не** суммировать severity как verified totals.
3. Ops: FA-06 Phase 1–3 (не код).
4. be-волна: project → messenger search → tms → crm → calls → imex (после spot-verify fallback findings).
