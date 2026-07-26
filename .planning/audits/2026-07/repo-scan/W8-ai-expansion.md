# W8-ai-expansion — AI-expansion acceptance

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** product-ver (readonly) · **Дата:** 2026-07-23  
> **Scope:** `.planning/july-2026/ai-expansion/07-ACCEPTANCE.md` · **Код не менялся**

## Summary

**Must (G1–G6) = PASS.** Maximal expand: M1–M5 PASS; **M6 PARTIAL**, **M7 open/PARTIAL**. Не блокирует Must; polish → 15.08.

**Вердикт:** Must `verified: false` · **historical_self_check:** PASS (Must structural scope) · maximal `verified: false` (M6/M7).

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **PASS** | `07-ACCEPTANCE.md` G1–G6 | Must закрыты | — |
| **PASS** | M1–M5 | prompts 112, PE agents, dry-run, P0/P1 status, gsd-ru core | — |
| **HIGH** | M6 | 15/21 P0 только в agents (не gsd-ru skill) — PARTIAL | Добить паттерны по GSD-INTEGRATION |
| **HIGH** | M7 | P1 routes 35 skills — predominantly route-only / open | W2–W4 polish |
| **MEDIUM** | W8-prompt HIGH | Битая ссылка `corporate-delivery.md` (соседний отчёт) | Создать playbook или убрать ссылки |
| **INFO** | Models cutover W4 | Open, не блокер Must | Day-X checklist |
| **PASS** | Prod code | AI-expansion wave не трогала backend/frontend product | — |

## Matrix

| Tier | Status |
|------|--------|
| Must G1–G6 | **PASS** |
| M6 | **PARTIAL** |
| M7 | **PARTIAL/open** |
