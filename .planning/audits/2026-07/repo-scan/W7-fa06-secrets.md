# W7-fa06-secrets — FA-06 / PG-03 secret rotation

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** devops-engineer + security-auditor (readonly) · **Дата:** 2026-07-23
> **Scope:** `.planning/production/FA-06-EXECUTION-LOG.md`, PG-03 · **Код не менялся**

## Summary

**Единственный prod-блокер.** Выполнена только **Phase 0** (git history scan). Phase 1–3 rotation (Resend/Telegram P1, JWT/LiveKit P2) — **не закрыты**. Working tree clean (test mocks only); history INF-03/SEC-03 exposure documented.

**Вердикт:** `verified: false` · **open prod blocker**.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **CRITICAL** | `FA-06-EXECUTION-LOG.md` Phase 1–3 empty | Rotation **не выполнена** — secrets historically exposed | Manual ops: PG-03 Phase 1–3; заполнить лог без значений |
| **HIGH** | History INF-03 | RESEND_API_KEY + TELEGRAM_BOT_TOKEN в compose history pre-`859ecef` | Rotate P1 + invalidate old |
| **HIGH** | History SEC-03 | `livekit-prod.yaml` / `devkey` в git history | Rotate LiveKit keys; filter-repo только с CTO |
| **PASS** | Working tree grep | Secret patterns = test mocks only | — |
| **PASS** | Phase 0 checklist | Commands executed, recorded | — |
| **INFO** | Deploy DEV_MODE | Fixed; нет в deploy.yml | — |
| **MEDIUM** | Maintenance window | Не анонсирован | Ops schedule |

## Status

| Phase | Status |
|-------|--------|
| 0 Pre-flight scan | ✅ done |
| 1 Rotate P1 (Resend/TG) | ❌ open |
| 2 Rotate P2 (JWT/LiveKit/…) | ❌ open |
| 3 Verify + close log | ❌ open |
