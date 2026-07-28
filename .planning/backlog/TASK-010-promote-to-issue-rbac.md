# TASK-010 — promote-to-issue: обязательный PermTaskCreate в целевом проекте

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P0** · Статус: Исправлена  
> Источник аудита: W3-A2-001 · TASK-008 стадия B

## [OUTCOME]

Участник канала не может создать задачу в чужом проекте через
`POST /messages/:id/promote-to-issue` без `PermTaskCreate`; при отсутствии
checker prod fail-closed (403), а не silent skip.

## [ROLE]

- Role: be1
- Responsibility: вызов `promoteCheck` + wiring адаптера RBAC + denial-тест
- Handoff to: стадия C (мутация)

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - `backend/internal/messenger/service_promote.go`
  - `backend/cmd/server/wiring_messenger.go`
  - адаптер `CanCreateTaskInProject` (rbac или тонкий wrapper рядом с wiring)
  - `backend/testframework/messenger/**`
  - backlog README + этот файл
- Do not touch:
  - контракт IssuePromoter / CreateIssueFromPromotion (оставить без RBAC внутри —
    guard живёт в messenger, как в комментарии tracker)
  - UI promote (если нет отдельной утечки)
- Constraints:
  - Сейчас `SetPromotePermissionChecker` есть, но **нигде не вызывается**
    в `PromoteToIssue`; wiring только `SetIssuePromoter`
  - `promoteCheck == nil` → Forbidden (fail-closed), не «пропустить»
  - Docker + hubtest
- Approval record: N/A

## [REFERENCES]

- Находка: `.planning/audits/total-2026-07/A3-2-messenger-services.md` W3-A2-001
- Интерфейс: `PromotePermissionChecker` в `service_promote.go:36-38`
- Tracker: `CreateIssueFromPromotion` явно без RBAC (`service_issues.go:92-98`)
- Stub в тестах: `testframework/messenger/service_wave123_test.go`
- Source precedence: код > комментарий контракта (комментарий врёт)

## Суть дефекта

Контракт утверждает проверку `PermTaskCreate`, но `promoteCheck` не вызывается,
а в prod не wired. Любой member канала создаёт issue в любом `project_id`.

## [ACCEPTANCE]

- Done criteria:
  - Перед `CreateIssueFromPromotion`: `promoteCheck.CanCreateTaskInProject`;
    false/err → 403
  - nil checker → 403 (или 500 Internal «не сконфигурирован» — выбрать
    Forbidden для PrivEsc-пути; зафиксировать в тесте)
  - `wiring_messenger.go`: `SetPromotePermissionChecker` с реализацией через
    `rbacSvc` + `PermTaskCreate`
  - Denial-тест: user member канала A, без прав в проекте B → 403, issue не создана
- **Мутация (C):** удалить блок `promoteCheck` → тест краснеет
- Regression: `./testframework/messenger/...`
- Stop when: denial green + wiring + статус «Исправлена»

## [OUTPUT SCHEMA]

- Коммит `fix(messenger): ...`
- Тест denial
- README + статус файла

## [OPERATING RULES]

- Guard на слое service messenger (не только handler)
- Co-authored-by запрещён

## Stage C (W18-C1 mutation) — PASS

- Mutation: remove `promoteCheck` block in `PromoteToIssue`
- RED: `TestPromoteToIssue_ValidationAndFlow/nil_promote_checker_forbidden` + `denied_by_promote_checker` — expected error, got nil
- GREEN: after restore
- Mock-only?: no (guard in service layer; mock repo is appropriate)
