# PLANNING-FIXED-SUMMARY — статус 11 находок аудита

**Дата сводки:** 2026-07-27  
**Источник аудита:** `00-PLANNING-AUDIT.md`  
**Ограничение работ:** только `.planning/` (+ согласованные docs вроде AGENTS/rules для дат/инвентаря). Продакшн-код не трогали.

| # | Находка | До | После | Закрыто? |
|---|---------|----|-------|----------|
| 1 | 51 фаза без VERIFICATION при статусе 100% | 52 без артефакта (скан) | Артефакты на диске для всех phase dirs; **часть не в git** из‑за `.gitignore` → `.planning/**/VERIFICATION.md` | **PARTIAL** — файлы есть; remote/git visibility broken until owner clears ignore / approves `git add -f` |
| 2 | Фазы 122/124 закрыты с недостигнутой целью coverage | Goal missed, статус complete | VERIFICATION с FAIL/PARTIAL (не подгонка под PASS) | **PARTIAL** (честный статус; gap не «исправлен» кодом) |
| 3 | Production-gate deploy blocked при 100% | Скрыто в агрегате | STATE `deploy_blockers` + PG-03/FA-06 раскрыты | **YES** (прозрачность) |
| 4 | COVERAGE.md протух (2026-06-22 как текущий) | 83.0% без свежего прогона | **81.9% measured 2026-07-27** Docker; historical section | **YES** (метрика честная; gate MISS) |
| 5 | NFR-24 version 17/83 | Разрыв без классификации | Классификация + 2 опции владельцу | **PARTIAL** — ждёт решения владельца |
| 6 | NFR-11 trash 15 схем / 4 типа API | Разрыв без классификации | Классификация + опции | **PARTIAL** — ждёт решения владельца |
| 7 | FR-INTR-08 newspost невосстановим | Нет фазы | PLAN `126-intranet-newspost-trash` | **PARTIAL** — план есть, кода нет (запрещён в этой волне) |
| 8 | FR-CAL-05 series без soft-delete | Нет фазы | PLAN `127-calendar-series-trash` | **PARTIAL** — план есть, кода нет |
| 9 | Wiki в спеке, вне роадмапа/кода | Молчание | `WIKI-SCOPE-OPTIONS.md` | **PARTIAL** — ждёт владельца |
| 10 | 14 `[default]` gray-zone решений | Без подтверждения | `DECISIONS-PENDING.md` | **PARTIAL** — ждёт владельца |
| 11 | ROADMAP обрывался на 14/17 | 18–125 отсутствовали | Секции восстановлены + Progress sync | **YES** |

## Ключевые live-метрики (не из старых доков)

| Метрика | Значение | Дата |
|---------|----------|------|
| Coverage overall | **81.9%** | 2026-07-27 |
| goals_verified (PASS at matrix scan) | **55** | 2026-07-27 |
| tasks_completed / phases_percent | 125 / 100 | legacy checkbox claim |
| spec_coverage | **9/28 = 32.1%** | 2026-07-27 |
| Routes `.tsx` / tests / stores / Go | 72 / 515 / 13 / 1.25.0 | 2026-07-27 |

## Что закрыть не удалось в этой волне

1. **Довести до нуля** фазы без verification (остаток **22–37** на момент сводки).
2. **Per-package statement coverage** таблицы 2026-07-27 (coverprofile Docker flake).
3. **Выполнить** FR-INTR-08 / FR-CAL-05 / Wiki / NFR раскатку — это код, вне scope.
4. **Подтвердить** 14 default решений — нужен владелец.
5. C6 PROJECT.md — проверить отдельно, если агент не доехал.

## Отчёты волн

- `PW1-REPORT.md` … `PW5-REPORT.md` в этой же папке.
