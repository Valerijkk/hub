# W6-tests-hubtest-core — Backend hubtest core

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** qa-auto-ver (readonly) · **Дата:** 2026-07-23
> **Scope:** `backend/testframework/` inventory, tags, race · **Код не менялся**

## Summary

**499** `*_test.go` в `backend/testframework/` (AGENTS/COVERAGE ещё 495 — drift HIGH в W6-coverage). Теги `-tags hubtest` обязательны. Ядро env/factory/mongotest/platform/steps — есть.

**Вердикт:** `verified: false` (inventory drift + coverage gate).

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **HIGH** | COVERAGE/AGENTS vs disk | 495 vs **499** | Sync docs → 499 |
| **HIGH** | CI coverage-gate | 6×80% security не enforced (W6-coverage) | CI per-package |
| **PASS** | layout | Домены в testframework/, 0 internal tests | — |
| **PASS** | `-tags hubtest` | export_hooks под тегом | Документировать в TROUBLESHOOTING |
| **MEDIUM** | race timeout | 600s канон — flaky риск на CI minutes | Шаринг контейнеров |
| **INFO** | packages | ~35 domain dirs + factory/env | — |
