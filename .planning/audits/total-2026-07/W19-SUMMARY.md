# W19 SUMMARY — перекрёстная проверка Stage C

> 2026-07-28 · TASK-008 · агент **W19-A5** · ветка `cursor-grok-4.5`  
> Источники: [`W19-A1-ORPHAN-P0P1.md`](W19-A1-ORPHAN-P0P1.md) · [`W19-A2-EVIDENCE-GAPS.md`](W19-A2-EVIDENCE-GAPS.md) · [`W19-A3-COMBOS.md`](W19-A3-COMBOS.md) · [`W18-SUMMARY.md`](W18-SUMMARY.md)  
> **Частичная сводка:** `W19-A4-REDTEAM.md` на момент сборки **отсутствует** → § Red team = TODO + interim из A3 CRITICAL/HIGH.

## Статус sibling-артефактов

| Агент | Артефакт | Статус |
|-------|----------|--------|
| W19-A1 | `W19-A1-ORPHAN-P0P1.md` | ✅ |
| W19-A2 | `W19-A2-EVIDENCE-GAPS.md` | ✅ |
| W19-A3 | `W19-A3-COMBOS.md` | ✅ |
| W19-A4 | `W19-A4-REDTEAM.md` | ❌ TODO — resume W19-A5 после появления |
| W18 | `W18-SUMMARY.md` | ✅ (вход: 39 PASS / 9 FAIL) |

---

## 1. Orphans summary (A1)

| Метрика | Значение |
|---------|---------:|
| P0 сироты | **0** (все 4 → TASK-009…012) |
| P1 mutation-классы без TASK | **0** (BE 43 + FE 10 замаплены) |
| P1 «сироты» = U-section | **10** единиц / **24** сырых Source ID |
| Missed (готово к TASK, забыто) | **0** |
| Stage A P0/P1 вне W12-реестров | **0** |

**Вывод:** дыр постановки нет. Незамапленные P0/P1 — только осознанный блок W12 §«Требует уточнения» (ADR / PO / scope), также в `backlog/README.md` (`Blocked ADR: U01–U06`).

### Top orphans (все U-section)

| ID | Суть | Блокер |
|----|------|--------|
| P1-BE-U01 | Ollama consent vs SSRF loopback | ADR offline dialer |
| P1-BE-U02 | Мёртвый Telegram-login тест | ADR-AUTH-01 cleanup |
| P1-BE-U03 | FR-INTR news polls greenfield | product scope |
| P1-BE-U04 | Backup gate multi-replica | ADR / harness |
| P1-BE-U05 | Project PermanentDelete OnDelete | ADR cascade |
| P1-BE-U06 | `version`/`deleted_at` TZ-gap (15 schema ID) | ADR VersionMixin |
| W10-A3-001 | Sprint backlog без `sprint_id` API | BE+FE confirm |
| W10-A5-002 | Trash «удалить навсегда» fake success | PO / API |
| W11-A5-005 | Docs webcams = Production vs Phase 16 | PO docs |
| W11-A5-006 | G-07 wiki KPI drift | PO |

---

## 2. Evidence gaps (A2)

| Метрика | Значение |
|---------|---------:|
| TASK-009…056 | 48 |
| «Исправлена» (W18 PASS) | 39 |
| «Открыта» (W18 FAIL) | 9 |
| **Evidence-gap (incomplete triad)** | **39 / 39 PASS** |
| Нет SHA фикса | **39** |
| Нет имени теста (union TASK∪W18) | **0** |
| Нет mutation log (union) | **0** |
| Thin in-file (тест только в W18) | **28** |
| Статус-дрейф Progress | TASK-038 («На проверке» vs README «Исправлена») |

**Вердикт:** mutation RED→GREEN и имена тестов есть; **системный пробел — отсутствие SHA** Stage B коммитов в evidence. Не считать «Исправлена» финальным закрытием Stage C до дописи SHA.

W18 FAIL (не gap, статус «Открыта»): **014, 019, 041, 042, 044, 045, 046, 052, 054**.

---

## 3. Combos (A3)

| Severity | Count | IDs |
|----------|------:|-----|
| CRITICAL | 1 | C-001 |
| HIGH | 5 | C-002…C-006 |
| MEDIUM | 5 | C-007…C-011 |
| LOW | 2 | C-012, C-013 (C-013 = closed ref) |
| **Actionable** | **12** | без C-013 |

### Top combos (тяжелее суммы)

1. **C-001 CRITICAL** — GDPR Mongo no-op (`gdpr_mongo` nil→`(0,nil)`) + HardDelete cascade → erase «успешен» при живых Mongo-сообщениях (TASK-079 / residual).
2. **C-002 HIGH** — soft-delete project + webhooks `ListActive` без `project.deleted_at` → доставка на URL корзины до 30д.
3. **C-003 HIGH** — AI consumers fail-open при `access==nil` поверх P1-BE-31 PASS.
4. **C-004 HIGH** — Storage Confirm empty entity + TASK-045 FAIL (Confirm без object).
5. **C-005 HIGH** — CallChat fail→empty + send enabled (TASK-073).
6. **C-006 HIGH** — SQL soft-delete OK + Meili upsert без delete → утечка корзины в search index.

Фокус-матрица: authz chains · webhook+scope · soft-delete+search · FE empty+mutate · GDPR.

---

## 4. Red team top risks

> **TODO:** полный отчёт `W19-A4-REDTEAM.md` ещё не записан. Ниже — **interim** (A3 CRITICAL/HIGH + W18 FAIL residual), не замена A4. После появления A4 — resume W19-A5 и заменить §4.

| # | Риск (обычный user / текущий код) | Источник |
|--:|-----------------------------------|----------|
| 1 | GDPR erase → HardDelete user при живых Mongo messages | C-001 |
| 2 | Soft-deleted project продолжает слать webhook events | C-002 |
| 3 | AI issue/review context без checker = IDOR | C-003 |
| 4 | Confirm upload без entity / без S3 object (NF-09-класс) | C-004 + TASK-045 |
| 5 | CallChat: сеть упала → «пустая» лента + отправка вслепую | C-005 + TASK-073 |
| 6 | Meili держит soft-deleted issues после SQL-фильтра | C-006 |
| 7 | W18 FAIL residual authz/race/docs: 014 teams ACL mock, 041 sprint race, 042 via_invite, 044 Meili SSRF hubtest, 045 Confirm 404, 046 schema≠runtime | W18-SUMMARY |

---

## 5. Recommended next actions before W20

1. **Дождаться / принять `W19-A4-REDTEAM.md`** и обновить эту сводку (resume W19-A5) — без A4 §«чему верить нельзя» в W20 неполный.
2. **be P0/P1 residual:** завести или закрыть TASK на C-001 (TASK-079), C-002 (webhook×project deleted), C-003 (fail-closed nil access); reopen mutation для **TASK-045** (и при необходимости 014/019/041/042/044/046).
3. **Evidence hygiene:** один проход `git log --grep=TASK-0NN` → дописать **SHA** в 39 PASS-задач (или таблицу в W18-SUMMARY); thin TASK (28) — имя RED-теста в тело; синхронизировать Progress TASK-038.
4. **fe:** TASK-073 (CallChat / BYOK / silent-empty) + webhooks list `isError` (C-005, C-008, C-009).
5. **search:** Meili delete-on-soft-delete (C-006); hard-error на total shard outage (C-007 / TASK-044).
6. **Не разблокировать U01–U06** в бэклог без ADR/PO — orphans осознанные.
7. **W20 вход:** `00-FINAL.md` опирается на: покрытие файлов Stage A · реестр исходов (39 PASS / 9 FAIL / 12 combos) · A4 attack chains · полный `go test -tags hubtest` + COV vs 81.9%.

---

## Артефакты волны

`W19-A1-ORPHAN-P0P1.md` · `W19-A2-EVIDENCE-GAPS.md` · `W19-A3-COMBOS.md` · `W19-A4-REDTEAM.md` (**TODO**) · `W19-SUMMARY.md` (этот файл) · вход `W18-SUMMARY.md`

*Собрано W19-A5 · partial · 2026-07-28*
