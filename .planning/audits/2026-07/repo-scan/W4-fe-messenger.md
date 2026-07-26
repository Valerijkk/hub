# W4-fe-messenger — Frontend messenger

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** fe-ver (readonly) · **Дата:** 2026-07-23  
> **Scope:** `modules/messenger`, `stores/messenger-store.ts`, `api/messenger.ts` · **Код не менялся**

## Summary

Messenger UI + Zustand + WS. Unit coverage почти отсутствует (W6-vitest HIGH). Backend CRITICAL search/delete/polls — FE может отображать чужие данные при успешном API.

**Вердикт:** `verified: false`.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **CRITICAL** | BE messenger search (W1) | UI search без channel filter → полный dump если API отдаёт | BE scope; FE: требовать channel context |
| **HIGH** | Vitest gap (W6) | Только `channel-icon` unit; нет composer/list/unread | Компонентные тесты + TC-MSG |
| **MEDIUM** | `messenger-store` | Локальный optimistic state может расходиться с membership 403 | Rollback store on 403 |
| **MEDIUM** | reactions/pins UI | Вызовы API без client-side member guard (доверие к BE) | Disable actions если !member |
| **PASS** | `api/messenger.ts` | Типизированные эндпоинты | — |
| **LOW** | infinite scroll cursor | Соответствует TZ chat cursor | — |
