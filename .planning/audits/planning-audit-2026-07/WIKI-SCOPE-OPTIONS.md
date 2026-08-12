# Wiki scope options — owner decision (C4 / B5)

**Date:** 2026-07-27  
**Audience:** product owner  
**Status:** **UNDECIDED** — options only; no recommendation  
**Audit refs:** `00-PLANNING-AUDIT.md` §9 / finding P2; `01-PLANNING-FIXES.md` B5; `DECISIONS-PENDING.md` D5  

---

## 1. Situation (facts)

| Layer | State |
|-------|--------|
| **Spec** | `docs/system_specification.md` § «Wiki — новый модуль (FR-WIKI)» — FR-WIKI-01…09, NFR-37…39, ERD (6 сущностей), ~22 REST endpoints, NATS `hub.wiki.*` |
| **User stories** | `docs/vision_and_scope.md` — US-WIKI-01…05, KPI G-07, MoSCoW |
| **Product code** | **Нет** модуля: `backend/internal/wiki` отсутствует; схем wiki в `backend/ent/schema/` нет; routes `/api/wiki/*` не зарегистрированы |
| **Root planning** | `.planning/ROADMAP.md` и `.planning/PROJECT.md` — слово **wiki не встречается** |
| **July roadmap** | `.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md` — wiki в Волнах **5.e** (старт) и **8.a** (завершение); handbook = wiki-space (Волна 6.b, зависит от 5.e) |
| **Gray zones** | GZ-SEC-07/11 (space RBAC + restrict) и GZ-V4-12/13 (handbook = общая wiki) — **приняты** 2026-07-03 |
| **GDPR defaults** | GZ-SEC-17/18 опираются на «wiki каскадит» / anonymize авторства правок — **невалидны** без модуля (см. `DECISIONS-PENDING.md` D5) |

**Не путать:** `docs/wiki/` — это **документация репозитория** (API.md, DEPLOY, QUICKSTART).  
**Продуктовый** wiki-модуль (spaces/pages для пользователей Hub) — отдельная сущность; его в коде нет.

Фокус решения аудита — **FR-WIKI-01…04** (ядро: spaces, дерево страниц, optimistic lock, trash-каскад, space RBAC).  
FR-WIKI-05…09 (restrict, версии, issue-links, поиск) — та же спека; любой выбор A/B/C обычно применяется ко **всему** блоку FR-WIKI, иначе останется половинчатый drift.

---

## 2. Core requirements (FR-WIKI-01…04)

| ID | Поведение (кратко) |
|----|--------------------|
| **FR-WIKI-01** | Space CRUD; страницы в иерархии `parent_id`; markdown + preview; move по дереву; цикл → 422; глубина >10 → 422 |
| **FR-WIKI-02** | Optimistic lock `version` → 409; изображения через storage presigned (не base64) |
| **FR-WIKI-03** | Удаление страницы → каскад **поддерева** в корзину 30д; restore поддерева; space удаляется только пустым (иначе 422) |
| **FR-WIKI-04** | Роли space `read`/`write`/`admin`; наследование на страницы; не-участник → 403; ≥1 admin (снятие последнего → 422) |

Источник: `docs/system_specification.md:515–529` (таблица FR; далее 05…09 и ERD/API).

---

## 3. Options (choose one)

### A) Build wiki (реализация)

**Что сделать:** завести фазы в корневом ROADMAP/PROJECT, реализовать модуль по спеке (минимум FR-WIKI-01…04; полный паритет спеки = 01…09).

**Ориентир по фазам** (грубо, по уже описанным July-волнам + типичный Hub GSD-срез):

| Срез | Содержание | Оценка |
|------|------------|--------|
| **Phase W1** | Ent: `wiki_space`, `wiki_space_member`, `wiki_page` (+ soft delete); BE CRUD spaces/pages/tree/move; space RBAC; cycle/depth checks | ~1 фаза · be1 (+ qa-manual TC) |
| **Phase W2** | FE: дерево + markdown editor/preview + space members UI; wiring routes | ~1 фаза · fe1 |
| **Phase W3** (полный модуль) | Restrict (05), versions/rollback (06–07), issue-links (08), search `types=wiki_page` + NF-03 filter (09), NATS outbox, handbook space | ~1–2 фазы · be+fe (+ search) |
| **QA / gate** | testframework unit + E2E smoke; docs-sync API.md; security-auditor на RBAC/restrict | вшито в фазы или +0.5 фазы |

| Объём | Rough effort |
|-------|----------------|
| **MVP ядро (01…04)** | **2–3 GSD-фазы** · порядка **2–3 недели** calendar при параллели be1∥fe1 (после разморозки July code freeze) |
| **Полный модуль (01…09)** | **3–5 GSD-фаз** · порядка **4–6 недель** (как Волны 5.e + 8.a в July roadmap) |
| Docs/planning only до кода | 0.5–1 день: фазы в ROADMAP/PROJECT + STATE |

**Плюсы:** закрывает G-07 / паритет Confluence-lite; разблокирует handbook (GZ-V4-13); делает GDPR-wiki ветки исполняемыми.  
**Минусы:** крупный новый домен (RBAC, tree, trash cascade, search scope); конкурирует с intranet/calendar/CRM/HR/SSO в July backlog; code freeze сейчас — старт только после kickoff.

**Не решает само:** нужно явное включение в `.planning/ROADMAP.md` / `PROJECT.md` (сейчас тишина) и согласование с July freeze.

---

### B) Explicitly out of v4 scope (annotation only)

**Что сделать:** оставить текст FR-WIKI в спеке, но **пометить** как out-of-scope для текущего milestone / до отдельного ADR/фазы.

Пример аннотации (черновик формулировки, не применять без решения):

> **Status: OUT OF SCOPE for v4 ship / July Rebuild.** Spec retained for Wave 5+/future. No implementation commitment in `.planning/ROADMAP.md` until owner re-opens. Handbook (GZ-V4-13) deferred or replaced by interim (intranet/HR markdown).

Параллельно: в `vision_and_scope.md` у US-WIKI — статус «deferred»; в `docs/wiki/API.md` таблица planned — баннер «not scheduled»; GDPR D5 — wiki-ветки `BLOCKED: waits FR-WIKI` (уже намечено в `DECISIONS-PENDING.md`).

| Работа | Rough effort |
|--------|----------------|
| Аннотации в spec + vision + API.md + STATE one-liner | **0.5–1 день** docs |
| Опционально: правка July roadmap 5.e/8.a → «deferred» | +0.5 дня |
| Код | **0** |

**Плюсы:** честный статус без потери проработанной спеки; GZ/US остаются как backlog-контракт.  
**Минусы:** G-07 и handbook остаются открытыми; drift «спека есть → кто-то думает, что строится» сохраняется, если баннер слабый; GDPR-wiki defaults всё ещё «мёртвые».

---

### C) Remove from spec

**Что сделать:** удалить (или перенести в archive/seed) блок FR-WIKI / US-WIKI / planned API / NFR-37…39; снять wiki из G-07; пересмотреть GZ-V4-12/13 (handbook) и упоминания wiki в GZ-SEC-17/18.

| Работа | Rough effort |
|--------|----------------|
| Вычистить `system_specification.md`, `vision_and_scope.md`, `docs/wiki/API.md` (секция Wiki planned) | **1–2 дня** docs + cross-link grep |
| Обновить gray-zone decisions / July roadmap ссылки на wiki | **0.5–1 день** |
| Код | **0** (нечего удалять) |

**Плюсы:** ноль drift «спека vs код vs ROADMAP»; планирование не врёт.  
**Минусы:** теряется готовая аналитика (ba/sa Wave 1–2); повторная спека дороже, если wiki вернётся; нужно явное новое решение по handbook/HR.

---

## 4. Side effects common to all options

| Артефакт | A Build | B Annotate | C Remove |
|----------|---------|------------|----------|
| `.planning/ROADMAP.md` / `PROJECT.md` | добавить фазы wiki | опционально «deferred» | без wiki (как сейчас по тексту) |
| July `ROADMAP-FINAL` 5.e / 8.a | исполнять | пометить deferred | вычеркнуть / seed |
| GZ-V4-13 handbook | зависит от wiki | interim или defer | пересмотреть GZ |
| GZ-SEC-17/18 wiki cascade | реализовать с модулем | `BLOCKED` до модуля | убрать wiki из GDPR narrative |
| `docs/wiki/` (repo docs) | не трогать (другой «wiki») | не трогать | не трогать |

---

## 5. Decision needed from owner

Отметить **одно**:

- [ ] **A** — Build (уточнить: MVP 01…04 only / full 01…09; когда относительно July kickoff)
- [ ] **B** — Out of v4 scope (spec annotation; keep text)
- [ ] **C** — Remove FR-WIKI / US-WIKI from active specs

После выбора — отдельная docs-задача (и при A — tech-lead фазы).  
**Этот файл решения не принимает.**

---

*Generated for planning-audit-2026-07 · C4 / B5 · documentation only*
