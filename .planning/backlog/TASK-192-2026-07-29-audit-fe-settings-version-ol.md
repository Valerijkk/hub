# 🔒 Audit FE: settings webhooks/AI privacy без version (OL)

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`AUDIT-FE-settings-mp-accounting-2026-07-29.md`](../audits/total-2026-07/AUDIT-FE-settings-mp-accounting-2026-07-29.md)  
> Исполнитель: **fe1** · Finding IDs: `FE-SMA-001`, `FE-SMA-002` (+ optional `FE-SMA-003`)

## [OUTCOME]

Toggle webhook и save AI privacy шлют `version` (BE VersionEQ); UI не ломается 400/409. Optional: вкладка marketplace «Мои» не показывает empty во время loading.

## [ROLE]

- Role: fe1 · Handoff: fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/src/modules/settings/webhooks.tsx`, `ai-privacy.tsx`; `frontend/src/api/webhooks.ts`, `ai.ts`; optional `marketplace-page.tsx`; Vitest
- Do not touch: BE VersionEQ (уже FIXED); accounting (чисто)
- Commits EN/RU; no AI co-author

## [ACCEPTANCE]

- FE-SMA-001: `WebhookEndpoint`/`UpdateWebhookRequest` + `version` в toggle; toast на fail
- FE-SMA-002: `AIUserPreference.version`; save с `version`; 409 → toast + refresh
- Optional FE-SMA-003: mine tab `isPending` → loading ≠ «нет объявлений»
- Mutation: omit version → RED (или UI stub без version ломает toggle/save)
- Stop: FIXED · «На проверке»
