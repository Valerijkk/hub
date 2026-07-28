# TASK-008 — FINAL (Stage C / W20)

> Дата: **2026-07-28** · Ветка: `cursor-grok-4.5` · Оркестратор: Grok  
> Программа: [`.planning/backlog/TASK-008-total-audit-program.md`](../../backlog/TASK-008-total-audit-program.md)  
> **Вердикт: Stage A ✅ · Stage B (P0–P2 code) ✅ · Stage C ⚠️ PARTIAL · TASK-008 НЕ ЗАКРЫВАТЬ**

---

## 1. Вердикт по стадиям

| Стадия | Статус | Комментарий |
|--------|--------|-------------|
| **A** (W1–W11) | ✅ COMPLETE | 1405 файлов в отчётах; сырых 812 (P0=4·P1=141·P2=422·P3=245) |
| **B** P0–P2 | ✅ код + backlog | TASK-009…056; P3 = TASK-057…078 backlog only |
| **B** P3 | ⏸ backlog | Вне mutation-wave C |
| **C** W18 mutation | ⚠️ 39/48 PASS | 9 FAIL → «Открыта» |
| **C** W19 cross | ✅ артефакты | orphans / evidence / combos / redteam |
| **C** W20 FINAL | ✅ этот файл | hubtest green (local); COV vs 81.9% — **не сравним** в этой сессии |

**P0:** все 4 закрыты мутацией PASS (TASK-009…012) → статус «Исправлена».

---

## 2. Сводка объёма

| Метрика | Значение |
|---------|----------|
| Файлов в Stage A (волны) | **1405** / оценка TASK ~2031 (wiki/docs drift) |
| Добор product (W18) | **7/7** → TASK-079, TASK-080 |
| Сырых находок A | **812** |
| Задач аудита 009–056 | **48** (mutation scope) |
| Mutation PASS → «Исправлена» | **39** |
| Mutation FAIL → «Открыта» | **9** (014, 019, 041, 042, 044, 045, 046, 052, 054) |
| P3 backlog | **057–078** (Открыта) |
| Catch-up | **079, 080** (Открыта) |
| P1 orphans (U-section/ADR) | **10** (W19-A1) — не код-баги без ADR |

---

## 3. Реестр исходов (mutation wave)

Источники: `W18-C{1,2,2b,3,4,5}-MUTATION.md`, `W18-SUMMARY.md`.

### Закрыто мутацией (39)

P0: 009–012 · P1: 013, 015–018, 020–025, 026–035 · P2: 036–040, 043, 047–051, 053, 055, 056 (+ часть gap C2b).

### Вернулось после мутации (9) — приоритет reopen

Порядок из W19:

1. **014** — team ACL / AddTeamToProject  
2. **079** — GDPR Mongo eraser fail-closed (catch-up, не из 48)  
3. **045** — storage Confirm ghost  
4. **044 / 019** — AI fail-open / SSRF residuals  
5. **041 + 046** — ent/calls + schema; SQL migrate = **owner approval**  
6. Остальное FAIL: **042, 052, 054**

Evidence SHA gaps: **39** PASS без SHA в теле задачи (`W19-A2-EVIDENCE-GAPS.md`) — факты в W18-*-MUTATION.md.

---

## 4. Evidence: полный hubtest

### Команда (local, Windows)

```text
cd backend
$env:CGO_ENABLED = "1"
go test -tags hubtest -count=1 -timeout=600s ./testframework/...
```

### Результат (после hotfix regression)

| | |
|--|--|
| Исход | **PASS** |
| Пакетов `ok` | **47** |
| `FAIL` | **0** |
| Лог | [`W20-HUBTEST-OUTPUT.txt`](W20-HUBTEST-OUTPUT.txt) |

### Hotfix, потребовавшиеся для зелёного прогона (регрессия Stage B)

Не отдельный audit-finding wave — починка тестов/export под уже влитые guards:

| Причина | Фикс |
|---------|------|
| P2-BE-12 `GetProject` требует membership; mock `Create` не добавлял owner → 500 в handler tests | `mockRepo.Create` зеркалит Ent (+owner); обновлены assert’ы в `service_extended_test.go` |
| TASK-042: `ExportGenerateGuestToken` → `h.service.room` nil panic | токен из `CallRoomService{cfg: h.cfg}` |
| TASK-039: `webhook_secret` NotEmpty; тест без секрета | `WebhookSecret: "gh-secret"` в github ent test |

Файлы: `testframework/project/service_test.go`, `service_extended_test.go`, `internal/calls/export_hooks.go`, `testframework/github/ent_repository_test.go`.

---

## 5. Coverage vs 81.9% (2026-07-27)

| Замер | Overall `internal/` | Сравнимость |
|-------|--------------------:|-------------|
| Baseline 2026-07-27 Docker | **81.9%** | канон (`.planning/COVERAGE.md`) |
| Local Windows (sandbox Go 1.25) | — | **блокер:** нет `covdata` в toolchain → coverprofile merge = 0% |
| Docker `golang:1.25-bookworm` (эта сессия) | **80.1%** | **не сравним:** 5 FAIL пакетов без PG/Mongo (`accounting`, `imex`, `platform/cron`, `sprint`, `tms`) — см. `W20-COVER-RUN.txt` / `W20-COVER-TOTAL.txt` |

**Вывод COV:** критерий «не упало относительно 81.9%» **не подтверждён и не опровергнут** method-comparable замером. Нужен повтор в `deployments/docker-compose.test.yml` + полный hubtest+coverprofile (как в COVERAGE.md).

---

## 6. Чему верить нельзя

| Тема | Docs / ожидание | Код / факт |
|------|-----------------|------------|
| Stage C «все фиксы закрыты» | 48 «На проверке» → все PASS | **9 FAIL** reopen; статус «Исправлена» ≠ mutation-proof |
| GDPR cascade complete | eraser всегда чистит Mongo | `NewMongoMessageEraser(nil)` → silent success (**TASK-079**) |
| Coverage gate 81.9% | «свежий замер W20» | Docker 80.1% без БД **не** baseline; local covdata отсутствует |
| Ent unique/CHECK (025/039/046) | «исправлено в schema» | SQL migrate **не** в prod без owner approval |
| U-section / SSO orphans | «все P1 в задачах» | 10 orphans ждут ADR/PO (`W19-A1`) |
| AI `project_access` godoc | «или system admin» | admin в RBAC mask upstream (**P3** catch-up) |
| Metrics / docs FE | цифры в docs | drift → **TASK-054/080** |

Combos CRITICAL (W19-A3): GDPR eraser × HardDelete; dual StartSprint × no unique active.

---

## 7. Риски и непокрытое

- Red-team top: 014 ACL · 045 Confirm · 079 PII · 019/044 SSRF/AI · см. `W19-A4-REDTEAM.md`
- P3 057–078 не мутированы
- Residual TF/docs из Stage A catch-up list — не полностью разобраны
- FA-06 secret rotation — вне scope аудита, остаётся prod-блокер ops

---

## 8. Handoff team-lead

1. **Не закрывать TASK-008** до reopen-фиксов 014→079→045 и method-comparable COV.  
2. Приоритет кода: **014, 079, 045, 044/019, 041+046**.  
3. Owner: approve SQL migrate для 025/039/046 перед prod.  
4. Повторить cover в Docker test compose; обновить `.planning/COVERAGE.md`.  
5. Дописать SHA в тела 39 PASS задач (evidence hygiene).

---

## 9. Verification block (rule 40)

| Check | Result |
|-------|--------|
| Stage A file lines present | PASS (1405; добор 7) |
| P0 closed or owner-accepted | PASS (009–012 mutation PASS) |
| Every Stage B fix independent mutation | **FAIL** (9/48) |
| Full hubtest green | **PASS** (local 47 ok) |
| Coverage ≥ 81.9% method-comparable | **NOT RUN** (env/tooling) |
| `00-FINAL.md` assembled | PASS (этот файл) |

**TASK-008 status:** остаётся **В работе** (C partial).
