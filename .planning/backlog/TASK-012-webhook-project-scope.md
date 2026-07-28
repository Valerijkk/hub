# TASK-012 — webhook dispatch: фильтр по проекту эндпоинта

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P0** · Статус: На проверке  
> Источник аудита: W4-A5-002 · TASK-008 стадия B

## [OUTCOME]

Активный вебхук проекта A не получает payload событий проекта B
(в т.ч. при `events:["*"]` и после фикса матчинга TASK-011).

## [ROLE]

- Role: be1
- Responsibility: project-scope в `dispatchEvent` + denial-тест
- Handoff to: стадия C

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - `backend/internal/apigateway/webhook.go`
  - `backend/internal/apigateway/repository.go` (WithProject / ProjectID в List)
  - `backend/testframework/apigateway/**`
  - backlog README + этот файл
- Do not touch:
  - signing, SSRF guards
- Constraints:
  - `ListActiveWebhookEndpoints` сейчас без edge Project
  - Парсить projectID из subject `project.<uuid>.…` и/или payload
  - События без project (если есть) — задокументировать поведение (не слать
    project-scoped hooks / или только org-level — выбрать fail-closed:
    не доставлять в project hooks без extractable project)
- Approval record: N/A

## [REFERENCES]

- Находка: A4-5 W4-A5-002
- `dispatchEvent`: `webhook.go:209-234`
- `ListActiveWebhookEndpoints`: `repository.go:116-119`
- `WebhookEndpointProjectID` уже есть для CRUD path
- Source precedence: код

## Суть дефекта

Все active endpoints получают любой матчащий event без сверки
`ep.ProjectID` с проектом события → cross-project leak.

## [ACCEPTANCE]

- Done criteria:
  - При dispatch: endpoint проекта A не ставится в очередь на event проекта B
  - Denial-тест с двумя проектами / двумя endpoints
  - List active грузит ProjectID (edge или field)
- **Мутация (C):** убрать фильтр project → тест краснеет
- Regression: `./testframework/apigateway/...`
- Stop when: green + статус «Исправлена»

## [OUTPUT SCHEMA]

- Коммит `fix(apigateway): ...` (можно атомарно с TASK-011 или отдельно)
- Denial-тест
- README

## [OPERATING RULES]

- Делать в одной волне с TASK-011
- Co-authored-by запрещён
