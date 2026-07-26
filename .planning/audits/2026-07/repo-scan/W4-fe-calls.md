# W4-fe-calls — Frontend calls

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** fe-ver (readonly) · **Дата:** 2026-07-23
> **Scope:** `modules/calls`, `stores/call-store.ts`, `api/calls.ts` · **Код не менялся**

## Summary

LiveKit UI (lazy). BE CreateCall без membership (W2 CRITICAL) — FE может получить RoomJoin JWT на чужой channel. **0 Vitest** на calls (W6 HIGH).

**Вердикт:** `verified: false`.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **CRITICAL** | BE CreateCall (W2) | FE create-call по channelId без member UX gate | BE IsChannelMember; FE: только из channel context |
| **HIGH** | Vitest = 0 | Нет unit на dialog/recording controls | Мок LiveKit + RTL |
| **HIGH** | guest-join UI | Unauth guest path зависит от слабого invite (BE) | Не ослаблять client checks |
| **MEDIUM** | recording start/stop | UI кнопки без RBAC hint | Hide unless canManage |
| **PASS** | lazy LiveKit | Не в main bundle | — |
| **INFO** | SFU = LiveKit | mediasoup REJECTED — FE не мигрировать | — |
