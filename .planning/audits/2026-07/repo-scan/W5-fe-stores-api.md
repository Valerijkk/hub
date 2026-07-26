# W5-fe-stores-api — Zustand stores + API clients

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** fe-ver (readonly) · **Дата:** 2026-07-23  
> **Scope:** `frontend/src/stores/*` (13), `frontend/src/api/*` (35) · **Код не менялся**

## Summary

13 Zustand stores (доменные), 35 API-модулей, `client.ts` с single-flight refresh и `ApiError.request_id`. Паттерн один файл = один backend-модуль — PASS.

**Вердикт:** `verified: false` · **historical_self_check:** PASS (file scope) с LOW drift notes.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **PASS** | `api/client.ts` | JWT + refresh mutex + request_id | — |
| **PASS** | stores count | 13 доменных store | Не делать mega-store |
| **PASS** | api modules | 35 typed clients | — |
| **LOW** | docs «11 stores» | AGENTS/rules говорят 11, диск **13** | Docs sync |
| **MEDIUM** | ws-store + messenger | Дублирование connection state риск | Единый source of truth |
| **INFO** | timer/sidebar/command | UI chrome stores — OK | — |
| **LOW** | stale cache | Query+Zustand overlap | Invalidate on logout |
