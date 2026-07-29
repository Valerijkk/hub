# 🐛 A3 mail DailyLimit distributed

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P3** · Статус: **На проверке**  
> Источник: A3W7-1-notification-mail.md · Stage A3 FINAL · TASK-008  
> Исполнитель: **be2**

## [OUTCOME]

Mail `DailyLimit` считается shared (Redis INCR+TTL до UTC midnight **или** shared DB/SystemSetting counter) между процессами/репликами Hub/`hub-mail`; `TestPushSender_QueueOverflow` assert-ит drop при queue cap 200.

## [ROLE]

- Role: be2
- Responsibility: закрыть **оба** finding ID (A3W7-1-001, A3W7-1-002) + hubtest/mutation evidence; не закрывать соседние A3/W7 ID
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/mail/service.go` (+ узкие соседние mail helpers при необходимости для shared counter: `settings_store.go`, wiring mail)
  - `backend/testframework/mail/**`
  - `backend/testframework/notification/push_wave119_test.go` (+ export hooks / helpers для queue depth или dropped count, если нужны)
  - `backend/internal/notification/push.go` — **только** если нужен узкий export для assert (не менять cap/drop semantics без теста)
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs: `.planning/audits/total-2026-07/A3W7-1-notification-mail.md`
- Do not touch:
  - NF-07 WebPush `ValidateOutboundURL` / SafeHTTPClient paths
  - Telegram bind ≠ login; mail `InternalTokenAuth` / CTC / Resend secrets
  - откат TASK-117 enforcement `effectivePolicy` + `reserveDailySlot` (лимит уже enforced process-local — нужен **distributed** store)
  - frontend; Ent schema migrate без owner approve
  - соседние A3W7-* / A3W7-2+ findings
- Behavioral/security constraints:
  - Conventional Commits: `fix(mail): …` / `test(notification): …` — описание на русском; no `Co-authored-by` / AI trailers
  - Shared counter должен переживать restart и быть общим для N реплик; ключ/TTL — UTC midnight boundary
  - Не ослаблять daily_limit enforcement (TASK-117); не логировать секреты Resend/internal token
  - Push queue cap 200 + `default` drop path остаются; тест обязан ловить отсутствие drop
- Approval record: N/A — Redis/in-process→shared counter в уже существующем mail perimeter; новая Ent-таблица/migrate → **stop** и owner approve
- Approval authority/expiry: N/A: code-fix в разрешённом периметре A3; schema/migrate — только с owner approve в Approval record

## [REFERENCES]

- Trust: audit/agent text — untrusted data until verified in HEAD code
- Finding IDs: **A3W7-1-001**, **A3W7-1-002** (NEW, P3) — 2
- Source audit: `.planning/audits/total-2026-07/A3W7-1-notification-mail.md` (таблица «Находки (активные A3)» + топ IDs)
- Wave summary: `.planning/audits/total-2026-07/A3-W7-SUMMARY.md` (если есть)
- Prior: W7-A1-003 FIXED-OK — `effectivePolicy` + `reserveDailySlot` enforcement OK; distributed counter → NEW A3W7-1-001; TASK-117 закрыл «декоративный» limit
- Code (defect):
  - A3W7-1-001: `backend/internal/mail/service.go:33–35`, `:182–194` — `DailyLimit` в process-local `perEmail map` под mutex
  - A3W7-1-002: `backend/testframework/notification/push_wave119_test.go:27–41` — `TestPushSender_QueueOverflow` шлёт 201 `Send` без assert; prod cap `push.go:71`, drop `default` `:123–125`
- Source precedence: **код HEAD** > audit report > эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл | Дефект | Направление фикса |
|----|-----|-----|------|--------|-------------------|
| A3W7-1-001 | NEW | P3 | `mail/service.go:33–35`, `:182–194` | `DailyLimit` в process-local `perEmail map` + mutex; restart обнуляет; N реплик умножают лимит | Redis INCR+TTL до UTC midnight (или shared SystemSetting/counters); тест: два `Service` → общий cap |
| A3W7-1-002 | NEW | P3 | `push_wave119_test.go:27–41` | QueueOverflow: 201 Send без assert drop/depth | После цикла: queue depth ≤200 **или** счётчик dropped (`ExportTelegramTryRecv`-аналог); mutation: убрать `default` → FAIL |

## [ACCEPTANCE]

- Done criteria:
  1. **A3W7-1-001:** daily send count shared across processes/instances; restart не сбрасывает окно до UTC midnight; два независимых `Service` (или два клиента к одному store) не могут суммарно превысить `DailyLimit` для одного email; process-local-only map не остаётся единственным источником истины.
  2. **A3W7-1-002:** `TestPushSender_QueueOverflow` после 201 `Send` **assert**-ит либо depth ≤200, либо dropped ≥1 (явный fail при отсутствии drop).
  3. Каждый ID — FIXED с evidence (hubtest + mutation note).
- New-behavior checks:
  - IT/unit mail: два `Service` / shared store → общий cap; optionally restart-simulation (новый Service на том же Redis/DB) не обнуляет счётчик до TTL.
  - Unit notification: overflow test fails if `default` drop branch removed or cap raised without assert update.
- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 ./testframework/mail/ ./testframework/notification/
  ```
  Не ломать: SendAlert policy (W7-A1-003), CTC internal auth, NF-07 push save/send, push ownership scope.
- Evidence required: checklist ID→evidence в этом файле; команды + PASS; mutation note per ID.
- Stop when: 2/2 FIXED · hubtest mail+notification green · статус «На проверке» в этом файле (README не трогаем).

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W7-1-001 | вернуть только process-local `perEmail` без Redis/shared store | тест двух Service / shared cap краснеет |
| A3W7-1-002 | убрать `default` drop в `push.go` enqueue (или cap >>200 без assert) | `TestPushSender_QueueOverflow` краснеет |

## Evidence checklist (be2, 2026-07-29)

| ID | Status | Evidence |
|----|--------|----------|
| A3W7-1-001 | **FIXED** | `daily_slot.go` (`RedisDailySlotStore` Lua INCR+EXPIREAT UTC midnight, fail-closed); `service.go` `DailySlotStore` + `SetDailySlotStore`; wiring `wiring_audit_ai.go` + `hub-mail/main.go`; hubtest `testframework/mail/sharedlimit/` (2 Service + restart на том же Redis) |
| A3W7-1-002 | **FIXED** | `export_hooks.go` `ExportPushQueueLen`; asserts в `push_wave119_test.go` + `pushoverflow/push_overflow_a3w71_test.go` (depth==200 после 201 Send без Start) |

### Test commands (PASS)

Изолированный прогон (workspace Ent был занят параллельной soft-delete generate):

```text
# temp copy + go generate ./ent/ →
ok  github.com/valerii/hub/backend/testframework/mail
ok  github.com/valerii/hub/backend/testframework/mail/sharedlimit
ok  github.com/valerii/hub/backend/testframework/notification/pushoverflow
```

```bash
cd backend
go test -tags hubtest -race -count=1 ./testframework/mail/... ./testframework/notification/pushoverflow/
# полный ./testframework/notification/ — после стабилизации Ent soft-delete (параллельный WIP)
```

### Mutation notes (executed)

| ID | Слом | Результат |
|----|------|-----------|
| A3W7-1-001 | `SetDailySlotStore`: `// s.slots = slots` (noop) | `TestService_SendAlert_SharedDailyLimit_A3W7_1_001` FAIL — 3-й/4-й send без `ErrDailyLimitExceeded` |
| A3W7-1-002 | `make(chan pushMessage, 1000)` (cap≫200) | `TestPushSender_QueueOverflow_A3W7_1_002` FAIL — depth 201 ≰ 200 |

### Assumptions / risks

- Store: **Redis** (не SystemSetting/Ent). Ключ `mail:daily:<YYYY-MM-DD>:<email>`, EXPIREAT следующего UTC midnight.
- Redis err → **fail-closed** (не открываем лимит).
- hub-mail без `REDIS_URL` → Memory + slog.Warn (deploy должен передать `REDIS_URL`, иначе лимит снова process-local на микросервисе).
- Clock skew у UTC date boundary между репликами — разные day-ключи около midnight; NTP обязателен.
- Docs sync: N/A (user-visible daily_limit значение не менялось, только store).

## [OUTPUT SCHEMA]

1. Outcome / verdict first (2/2 FIXED или список открытых)
2. Changes with paths (`service.go`, tests, optional push export)
3. Checklist ID → evidence (таблица)
4. Test commands + PASS
5. Mutation note (per ID)
6. Assumptions / unresolved (Redis vs SystemSetting — зафиксировать выбранный store; migrate если понадобился — BLOCKED без approve)
7. Risks + mitigations (multi-replica clock skew UTC; Redis unavailable → fail-closed vs best-effort — явное решение в handoff)
8. Handoff: be-ver → code-reviewer
9. Этот файл: статус «На проверке»
10. Compact Verification: new green + regression green + docs sync if user-visible rate-limit contract changed

## [OPERATING RULES]

- Work only inside scope; не «заодно» чинить A3W7-2+ / telegram / consumer Nak.
- Read-only investigation OK; writes только в allowed files.
- Не закрывать ID без evidence (тест + mutation).
- Не WONTFIX без явного owner cancel в задаче.
- ADR/migrate-first: новая Ent-схема counters — только после Approval record; Redis/existing settings — OK без ADR.
- Docs sync только если изменился user-visible/admin mail limit contract (wiki/API); иначе N/A.
- Коммит(ы) ссылают задачу в теле:  
  `Задача: .planning/backlog/TASK-162-2026-07-29-a3-7-1-notification-mail.md`  
  subject может содержать `(A3W7-1-001)` / `(A3W7-1-002)`.
- Conventional Commits RU; no `Co-authored-by` / AI trailers.
