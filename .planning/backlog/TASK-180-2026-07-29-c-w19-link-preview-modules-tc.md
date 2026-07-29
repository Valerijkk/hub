# 🔒 C-W19: modules LinkPreview XSS + test_case preview membership

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`C-W19-fe-ai-platform.md`](../audits/total-2026-07/C-W19-fe-ai-platform.md) §C-W19-002  
> Исполнители: **fe1** + **be2** · Finding ID: `C-W19-FE-002`

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
