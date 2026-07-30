# 🔒 Audit FE: settings webhooks/AI privacy без version (OL)

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: [`AUDIT-FE-settings-mp-accounting-2026-07-29.md`](../../../audits/total-2026-07/AUDIT-FE-settings-mp-accounting-2026-07-29.md)  
> Исполнитель: **fe1** · Mutation: **fe2** W10 · Finding IDs: `FE-SMA-001`, `FE-SMA-002` (+ optional `FE-SMA-003`)  
> Evidence: [`MUTATION-2026-07-30-W10-fe2.md`](../../../audits/total-2026-07/MUTATION-2026-07-30-W10-fe2.md) · 3/3 PROVEN

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
- Stop: FIXED · **Закрыта** (mutation W10 fe2 · 3/3 PROVEN)

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30 · W7 slot 4 · fe1). Слой **frontend**. Typecheck: `cd frontend && npx tsc --noEmit` — **PASS**. Удаление `version` из типа/вызова **может** поймать tsc, если callers ещё передают поле; primary RED — Vitest ниже (не полагаться только на tsc).

| ID | что ломать (file:line + фрагмент) | какой тест (`it` / команда) | слой |
|----|-----------------------------------|-----------------------------|------|
| FE-SMA-001 | `webhooks.tsx:114-116` — `updateWebhook(ep.id, { is_active: !ep.is_active })` **без** `version: ep.version`; **и/или** `api/webhooks.ts` — убрать `version` из `UpdateWebhookRequest` / `WebhookEndpoint` | describe `WebhooksPage version OL (FE-SMA-001 / TASK-192)` → `it("toggle active sends is_active + version")`; `it("updateWebhook PUT body includes version from caller")`; `it("UpdateWebhookRequest requires version…")` | frontend |
| FE-SMA-002 | `ai-privacy.tsx:48-50` — `updateAIPreferences({ allow_external_ai })` без `version`; **и/или** убрать 409→refresh (`:57+` handler) | describe `UserAISettings version OL (FE-SMA-002 / TASK-192)` → `it("save sends allow_external_ai + version from load")`; `it("409 on save refreshes preferences")` | frontend |
| FE-SMA-003 | `marketplace-page.tsx:181` / `:493` — empty «У вас нет объявлений» гейтить через `!browseLoading` вместо `!listLoading` (mine tab показывает empty while pending) | describe `MarketplacePage mine loading≠empty (FE-SMA-003 / TASK-192)` → `it("shows loading on mine tab while listMyListings pending, not empty copy")` | frontend |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| webhook toggle + version body | ✅ RTL spy `api.put` | |
| `UpdateWebhookRequest.version` type | ✅ `expectTypeOf` (+ optional tsc) | type-only — не runtime BE VersionEQ |
| AI privacy save version + 409 refresh | ✅ RTL | |
| marketplace mine loading≠empty | ✅ RTL | |
| BE VersionEQ reject без version | ❌ FE suite | **не layer_mismatch для FE-критерия**; BE уже FIXED OOS |
| `npx tsc --noEmit` alone | ⚠️ | может поймать omit из типа; **не** ловит wrong-tab loading gate (003) |

**layer_mismatch:** нет (FE guards ↔ FE tests).  
**Confidence:** CONFIRMED.
