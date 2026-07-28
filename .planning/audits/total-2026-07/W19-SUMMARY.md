# W19 SUMMARY — перекрёстная проверка (Stage C)

> 2026-07-28 · TASK-008 · агент **W19-A5** (синтез) · ветка `cursor-grok-4.5`  
> Входы: [`W18-SUMMARY.md`](W18-SUMMARY.md), [`W18-CATCHUP-PRODUCT.md`](W18-CATCHUP-PRODUCT.md),  
> [`W19-A1-ORPHAN-P0P1.md`](W19-A1-ORPHAN-P0P1.md), [`W19-A2-EVIDENCE-GAPS.md`](W19-A2-EVIDENCE-GAPS.md),  
> [`W19-A3-COMBOS.md`](W19-A3-COMBOS.md), [`W19-A4-REDTEAM.md`](W19-A4-REDTEAM.md)  
> Re-verify входов: A1 (19:54) · A2 (count=39) · A3 (19:52) · A4 (19:54) UTC+3.

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
| Сырые Source ID без задачи | **24** (20 BE U + 4 FE U) |

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

U\* не разблокировать без ADR/PO; держать в `backlog/README.md` (Blocked ADR: BE U01–U06; **добавить FE U** рядом).  
Не путать с W18 FAIL (014, 019, 041…) — там задачи есть, это residual fix, не orphan.

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
Кандидаты SHA — приложение B в A2 (`git log --grep=TASK-0NN`); ещё не evidence.

---

## 3. Combinatorial defects

Источник: [`W19-A3-COMBOS.md`](W19-A3-COMBOS.md) (re-verify path:line 19:52)

| Severity | Count | Actionable |
|----------|------:|-----------:|
| CRITICAL | 2 | 2 |
| HIGH | 6 | 6 |
| MEDIUM | 5 | 5 |
| LOW | 3 | 3 |
| **Total** | **16** | **15** (C-016 = FE adjacent reference) |

### Top risks (CRITICAL + HIGH)

| ID | Цепочка | Связь |
|----|---------|--------|
| **C-001** CRITICAL | GDPR Mongo eraser `(0,nil)` + HardDelete cascade → erase без удаления сообщений | TASK-079 · W18-C5-001 |
| **C-002** CRITICAL | Dual StartSprint race unproven + schema без unique active → persistent dual active | **041** + **046** |
| **C-003** HIGH | AddTeamToProject без team ACL + только PermTeamManage → forced membership | **014** |
| **C-005** HIGH | To4 mutation-blind + Meili SSRF hubtest gap | **019** + **044** |
| **C-006** HIGH | AI `access==nil` fail-open + stream cancel unproven | **044** |
| **C-007** HIGH | Confirm empty entity + Confirm-without-object | **045** |
| **C-008** HIGH | Soft-delete project + webhooks ListActive без `deleted_at` | residual |
| **C-009** HIGH | SQL soft-delete OK + Meili upsert orphans | **044** adjacent |

MEDIUM (не top, но в матрице): C-004 (014 mid-fail proof), C-010 (014×041 blast), C-011 (042 via_invite), C-012 (search outage UX), C-013 (045 legacy entity).  
LOW: C-014 (052), C-015 (054/080), C-016 (073 CallChat).

Все W18 FAIL 014/019/041/042/044/045/046/052/054 + TASK-079 в матрице A3.  
Порядок фикса A3: **C-001/079 → C-002 (041+046) → C-003/004 (014) → C-007/013 (045) → C-005/006 (019/044)**.  
`verified: false` для чистого Stage C — CRITICAL/HIGH без mitigation.

---

## 4. Red team (auth non-admin) — top risks

Источник: [`W19-A4-REDTEAM.md`](W19-A4-REDTEAM.md) (re-verify 19:54)

**Вердикт A4:** 0 новых CRITICAL как single-step non-admin exploit; CRITICAL combo **TASK-079** (admin erase + residual PII for members); остаточные Med/High → W18 FAIL.  
`verified: true` = audit complete with file:line — **не** «product secure».

### Top 5 chains

| # | Цепочка | L×I | Reopens |
|--:|---------|-----|---------|
| 1 | Team attach → forced project membership (нет `requireTeamMember`) | Med × High | **014** |
| 2 | ConfirmUpload ghost / weak entity bind | Med–High × Med | **045** |
| 3 | GDPR Mongo eraser `(0,nil)` → HardDelete + live messages | Low–Med × Critical | **079** |
| 4 | User SSRF (push / link-preview) + To4 mutation blind | Low × High | **019** (+**044**) |
| 5 | AI privacy / nil access fail-open + stream/Meili gaps | Low–Med × High | **044** |

Dual StartSprint (**041**) — residual #6 (Low–Med × Med); stacks with **046** → A3 C-002 CRITICAL.

### W18 FAIL → reopen map (сводка)

| TASK | Red-team | Приоритет до W20 |
|------|----------|------------------|
| **014** | YES | **P0 reopen** |
| **079** | YES (catch-up) | **P0 reopen** |
| **045** | YES | **P0 reopen** |
| 044 | YES | next |
| 019 | YES (residual) | next |
| 041 | YES | next |
| 042 | PARTIAL | next |
| 046 | PARTIAL | next |
| 052 | LOW | hygiene |
| 054 | NO (exploit) | docs → TASK-080 |

Порядок reopen-fix (A4): **014 → 079 → 045 → 044/019 → 041/046**.

---

## 5. Recommended actions before W20

Порядок — блокеры «mutation-closed» и CRITICAL combo первыми.

1. **Не закрывать TASK-008.** W20 = FINAL + полный hubtest + coverage; W19 не гейтит закрытие программы.
2. **Reopen / дожать mutation (be):** **014 → 079 → 045** (red-team blockers), затем **044, 019, 041+046, 042**.
3. **CRITICAL combos (be):** **C-001 / TASK-079** (Mongo eraser fail-closed); **C-002** dual-sprint/unique (**041**/**046**).
4. **HIGH combos (be):** team ACL (**014**), Confirm (**045**), SSRF/AI (**019**/**044**), soft-delete webhook/Meili (A3 C-008/009).
5. **FE/docs:** 052, 054/080, CallChat empty+mutate (073 / C-016).
6. **Evidence hygiene (optional, docs):** backfill SHA Stage B в 39 PASS TASK (A2 app. B); thin 28 — имя RED-теста в тело TASK; Progress TASK-038.
7. **U\* orphans:** не заводить TASK до ADR/PO; Blocked в README + FE U рядом с BE U01–U06.
8. **TASK-080** (docs drift autotest) — не security-блокер; можно параллельно с W20 quality gate.
9. **W20 входной критерий (рекомендация):** нет открытых CRITICAL combo; mutation-closed по **014, 079, 045**; hubtest green на затронутых пакетах; coverage baseline без регрессии.

---

## Артефакты волны

| Агент | Файл | Результат |
|-------|------|-----------|
| W19-A1 | `W19-A1-ORPHAN-P0P1.md` | P0=0 · P1 orphans=10 (U only) · missed=0 |
| W19-A2 | `W19-A2-EVIDENCE-GAPS.md` | 39 SHA-gaps · triad иначе OK |
| W19-A3 | `W19-A3-COMBOS.md` | 16 combos · 2 CRITICAL · 6 HIGH |
| W19-A4 | `W19-A4-REDTEAM.md` | top-5 · blockers 014/079/045 |
| W19-A5 | **этот файл** | синтез + handoff → W20 |

---

*Синтез W19-A5 · 2026-07-28 · TASK-008 · не закрывать*
