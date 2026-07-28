# W18-C1 — мутация BE (be2 runtime)

> 2026-07-28 · Автор Stage B: **be1** · Проверяющий: **be2 (C1)** · Правило: проверяет ≠ автор  
> Задачи: TASK-009, 010, 011, 012, 013, 017, 018, 021 (8)

## Итог

| Метрика | Значение |
|---------|----------|
| PASS | **8 / 8** |
| FAIL | **0** |
| FAIL IDs | — |

Все мутации: сломан guard в `backend/internal/**` → hubtest RED → restore → GREEN.  
Ни один кейс не «guard в Ent/repo при mock-only тесте» → insufficient нет.

## Протокол

1. `[ACCEPTANCE]` → критерий мутации  
2. Слом ровно этой защиты в production  
3. `cd backend && go test -tags hubtest -race -count=1 -timeout=180s ./testframework/<pkg>/ -run <Test>`  
4. RED → restore → GREEN  
5. Mock-only Ent-gap? → FAIL даже при RED на mock  
6. PASS → «Исправлена»

## Результаты

| TASK | Mutation site | Test RED | Test GREEN | Mock-only? | Verdict PASS/FAIL | Notes |
|------|---------------|----------|------------|------------|-------------------|-------|
| 009 | `project/handler.go` Delete route: strip `RequirePermission` | `TestProjectRoutes_RBAC_OutsiderForbidden/delete` — expected 403 got **200** | ✅ same suite | no | **PASS** | Fiber + real RBACService |
| 010 | `messenger/service_promote.go`: remove `promoteCheck` block | `TestPromoteToIssue_ValidationAndFlow/{nil_promote_checker_forbidden,denied_by_promote_checker}` — error expected, got nil | ✅ | no | **PASS** | Guard in service; mock repo OK |
| 011 | `apigateway/webhook.go` `eventMatchesSubscription` → exact-only | `TestContainsEvent_UIShortName_MatchesProjectSubject`; `TestWebhookDispatcher_dispatchEvent_EnqueuesOnUIShortMatch` | ✅ | no | **PASS** | Real containsEvent via export |
| 012 | `apigateway/webhook.go` `dispatchEvent`: drop project-scope filter | `TestWebhookDispatcher_dispatchEvent_ProjectScope_Denial` — A+B endpoints enqueued | ✅ | no | **PASS** | Guard in dispatcher |
| 013 | `tracker/repo_custom_fields.go` `DeleteCustomField`: drop `HasProjectWith` | `TestDeleteCustomField_RejectsForeignProject_P1BE02` — expected NotFound, got nil | ✅ | no | **PASS** | Ent sqlite; not mock-only |
| 017 | `messenger/ent_repository.go` `RemoveBookmark`: drop `UserIDEQ` | `TestRemoveBookmark_RejectsForeignOwner_P1BE15` — foreign delete OK | ✅ | no | **PASS** | Ent layer |
| 018 | `ws/hub.go` `registerClient`: drop done-guard; `calls/scheduled.go` Update: skip `authorizeScheduledCallMutate` | `TestHub_UnregisterBeforeRegister_NoZombie_P1_BE_21` (zombie); `TestScheduledCall_Update_RejectsStranger_P1_BE_22` (nil err) | ✅ both | no | **PASS** | Real Hub + Ent calls fixture |
| 021 | `ai/content_policy.go`: only `role==user`; `ai/messenger_bot.go` `shouldRespond`: drop bot membership | `TestCheckContentPolicy_BlocksJailbreakInSystemRole`; `TestMessengerBot_shouldRespond` | ✅ | no | **PASS** | Real funcs via export |

## Команды (канон)

```bash
cd backend
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/project/ -run 'TestProjectRoutes_RBAC_OutsiderForbidden'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/messenger/ -run 'TestPromoteToIssue_ValidationAndFlow/(nil_promote_checker_forbidden|denied_by_promote_checker)'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/apigateway/ -run 'TestContainsEvent_UIShortName_MatchesProjectSubject|TestWebhookDispatcher_dispatchEvent_EnqueuesOnUIShortMatch|TestWebhookDispatcher_dispatchEvent_ProjectScope_Denial'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/tracker/ -run 'TestDeleteCustomField_RejectsForeignProject_P1BE02'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/messenger/ -run 'TestRemoveBookmark_RejectsForeignOwner_P1BE15$'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/ws/ -run 'TestHub_UnregisterBeforeRegister_NoZombie_P1_BE_21$'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/calls/ -run 'TestScheduledCall_Update_RejectsStranger_P1_BE_22$'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/ai/ -run 'TestCheckContentPolicy_BlocksJailbreakInSystemRole|TestMessengerBot_shouldRespond$'
```

Production code restored after each RED; working tree production files clean for W18-C1 mutations.
