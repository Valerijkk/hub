# W19 SUMMARY — перекрёстная проверка (Stage C)

> 2026-07-28 · TASK-008 · агент **W19-A5** (синтез) · ветка `cursor-grok-4.5`  
> Входы: [`W18-SUMMARY.md`](W18-SUMMARY.md), [`W18-CATCHUP-PRODUCT.md`](W18-CATCHUP-PRODUCT.md),  
> [`W19-A1-ORPHAN-P0P1.md`](W19-A1-ORPHAN-P0P1.md), [`W19-A2-EVIDENCE-GAPS.md`](W19-A2-EVIDENCE-GAPS.md),  
> [`W19-A3-COMBOS.md`](W19-A3-COMBOS.md), [`W19-A4-REDTEAM.md`](W19-A4-REDTEAM.md)

## Контекст после W18

| Метрика | Значение |
|---------|---------:|
| Mutation PASS → «Исправлена» | **39** |
| Mutation FAIL → «Открыта» | **9** |
| FAIL IDs | **014, 019, 041, 042, 044, 045, 046, 052, 054** |
| Catch-up product (W18-C5) | P0=0 · P1=0 · P2=2 · P3=2 → **TASK-079**, **TASK-080** |

**TASK-008 не закрывать.** W19 = orphans / evidence / combos / red-team перед W20 FINAL + полный hubtest + coverage.

---

## 1. Orphans P0/P1

Источник: [`W19-A1-ORPHAN-P0P1.md`](W19-A1-ORPHAN-P0P1.md)

| Метрика | Значение |
|---------|---------:|
| **P0 сироты** | **0** (009…012 покрывают реестр) |
| **P1 сироты (единицы реестра)** | **10** — все U-section |
| Missed (готовы к TASK, забыты) | **0** |
| Stage A P0/P1 вне W12 | **0** |

**Вывод:** дыр постановки нет. Незамапленные P1 = блок «Требует уточнения» (ADR / PO / greenfield):

| # | ID | Суть |
|--:|----|------|
| 1 | P1-BE-U01 | Offline/Ollama vs SafeHTTP loopback — нужен ADR dialer |
| 2 | P1-BE-U02 | Мёртвый Telegram-login тест / handleLogin — ADR-AUTH-01 |
| 3 | P1-BE-U03 | News polls FR-INTR-07/08 — greenfield, модуля нет |
| 4 | P1-BE-U04 | Backup gate process-local — multi-replica ADR |
| 5 | P1-BE-U05 | PermanentDelete / OnDelete edges — cascade ADR |
| 6 | P1-BE-U06 | `version`/`deleted_at` на mutable-схемах — VersionMixin ADR |
| 7 | W10-A3-001 | Sprint backlog UI без `sprint_id` — BE confirm |
| 8 | W10-A5-002 | Trash «навсегда» fake success — API/роль? |
| 9 | W11-A5-005 | Webcams docs vs TZ Phase 16 — решение PO |
| 10 | W11-A5-006 | G-07 wiki KPI vs «планируется» — решение PO |

U\* не разблокировать без ADR/PO; держать в `backlog/README.md` (Blocked ADR: U01–U06).

---

## 2. Evidence gaps

Источник: [`W19-A2-EVIDENCE-GAPS.md`](W19-A2-EVIDENCE-GAPS.md)

| Метрика | Значение |
|---------|---------:|
| PASS без полной триады (test + SHA + mutation log) | **39 / 39** |
| Нет **SHA** коммита фикса | **39** (системный пробел) |
| Нет имени теста (union TASK ∪ W18) | **0** |
| Нет mutation log (union) | **0** |
| Thin in-file (тест только в W18, не в теле TASK) | **28** |
| Статус-дрейф | **TASK-038** Progress = «На проверке», README = «Исправлена» |

Gap IDs = все PASS: 009–013, 015–018, 020–040, 043, 047–051, 053, 055, 056  
(не gap: FAIL 014, 019, 041, 042, 044, 045, 046, 052, 054).

**Вердикт:** mutation-доказательства есть; SHA нигде не зафиксирован → «Исправлена» ≠ финальное закрытие Stage C по ACCEPTANCE.

---

## 3. Combinatorial defects

Источник: [`W19-A3-COMBOS.md`](W19-A3-COMBOS.md)

| Severity | Count | Actionable |
|----------|------:|-----------:|
| CRITICAL | 1 | 1 |
| HIGH | 5 | 5 |
| MEDIUM | 5 | 5 |
| LOW | 2 | 1 (C-013 = closed ref) |
| **Total** | **13** | **12** |

### Top risks (CRITICAL + HIGH)

| ID | Цепочка | Связь |
|----|---------|--------|
| **C-001** CRITICAL | GDPR Mongo eraser `(0,nil)` + HardDelete cascade → erase без удаления сообщений | TASK-079 · W18-C5-001 |
| **C-002** HIGH | Soft-delete project + webhooks ListActive без `project.deleted_at` → события в корзину 30д | residual после 009/012 |
| **C-003** HIGH | AI `access==nil` fail-open + issue/review consumers → IDOR | W18-C5-004 · residual 021 |
| **C-004** HIGH | Confirm empty entity + Confirm без object (TASK-045 FAIL) | **045** reopen |
| **C-005** HIGH | CallChat empty при error + send enabled | TASK-073 |
| **C-006** HIGH | SQL soft-delete OK + Meili upsert orphans | residual после 022 |

Catch-up усиливает C-001/C-003: P2 eraser no-op (079) и P3 consumer fail-open уже в product-коде.

---

## 4. Red team (auth non-admin)

Источник: [`W19-A4-REDTEAM.md`](W19-A4-REDTEAM.md)

**Вердикт A4:** 0 новых CRITICAL, доказанно открытых для ordinary non-admin; остаточные Med/High привязаны к W18 FAIL.

### Top 5 chains

| # | Цепочка | L×I | Reopens |
|--:|---------|-----|---------|
| 1 | Team attach → forced project membership (нет `requireTeamMember`) | Med × High | **014** |
| 2 | ConfirmUpload ghost / weak entity bind | Med–High × Med | **045** |
| 3 | Concurrent StartSprint → dual active | Low–Med × Med | **041** |
| 4 | User SSRF (push / link-preview) + To4 mutation blind | Low × High | **019** |
| 5 | AI privacy / nil access fail-open + stream cancel gap | Low–Med × High | **044** |

### W18 FAIL → reopen map (сводка)

| TASK | Red-team | Приоритет до W20 |
|------|----------|------------------|
| **014** | YES | **P0 reopen** |
| **041** | YES | **P0 reopen** |
| **045** | YES | **P0 reopen** |
| 019 | YES (residual) | next |
| 044 | YES | next |
| 042 | PARTIAL | next |
| 046 | PARTIAL | next |
| 052 | LOW | hygiene |
| 054 | NO (exploit) | docs → TASK-080 |

---

## 5. Recommended actions before W20

Порядок — блокеры «mutation-closed» и CRITICAL combo первыми.

1. **Не закрывать TASK-008.** W20 = FINAL + полный hubtest + coverage; W19 не гейтит закрытие программы.
2. **Reopen / дожать mutation (be):** **014 → 041 → 045** (red-team blockers), затем **019, 042, 044, 046**.
3. **CRITICAL combo (be):** **C-001 / TASK-079** — fail-closed Mongo eraser + cascade abort; hubtest на реальном eraser (W18-C5-002).
4. **HIGH combos (be):** C-002 (webhook filter на soft-deleted project), C-003 (fail-closed nil access в consumers), C-006 (Meili delete на soft-delete), C-004 вместе с 045.
5. **FE empty+mutate:** C-005 / C-008 / C-009 через **TASK-073** (+ WebhooksPage `isError`).
6. **Evidence hygiene (optional, docs):** backfill SHA Stage B в 39 PASS TASK (`git log --grep=TASK-0NN`); thin 28 — скопировать имя RED-теста из W18 в тело TASK; синхронизировать Progress TASK-038.
7. **U\* orphans:** не заводить TASK до ADR/PO; оставить Blocked в README.
8. **TASK-080** (docs drift autotest) — не security-блокер; можно параллельно с W20 quality gate.
9. **W20 входной критерий (рекомендация):** нет открытых CRITICAL combo; mutation-closed по **014, 041, 045**; hubtest green на затронутых пакетах; coverage baseline без регрессии.

---

## Артефакты волны

| Агент | Файл | Результат |
|-------|------|-----------|
| W19-A1 | `W19-A1-ORPHAN-P0P1.md` | P0=0 · P1 orphans=10 (U only) · missed=0 |
| W19-A2 | `W19-A2-EVIDENCE-GAPS.md` | 39 SHA-gaps · triad иначе OK |
| W19-A3 | `W19-A3-COMBOS.md` | 13 combos · 1 CRITICAL · 5 HIGH |
| W19-A4 | `W19-A4-REDTEAM.md` | top-5 · blockers 014/041/045 |
| W19-A5 | **этот файл** | синтез + handoff → W20 |

---

*Синтез W19-A5 · 2026-07-28 · TASK-008*
