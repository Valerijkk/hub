# 🔒 Audit BE: PushSender без *ent.Client + Telegram link rate limit

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: [`AUDIT-BE-auth-storage-notif-2026-07-29.md`](../audits/total-2026-07/AUDIT-BE-auth-storage-notif-2026-07-29.md)  
> Исполнитель: **be2** · Finding IDs: `ASN-001`, `ASN-002`  
> Mutation prove: **be1** W13 slot 1 refill · 2026-07-30 · [`MUTATION-2026-07-30-W13-be1-task188.md`](../audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task188.md) · **2/2 PROVEN**

## [OUTCOME]

`PushSender` не держит `*ent.Client` (layering через repo/port); Telegram link-code confirm под rate limit (как magic-link семейство).

## [ROLE]

- Role: be2 · Handoff: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/internal/notification/**` PushSender; `backend/internal/auth/**` telegram link confirm; middleware rate limit; tests
- Do not touch: ослаблять NF-07 ValidateOutboundURL; Magic Link rate contract
- TZ: no `*ent.Client` in Service/Handler layering
- Commits EN/RU; no AI co-author

## [ACCEPTANCE]

- ASN-001: PushSender без `*ent.Client`; hubtest/layering assert
- ASN-002: confirm link-code rate-limited; mutation убрать RL → RED
- Stop: 2/2 PROVEN (mutation) · «Закрыта»

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 1 refill / be1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| ASN-001 | `backend/internal/notification/push.go:43-76` — `PushSender` хранит `PushSubscriptionStore`, **нет** `*ent.Client` / import `ent` | PushSender (layering) | `TestPushSender_NoEntImport_ASN001` · `backend/testframework/notification/push_layering_asn001_test.go` | вернуть `db *ent.Client` / `import …/ent` (+ pushsubscription) в `push.go` → AST import-assert RED |
| ASN-002 | `telegram_bind.go:111-147` — `ConfirmLinkCode`: `failsKey` ≥ `tgLinkCodeVerifyMax` → `ErrTelegramConfirmLimited`; `recordLinkCodeFailure` на invalid; `handler_telegram_bind.go:60-61` → 429 | service (+ handler map) | `TestTelegramBind_ConfirmLinkCode_BruteForceLockout_ASN002` · `backend/testframework/notification/telegram_linkcode_rl_asn002_test.go` | убрать fails-check / `recordLinkCodeFailure` → brute-force без `ErrTelegramConfirmLimited` / HTTP≠429 → RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | AST `parser.ImportsOnly` на prod `push.go` | **OK** — static layering gate (не Ent runtime) |
| 002 | TelegramBindService + miniredis fails counter + Ent user fixture; Fiber confirm → 429 | **OK** — Redis RL + Ent user; не mock≠Ent |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A** — оба ID code/runtime (001 = static import).
- Do not touch: NF-07 `ValidateOutboundURL` в push send path; Magic Link rate contract (`magiclink:fails:`) — зеркало, не ослаблять.
