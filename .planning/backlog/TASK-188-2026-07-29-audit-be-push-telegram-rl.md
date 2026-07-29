# 🔒 Audit BE: PushSender без *ent.Client + Telegram link rate limit

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`AUDIT-BE-auth-storage-notif-2026-07-29.md`](../audits/total-2026-07/AUDIT-BE-auth-storage-notif-2026-07-29.md)  
> Исполнитель: **be2** · Finding IDs: `ASN-001`, `ASN-002`

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
- Stop: 2/2 FIXED · «На проверке»
