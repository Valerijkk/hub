# W5-fe-a11y — Accessibility spot-check

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** fe-ver / designer-ver (readonly) · **Дата:** 2026-07-23
> **Scope:** login, command palette, E2E a11y · **Код не менялся**

## Summary

E2E `platform/a11y.spec.ts` есть; NFR-16 WCAG full — **не** sign-off (AGENTS: login axe PASS, full open). reduced-motion в app.css — OK.

**Вердикт:** `verified: false` (полный WCAG не закрыт).

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **HIGH** | NFR-16 | Full WCAG sign-off open; только login axe | axe/playwright на shell+tracker+messenger |
| **MEDIUM** | modals/dialogs | Focus trap / Esc — проверить Shadcn Dialog везде | Audit critical modals |
| **MEDIUM** | live regions | Toasts/ошибки не везде `aria-live` | Добавить polite/assertive |
| **PASS** | `app.css` reduced-motion | `!important` exception | — |
| **PASS** | E2E `a11y.spec.ts` | Smoke существует | Расширить |
| **LOW** | icon-only buttons | Нужен `aria-label` spot-check | Lint jsx-a11y |
