# GitHub / GitLab / Integrations / Imex / Marketplace

**Дата:** 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Зона A5:** `backend/internal/{github,gitlab,integrations,imex,marketplace}`  
**Режим:** READ-ONLY (кроме этого отчёта)

## Что проверено

| Пакет | Файлы (production `.go`) | Фокус |
|-------|--------------------------|--------|
| `github/` | handler, service, repository, preview, pr_review, converters, dto | Webhook HMAC (SEC-02), SafeHTTP, RBAC маршрутов, IDOR Delete/CreateBranch, cross-secret |
| `gitlab/` | handler, service, repository, mr_review, dto | X-Gitlab-Token + ConstantTimeCompare, SSRF ValidateOutboundURL, TriggerPipeline scope, URL build |
| `integrations/onec/` | handler, service, settings_store, dto, user_lookup | Admin gate, SSRF на webhook URL, маскирование bearer |
| `imex/` | handler, service, repository, jira_import, pdf, dto | Authz на import/export, CSV mappings, tx AllocateIssueKey, PDF project scope |
| `marketplace/` | handler, service, repository, dto | Admin/owner checks, optimistic lock, pagination cap |

**Не найдено / чисто по чеклисту:** `http.DefaultClient` отсутствует; исходящий HTTP — `safehttp.NewSafeHTTPClient` / `ValidateOutboundURL`; webhook без skip при пустой подписи/токене (403); access tokens в ответах интеграций не утекают; NATS `PublishJSON` nil-safe; marketplace `per_page` capped ≤100 в repository.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A5-001 | P0 | `imex/handler.go:80` | `ImportIssuesCSV(..., nil, preview)` — mappings всегда `nil` | В import-режиме `buildFieldMap` (`service.go:286-298`) при пустых mappings даёт пустой `fieldMap` → `getField(..., "title")` всегда `""` → все строки skip. Тесты сервиса передают mappings явно (`testframework/imex/service_extended_test.go:69`), HTTP-путь — нет. CSV-импорт через API фактически мёртв. | Парсить mappings из form/JSON; либо auto-map по именам колонок (`title`/`Title`/`Summary` → `title`) когда mappings пуст. Добавить handler-тест `mode=import`. |
| A5-002 | P1 | `imex/handler.go:26-35` | Import/export только `authMw`, без `rbac.RequirePermission` | Любой аутентифицированный пользователь с UUID чужого проекта может импортировать/экспортировать задачи, ZIP, PDF. GitHub/GitLab на тех же `/projects/:projectId` требуют `PermProjectView` / `ManageWorkflow`. | Навесить `PermProjectView` (export) / `PermTaskCreate` или `PermProjectEdit` (import) как у tracker; membership через существующий RBAC. |
| A5-003 | P1 | `github/handler.go:202-214`, `github/service.go:121-122`, `github/repository.go:215-216` | `DeleteIntegration(id)` без привязки к `:projectId` | RBAC проверяет право на project из URL, удаление идёт по голому UUID. Пользователь с `manage_workflow` на проекте A удаляет интеграцию проекта B: `DELETE /projects/{A}/github/integrations/{B-id}`. | `DeleteIntegration(ctx, projectID, id)` + `Where(HasProjectWith(project.ID(projectID)))`; 404 если не принадлежит. |
| A5-004 | P1 | `gitlab/handler.go:185-196`, `gitlab/service.go:113-114`, `gitlab/repository.go:162-163` | Тот же IDOR на удаление GitLab-интеграции | Аналогично A5-003. | Scoped delete по `projectID` + `integrationID`. |
| A5-005 | P1 | `gitlab/service.go:347-356` | `TriggerPipeline` берёт интеграцию по ID, `projectID` из URL не сверяется | Юзер с правом на проект A может триггерить CI чужой интеграции (токен репо B), передав её UUID. | После `GetIntegration` проверить `integration.ProjectID == projectID` (или query с filter по project). |
| A5-006 | P1 | `github/handler.go:51`, `github/service.go:310-325` | `CreateBranch`: только `authMw`; нет проверки, что issue и integration из одного проекта | Любой залогиненный юзер, зная `issueKey` + `integrationId`, создаёт ветку в чужом GitHub-репо через чужой PAT. | RBAC на проект issue; `integration.Edges.Project.ID == issue.ProjectID`; иначе 403/404. |
| A5-007 | P1 | `gitlab/handler.go:43`, `gitlab/service.go:444-459` | То же для GitLab `CreateBranch` | Чужой PAT / чужой репозиторий. | Как A5-006. |
| A5-008 | P1 | `github/handler.go:119-126` + `service.go:152` / `208` | HMAC принимается, если совпал **любой** секрет инстанса; репозиторий берётся из payload | Утечка/знание webhook secret интеграции A позволяет подписать payload с `repo` интеграции B → `FindIntegrationByRepo` пишет links / меняет статусы чужого проекта. Signature check не привязан к найденной интеграции. | Валидировать подпись секретом **найденной** по repo интеграции (или вернуть matching integration id из verify и использовать только его). Не принимать «any-of secrets». |
| A5-009 | P1 | `gitlab/handler.go:99-102` + `service.go:145` / `186` | `X-Gitlab-Token` сравнивается с любым секретом; `project_path` из payload | Аналог A5-008 для GitLab. | Сверять token с секретом интеграции, найденной по `project_path` (constant-time), либо двухфазно: path → secret → compare. |
| A5-010 | P1 | `imex/pdf.go:22-29`, `74-81` | `projectID` аргумент не используется; `GetTestRun`/`GetSprint` по ID без scope | При наличии auth (и отсутствии RBAC — A5-002) экспорт PDF чужого прогона/спринта через URL другого `projectId`. Даже с RBAC на path — IDOR внутри org. | Загрузить run/sprint с `Where(HasProjectWith(project.ID(projectID)))`; mismatch → 404. |
| A5-011 | P2 | `marketplace/service.go:121`, `216`, `270`; `repository.go:136-137` | Optimistic lock: `if req.Version > 0` — при `version: 0` / omit проверка и WHERE пропускаются | Клиент может обойти conflict detection; lost update при concurrent edit. | Требовать `version >= 1` (ErrInvalidInput); всегда `Where(Version(req.Version))`. |
| A5-012 | P2 | `gitlab/service.go:475-476` | `branch`/`ref` вставляются в query без `url.QueryEscape` | `BranchName` / `Ref` с `&`, `=`, пробелами ломают/инжектят query GitLab API. | `url.Values{}.Set("branch", ...).Encode()` или `QueryEscape`. |
| A5-013 | P2 | `github/handler.go:48`, `preview.go:55-56`, `94-96` | `ListGitHubLinks` / issue link-preview: auth без membership | Утечка title/status/assignee/links по `issueKey` любому authenticated user. | Проверять `PermProjectView` на проект issue (как tracker). |
| A5-014 | P2 | `gitlab/service.go:320-321` (и аналоги success-path Decode: ~418, ~560) | Успешный ответ GitLab читается `json.NewDecoder(resp.Body)` без `LimitReader` | Error-path лимитирован 10MB; success — нет → memory DoS от self-hosted GitLab. | Обернуть `io.LimitReader(resp.Body, maxGitLabResponseBodyBytes)` и на success. |
| A5-015 | P2 | `github/handler.go:147-157` | CreateIntegration: нет проверки пустых `access_token` / `webhook_secret` (в отличие от GitLab handler) | Сохраняется бесполезная/дыра в webhook (пустой secret пропускается в `ListAllWebhookSecrets`, но битый token). | Валидировать required поля как `gitlab/handler.go:133-138`. |
| A5-016 | P3 | `gitlab/handler.go:100` | `ConstantTimeCompare([]byte(token), []byte(secret))` при разной длине — early exit (не constant-time) | Timing по длине токена; при нескольких секретах разной длины — доп. leak. | Сравнивать HMAC/SHA256 обоих, или pad до max length перед compare. |
| A5-017 | P3 | `imex/pdf.go:204-206` | Обрезка ячейки PDF: `cellVal[:47]` по байтам | Mid-rune truncate для UTF-8 (кириллица/emoji) → битый PDF text. | Обрезать по runes (`[]rune`). |
| A5-018 | P3 | `imex/service.go:267-276` | `_ =` на `MarshalIndent` metadata / `zw.Create` / `Write` | При ошибке записи ZIP клиент может получить обрезанный/пустой архив с HTTP 200. | Проверять err; fail export. |

## Чисто

- **SEC-02 (skip signature):** нет — пустой `X-Hub-Signature-256` / `X-Gitlab-Token` → 403 (`github/handler.go:100-101`, `gitlab/handler.go:90-91`).
- **SSRF:** GitHub/GitLab/1C/preview используют SafeHTTP / `ValidateOutboundURL`; `http.DefaultClient` в зоне нет.
- **Secrets in API responses:** `toIntegrationResponse` не отдаёт tokens; 1C `BearerTokenSet` boolean only.
- **Webhook body DoS:** `LimitReader(..., max+1)` + size check на GitHub/GitLab/imex upload.
- **Marketplace authz:** admin на pending/approve/reject; owner на update/sold/withdraw; soft-delete owner|admin.
- **imex AllocateIssueKey:** транзакция + rollback на ошибках (`repository.go:50-77`).
- **GitHub HMAC compare:** `hmac.Equal` (не `!=`).

## Оценка зоны: 5.5/10

**Обоснование:** Инфраструктурный security baseline (SafeHTTP, HMAC/token reject, encrypt-at-rest tokens, admin gate 1C, marketplace ownership) на уровне ~7–8. Сильно тянут вниз: полностью сломанный HTTP CSV-import (P0), отсутствие RBAC на всём imex, классические IDOR на delete/trigger/create-branch/PDF, и webhook «any secret → any repo» (P1). После фикса A5-001…A5-010 зона поднялась бы к ~7.5–8/10.
