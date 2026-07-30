# 🔒 C-W19: modules LinkPreview XSS + test_case preview membership

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: [`C-W19-fe-ai-platform.md`](../audits/total-2026-07/C-W19-fe-ai-platform.md) §C-W19-002  
> Исполнители: **fe1** + **be2** · Finding ID: `C-W19-FE-002`  
> Mutation: [`MUTATION-2026-07-30-W12-fe2.md`](../audits/total-2026-07/MUTATION-2026-07-30-W12-fe2.md) · FE **PROVEN** · BE **N/A** (SKIP fe2 slot)

## [OUTCOME]

Dead/parallel `modules/messenger/link-preview` не даёт `href` без whitelist; BE `test_case` link-preview не раскрывает тип/title чужому non-member (как issue path).

## [ROLE]

- fe1: FE modules LinkPreview → `@/lib/safe-url` / удалить dead path; не ломать `components/messenger` (TASK-174)
- be2: `github/preview.go` test_case → `IsProjectMember` или fail-closed generic
- Handoff: fe-ver + be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed FE: `frontend/src/modules/messenger/link-preview.tsx`, `message-bubble.tsx` (если wire); tests
- Allowed BE: `backend/internal/github/preview.go`, handler wiring; `testframework/github/**`
- Do not touch: ослаблять components whitelist; OOS OG img P3
- Commits EN type/scope + RU; no AI co-author

## [REFERENCES]

- Prod components FIXED (TASK-174); modules path `href={url}` без sanitize
- BE issue preview membership OK; test_case без membership → enum keys
- Audit: C-W19-fe-ai-platform.md

## [ACCEPTANCE]

- FE: `javascript:` / unsafe URL → no href / null (mutation)
- BE: non-member TC URL → не `type=test_case` с title (mutation)
- Stop: 2/2 FIXED · «На проверке»

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30, fe2 W7 slot 5). **FE only** (слот fe2); BE findings → skip.

| ID | что ломать (file:line + фрагмент) | какой тест (`it` / команда) | слой |
|----|-----------------------------------|-----------------------------|------|
| C-W19-FE-002 (FE) | `modules/messenger/link-preview.tsx:192` / `:225` — убрать `isSafeUrl` early-return; `:240` — `href={url}` без `sanitizeUrl` **и/или** `lib/safe-url.ts:5` ослабить `SAFE_URL_RE` (добавить `javascript:`) | describe `modules LinkPreview URL whitelist (C-W19-FE-002)` → `it("does not render preview for javascript: URL")`; `it("does not render preview for data: URL")`; `it("safe https URL uses sanitizeUrl href…")`; `it("mutation evidence: isSafeUrl rejects javascript:")` · `domains/messenger/modules-link-preview-xss.test.tsx` | frontend |
| C-W19-FE-002 (BE) | `backend/internal/github/preview.go:138` — убрать `IsProjectMember` / вернуть `type=test_case`+title non-member | **BE SKIP** (fe2 slot) — be2 / W8 BE agent | backend |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| modules LinkPreview isSafeUrl + sanitizeUrl href | ✅ RTL + unit | FE |
| BE test_case membership | ⏭ BE SKIP | не зона fe2; см. `preview.go` + `testframework/github/preview*.go` |
| components/messenger LinkPreview | OOS | TASK-174 |
| layer_mismatch | нет | RTL импортирует prod `modules/.../link-preview` |
| `npx tsc --noEmit` | ❌ | не ловит XSS href |

**Confidence:** CONFIRMED (FE); BE N/A для этого слота.
