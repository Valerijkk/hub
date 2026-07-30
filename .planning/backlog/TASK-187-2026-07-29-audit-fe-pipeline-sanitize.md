# 🔒 Audit FE: pipeline-status sanitizeUrl + tracker residuals

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Mutation W11 slot 5 (fe1, 2026-07-30): FE-TA-001…004 **PROVEN**; FE-TA-005 **N/A** — returned Открыта.  
> Mutation W13 slot 5 refill (fe2, 2026-07-30): FE-TA-005 FIXED + Vitest · **5/5 PROVEN** → **Закрыта**. Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe2-task187.md`  
> Источник: [`AUDIT-FE-tracker-admin-2026-07-29.md`](../audits/total-2026-07/AUDIT-FE-tracker-admin-2026-07-29.md)  
> Исполнитель: **fe2** (refill) / fe1 (W11) · Finding IDs: `FE-TA-001`…`FE-TA-005`

## [OUTCOME]

`pipeline-status` / CI run links проходят `sanitizeUrl` (как github/gitlab TASK-125); fail planning-poker / automation secondary не маскируются под empty; CI chrome tokens — hub `var(--*)`.

## [ROLE]

- Role: fe2 · Handoff: fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/src/modules/tracker/pipeline-status.tsx` (или актуальный path); related tracker admin components из отчёта FE-TA-*; Vitest
- Do not touch: TASK-174 LinkPreview components; backend
- Commits EN/RU; no AI co-author

## [ACCEPTANCE]

- FE-TA-001: `href` только после sanitize / null на unsafe ✅
- FE-TA-002…004: attachments / poker / automation secondary ✅
- FE-TA-005: hub tokens в `pipeline-status.tsx` + Vitest source-guard ✅
- Mutation: raw `javascript:` url → нет href RED; raw tokens → source-guard RED ✅
- Stop: FIXED · **Закрыта**

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30 · W13 slot 5 refill · fe2). Слой **frontend**.

| ID | что ломать (file:line + фрагмент) | какой тест (`it` / команда) | слой |
|----|-----------------------------------|-----------------------------|------|
| FE-TA-001 | `pipeline-status.tsx` — убрать `if (!isSafeUrl(run.html_url))` → всегда `<a href={run.html_url}>` | describe `FE · PipelineStatus sanitizeUrl (FE-TA-001 / TASK-187)` → `it("mutation: javascript: html_url → нет href…")`; source guard | frontend |
| FE-TA-002 | `issue-attachments.tsx` — убрать `isSafeUrl` gate; `href={candidate}` raw | describe `FE · IssueAttachments sanitizeUrl (FE-TA-002 / TASK-187)` → `it("mutation: javascript: att.url → нет <a href>")` | frontend |
| FE-TA-003 | `planning-poker.tsx` — `loadFailed \|\| !pokerSession` → «Сессия не найдена…» | describe `PlanningPoker error≠empty (FE-TA-003 / TASK-187)` → `it("network fail → alert+retry, not «сессия не найдена»")` | frontend |
| FE-TA-004 | `automation-rules-panel.tsx` — не читать `statusesError`/`channelsError`; убрать secondary `QueryErrorAlert` | describe `AutomationRulesPanel secondary error≠empty (FE-TA-004 / TASK-187)` → `it("statuses fail…")`; `it("channels fail…")` | frontend |
| FE-TA-005 | вернуть raw `text-green-600`/`text-red-600`/`bg-gray-*` в `pipeline-status.tsx` | `it("source guard: hub tokens — нет text-green/red-600 / bg-gray-* (FE-TA-005)")` | frontend |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| pipeline `isSafeUrl`/`sanitizeUrl` | ✅ RTL + source guard | primary XSS |
| attachments sanitize | ✅ RTL | |
| planning-poker loadFailed≠empty | ✅ RTL | |
| automation secondary statuses/channels | ✅ RTL | |
| FE-TA-005 hub-tokens `pipeline-status` | ✅ source-guard | FIXED W13 fe2 |
| Residual cluster (gitlab/github/triage/header/sidebar `#hex`) | ❌ | hygiene leftover; вне kill-recipe |
| `npx tsc --noEmit` | ❌ | не ловит 001–005 |

**layer_mismatch:** нет.  
**Confidence:** CONFIRMED (001–005 PROVEN).
