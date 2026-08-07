# Справочник: cookbook, примеры, каталоги

> Часть плейбука тестирования. Канон и оглавление — [`../testing.md`](../testing.md).

Приложения A–G: cookbook по доменам, сквозной пример, каталог
security-тестов, команды и карта CI, FAQ, расширенные пары, рубрика ревью.
Открывается по конкретному вопросу, а не читается целиком.

---

## Приложение A. Cookbook по доменам

Для каждого домена — что тестировать, какие паттерны применять, реальные имена тестов из `backend/testframework/` и `docs/TESTING.md`. Используй как отправную точку при работе над доменом.

### A.1. auth (Magic Link only, v2.5)

**Инварианты (TZ):** только Magic Link (без паролей/SSO); Telegram — только notify bind, не login; JWT access 15 мин; refresh Redis 30 дней с instant revocation; rate limit 3/10 мин; сравнение секретов через `subtle.ConstantTimeCompare`.

| Что тестировать | Слой | Тест / TC |
|-----------------|------|-----------|
| Запрос кода magic link | handler | TC-AUTH-01 |
| Верификация кода → токены | service | `TestVerifyMagicLink_*`, TC-AUTH-02 |
| Неверный код (constant-time) | service | TC-AUTH-03, SEC-07 |
| JWT round-trip, expiry | service | `TestJWTExpired`, TC-AUTH-05 |
| Refresh обновление/отзыв | service | TC-AUTH-04 |
| Rate limit 3/10 мин | middleware | TC-AUTH-10 |
| Онбординг, logout | e2e | `onboarding-profile.spec.ts`, `logout.spec.ts` |

**Ключевой паттерн — проверка constant-time сравнения (SEC-07):**

```go
// TC-AUTH-03: неверный код отклоняется; сравнение через ConstantTimeCompare (timing-safe).
func TestVerifyMagicLink_WrongCode(t *testing.T) {
	svc, _ := newAuthTestService()
	ctx := context.Background()

	_, err := svc.VerifyMagicLink(ctx, "user@hub.local", "000000")
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	// 400/401 по контракту; главное — не 500 и не паника.
	assert.Contains(t, []int{400, 401}, appErr.Code)
}
```

**Anti-regression:** endpoints `/auth/telegram*` **удалены** (ADR-AUTH-01). Тест не должен их воскрешать; если встречаешь — это устаревший тест (ROOT=test).

### A.2. tracker (задачи, workflow, доски)

**Инварианты:** optimistic locking через `version`; issue_key автоинкремент (`HUB-1`, `HUB-2`); soft delete; 9 типов задач; workflow-переходы валидируются.

| Что тестировать | Слой | Тест / TC |
|-----------------|------|-----------|
| Создание, автоинкремент | service | `TestCreateIssue`, `TestCreateIssueAllTypes`, TC-TRCK-01/02 |
| Optimistic locking, конфликт | service | `TestUpdateIssueOptimisticLocking`, TC-TRCK-03/04 |
| Workflow валидный/невалидный переход | service | `TestMoveIssueValidTransition`, `TestMoveIssueInvalidTransition`, TC-TRCK-10/11 |
| Связи задач | service | `TestCreateIssueLink`, TC-TRCK-06 |
| JQL поиск | service | `TestParseJQLSimple`, `TestToEntPredicates`, TC-TRCK-09 |
| Метки, доски, кастомные поля | service | `TestCreateLabel`, `TestBoardData`, `TestCustomFieldValues` |
| Пагинация (cursor) | service | `TestListIssuesPagination` |

Паттерн — `mockRepo` с `projectCounters` для автоинкремента (см. раздел 4.7). Optimistic locking — эталонный тест раздела 4.4.

### A.3. messenger (каналы, DM, треды, реакции)

**Инварианты:** DM дедупликация; read-only канал (403 для не-Owner); unread не инкрементит отправителя; WS room только для участника (NF-02).

| Что тестировать | Тест / TC |
|-----------------|-----------|
| DM дедуп | `TestCreateDM_Dedup`, TC-MSG-02 |
| Отправка с проверкой участника | `TestSendMessage_MemberCheck`, TC-MSG-03 |
| Реакция toggle | `TestToggleReaction_AddRemove`, TC-MSG-04 |
| Тред + счётчик ответов | `TestThreadReply_IncrementsReplyCount`, TC-MSG-06 |
| Read-only (таблица решений) | `TestSendMessage_ReadOnlyBlocked_RegularUser`, `TestSendMessage_ReadOnlyAllowed_Admin`, TC-MSG-07 |
| Unread (skip sender) | `TestIncrement_SkipsSender`, `TestMarkRead_ResetsToZero`, TC-MSG-09 |
| Presence | `TestHeartbeat_SetsScore`, `TestGetStatus_Online/Away/Offline` |

**Паттерн — read-only как таблица решений (раздел 3.4):**

```go
// TC-MSG-07: матрица (is_readonly × роль) → разрешение отправки.
func TestSendMessage_ReadOnly_DecisionTable(t *testing.T) {
	cases := []struct {
		name       string
		isReadonly bool
		role       string
		wantCode   int // 0 = успех
	}{
		{"open channel, member", false, "member", 0},
		{"readonly, admin allowed", true, "admin", 0},
		{"readonly, member blocked", true, "member", 403},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange канал с tc.isReadonly, пользователь с tc.role
			// act: отправка
			// assert: tc.wantCode
		})
	}
}
```

### A.4. tms (тест-кейсы, прогоны, Allure)

**Инварианты:** версионирование кейсов; case_key `HUB-TC-N`; Allure matching по стратегиям (раздел 8.2); webhook требует API Key.

| Что тестировать | Тест / TC |
|-----------------|-----------|
| CRUD кейса (classic/BDD) | `TestCreateTestCase_Classic`, `TestCreateTestCase_BDD`, TC-TMS-01/02 |
| Версии + rollback | `TestUpdateTestCase_CreatesVersion`, `TestRollbackTestCase`, TC-TMS-03/04 |
| Прогон + результаты | `TestAddTestCasesToRun`, `TestExecuteTestResult_UpdatesStats`, TC-TMS-05/06 |
| Allure parse | `TestParseAllureResults_*`, TC-ALLURE-01..04 |
| Allure matching | `TestMatchTestCase_By*`, `TestMatchTestCase_PriorityOrder`, TC-ALLURE-05 |
| Bug from failed | `TestCreateBugFromFailedResult`, TC-TMS-09 |
| Webhook API Key | `handler_webhook_rbac_test.go`, SEC |

Allure — центральный домен playbook, см. раздел 8. Паттерн частичного мока (`mockMatcherRepo` с embedded interface) — раздел 4.7.

### A.5. calls (LiveKit, запись, webhook)

**Инварианты:** webhook signature обязательна (SEC-02, не skip); лимит активных записей; NATS nil-safe publish; эфемерный чат с лимитом длины.

| Что тестировать | Тест / TC |
|-----------------|-----------|
| LiveKit токен | `TestGenerateToken`, `TestGenerateToken_DifferentRooms`, TC-CALL-01/02 |
| Webhook подпись (reject invalid) | `TestVerifyWebhookSignature_EmptyAuth`, `TestVerifyWebhookSignature_InvalidJWT`, SEC-02 |
| Лимит записей | `TestMaxActiveRecordingsLimit`, TC-CALL-07 |
| Эфемерный чат + лимит длины | `TestSendMessage`, `TestSendMessage_TextTooLong`, TC-CALL-09 |

**Паттерн — webhook подпись (SEC-02, нельзя skip):**

```go
// SEC-02: webhook без валидной подписи отклоняется даже при отсутствии claim.
func TestVerifyWebhookSignature_EmptyAuth(t *testing.T) {
	svc := newCallsTestService()
	err := svc.VerifyWebhookSignature("", []byte(`{"event":"x"}`))
	require.Error(t, err) // пустая подпись → отказ, не пропуск
}
```

### A.6. sprint (state machine)

**Инварианты:** один active-спринт на проект; переходы planning→active→completed; перенос незакрытых при complete.

| Что тестировать | Тест / TC |
|-----------------|-----------|
| Создание, дата конца < начала | `TestCreateSprint`, `TestCreateSprint_EndDateBeforeStartDate`, TC-SPRINT-01 |
| Start, запрет второго active | `TestStartSprint_Success`, `TestStartSprint_RejectsWhenActiveExists`, TC-SPRINT-02/03 |
| Burndown, velocity | `TestGetBurndown`, `TestGetVelocity`, TC-SPRINT-04/05 |
| Complete + перенос | `TestCompleteSprint_MovesIncompleteIssues`, TC-SPRINT-07 |

Эталон техники «переходы состояний» — раздел 3.5.

### A.7. ws (real-time)

**Инварианты:** send after disconnect без паники (SEC-05); non-member subscribe отклоняется (NF-02); Fiber WS (не gorilla/net/http).

| Что тестировать | Тест |
|-----------------|------|
| Message-типы, JSON | `TestMessageTypeConstants`, `TestClientMessage_JSON`, TC-WS-01/02 |
| Ping без room | `TestClientMessage_PingHasNoRoom`, TC-WS-03 |
| Authz подписки (NF-02) | `authorizer_test.go`, `handler_auth_test.go` |
| Shutdown без паники (SEC-05) | `hub_shutdown_test.go`, `pump_test.go` |

`-race` здесь критичен — домен полон горутин.

### A.8. rbac, search, notification (кратко)

| Домен | Ключевые тесты | Инвариант |
|-------|----------------|-----------|
| rbac | `TestPermission_Has/Add/Remove`, `TestDefaultRolePermissions` | битовые операции идемпотентны; матрица ролей |
| search | `TestMergeResultsByScore`, JQL-тесты | user-scope фильтр (NF-03) — результаты только по проектам/каналам пользователя |
| notification | `TestSubjectToType`, `TestShouldDispatch_Debounce` | quiet hours; urgent пробивает; WebPush endpoint валидируется (NF-07) |

**Паттерн — user-scope фильтр поиска (NF-03):**

```go
// NF-03: результаты поиска ограничены проектами/каналами пользователя.
func TestSearch_UserScoped(t *testing.T) {
	// arrange: две сущности в разных проектах, пользователь — участник только одного
	// act: поиск
	// assert: возвращена только сущность из «своего» проекта
}
```

---
## Приложение B. Полный worked-example: фича end-to-end

Задача: добавить **WIP-лимит на колонку Kanban** (нельзя перенести задачу в колонку, где достигнут лимит). Показываем полный QA-цикл: TC → BE unit → FE unit → E2E → Allure → docs sync.

### B.1. Шаг 1 — `qa-manual` пишет TC

В `docs/TESTING.md`, модуль TRCK:

```markdown
| TC-TRCK-13 | WIP-лимит: перенос в полную колонку | Колонка "In Progress" wip_limit=2, в ней 2 задачи | Перетащить 3-ю задачу в колонку | 409/400, перенос отклонён, показана ошибка лимита | — (ручной) |
| TC-TRCK-14 | WIP-лимит: перенос при свободном месте | wip_limit=2, в колонке 1 задача | Перетащить 2-ю задачу | 200, задача перенесена | — (ручной) |
```

Границы (раздел 3.2): limit−1 (успех), limit (граница — ещё успех при перемещении внутри?), limit+1 (отказ).

### B.2. Шаг 2 — `qa-auto1` пишет BE unit

```go
// TC-TRCK-13, TC-TRCK-14: WIP-лимит колонки при перемещении задачи.
func TestMoveIssue_WIPLimit(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	defaultID := repo.getDefaultStatusID()
	inProgress := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, defaultID)
	repo.setWIPLimit(inProgress, 2) // хелпер мока: лимит колонки = 2

	// Заполняем колонку до лимита.
	for i := 0; i < 2; i++ {
		iss, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
			Title: fmt.Sprintf("Filler %d", i), IssueType: "task", Priority: "medium",
		})
		require.NoError(t, err)
		_, err = svc.MoveIssue(ctx, iss.IssueKey, userID, tracker.MoveIssueRequest{StatusID: inProgress.String(), Rank: "a"})
		require.NoError(t, err)
	}

	// TC-TRCK-13: третья задача — отказ по лимиту.
	third, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Third", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)
	_, err = svc.MoveIssue(ctx, third.IssueKey, userID, tracker.MoveIssueRequest{StatusID: inProgress.String(), Rank: "b"})
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Contains(t, []int{400, 409}, appErr.Code, "перенос сверх WIP-лимита должен отклоняться")
}
```

### B.3. Шаг 3 — `qa-auto2` пишет FE unit

```typescript
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { KanbanColumn } from "@/components/tracker/kanban-column";

describe("KanbanColumn WIP limit", () => {
	it("shows count vs limit and highlights when full", () => {
		render(
			<KanbanColumn
				name="In Progress"
				wipLimit={2}
				issues={[{ id: "1", issue_key: "HUB-1", title: "A", issue_type: "task", priority: "low", status: "in_progress" },
				         { id: "2", issue_key: "HUB-2", title: "B", issue_type: "task", priority: "low", status: "in_progress" }]}
			/>,
		);
		// Счётчик 2/2 виден.
		expect(screen.getByText("2/2")).toBeInTheDocument();
	});
});
```

### B.4. Шаг 4 — E2E (только если критический путь)

WIP-лимит критичен для UX доски → добавляем шаг в существующий сценарий доски, а не отдельный тяжёлый spec. Используем POM/Steps, dev-login, web-first assertions.

```typescript
// В домене tracker; проверяем сообщение об отказе при переносе сверх лимита.
test("WIP limit blocks over-fill", async ({ page, backendAvailable, devModeAvailable }) => {
	skipUnlessBackend(backendAvailable);
	skipUnlessDevMode(devModeAvailable);
	const auth = new AuthSteps(page);
	await auth.loginAsAdmin(page);
	// ... открыть доску с настроенным лимитом, перетащить сверх лимита,
	await expect(page.getByText(/лимит/i)).toBeVisible({ timeout: 10_000 });
});
```

### B.5. Шаг 5 — Allure-аннотация (при заливке в TMS)

```typescript
await allure.epic("Tracker");
await allure.feature("Kanban WIP limit");
await allure.label("AS_ID", "HUB-TC-TRCK-13");
await allure.severity("normal");
```

### B.6. Шаг 6 — docs sync и привязка

Обновить колонку «Автотест» и сводку:

```markdown
| TC-TRCK-13 | WIP-лимит: перенос в полную колонку | ... | 409/400 | `TestMoveIssue_WIPLimit` |
| TC-TRCK-14 | WIP-лимит: перенос при свободном месте | ... | 200 | `TestMoveIssue_WIPLimit` |
```

`TRCK: N → N+2 автоматизировано`. Заметка в `.planning/STATE.md`.

### B.7. Шаг 7 — регрессия и хендофф

```bash
cd backend && go test -tags hubtest -race -count=1 ./testframework/tracker/...
cd frontend && npx tsc --noEmit && npx vitest run ../testframework/stages/unit/frontend-test/domains/tracker
/qa-gate
```

Пирамида соблюдена: 1 BE unit (закрывает границы через один тест с наполнением), 1 FE unit, 0-1 E2E. Traceability двусторонняя. Артефакты вычищены.

---
## Приложение C. Каталог security-critical тестов

Эти тесты — приоритет №1 (из `docs/test_plan.md`). Их нельзя удалять/ослаблять; изменение — только с `security-auditor`.

| ID | Модуль | Требуемая проверка | Реальный тест |
|----|--------|--------------------|---------------|
| SEC-01 | apigateway | API key middleware order → 401 без ключа | `apikey_test.go`, `handler_webhook_rbac_test.go` |
| SEC-02 | calls | Webhook reject missing/tampered signature | `TestVerifyWebhookSignature_*` |
| SEC-05 | ws | send after disconnect — без паники | `hub_shutdown_test.go`, `pump_test.go` |
| SEC-06 | auth | concurrent register / first-admin lock | auth concurrency-тесты |
| SEC-07 | auth | `ConstantTimeCompare` на magic link | verify-тесты |
| NF-02 | ws | non-member subscribe → reject | `authorizer_test.go` |
| NF-03 | search | результаты в user-scope | search service-тесты |
| NF-07 | notification | WebPush invalid endpoint → 400 | notification-тесты |

**Правило:** при работе рядом с любым из этих тестов — прогнать security-пакеты и убедиться, что покрытие ≥80% держится. Изменение поведения security-инварианта требует `security-auditor` + `gsd-secure-phase`.

**Запрещённые паттерны (TZ FORBIDDEN), которые ловятся тестами/ревью:**

| Паттерн | Риск | Проверка |
|---------|------|----------|
| `http.DefaultClient` для user-URL | SSRF | `ValidateOutboundURL`/safehttp |
| `io.ReadAll` без `io.LimitReader` | memory DoS | лимитированное чтение тела |
| `!=` для секретов | timing attack | `subtle.ConstantTimeCompare` |
| `*ent.Client` в service/handler | нарушение layering | repository interface |
| `panic()` в prod path | краш | возврат ошибки |
| skip webhook signature | подделка события | обязательная проверка |

---
## Приложение D. Справочник команд и карта CI

### D.1. Команды прогона

```bash
# Backend — как CI (job backend-test)
cd backend
go generate ./ent
go vet -tags hubtest ./...
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Backend — один домен / один тест
go test -tags hubtest -race ./testframework/tracker/...
go test -tags hubtest -run TestCreateIssue ./testframework/tracker/...

# Backend — coverage (профиль в TMP, rule 45)
go test -tags hubtest -race -covermode=atomic -coverprofile="${TMPDIR:-/tmp}/hub-cov.out" ./testframework/...
go tool cover -func="${TMPDIR:-/tmp}/hub-cov.out" | tail -1

# Backend — тайминги / медленные тесты
go test -tags hubtest -v ./testframework/tracker/... 2>&1 | rg -- "--- (PASS|FAIL)"
go test -tags hubtest -json ./testframework/... > "${TMPDIR:-/tmp}/gotest.json"

# Backend — бенчмарк
go test -tags hubtest -run '^$' -bench . -benchmem ./testframework/tms/...

# Backend — детект флейка
go test -tags hubtest -race -run TestSuspect -count=20 ./testframework/<pkg>/...
go test -tags hubtest -shuffle=on ./testframework/<pkg>/...

# Prod build check
go build -tags release -o bin/server ./cmd/server

# Frontend — как CI (job frontend-test)
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit
npm run lint
npx vitest run
# один домен
npx vitest run ../testframework/stages/unit/frontend-test/domains/auth

# E2E
npm run test:e2e:smoke   # job e2e-smoke (3 spec, без backend)
npm run test:e2e         # job e2e-full (20 spec, с backend)
npx playwright show-trace test-results/<...>/trace.zip

# Full gate
/qa-gate
```

### D.2. Карта CI job → путь → домены

| CI Job | Стадия | Путь | Домены |
|--------|--------|------|--------|
| `backend-test` | unit | `backend/testframework/{domain}/` | auth, tracker, tms, messenger, calls, ws, sprint, … |
| `frontend-test` | unit | `stages/unit/frontend-test/domains/{domain}/` | auth, admin, ai, tracker, tms, v4, … |
| `build` | build | — | только сборка (`tsc`, `vite`, `-tags release`) |
| `e2e-smoke` | e2e | `stages/e2e/domains/` (3) | auth, platform |
| `e2e-full` | e2e | `stages/e2e/domains/` (20) | все 9 доменов |
| `coverage-gate` | quality | `backend/testframework/` | все BE, ≥82% |

### D.3. Docker-прогон backend (канонический, как в CI)

```bash
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --cpus 2.0 --memory 4g --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."
```

### D.4. Быстрая навигация по документам

| Нужно | Документ |
|-------|----------|
| TC-матрица, API-коды, manual checklist | `docs/TESTING.md` |
| Стратегия, пирамида, CI pipeline | `docs/test_plan.md` |
| Карта testframework | `testframework/PIPELINE.md` |
| Паттерны тестов (rule) | `.cursor/rules/40-testing.mdc` |
| Cleanup артефактов | `.cursor/rules/45-artifacts-cleanup.mdc` |
| Регрессия и верификаторы | `.cursor/rules/56-regression-verify.mdc` |
| Triage FAIL | `.cursor/skills/test-triage/SKILL.md` |
| Полный gate | `.cursor/skills/qa-gate/SKILL.md` |
| Coverage-аудит | `.cursor/skills/coverage-audit/SKILL.md` |
| Матрица покрытия | `.planning/COVERAGE.md` |
| Session state | `.planning/STATE.md` |

---
## Приложение E. FAQ и типовые ошибки агентов

**Q: Тест падает, я вижу почему — можно сразу поправить assertion?**
Нет. Сначала triage (раздел 12): воспроизведи функционал по спеке, определи ROOT=code|test, запиши. Правка assertion без этого = фиксация бага.

**Q: Где писать новый Go-тест — в `internal/` рядом с кодом?**
Нет. `internal/*_test.go` = 0 (TF-M08). Все тесты в `backend/testframework/{domain}/`. Приватные символы — через `export_hooks.go` (`//go:build hubtest`).

**Q: Нужен реальный Postgres для теста сервиса?**
Обычно нет. Мок (`mockRepo`) для логики; sqlite `env.Builder` для реальной схемы Ent. Реальный PG/Mongo — только для их специфики и только в CI.

**Q: Как гарантировать, что мой Allure-результат сматчится на нужный TC?**
Используй `AS_ID` label или tms-link с ключом `HUB-TC-*` (стратегия `MatchByAnnotation`, высший приоритет). Не полагайся на fullName/className.

**Q: E2E флейки — накрутить retries?**
Нет. Найди причину (обычно sleep вместо web-first assertion или общие данные). Ретрай — смягчение инфры, не замена фикса. Unit-тесты не ретраим вовсе.

**Q: Забыл обновить колонку «Автотест» — критично?**
Да. Без этого нет traceability (rule 55). Задача не готова к хендоффу.

**Q: qa-auto1 и qa-auto2 хотят тронуть один файл — как быть?**
Один файл — один владелец в рамках задачи. Делите по слою (BE/FE) и домену (раздел 1.4); пересечение = хендофф через STATE.md.

**Q: Куда писать coverage-профиль?**
В TMP (`${TMPDIR:-/tmp}`), никогда в `backend/`. После — cleanup (rule 45), `git status` чист.

**Q: Что делать, если спецификация неясна?**
Stop. Не угадывать поведение. Уточнить у SA/BA, обновить docs, потом писать код/тест (triage, шаг 3).

---
## Приложение F. Расширенные good/bad по частым сценариям

Быстрый справочник «как надо / как не надо» для ситуаций, которые встречаются чаще всего при работе над тестами Hub.

### F.1. Проверка ошибки создания сущности

**Bad — проверка по тексту ошибки (хрупко к переводам/рефакторингу):**

```go
_, err := svc.CreateIssue(ctx, projectID, userID, req)
assert.Equal(t, "issue validation failed: title required", err.Error())
```

**Good — по коду и подстроке поля:**

```go
_, err := svc.CreateIssue(ctx, projectID, userID, req)
require.Error(t, err)
appErr, ok := err.(*apperrors.AppError)
require.True(t, ok)
assert.Equal(t, 400, appErr.Code)
assert.Contains(t, err.Error(), "title") // указываем поле, не весь текст
```

### F.2. Тест с несколькими похожими случаями

**Bad — копипаста трёх почти одинаковых функций:**

```go
func TestCreateBug(t *testing.T)  { /* ... IssueType: "bug" ... */ }
func TestCreateTask(t *testing.T) { /* ... IssueType: "task" ... */ }
func TestCreateEpic(t *testing.T) { /* ... IssueType: "epic" ... */ }
```

**Good — table-driven + `t.Run` (см. реальный `TestCreateIssueAllTypes`):**

```go
for _, it := range []string{"bug", "task", "epic"} {
	t.Run(it, func(t *testing.T) {
		resp, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
			Title: "X", IssueType: it, Priority: "medium",
		})
		require.NoError(t, err)
		assert.Equal(t, it, resp.IssueType)
	})
}
```

### F.3. Ожидание асинхронного результата (FE)

**Bad — синхронный поиск сразу после рендера:**

```typescript
renderWithProviders(<IssueList projectId="p1" />);
expect(screen.getByText("HUB-42")).toBeInTheDocument(); // упадёт: данные ещё грузятся
```

**Good — `findBy*` (ждёт появления):**

```typescript
renderWithProviders(<IssueList projectId="p1" />);
expect(await screen.findByText("HUB-42")).toBeInTheDocument();
```

### F.4. Изоляция стора между тестами

**Bad — стейт течёт между тестами:**

```typescript
it("test A", () => { useAuthStore.getState().setTokens("a", "b"); /* ... */ });
it("test B", () => { /* isAuthenticated неожиданно true из test A */ });
```

**Good — сброс в `beforeEach`:**

```typescript
beforeEach(() => {
	sessionStorage.clear();
	useAuthStore.setState({ accessToken: null, refreshToken: null, user: null, isAuthenticated: false });
	vi.restoreAllMocks();
});
```

### F.5. E2E-навигация после действия

**Bad — sleep + проверка url строкой:**

```typescript
await createButton.click();
await page.waitForTimeout(2000);
expect(page.url()).toContain("/projects/");
```

**Good — web-first assertion с regex и таймаутом:**

```typescript
await createButton.click();
await expect(page).toHaveURL(/\/projects\/[0-9a-f-]+/, { timeout: 30_000 });
```

### F.6. Мок HTTP-клиента vs реальная сеть (FE)

**Bad — тест ходит в реальный backend (флейк, зависимость от окружения):**

```typescript
it("loads projects", async () => {
	const data = await fetchProjects(); // реальный fetch
	expect(data.length).toBeGreaterThan(0);
});
```

**Good — `vi.spyOn` на api-клиент (детерминизм):**

```typescript
const getSpy = vi.spyOn(api, "get").mockResolvedValue([{ id: "1", name: "Hub" }]);
const data = await fetchProjects();
expect(getSpy).toHaveBeenCalledWith("/projects");
expect(data).toHaveLength(1);
```

### F.7. Проверка события/побочного эффекта

**Bad — проверка, что метод «как-то отработал»:**

```go
err := svc.SendMessage(ctx, channelID, userID, "hi")
require.NoError(t, err) // но не проверили, что участник проверен и событие отправлено
```

**Good — проверка наблюдаемого контракта:**

```go
// TC-MSG-03: отправка проверяет членство и публикует NATS-событие.
msg, err := svc.SendMessage(ctx, channelID, memberID, "hi")
require.NoError(t, err)
assert.Equal(t, "hi", msg.Content)
// не-участник → отказ
_, err = svc.SendMessage(ctx, channelID, strangerID, "hi")
require.Error(t, err)
```

### F.8. Уникальность тестовых данных

**Bad — фиксированный ключ (конфликт при параллели/повторе):**

```go
proj, _ := client.Project.Create().SetKey("TST").Save(ctx) // второй тест упадёт на unique
```

**Good — уникальность через uuid (реальный env.Builder):**

```go
email := b.prefix + "-" + uuid.New().String()[:8] + "@test.com"
```

```typescript
const suffix = Date.now().toString(36).toUpperCase().slice(-6);
const projectKey = `E${suffix}`.slice(0, 10);
```

---
## Приложение G. Рубрика ревью для верификаторов

`qa-auto-ver` и `qa-manual-ver` используют эту рубрику для приёмки. Каждый пункт — PASS/FAIL; любой FAIL блокирует `verified: true`.

### G.1. Рубрика `qa-auto-ver`

| # | Критерий | PASS если | FAIL если |
|---|----------|-----------|-----------|
| 1 | Детерминизм | нет sleep/сети/общего состояния; повторный прогон зелёный | есть `waitForTimeout`/`time.Sleep`/реальный fetch |
| 2 | Race | `-race` зелёный | гонки или `-race` не запускался |
| 3 | Привязка TC | имя в колонке «Автотест» существует в коде | «✅» без имени, или имя не находится через `rg` |
| 4 | Осмысленность ассертов | проверяют поведение/контракт | проверяют мок или `nil,nil`; тест без ассертов |
| 5 | Покрытие | новый код покрыт, порог держится | падение ниже 82%/80% |
| 6 | Границы и негатив | есть позитив+негатив+границы | только happy path |
| 7 | Triage при фиксах | записан ROOT=code\|test | assertion подогнан под вывод |
| 8 | Время | нет деградации; медленные оправданы | новый тест тормозит без причины |
| 9 | Артефакты | git чист, профили в TMP | `*.out`/`cover*`/`test-results` в diff |
| 10 | Docs sync | TESTING.md + STATE.md обновлены | колонка не обновлена |

### G.2. Рубрика `qa-manual-ver`

| # | Критерий | PASS если | FAIL если |
|---|----------|-----------|-----------|
| 1 | Полнота TC | покрыты позитив/негатив/границы требования | дыры в сценариях |
| 2 | Техники тест-дизайна | классы/границы/таблицы применены осмысленно | случайный набор шагов |
| 3 | Трассируемость | каждое требование → TC | требование без TC |
| 4 | Воспроизводимость | шаги конкретны, результат наблюдаем | размытые «проверить, что работает» |
| 5 | Формат | все колонки заполнены, ID корректен | пропущены предусловие/ожидаемый результат |
| 6 | Честность статуса | ручные помечены «— (ручной)» | ручной выдан за авто |

### G.3. Как верификатор проверяет привязку (практика)

```bash
# 1. Извлечь имена из колонки «Автотест» в docs/TESTING.md и проверить их существование:
rg -o '`Test\w+`' docs/TESTING.md | tr -d '`' | sort -u | while read fn; do
  rg -q "func $fn" backend/testframework && echo "OK  $fn" || echo "MISS $fn"
done

# 2. Проверить обратную связь: тесты ссылаются на TC-ID
rg -c "TC-[A-Z]+-[0-9]+" backend/testframework | rg -v ":0"

# 3. Прогнать на флейк подозрительный тест
go test -tags hubtest -race -run <Name> -count=20 ./testframework/<pkg>/...
```

`MISS` в шаге 1 = битая привязка MD→код (FAIL критерия 3 рубрики G.1).

---

> **Итог.** Этот playbook — контракт качества QA Hub. `qa-manual` владеет описанием (TC), `qa-auto1/2` — реализацией (автотесты) и привязкой, верификаторы — приёмкой. Пирамида (много unit, мало E2E), детерминизм (без sleep/сети/общего состояния), traceability (requirement → TC → autotest) и Allure-матчинг по `HUB-TC-*` — четыре опоры, на которых держится доверие к зелёному CI. Экономь CI-минуты, чисти артефакты, и никогда не делай тест зелёным, не проверив функционал по спецификации.
