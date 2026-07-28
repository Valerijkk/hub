# W19-A1 — сироты P0/P1 без задач бэклога (Stage C cross-check)

> TASK-008 · Stage C · агент **W19-A1** · 2026-07-28 · ветка `cursor-grok-4.5`  
> Режим: **read-only по product code**; артефакт только в `audits/total-2026-07/`.  
> Перепроверка: инвентарь W12 ↔ `TASK-009`…`TASK-080` + сырые `| … | P0/P1 |` в `A*-*.md`.

## Метод

1. Канон Stage A → B:  
   [`W12-REGISTRY-P0.md`](W12-REGISTRY-P0.md),  
   [`W12-REGISTRY-P1-BACKEND.md`](W12-REGISTRY-P1-BACKEND.md),  
   [`W12-REGISTRY-P1-FRONTEND.md`](W12-REGISTRY-P1-FRONTEND.md),  
   плюс [`00-STAGE-A-COMPLETE.md`](00-STAGE-A-COMPLETE.md) (Σ P0=4 · P1=141) и [`W18-SUMMARY.md`](W18-SUMMARY.md).
2. Покрытие: `.planning/backlog/TASK-009`…`TASK-080` + колонка «Источник» в  
   [`backlog/README.md`](../../backlog/README.md)  
   (ClassID, диапазоны `P1-BE-01..04` / `P1-BE-35..39,43`, Source ID, slug `FE-P1-*`).
3. «Сирота» = элемент реестра P0/P1 (или сырой Stage A P0/P1 вне реестра),  
   **не** упомянутый ни в одном TASK-009…080 как владелец фикса.

### Контроль (2026-07-28, re-verify W19-A1 · 19:58 UTC+3)

Независимый scan: `A*-*.md` (`| W*-A*-* | P0/P1 |`) ↔ `W12-REGISTRY-P0/P1-*` ↔ тела `TASK-009`…`080` + колонка «Источник» в `backlog/README.md` (раскрытие `P1-BE-01..04`, `P1-BE-35..39,43`) + `W18-SUMMARY.md`.

| Проверка | Результат |
|----------|-----------|
| Сырые P0 в `A*-*.md` | ровно **4**: W1-A5-001, W3-A2-001, W4-A5-001, W4-A5-002 |
| Сырые P1 в `A*-*.md` | ровно **141** (= 00-STAGE-A-COMPLETE Σ); дубликатов ID **0** |
| Checksum P1 | BE 79+20 + FE 38+4 = **141** |
| P0 → TASK | 009…012 (все 4 в README + файлах задач) |
| `P1-BE-01`…`43` после раскрытия `..` / `,43` | **43/43** в TASK-013…025 |
| `FE-P1-*` ×10 | **10/10** в TASK-026…035 |
| U\* ClassID / FE U Source ID в телах TASK-009…080 | **0** вхождений |
| Stage A P0/P1 вне W12-реестров | **0** (W18 catch-up product: P0=0 · P1=0 · P2=2 → TASK-079/080) |
| README Blocked ADR | только BE U01–U06; FE U (4 ID) — тоже сироты, без строки Blocked |
| Re-scan итог | P0 orphan **0**; mutation BE/FE **0** missed; U\* → **10** orphan units / **24** сырых Source ID |

---

## Сводка

| Метрика | Значение |
|---------|---------:|
| P0 в реестре | **4** |
| P0 с задачей | **4** (TASK-009…012) |
| **P0 сироты** | **0** |
| P1 BE классов (мутация) | **43** → все в TASK-013…025 |
| P1 FE классов (мутация) | **10** → все в TASK-026…035 |
| P1 U-секция BE | **6** классов / **20** сырых ID |
| P1 U-секция FE | **4** Source ID |
| **P1 сироты (единицы реестра)** | **10** (все = U-section) |
| P1 сироты «missed» (готовые к TASK, но забыты) | **0** |
| P1 сырых Source ID без задачи | **24** (20 BE U + 4 FE U) |
| Stage A P0/P1 вне W12-реестров | **0** |

**Вывод:** дыр постановки нет. Единственные незамапленные P0/P1 — осознанный блок  
W12 §«Требует уточнения» (ADR / product-scope / BE+PO), зафиксированный также в  
[`backlog/README.md`](../../backlog/README.md) (`Blocked ADR: U01–U06` — BE; FE U держать рядом до PO/BE confirm).

---

## Сироты (полный список = top 10)

| # | Source ID | Severity | Title | Почему нет задачи |
|--:|-----------|----------|-------|-------------------|
| 1 | **P1-BE-U01** (`W5-A2-003`) | P1 | Offline/Ollama consent vs `NewSafeHTTPClient` loopback block | **U-section** — нужен ADR (trusted offline dialer vs sidecar); без выбранного поведения нет mutation criterion |
| 2 | **P1-BE-U02** (`W6-A2-001`) | P1 | Тест `HandleLogin_StoresCode` держит мёртвый Telegram-login путь | **U-section** — решение: удалить только тест или ещё `handleLogin`+export (ADR-AUTH-01) |
| 3 | **P1-BE-U03** (`W7-A2-001`) | P1 | FR-INTR-07/08 news polls + author-delete / restore | **U-section** — greenfield scope; модуля polls нет → TASK по rule 57 нельзя |
| 4 | **P1-BE-U04** (`W7-A4-003`) | P1 | Backup maintenance gate process-local (multi-replica) | **U-section** — нужен multi-instance harness или ADR single-replica affinity |
| 5 | **P1-BE-U05** (`W8-A4-002`) | P1 | Project `PermanentDelete` / 25+ edges без OnDelete | **U-section** — ADR cascade policy (staged soft→30d vs OnDelete per-edge) |
| 6 | **P1-BE-U06** (15× W8 schema IDs*) | P1 | Системный TZ-gap `version` / `deleted_at` на mutable-схемах | **U-section** — ADR VersionMixin + исключения; без списка scope задача бесконечна |
| 7 | **W10-A3-001** | P1 | Sprint backlog UI «без спринта» без `sprint_id` в API/DTO | **U-section** — нужен BE confirm / кросс-задача; FE-only неполон |
| 8 | **W10-A5-002** | P1 | Trash «Удалить навсегда» — fake success без API | **U-section** — есть ли admin permanent-delete + роль; иначе только UX-убрать кнопку |
| 9 | **W11-A5-005** | P1 | Docs: webcams = Production vs TZ Phase 16 | **U-section** — решение PO (правка docs vs признать partial) |
| 10 | **W11-A5-006** | P1 | G-07 «US-WIKI реализованы» ↔ KPI wiki «планируется» | **FIXED** TASK-105 (G-07 Partial; A2W11-5-001) |

\* U06 сырые: `W8-A1-002`, `W8-A1-003`, `W8-A2-001`…`004`, `W8-A3-002`, `W8-A4-001`, `W8-A4-003`, `W8-A5-002`, `W8-A5-005`, `W8-A5-006`, `W8-A5-008`…`010` (15 ID).

### Классификация причин

| Причина | P0 | P1 ед. | Примечание |
|---------|---:|-------:|------------|
| U-section (blocked ADR/PO/scope) | 0 | **10** | не duplicate, не missed |
| duplicate (уже в другой TASK) | 0 | 0 | — |
| missed (готово к постановке, забыто) | 0 | **0** | — |
| ADR (явный подтип U) | 0 | 6 BE U | U01–U06; FE U — PO/BE confirm |

---

## Контроль покрытия (не сироты)

| Группа | TASK |
|--------|------|
| P0 ×4 | 009…012 |
| P1-BE-01…43 (mutation) | 013…025 (группировка W12 ~13 slug’ов) |
| FE-P1-* ×10 | 026…035 |
| P2/P3 | 036…078 (вне scope этого отчёта) |
| Catch-up Stage C | 079…080 (не P0/P1 Stage A; P2 residual) |

---

## Handoff

- **P0 сирот нет** — Stage C mutation по P0 опирается на TASK-009…012.
- **P1 сироты = только U\*** — заводить TASK после ADR/PO; до того оставлять в Blocked README (добавить FE U рядом с BE U01–U06).
- Не путать с FAIL мутации W18 (014, 019, 041…): у них задачи есть — это residual fix, не orphan.
- Следующий кросс-чек Stage C: P2/P3 orphans, evidence gaps, combos, redteam — по назначению оркестратора.

---

*Артефакт W19-A1 · 2026-07-28 · verified against TASK-009…080 · raw Σ P0=4 P1=141*
