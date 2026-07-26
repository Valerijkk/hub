# W5-fe-design-system — Design system / Shadcn

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** designer-ver (readonly) · **Дата:** 2026-07-23  
> **Scope:** `frontend/src/components/ui/`, `--hub-*`, `!important` · **Код не менялся**

## Summary

**`!important` в `components/ui/` = 0** — PASS (Phase 8). Единственные `!important` — `app.css` prefers-reduced-motion (a11y). Shadcn: className-only канон.

**Вердикт:** `verified: false` · **historical_self_check:** PASS (file scope).

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **PASS** | `components/ui/**` | `!important` count = **0** | Не добавлять |
| **PASS** | tokens `--hub-*` | Используются в UI | Не плодить параллельную палитру |
| **INFO** | `app.css:1359-1362` | `!important` только reduced-motion | Accepted a11y |
| **LOW** | возможные raw hex вне tokens | Spot-check drift | designer-ver на PR |
| **INFO** | Radix behavior | Не менять — только className | — |
