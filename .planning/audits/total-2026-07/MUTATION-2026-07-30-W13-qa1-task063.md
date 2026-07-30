# MUTATION W13 slot 4 refill — TASK-063 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/archive/legacy/TASK-063-be-p3-errors-input-hygiene.md` `[MUTATION CRITERIA]`  
> Classes: **P3-BE-08**, **P3-BE-09** · Коммит: **не создавался** (инструкция слота).

## Verdict

**13/13 PROVEN** (primary runnable) · **3 GAP_NOTED** (sibling paths) · **0 NOT_PROVEN** · TASK-063 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| W2-A1-011 | **PROVEN** | `TestWebhookAllure_InvalidJSON_StableMessage_P3BE08`; `TestWebhookJUnit_InvalidXML_StableMessage_P3BE08` | 400 msg → `err.Error()` | handler TMS webhook |
| W5-A1-008 | **PROVEN** | `TestSuggestComposer_PrefixTooShort` | sentinel → plain `fmt.Errorf` | service AI (handler HTTP **GAP**) |
| W5-A4-011 | **PROVEN** | `TestHandler_CreateFilter_SafeErrorBody_P3BE08` | 500 body → `err.Error()` | handler search create (**list GAP**) |
| W6-A4-011 | **PROVEN** | `TestHandler_requireAdmin_RepoError_500_P3BE08` | any IsSystemAdmin err → 401 | handler audit (**mock≠Ent**) |
| W7-A2-011 | **PROVEN** | `TestMapServiceError_AccessDenied_P3BE08` | map → `"admin access required"` | handler map HR |
| W7-A5-008 | **PROVEN** | `TestMapServiceError_NotOwner_P3BE08` | map NotOwner → admin msg; service → ErrForbidden | service/handler marketplace |
| W1-A5-011 | **PROVEN** | `TestCreateInvite_FrontendBase_P3BE09`; `…_FrontendURLEnv_P3BE09` | `toInviteResponse(..., "")` | service project |
| W2-A4-012 | **PROVEN** | `TestDetectImportFormat_CaseInsensitive_P3BE09` | HasSuffix без ToLower | importer TMS |
| W2-A5-008 | **PROVEN** | `TestCreateColumn_InvalidColType_P3BE09` | drop `isValidColumnType` | service bases |
| W2-A5-011 | **PROVEN** | `TestUpdateRecord_VersionZero_P3BE09`; `TestUpdateBase_OptimisticLock_VersionEQ_W2A5004` | drop `version < 1` | service bases |
| W3-A1-011 | **PROVEN** | `TestQueryInt` / `negative falls back to default` | drop clamp `v < 1 \|\| v > max` | handler helper messenger |
| W5-A1-009 | **PROVEN** | `TestAdminUpdateProvider_NilUUID_P3BE09` | allow `uuid.Nil` | handler AI admin (**Test/user GAP**) |
| W7-A5-010 | **PROVEN** | `TestListApproved_InvalidCategory_P3BE09` | drop category allowlist | service marketplace |

Production restored (`git checkout --` touched guards) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W6-A4-011` · `GAP IDs: W5-A1-008 handler HTTP; W5-A4-011 listFilters; W5-A1-009 TestProvider+user Update`.

## Per-ID evidence

### W2-A1-011 — TMS webhook stable 400

```text
BASELINE Webhook.*StableMessage_P3BE08 → PASS (~8.1s)

BREAK: backend/internal/tms/handler_webhook.go
  "невалидный JSON" / "невалидный JUnit XML" → err.Error() // MUTATED W2-A1-011
→ FAIL TestWebhookAllure_InvalidJSON_StableMessage_P3BE08
        body leaks "invalid character"; NotContains RED
→ FAIL TestWebhookJUnit_InvalidXML_StableMessage_P3BE08
        body leaks XML syntax error
RESTORE → PASS (~9.1s)
```

### W5-A1-008 — SuggestComposer sentinel (+ handler GAP)

```text
BASELINE TestSuggestComposer_PrefixTooShort → PASS (~10.3s)

BREAK: backend/internal/ai/composer_suggest.go
  return ErrInvalidPrefix → fmt.Errorf("prefix too short") // MUTATED W5-A1-008
→ FAIL ErrorIs ErrInvalidPrefix
RESTORE → PASS (~7.2s)

GAP_NOTED — handler HTTP errors.Is map: нет Exact RED (criteria).
```

### W5-A4-011 — createFilter safe 500 (+ list GAP)

```text
BASELINE CreateFilter_SafeErrorBody_P3BE08 → PASS (~6.2s)

BREAK: backend/internal/search/handler.go createFilter
  stable msg → err.Error() // MUTATED W5-A4-011
→ FAIL Equal "не удалось создать фильтр"; NotContains "create failed"
RESTORE → PASS (~7.2s)

GAP_NOTED — listFilters Exact RED отсутствует (criteria).
```

### W6-A4-011 — requireAdmin 500 vs 401 (**mock≠Ent**)

```text
BASELINE requireAdmin_*_P3BE08 → PASS (~8.7s)

BREAK: backend/internal/audit/handler.go
  ErrUserNotFound→401 / other→500 → any err→401 // MUTATED W6-A4-011
→ FAIL TestHandler_requireAdmin_RepoError_500_P3BE08
        expected 500, got 401
RESTORE → PASS (~7.6s)

mock≠Ent: adminErrRepo; handler guard killed → PROVEN (не reopen).
```

### W7-A2-011 — HR AccessDenied map

```text
BASELINE MapServiceError_AccessDenied_P3BE08 → PASS (~7.6s)

BREAK: backend/internal/hr/handler.go
  "leave access denied" → "admin access required" // MUTATED W7-A2-011
→ FAIL Equal "leave access denied"
RESTORE → PASS (~5.9s)
```

### W7-A5-008 — marketplace NotOwner

```text
BASELINE MapServiceError_NotOwner + DeleteListing_Forbidden → PASS (~7.1s)

BREAK: marketplace/handler.go map + service.go stranger Delete
  NotOwner msg → admin; ErrNotOwner → ErrForbidden // MUTATED W7-A5-008
→ FAIL TestMapServiceError_NotOwner_P3BE08 Equal "not listing owner"
RESTORE → PASS (~6.4s)
```

### W1-A5-011 — invite frontend base

```text
BASELINE CreateInvite_Frontend*_P3BE09 → PASS (~10.2s)

BREAK: backend/internal/project/invite.go
  toInviteResponse(inv, resolveFrontendURL()) → ("",) // MUTATED W1-A5-011
→ FAIL both FrontendBase / FrontendURLEnv (Link missing https://… base)
RESTORE → PASS (~7.8s)
```

### W2-A4-012 — detectImportFormat case

```text
BASELINE DetectImportFormat_CaseInsensitive_P3BE09 → PASS (~9.9s)

BREAK: backend/internal/tms/importer.go
  ToLower+HasSuffix → HasSuffix(filename) only // MUTATED W2-A4-012
→ FAIL .CSV/.Json/.XML/.XLSX → ""
RESTORE → PASS (~9.2s)
```

### W2-A5-008 — ColType allowlist

```text
BASELINE CreateColumn_InvalidColType_P3BE09 → PASS (~6.9s)

BREAK: backend/internal/bases/service.go
  drop isValidColumnType gate // MUTATED W2-A5-008
→ FAIL ErrorIs ErrInvalidInput (got non-InvalidInput / Ent path)
RESTORE → PASS (~6.1s)
```

### W2-A5-011 — version < 1

```text
BASELINE UpdateRecord_VersionZero + UpdateBase_…VersionEQ → PASS (~6.9s)

BREAK: backend/internal/bases/service.go UpdateBase+UpdateRecord
  drop req.Version < 1 // MUTATED W2-A5-011
→ FAIL both ErrorIs ErrInvalidInput on version=0
RESTORE → PASS (~6.2s)
```

### W3-A1-011 — queryInt clamp

```text
BASELINE TestQueryInt → PASS (~7.6s)

BREAK: backend/internal/messenger/handler.go
  clamp v<1||v>max removed // MUTATED W3-A1-011
→ FAIL negative_falls_back_to_default expected 25 got -5
RESTORE → PASS (~8.9s)
```

### W5-A1-009 — admin nil UUID (+ siblings GAP)

```text
BASELINE AdminUpdateProvider_NilUUID_P3BE09 → PASS (~10.9s)

BREAK: backend/internal/ai/admin_handler.go UpdateProvider
  id == uuid.Nil check dropped // MUTATED W5-A1-009
→ FAIL expected 400 "invalid id"; got provider not found
RESTORE → PASS (~13.9s)

GAP_NOTED — TestProvider + user UpdateProvider nil UUID: нет Exact RED (criteria).
```

### W7-A5-010 — category allowlist

```text
BASELINE ListApproved_InvalidCategory_P3BE09 → PASS (~7.8s)

BREAK: backend/internal/marketplace/service.go ListApproved
  drop isValidCategory // MUTATED W7-A5-010
→ FAIL ErrorIs ErrInvalidInput but got nil
RESTORE → PASS (~7.3s)
```

```text
leftover_mutated=[]
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=180s ./testframework/tms/ -run "TestWebhook(Allure|JUnit)_Invalid.*StableMessage_P3BE08"
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/ai/ -run TestSuggestComposer_PrefixTooShort
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/search/ -run TestHandler_CreateFilter_SafeErrorBody_P3BE08
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/audit/ -run "TestHandler_requireAdmin_(RepoError_500|UserNotFound_401)_P3BE08"
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/hr/ -run TestMapServiceError_AccessDenied_P3BE08
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/marketplace/ -run "TestMapServiceError_NotOwner_P3BE08|TestHandler_DeleteListing_Forbidden_W7A5016"
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/project/ -run "TestCreateInvite_Frontend(Base|URLEnv)_P3BE09"
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/tms/ -run TestDetectImportFormat_CaseInsensitive_P3BE09
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/bases/ -run TestCreateColumn_InvalidColType_P3BE09
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/bases/ -run "TestUpdateBase_OptimisticLock_VersionEQ_W2A5004|TestUpdateRecord_VersionZero_P3BE09"
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/messenger/ -run TestQueryInt
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/ai/ -run TestAdminUpdateProvider_NilUUID_P3BE09
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/marketplace/ -run TestListApproved_InvalidCategory_P3BE09
```

Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task063.py`  
JSON: `_mut-batch-w13-task063.json`.

## Gaps

- **W5-A1-008** handler HTTP `errors.Is` — **GAP_NOTED** (service path PROVEN).
- **W5-A4-011** listFilters Exact RED — **GAP_NOTED** (createFilter PROVEN).
- **W5-A1-009** TestProvider + user UpdateProvider nil UUID — **GAP_NOTED** (admin Update PROVEN).
- **W6-A4-011** mock≠Ent — flagged; handler mutation still RED → **PROVEN**, не reopen.

## Handoff

- Evidence: этот файл · TASK-063 **Закрыта** · README sync  
- be-ver → code-reviewer (fix/tests already FIXED; mutation PROVEN + GAP_NOTED siblings)  
- Коммит: не создавался
