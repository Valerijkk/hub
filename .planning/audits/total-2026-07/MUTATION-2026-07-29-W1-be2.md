# MUTATION — 2026-07-29 · W1 · be2-checker (независимый от авторов)

> Задача 2 / TASK-008 · Проверяющий: **be2** · Авторы rework: be1  
> Контекст W18 FAIL: `.planning/audits/total-2026-07/W18-C3-MUTATION.md` (041, 042, 044, 046)  
> Env: `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` (localhost hub:hub) · `CGO_ENABLED=1` · `-tags hubtest`

## Итог

| ID | ClassIDs | Verdict | Key RED test(s) | Mock-gap? | README |
|----|----------|---------|-----------------|-----------|--------|
| TASK-041 | P2-BE-11 | **PASS** | `TestStartSprintAtomic_Ent_SecondActiveConflicts` | нет (Ent sequential) | Закрыта |
| TASK-042 | P2-BE-12..15 | **PASS** | `TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13` | нет (NATS PublishJSON) | Закрыта |
| TASK-044 | P2-BE-17/18 | **FAIL** | Meili RED ✓; stream `ctx.Err` → GREEN | stream cancel | **Открыта** |
| TASK-046 | P2-BE-22 | **PASS** | `TestComponent_UniqueNamePerProject_P2BE22`, `TestJournalLine_DebitXorCredit_P2BE22` | нет (enttest runtime DDL) | Закрыта |

Production после restore: мутации в `backend/internal/**` и `backend/ent/migrate/schema.go` откатены.

---

## TASK-041 — Sprint poker/start · **PASS**

### ACCEPTANCE
vote after reveal→reject; два Start→один active

### Мутации → RED → restore GREEN

| Guard | Mutation | RED test |
|-------|----------|----------|
| P2-BE-11 second Start | remove `ErrActiveSprintConflict` check in `StartSprintAtomic` | `TestStartSprintAtomic_Ent_SecondActiveConflicts` — ErrorAs typed conflict → UNIQUE wrap / no typed err |

### Notes
- W18 gap (SQLite concurrent + mock embeds) закрыт sequential Ent-тестом.
- UNIQUE на active sprint всё ещё блокирует второй active; тест ловит именно typed `*ErrActiveSprintConflict`.

### GREEN restore
```
go test -tags hubtest -count=1 -timeout=90s -run 'TestStartSprintAtomic_Ent_SecondActiveConflicts|TestPoker_VoteAfterReveal_Rejected' ./testframework/sprint/
→ ok
```

### Mock-gap?
**Нет** для dual-Start ACCEPTANCE (Ent sequential).

---

## TASK-042 — perm/messenger/ws/calls · **PASS**

### ACCEPTANCE
CreateBoard без Edit→403; member.joined без token; revoke→Unregister; guest RL; **via_invite**

### Мутации → RED → restore GREEN

| Guard | Mutation | RED test |
|-------|----------|----------|
| P2-BE-13 via_invite | `"via_invite": true→false` in `AcceptInvite` PublishJSON | `TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13` RED |
| (control) handmade map | same mutation | `TestAcceptInvite_PayloadViaInviteBool_P2_BE_13` stays GREEN |

### Notes
- W18 gap (handmade JSON map) закрыт NATS-capture тестом.
- CreateBoard / revoke / guest — ранее mutation-proven в W18; rework фокус — via_invite.

### GREEN restore
```
go test -tags hubtest -count=1 -timeout=90s -run 'TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13' ./testframework/messenger/
→ ok
```

### Mock-gap?
**Нет** для via_invite (production NATS path).

---

## TASK-044 — AI stream/privacy + search JQL/SSRF · **FAIL**

### ACCEPTANCE
cancel→stream stop; AllowExternalAI default false; OR hits; Meili SSRF blocked

### Мутации → RED → restore GREEN

| Guard | Mutation | Result |
|-------|----------|--------|
| P2-BE-18 Meili SSRF | `validateMeiliHost` → `return nil` + ignore reject in `NewMeiliIndexerFromEnv` | **RED** `TestValidateMeiliHost_BlocksMetadata_P2BE18`, `TestNewMeiliIndexerFromEnv_RejectsMetadataHost_P2BE18` |
| P2-BE-17 stream cancel | remove `ctx.Err()` mid-loop in `adapter_stream.go` | **GREEN** `TestChatCompletionStream_CancelMidRead_P2BE17` — HTTP `NewRequestWithContext` + body Read всё ещё отдают `context.Canceled`; httptest server exits on `r.Context().Done()` |

### Notes
- Meili rework **закрыл** W18 SSRF gap.
- Stream cancel IT **не** mutation-proof для guard `ctx.Err()` в цикле Scan (как заявлено в комментарии теста). Поведение «cancel→stop» держится HTTP-ctx, не mid-loop check → ACCEPTANCE mutation для этого guard — **gap**.
- Prefer default / OR — W18 уже PASS; `./testframework/ai` сейчас не собирается без stub `FindTestCaseByKey` на `githubReviewRepo` (вне scope мутации; не чинил в финальном дереве).

### GREEN restore
```
go test -tags hubtest -count=1 -timeout=90s -run 'TestValidateMeiliHost_BlocksMetadata_P2BE18|TestNewMeiliIndexerFromEnv_RejectsMetadataHost_P2BE18' ./testframework/search/
→ ok
adapter_stream.go restored (ctx.Err present)
```

### Mock-gap?
**Да** (stream cancel): тест не краснеет при снятии production mid-loop `ctx.Err`. → **Открыта**.

---

## TASK-046 — Ent unique/FK/CHECK · **PASS**

### ACCEPTANCE
duplicate (project,name)→unique; XOR debit/credit CHECK

### Мутации → RED → restore GREEN

| Guard | Mutation (runtime DDL `ent/migrate/schema.go`, gitignored generated) | RED test |
|-------|----------|----------|
| unique (project,name) | `component_name_project_components` Unique true→false | `TestComponent_UniqueNamePerProject_P2BE22` (duplicate Create → nil err) |
| XOR CHECK | drop `journalline_debit_xor_credit` Checks map | `TestJournalLine_DebitXorCredit_P2BE22` (debit+credit Create → nil err) |

### Notes
- W18 FAIL был на `TestSchema_P2BE22_UniqueFKCheck` (source grep). Rework добавил `runtime_p2be22_task046_test.go` — слой как TASK-025.
- Orchestrator note в `MUTATION-2026-07-29-W1.md` про schema-only — **устарел**: runtime тесты есть и ловят мутацию DDL.

### GREEN restore
```
go test -tags hubtest -count=1 -timeout=90s -run 'TestComponent_UniqueNamePerProject_P2BE22|TestJournalLine_DebitXorCredit_P2BE22' ./testframework/ent/
→ ok
```
(`migrate/schema.go` Unique/Checks восстановлены вручную — файл не в git.)

### Mock-gap?
**Нет**

---

## Commands (canonical)

```text
cd backend
$env:CGO_ENABLED=1
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"
go test -tags hubtest -count=1 -timeout=90s -run '<TestName>' ./testframework/<pkg>/
```
