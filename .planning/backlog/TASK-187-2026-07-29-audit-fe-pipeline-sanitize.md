# 🔒 Audit FE: pipeline-status sanitizeUrl + tracker residuals

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`AUDIT-FE-tracker-admin-2026-07-29.md`](../audits/total-2026-07/AUDIT-FE-tracker-admin-2026-07-29.md)  
> Исполнитель: **fe1** · Finding IDs: `FE-TA-001` (+ optional P3 из отчёта)

## [OUTCOME]

`pipeline-status` / CI run links проходят `sanitizeUrl` (как github/gitlab TASK-125); fail planning-poker / automation secondary не маскируются под empty.

## [ROLE]

- Role: fe1 · Handoff: fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/src/modules/tracker/pipeline-status.tsx` (или актуальный path); related tracker admin components из отчёта FE-TA-*; Vitest
- Do not touch: TASK-174 LinkPreview components; backend
- Commits EN/RU; no AI co-author

## [ACCEPTANCE]

- FE-TA-001: `href` только после sanitize / null на unsafe
- Optional P3 из того же отчёта в том же PR
- Mutation: raw `javascript:` url → нет href RED
- Stop: FIXED · «На проверке»
