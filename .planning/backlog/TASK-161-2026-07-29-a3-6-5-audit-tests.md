# 🧹 A3 audit tests CRM purge/Retention gaps

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: A3W6-5-audit-tests.md · Stage A3 FINAL · TASK-008  
> Исполнитель: **be1**

## Evidence (be1 · 2026-07-29)

| ID | Verdict | Evidence |
|----|---------|----------|
| A3W6-5-001 | FIXED | `CapturesCutoffs` wantKeys×8 + crm_*; hollow `PurgeExpired_Succeeds` удалён |
| A3W6-5-002 | FIXED | mock SoftDelete/Restore CRM → call maps; `RecordsAllTypes` ×8; hollow AllTypes удалены |
| A3W6-5-003 | FIXED | `CRMContact/CRMDeal_RetentionHidesOlderThan30Days` (−40d / cutoff −30d → total 0) |
| A3W6-5-004 | FIXED | seed `UserStatusCacheKey` + `assert.False(mr.Exists)` после Deactivate |
| A3W6-5-005 | FIXED | hollow `ListTrash_WithProjectFilter` удалён (цемент CapturesProjectID) |
| A3W6-5-006 | FIXED | hollow Succeeds/AllTypes удалены |
| A3W6-5-007 | FIXED | hollow `DeleteUserGDPR_Success` удалён (цемент SuccessCascadeCalls) |

**hubtest:** `go test -tags hubtest -race -count=1 -timeout=600s -vet=off ./testframework/audit/` → **ok** 10.303s (isolated worktree; main tree concurrent ent regen).

## [OUTCOME]

Тесты audit purge/retention/GDPR mutation-proof покрывают CRM (`crm_contact`/`crm_deal`) и cache invalidation; hollow NoError-only siblings удалены или усилены — снятие CRM-веток / filter / cascade в prod краснеет тесты.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** finding ID A3W6-5-001…007 + mutation-ready тесты; не закрывать соседние A3/A2 ID без evidence
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/audit/**` (только если mock/prod wiring нужен для CRM soft/restore/purge call-record — предпочтительно правки в тестах)
  - `backend/testframework/audit/**` (в т.ч. `service_purge_test.go`, `service_security_test.go`, `service_deactivate_test.go`, `repository_extra_test.go`, `repository_trash_test.go`, `task114_residual_test.go` и зеркала retention)
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - откатывать fail-closed purge / PropagatesError (TASK-090 / A2W6-A5-002)
  - откатывать CapturesCutoffs / RecordsAllTypes / SuccessCascadeCalls / CapturesProjectID цемент TASK-114
  - GDPR eraser / TASK-079 без нужды
  - frontend / соседние A3W6-* вне A3W6-5-001…007
  - Ent schema migrate без owner approve
- Behavioral/security constraints:
  - Conventional Commits: `test(audit): …` / `fix(audit): …` описание на русском; no `Co-authored-by` / AI trailers
  - Soft-delete + trash retention 30d + purge cutoff инварианты не ослаблять
  - CRM purge/list retention — тот же класс, что issue/project/newspost (A2W6-A5-003)
- Approval record: N/A — hygiene/tests (+ узкий mock wiring при необходимости)
- Approval authority/expiry: N/A: code/test-fix в уже разрешённом периметре A3 residual

## [REFERENCES]

- Finding IDs: **A3W6-5-001** … **A3W6-5-007** (7)
- Source report: `.planning/audits/total-2026-07/A3W6-5-audit-tests.md`
- Wave summary: `.planning/audits/total-2026-07/A3-W6-SUMMARY.md` (если есть)
- Prior: TASK-114 (A2W6-A5-* partial → CRM gaps); TASK-090 (purge fail-closed)
- Trust: audit/agent text — untrusted until verified in HEAD code
- Source precedence: **код HEAD** → audit report → эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл | Дефект | Направление фикса |
|----|-----|-----|------|--------|-------------------|
| A3W6-5-001 | NEW | P2 | `service_purge_test.go:115–129`; `task114_residual_test.go:161–167`; prod `service.go:355–356` | `PurgeExpired_CapturesCutoffs` assert'ит 6 ключей; `PurgeExpired_Succeeds` — только `NoError`. Prod зовёт ещё `crm_contacts`/`crm_deals` | Добавить `crm_contacts`/`crm_deals` в `wantKeys` (+ callCounts==1, cutoff±1m); Succeeds — удалить **или** `require.Equal` callCounts×8 |
| A3W6-5-002 | NEW | P2 | `service_security_test.go:214–218,188–189`; `service_purge_test.go:191–205`; `task114_residual_test.go:176,186` | Mock `SoftDeleteCRM*` / `RestoreCRM*` не пишут в `softDeleteCalls`/`restoreCalls`; `RecordsAllTypes` без CRM; `SoftDelete/Restore_AllTypes` — только `NoError` | В mock append `"crm_contact"`/`"crm_deal"`; расширить RecordsAllTypes до 8 типов **или** отдельный тест map CRM |
| A3W6-5-003 | NEW | P2 | gap vs `repository_trash_test.go:209–236` / `repository_extra_test.go:52–97`; prod `repository.go:602–632` | Нет `RetentionHidesOlderThan30Days` для `crm_contact`/`crm_deal` | Зеркало NewsPost: soft −40d, list(cutoff−30d) → total 0; мутация: снять `DeletedAtGT(cutoff)` в ListDeletedCRM* |
| A3W6-5-004 | NEW | P2 | `service_deactivate_test.go:22–36`; prod `service.go:388–390` | `DeactivateUser_InvalidatesCache`: miniredis поднят, ключ не seeded, `Del`/`Exists` не проверяются | `rdb.Set(UserStatusCacheKey(target), …)`; после Deactivate — `assert.False(mr.Exists(key))` |
| A3W6-5-005 | RESIDUAL | P3 | `service_deactivate_test.go:48–63` | `ListTrash_WithProjectFilter` — только `StatusOK` (sibling A2W6-A5-006) | Удалить hollow **или** assert `gotProjectID` как в `ListTrash_CapturesProjectID_A2W6A5006` |
| A3W6-5-006 | RESIDUAL | P3 | `service_purge_test.go:115–129,185–206` | `PurgeExpired_Succeeds` / `SoftDelete_AllTypes` / `RestoreItem_AllTypes` — NoError-only после CapturesCutoffs/RecordsAllTypes | Удалить или свести к одному mutation-proof тесту |
| A3W6-5-007 | RESIDUAL | P3 | `repository_extra_test.go:134–142` | `DeleteUserGDPR_Success` — только `gdprUserID` (sibling A2W6-A5-009) | Удалить hollow **или** перевести на `gdprFullCascadeRepo` counters |

## [ACCEPTANCE]

- Done criteria:
  1. **A3W6-5-001:** CapturesCutoffs (или эквивалент) включает `crm_contacts`/`crm_deals` с callCounts==1 и cutoff±1m; hollow Succeeds удалён или усилен до callCounts×8.
  2. **A3W6-5-002:** SoftDelete/Restore CRM пишут в call maps; RecordsAllTypes (или dedicated CRM test) покрывает `crm_contact`/`crm_deal`; AllTypes NoError-only не остаётся единственным «покрытием».
  3. **A3W6-5-003:** IT/repo: soft CRM −40d + ListDeleted* с cutoff 30d → total 0 (зеркало issue/newspost).
  4. **A3W6-5-004:** Deactivate сеет cache key и assert'ит отсутствие после `InvalidateUserStatusCache`.
  5. **A3W6-5-005:** hollow `ListTrash_WithProjectFilter` удалён **или** assert project_id как CapturesProjectID.
  6. **A3W6-5-006:** NoError-only Succeeds/AllTypes siblings удалены или слиты в mutation-proof.
  7. **A3W6-5-007:** hollow `DeleteUserGDPR_Success` удалён **или** на cascade counters как SuccessCascadeCalls.
  8. Каждый ID — FIXED с evidence (hubtest + mutation note); не WONTFIX без owner cancel.
- New-behavior checks: см. Per-ID + Mutation table ниже.
- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/audit/
  ```
  Не ломать: PropagatesError, ListTrash cutoff, RetentionHides non-CRM types, ReassignTasks, SkipsOnCreateError, limits fail-closed, GDPR cascade abort HardDelete.
- Evidence required: checklist ID→evidence в этом файле / handoff; команды + PASS; mutation note per ID.
- Stop when: 7/7 FIXED · hubtest audit green · статус «Исправлена» в README + этом файле.

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W6-5-001 | убрать вызовы `crm_contacts`/`crm_deals` из `PurgeExpired` | CapturesCutoffs / callCounts×8 краснеет |
| A3W6-5-002 | CRM soft/restore → SoftDeleteIssue / без append CRM в mock | RecordsAllTypes / CRM map тест краснеет |
| A3W6-5-003 | снять `DeletedAtGT(cutoff)` в ListDeletedCRM* | RetentionHides CRM краснеет (total≠0) |
| A3W6-5-004 | убрать `InvalidateUserStatusCache` / не Del key | cache Exists assert краснеет |
| A3W6-5-005 | (если оставлен assert) снять parse `project_id` | CapturesProjectID / усиленный sibling краснеет; если удалён hollow — N/A: удаление |
| A3W6-5-006 | N/A: удаление hollow **или** та же мутация, что 001/002 на усиленном тесте | hollow не маскирует gap |
| A3W6-5-007 | (если усилен) убрать cascade call в mock path | cascade counters краснеют; если удалён hollow — N/A: удаление |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (7/7 FIXED или список открытых)
2. Changes with paths
3. Checklist ID → evidence (таблица)
4. Test commands + PASS
5. Mutation note (per ID)
6. Assumptions / unresolved (напр. удалили vs усилили hollow)
7. Risks + mitigations (ложные зелёные на CRM purge в prod)
8. Handoff: be-ver → code-reviewer
9. README + этот файл: статус «Исправлена»

## [OPERATING RULES]

- Work only inside scope; не закрывать соседние A3/A2 ID «заодно» без evidence
- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record (здесь N/A)
- Не ослаблять TASK-090 fail-closed purge и TASK-114 цемент-тесты
- Hollow siblings: предпочтительно **удалить**, если дубль цемента уже есть; усиливать только если нужен отдельный слой (repo vs service)
- Docs sync только если изменился user-visible API contract
- Conventional Commits RU; no Co-authored-by / AI trailers
- Коммит(ы) ссылают задачу в теле:  
  `Задача: .planning/backlog/TASK-161-2026-07-29-a3-6-5-audit-tests.md`  
  subject может содержать `(A3W6-5-001)` и т.п.
