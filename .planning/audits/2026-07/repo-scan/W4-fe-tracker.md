# W4-fe-tracker — Frontend tracker

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** fe-ver (readonly) · **Дата:** 2026-07-23
> **Scope:** `frontend/src/modules/tracker`, `api/tracker.ts`, routes issues · **Код не менялся**

## Summary

Tracker UI на TanStack Query + board-store; API типизирован. Security UI зависит от backend IDOR (W1-tracker HIGH). FE: нет `!important` в модуле; риск — оптимистичные мутации без жёсткой обработки 403/409.

**Вердикт:** `verified: false` (зависимость от BE IDOR + gaps UX ошибок).

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **HIGH** | BE `tracker` (W1) | FE показывает/мутирует issue по key в URL-проекте — при BE IDOR UI «успешен» на чужих | Сначала BE project-scope; FE: на 403 → toast + invalidate |
| **MEDIUM** | modules/tracker drag-and-drop | Move без явного version в UI → BE lost-update (W1 MoveIssue) | Передавать `version` в move API |
| **MEDIUM** | error UX | Не везде отличимы 403/404/409 conflict | Единый handler ApiError |
| **LOW** | JQL search UI | Обход service на BE — FE не виноват | Ждать BE refactor |
| **PASS** | `api/tracker.ts` | Типизированный client, не `any` | — |
| **INFO** | ErrorBoundary | Обязателен на tracker route | Проверить mount |
