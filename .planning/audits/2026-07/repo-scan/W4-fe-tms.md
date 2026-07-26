# W4-fe-tms — Frontend TMS

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** fe-ver (readonly) · **Дата:** 2026-07-23  
> **Scope:** `modules/tms`, `stores/tms-store.ts`, `api/tms.ts` · **Код не менялся**

## Summary

TMS UI опирается на project routes. Backend CRITICAL IDOR по caseKey/runId (W2-tms) — FE не валидирует принадлежность ресурса к `projectId` из URL (и не должен вместо BE, но должен корректно обрабатывать 403).

**Вердикт:** `verified: false`.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **CRITICAL** | BE tms (W2) | UI Get/Update по key/id при чужом project в URL «работает» | BE authorize*; FE redirect 403 |
| **HIGH** | optimistic version | UI может слать version=0 / игнор 409 | Required version + conflict toast |
| **MEDIUM** | Allure import UI | Импорт в run без client check project | Disable если run.project ≠ route |
| **PASS** | `api/tms.ts` | Typed client | — |
| **INFO** | E2E `tms.spec.ts` | Smoke есть; IDOR E2E нет | Добавить denial e2e после BE fix |
