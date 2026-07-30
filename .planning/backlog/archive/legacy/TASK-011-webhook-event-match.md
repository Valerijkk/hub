# TASK-011 — webhook: матчинг UI-событий с NATS subjects

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P0** · Статус: Закрыта
> Источник аудита: W4-A5-001 · TASK-008 стадия B

## [OUTCOME]

Вебхук, созданный в UI с событиями вида `issue.created`, получает доставку
при публикации JetStream subject `project.<uuid>.issue.created` (и аналоги
для tms/sprint).

## [ROLE]

- Role: be1
- Responsibility: нормализовать `containsEvent` + интеграционный тест enqueue
- Handoff to: стадия C

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - `backend/internal/apigateway/webhook.go` (`containsEvent` и хелперы)
  - `backend/internal/apigateway/export_hooks.go` при необходимости
  - `backend/testframework/apigateway/**`
  - опционально FE только если нужны aliases в UI (предпочтительно чинить match на BE)
  - backlog README + этот файл
- Do not touch:
  - HMAC signing path (SEC-02)
  - SafeDial / ValidateOutboundURL
- Constraints:
  - UI: `frontend/src/modules/settings/webhooks.tsx:13-18` — короткие имена
  - Subject примеры: `project.<id>.issue.created`, `tms.testrun.completed`, …
  - Сохранить `*` как wildcard
  - Exact match короткого имени на короткий subject оставить working
- Approval record: N/A

## [REFERENCES]

- Находка: `.planning/audits/total-2026-07/A4-5-apigateway.md` W4-A5-001
- `containsEvent`: `webhook.go:325-331`
- Source precedence: код

## Суть дефекта

`containsEvent` сравнивает только `==` или `*`. UI сохраняет `issue.created`,
NATS шлёт `project.<uuid>.issue.created` → продуктовый путь не доставляет.

## [ACCEPTANCE]

- Done criteria:
  - Match: exact **или** suffix/token (последний сегмент / `HasSuffix("."+short)`)
    **или** явная таблица aliases UI→subject-pattern
  - Тест: endpoint events=`["issue.created"]` + eventType=`project.<uuid>.issue.created`
    → job enqueued / containsEvent true
  - `*` по-прежнему матчит всё
- **Мутация (C):** вернуть exact-only → тест краснеет
- Regression: `./testframework/apigateway/...`
- Stop when: green + статус «Исправлена»

## [OUTPUT SCHEMA]

- Коммит `fix(apigateway): ...`
- Тест match
- README

## [OPERATING RULES]

- Не ломать существующие exact subjects в тестах
- Co-authored-by запрещён

## Stage C (W18-C1 mutation) — PASS

- Mutation: `eventMatchesSubscription` → exact-only (`*` / `==`)
- RED: `TestContainsEvent_UIShortName_MatchesProjectSubject` + `TestWebhookDispatcher_dispatchEvent_EnqueuesOnUIShortMatch`
- GREEN: after restore
- Mock-only?: no (real `containsEvent` via export)

## Зависимость

Делать **до или вместе** с TASK-012: после фикса match без project-scope
усилится leak (`*`). TASK-012 обязателен в той же волне P0.
