# 🐛 C-W19: FE sprint.completed после BE allowlist trim

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: Stage C W19 · [`C-W19-fe-ai-platform.md`](../audits/total-2026-07/C-W19-fe-ai-platform.md) §C-W19-001  
> Исполнитель: **fe1** · Finding ID: `C-W19-FE-001` (report C-W19-001 HIGH)

## [OUTCOME]

UI webhooks не предлагает `sprint.completed` после BE-trim (TASK-152); Create/Update не получает 400 «неизвестное событие» из-за мёртвой опции.

## [ROLE]

- Role: fe1 · Handoff: fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/src/modules/settings/webhooks.tsx`; wiki/API one-liner если обещает event; README + этот файл; Vitest settings/webhooks
- Do not touch: BE allowlist / вернуть `sprint.completed` producer без ADR
- Commits: type/scope EN, описание RU; no AI co-author

## [REFERENCES]

- BE trim: `apigateway/webhook.go` allowedWebhookEvents (TASK-152 A3W4-5-003)
- FE: `webhooks.tsx` `AVAILABLE_EVENTS` всё ещё с `sprint.completed`
- Audit: `.planning/audits/total-2026-07/C-W19-fe-ai-platform.md`

## [ACCEPTANCE]

- Убрать `sprint.completed` из FE `AVAILABLE_EVENTS` (+ docs если нужно)
- Vitest: опция отсутствует; mutation — вернуть event → тест краснеет / API 400 documented
- Stop: FIXED · «На проверке»

## [EVIDENCE]

- **Fix:** `frontend/src/modules/settings/webhooks.tsx` — `AVAILABLE_EVENTS` без `sprint.completed` (export + comment C-W19-FE-001). SHA: `4b4a409e`
- **Test:** `testframework/stages/unit/frontend-test/domains/platform/webhook-events-sprint-completed-c-w19.test.tsx` — 2/2 PASS. SHA: `a5cb26fd`
- **Mutation:** временно добавить `"sprint.completed"` в массив → **оба it FAIL** (not.toContain + queryByText null); restore → GREEN.
- **Регрессия:** trash-webhooks-p1 + tokens-dead-ui-authz-p2 + TASK-179 = 14/14 PASS.
- **BE contract:** Create/Update с неизвестным event → 400 (TASK-152); wiki/API.md уже документирует trim.
- **Docs:** `df67ac7d` · статус «На проверке» · README не трогали (constraint).
