# W19-A2 — пробелы evidence у задач TASK-009…056

> 2026-07-28 · Агент **W19-A2** · ветка `cursor-grok-4.5` · **no push master**  
> Задача: [TASK-008](../../backlog/TASK-008-total-audit-program.md) § Stage C / W19  
> Фокус: задачи в статусе **«Исправлена»** без адекватного evidence  
> Режим: read-only по product code; сверка task body ↔ `W18-*-MUTATION.md`  
> Перепроверка: повторный проход по README + телам TASK-009…056 + всем 6 mutation-логам

## Определение «адекватного evidence»

По `[ACCEPTANCE]` задач Stage B и протоколу W18
([`W18-AGENT-ASSIGNMENT.md`](W18-AGENT-ASSIGNMENT.md)):

| # | Артефакт | Где допускается |
|---|----------|-----------------|
| 1 | **Имя теста** (Go `Test…` / Vitest case title) | тело TASK и/или `W18-*-MUTATION.md` |
| 2 | **SHA коммита** фикса | тело TASK и/или mutation-лог |
| 3 | **Mutation log** (RED→restore→GREEN) | тело TASK и/или `W18-C{1,2,2b,3,4,5}-MUTATION.md` |

**Evidence-gap** = статус README/файла ∈ {«Исправлена», «На проверке»}
**и** отсутствует **хотя бы один** элемент триады при union(task body, W18 mutation).

«Закрыта без evidence» (формулировка TASK-008 W19) трактуется так же:
статус «Исправлена» без полной триады.

### Как считали (ловушки парсинга)

- W18-C1 в таблице пишет `009`/`010`, не всегда `TASK-009` — mutation-log
  для C1 PASS всё равно есть (короткий ID в колонке TASK).
- FE (C4/C5): имя RED-теста = **Vitest title** в backticks, не `Test…`.
- Строка ACCEPTANCE `Evidence: … SHA коммитов` — **шаблон**, не доказательство;
  в счёт SHA не входит.
- Фактических hex SHA (`[0-9a-f]{7,40}` привязанных к fix-коммиту) в телах
  TASK-009…056 и во всех `W18-*-MUTATION.md` — **0**.

## Сводка

| Метрика | Значение |
|--------|---------:|
| TASK-009…056 всего | 48 |
| «Исправлена» (W18 PASS) | 39 |
| «Открыта» (W18 FAIL) | 9 |
| Ещё явно «На проверке» в теле | 1 (TASK-038 Progress; README = Исправлена) |
| **Evidence-gap (incomplete triad)** | **39** |
| Gap: нет SHA | **39 / 39** |
| Gap: нет имени теста (union) | **0** |
| Gap: нет mutation log (union) | **0** |
| Нет имени теста **в теле TASK** (есть только в W18) | **28** |
| Полностью пустая триада (union) | **0** |

**Вердикт:** все 39 задач, помеченных «Исправлена» после W18 PASS, имеют
имя теста и mutation-лог (через `W18-*-MUTATION.md` и/или секцию в TASK),
но **ни одна не фиксирует SHA коммита фикса** — системный пробел evidence
относительно ACCEPTANCE («Evidence: … SHA коммитов»).

---

## Evidence-gap IDs (39)

```
TASK-009, TASK-010, TASK-011, TASK-012, TASK-013,
TASK-015, TASK-016, TASK-017, TASK-018, TASK-020,
TASK-021, TASK-022, TASK-023, TASK-024, TASK-025,
TASK-026, TASK-027, TASK-028, TASK-029, TASK-030,
TASK-031, TASK-032, TASK-033, TASK-034, TASK-035,
TASK-036, TASK-037, TASK-038, TASK-039, TASK-040,
TASK-043, TASK-047, TASK-048, TASK-049, TASK-050,
TASK-051, TASK-053, TASK-055, TASK-056
```

(Не в списке — W18 FAIL → «Открыта»: **014, 019, 041, 042, 044, 045, 046, 052, 054**.)

### Покрытость mutation-логом (не gap)

| W18 артефакт | PASS IDs в списке gap |
|--------------|----------------------|
| [W18-C1-MUTATION.md](W18-C1-MUTATION.md) | 009–013, 017, 018, 021 |
| [W18-C2-MUTATION.md](W18-C2-MUTATION.md) | 015, 016, 020, 022, 024 |
| [W18-C2b-MUTATION.md](W18-C2b-MUTATION.md) | 036, 037, 039, 043 |
| [W18-C3-MUTATION.md](W18-C3-MUTATION.md) | 023, 025, 038, 040 |
| [W18-C4-MUTATION.md](W18-C4-MUTATION.md) | 026–031, 035, 047, 049, 050 |
| [W18-C5-MUTATION.md](W18-C5-MUTATION.md) | 032–034, 048, 051, 053, 055, 056 |

---

## Детализация пробелов

### 1. SHA отсутствует везде (P0 evidence gap) — 39/39

В телах TASK-009…056 и во всех `W18-*-MUTATION.md` нет строк вида
`SHA: <hex>` / `commit <hex>` / привязки фикса к хешу.

ACCEPTANCE типовой задачи Stage B требует: *«Evidence: вывод тестов, SHA коммитов»*.
W18 фиксирует RED/GREEN и имя теста, но **не** SHA Stage B коммита.

### 2. Имя теста только в mutation-логе, не в теле TASK — 28

Протокол W18: PASS → «Исправлена» **(+ evidence в task file)**.
У следующих файлов в теле нет конкретного `Test…` / Vitest case-name
(есть ссылка на W18 и/или ClassID / «PASS»):

`015, 016, 020, 022, 024, 025, 026–035, 036, 037, 039, 043, 047–051, 053, 055, 056`

Типичный thin evidence (пример C5):

```markdown
## [W18-C5 EVIDENCE]
- Mutation: PASS — см. …/W18-C5-MUTATION.md
- Статус: Исправлена
```

Имена тестов **есть** в соответствующем `W18-*-MUTATION.md` (колонка RED test)
→ union-триада по пункту «test» закрыта; это **thin in-file evidence**, не полный пробел.

Обратный набор (имя теста уже в теле TASK):  
`009–013, 017, 018, 021, 023, 038, 040` (11 задач; в основном C1 + часть C3).

### 3. Статус-дрейф «На проверке» — TASK-038

| Поле | Значение |
|------|----------|
| README / frontmatter | **Исправлена** · W18-C3 PASS (partial residual OK) |
| Секция Progress в теле | **На проверке (partial)** — остаток Source IDs |

W18-C3 закрыл mutation-ready subset; Progress не синхронизирован.
SHA по-прежнему нет → в общем счётчике gap.

---

## Приложение: gap × RED test (для дописывания в TASK)

SHA у всех = **нет**. Колонка «в теле» = есть ли `Test…`/Vitest title в файле задачи.

| ID | В теле | W18 | RED test (из mutation-лога) |
|----|:------:|-----|------------------------------|
| TASK-009 | да | C1 | `TestProjectRoutes_RBAC_OutsiderForbidden/delete` |
| TASK-010 | да | C1 | `TestPromoteToIssue_ValidationAndFlow` |
| TASK-011 | да | C1 | `TestContainsEvent_UIShortName_MatchesProjectSubject` |
| TASK-012 | да | C1 | `TestWebhookDispatcher_dispatchEvent_ProjectScope_Denial` |
| TASK-013 | да | C1 | `TestDeleteCustomField_RejectsForeignProject_P1BE02` |
| TASK-015 | нет | C2 | `TestDeleteSuite_RejectsForeignProject_P1BE10` |
| TASK-016 | нет | C2 | `TestHandler_ImportTestCases_OversizedBody_413` |
| TASK-017 | да | C1 | `TestRemoveBookmark_RejectsForeignOwner_P1BE15` |
| TASK-018 | да | C1 | `TestHub_UnregisterBeforeRegister_NoZombie_P1_BE_21` |
| TASK-020 | нет | C2 | `TestWebhookCreate_SecretOnce_P1BE27` |
| TASK-021 | да | C1 | `TestCheckContentPolicy_BlocksJailbreakInSystemRole` |
| TASK-022 | нет | C2 | `TestPostgresSearch_Channels_SoftDeleteAndPrivateScope_P1BE32` |
| TASK-023 | да | C3 | `TestDemoteAdmin_RequirePermission403_P1BE33` |
| TASK-024 | нет | C2 | `TestNotificationConsumer_NakOnRepoErr_P1_BE_35` |
| TASK-025 | нет | C3 | `TestCoffeeEnrollment_UniquePerUser_TASK025` |
| TASK-026 | нет | C4 | `api client > one-retry: second 401…` |
| TASK-027 | нет | C4 | `subscribes to call:{callId}…` |
| TASK-028 | нет | C4 | `calls joinCall(call.id) on Войти…` |
| TASK-029 | нет | C4 | `sendMessage includes replyToMessage.id…` |
| TASK-030 | нет | C4 | `board page source has no monitorForElements dual mutate` |
| TASK-031 | нет | C4 | `shows loading before response…` |
| TASK-032 | нет | C5 | `listTrash calls GET /projects/trash…` |
| TASK-033 | нет | C5 | `ignores late getMe setUser after logout` |
| TASK-034 | нет | C5 | `collapsed SidebarUserProfile is button without nested link` |
| TASK-035 | нет | C4 | `hides comparison table after env change…` |
| TASK-036 | нет | C2b | `TestStructValidator_EmptyTitleEmail_Mutation/with_validator_rejected` |
| TASK-037 | нет | C2b | `TestDeleteComment_RejectsForeignAuthor_Ent_TASK037` |
| TASK-038 | да | C3 | `TestEntRepository_MoveIssue_OptimisticLock_W1A3004` |
| TASK-039 | нет | C2b | `TestSoftDeleteSuite_KeepsRowWithDeletedAt` |
| TASK-040 | да | C3 | `TestWSHandler_NilAuthorizer_RefusesUpgrade` |
| TASK-043 | нет | C2b | `TestLoad_productionRejectsDeriveDevKey_W4A4009` |
| TASK-047 | нет | C4 | `shows alert + retry instead of empty sidebar…` |
| TASK-048 | нет | C5 | `keeps dialog open when onConfirm rejects` |
| TASK-049 | нет | C4 | `PinnedMessages trigger has accessible name` |
| TASK-050 | нет | C4 | Presence tokens / `myTasks hides…` / webhooks viewer |
| TASK-051 | нет | C5 | `gif-picker clears debounceRef on unmount` |
| TASK-053 | нет | C5 | `refresh failure best-effort POSTs…` |
| TASK-055 | нет | C5 | `sidebar-user does not claim fake Online` |
| TASK-056 | нет | C5 | `locale=en does not return Russian…` |

---

## Не gap (контроль)

| Группа | IDs | Почему не evidence-gap |
|--------|-----|-------------------------|
| W18 FAIL → Открыта | 014, 019, 041, 042, 044, 045, 046, 052, 054 | Статус не «Исправлена»; mutation-лог документирует FAIL + причину |
| P3 TASK-057…078 | — | Вне scope W19-A2 (009…056) |

Сверка PASS/FAIL: [`W18-SUMMARY.md`](W18-SUMMARY.md) — **39 / 9**.

---

## Рекомендации (handoff)

1. **Дописать SHA** Stage B коммитов в тела 39 TASK (или таблицу в W18-SUMMARY) —
   один проход `git log --grep=TASK-0NN`.
2. Для 28 «thin» TASK — скопировать **имя RED-теста** из приложения выше
   (или из W18) в секцию evidence файла.
3. TASK-038: либо убрать «На проверке» из Progress, либо вернуть README в
   «На проверке (partial)» до закрытия residual Source IDs.
4. Не считать «Исправлена» финальным закрытием TASK-008 Stage C, пока SHA не в evidence.

---

## Метод сверки

1. Статусы: `.planning/backlog/README.md` + frontmatter/`**Статус:**` в TASK-009…056.
2. Триада: regex по телу TASK + строки/секции `TASK-NNN` / короткий `NNN` во всех
   `W18-C{1,2,2b,3,4,5}-MUTATION.md` (включая Vitest titles в колонке RED test).
3. Сверка итога PASS/FAIL с [`W18-SUMMARY.md`](W18-SUMMARY.md) (39 / 9).
4. Контроль: grep hex SHA по TASK + W18 mutation — пусто (кроме шаблона «SHA коммитов»).

---

*Артефакт W19-A2 · 2026-07-28 · count=**39** · gap IDs выше*
