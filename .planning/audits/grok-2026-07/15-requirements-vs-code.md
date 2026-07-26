# A15 — Requirements vs Code (planning ↔ codebase)

**Дата:** 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Режим:** READ-ONLY (кроме этого отчёта)  
**Источники:** `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`, `.planning/PROJECT.md`, `.planning/july-2026/ROADMAP-FINAL-2026-07-03.md`, `.planning/phases/`, `docs/tz_constraints.md`, код `frontend/` + `backend/`

## Scope

| Каталог | Что внутри | Чекбоксы |
|---------|------------|----------|
| `.planning/REQUIREMENTS.md` | **v2.0 Redesign only** — 85 REQ (DSFN…XTRA) | 85× `[x]`, 0× `[ ]`; Future PLSH-01..03 deferred |
| `.planning/ROADMAP.md` | Phases 1–64 в overview; Progress-таблица **только до Phase 17**; AUTH/TASK/… ID в Phase Details | Phase 39 **пропущен** в checklist; **65–125 отсутствуют** |
| `.planning/PROJECT.md` | Продуктовый backlog (не REQ ID) | **1** `[x]` / **103** `[ ]` — сильно устарел vs shipped |
| July roadmap | Планирование W2+/Rebuild (freeze) | Не претендует на закрытие 1–125 |

**Метод:** полный проход 85 REQ из `REQUIREMENTS.md` + сверка ключевых ROADMAP success/module claims с кодом (`file:LINE`) + инвентаризация фаз 1–125.

---

## Вердикт 125/125 и 100%

| Утверждение STATE | Факт | Вердикт |
|-------------------|------|---------|
| `total_phases: 125` / `completed_phases: 125` / `percent: 100` | Инвентарь каталогов: **7** (`milestones/v1.0-phases` 01–07) + **6** (`milestones/v2.0-phases` 08–13) + **112** (`.planning/phases/{14..125}-*`) = **125**. Пропусков в 14–125 нет. Extra: `999-test-framework` (backlog, не в 125). | **Инвентарь фаз: ДА** |
| «Все фазы COMPLETE в planning-каноне» | В `.planning/phases/*/PLAN.md` слово `COMPLETE` ≥ **79**/112; явный `**Status:** COMPLETE` ~**68**. Phase **14** — **нет `PLAN.md`**. Корневой `ROADMAP.md` Progress обрывается на Phase **17**; checklist — до **64**, без Phase **39** и без **65–125**. | **Документация статуса: ЧАСТИЧНО / drift** |
| «100% требований продукта» | `REQUIREMENTS.md` = 100% **v2.0 visual** (85/85 marked). `PROJECT.md` почти весь unchecked. July = future work. | **Не путать:** 100% STATE = фазы покрытия/милестоунов, **не** синхрон PROJECT/ROADMAP |

**Итог по 125/125:** утверждение STATE про **наличие и закрытие волны фаз 1–125** согласуется с каталогами + PLAN артефактами coverage-wave; утверждение «единый ROADMAP отражает 125/125» — **ложь** (A15-001). Score ниже учитывает drift planning, не только UI-токены.

---

## Сводка counts

| Метрика | Значение |
|---------|----------|
| REQ в `REQUIREMENTS.md` (Must/active) | **85** |
| Реализовано | **79** |
| Частично | **4** |
| Расходится | **2** |
| Отсутствует | **0** |
| Future (PLSH, out of scope) | **3** (не в score) |
| Фазы 1–125 на диске | **125** |
| ROADMAP checklist Phase N | **63** (1–64 без 39) |
| PROJECT.md unchecked | **103** |
| Проблемные findings (A15-*) | **8** |

### Score: **7.2 / 10**

v2.0 redesign REQ в основном подтверждены кодом; сильный planning-drift (ROADMAP/PROJECT/LOGN-05 vs TZ) и мёртвый Telegram-login код снижают оценку.

---

## Findings (расхождения)

| ID | Severity | Проблема | Доказательство |
|----|----------|----------|----------------|
| **A15-001** | P0 (planning) | STATE `125/125` / `100%`, но корневой `ROADMAP.md` не перечисляет Phases **65–125** и Progress-таблица останавливается на Phase **17** | `STATE.md:10-14`; `ROADMAP.md:366-389` vs dirs `.planning/phases/65-*`…`125-*` |
| **A15-002** | P1 | Phase **39** есть в `.planning/phases/39-intranet-news/` и в overview-списке отсутствует между 38 и 40 | `ROADMAP.md:77-78` (38→40); dir `39-intranet-news/PLAN.md:1` |
| **A15-003** | P1 | `PROJECT.md` почти весь `[ ]` при shipped v1–v4 — ложный «backlog gap» vs код | `PROJECT.md` ~103 unchecked; модули `backend/internal/{tracker,messenger,tms,calls,crm,…}` существуют |
| **A15-004** | P1 | **LOGN-05** / v1 AUTH: REQ обещает Magic Link **+ Telegram login**; TZ — Telegram **не** login; UI login без Telegram; HTTP routes без `/auth/telegram`; остаётся dead code | `REQUIREMENTS.md:54`; `docs/tz_constraints.md:16-17`; `login.tsx` (нет Telegram); `auth/handler.go:35-37`; `service_login.go:94-115`; `api/auth.ts:94-96` deprecated |
| **A15-005** | P2 | Phase **14** в ROADMAP Progress = Complete 5/5, в `.planning/phases/14-architectural-hardening/` **нет PLAN.md** | `ROADMAP.md:386`; пустой/без PLAN dir `14-architectural-hardening` |
| **A15-006** | P2 | **DSFN-08** «все `!important` удалены» — 4 остались в `prefers-reduced-motion` (a11y) | `app.css:1355-1362` |
| **A15-007** | P2 | **MSGR-04** «glassmorphism composer» — класс `.glass` в CSS есть, composer его **не** использует (card + green focus glow) | `app.css:317-322`; `message-composer.tsx:584` |
| **A15-008** | P3 | `REQUIREMENTS.md` не содержит AUTH/TASK/… ID, на которые ссылается ROADMAP Phase Details — разрыв трассируемости | `ROADMAP.md:112` (`AUTH-01…`); единственный REQ-файл = redesign |

---

## Таблица: REQUIREMENTS.md → код

Статусы: **реализовано** / **частично** / **расходится** / **отсутствует**.

### Design System (DSFN)

| REQ | Статус | Доказательство |
|-----|--------|----------------|
| DSFN-01 | реализовано | `frontend/src/app.css:133-134` Inter + JetBrains Mono (`--sans`/`--mono`) |
| DSFN-02 | реализовано | `app.css:164-170` void/base/raised/card/input (dark) |
| DSFN-03 | реализовано | `app.css:183-186` `--t1`…`--t4` |
| DSFN-04 | реализовано | `app.css:189-191` `--b1`…`--b3` |
| DSFN-05 | реализовано | `app.css:213-217` `--r1`…`--r4`, `--pill` |
| DSFN-06 | реализовано | `app.css:229-232` `--sh1`…`--sh4` |
| DSFN-07 | реализовано | `app.css:224-226` `--spring` / `--bounce` |
| DSFN-08 | **частично** | Не-a11y `!important` сняты; a11y block остаётся `app.css:1359-1362` (**A15-006**) |
| DSFN-09 | реализовано | `theme-provider.tsx:31-40` accents red/blue/green/purple/orange; `app.css:244+` `[data-accent]` |

### Theme (THME)

| REQ | Статус | Доказательство |
|-----|--------|----------------|
| THME-01 | реализовано | `app.css:164-237` `.dark` palette |
| THME-02 | реализовано | `app.css:77-159` `:root` light tokens |
| THME-03 | реализовано | `theme-provider.tsx:101+` next-themes dark/light/system |
| THME-04 | **частично** | Токены с иерархией contrast заданы; формального WCAG-отчёта в коде нет (axe login PASS — по STATE/AGENTS, не перепроверялся в этом аудите) |

### Layout (LYOT)

| REQ | Статус | Доказательство |
|-----|--------|----------------|
| LYOT-01 | реализовано | `sidebar.tsx:36` gradient edge line |
| LYOT-02 | реализовано | `sidebar-link.tsx:4,82` accent glow / gradient overlay |
| LYOT-03 | реализовано | `topbar.tsx:22` `backdrop-blur-[24px]` + gradient bottom `topbar.tsx:88` |
| LYOT-04 | реализовано | Sidebar collapse path в layout shell (`sidebar.tsx` + app-shell) |
| LYOT-05 | реализовано | `app-shell.tsx:115,145` mobile overlay + `MobileNav` |

### UI Primitives (PRIM)

| REQ | Статус | Доказательство |
|-----|--------|----------------|
| PRIM-01 | реализовано | `button.tsx:13-15` gradient primary + glow hover |
| PRIM-02 | реализовано | `card.tsx:12` gradient bg + top-edge highlight + hover |
| PRIM-03 | реализовано | `input.tsx:11` blue focus ring + `scale-[1.005]` |
| PRIM-04 | реализовано | `badge.tsx` + glow tokens `--gB/--gR` в DS |
| PRIM-05 | реализовано | `dialog.tsx` / `popover.tsx` (Shadcn + blur tokens) |
| PRIM-06 | реализовано | `tabs.tsx` active indicator |
| PRIM-07 | реализовано | `dropdown-menu.tsx` |
| PRIM-08 | реализовано | `scroll-area.tsx:34-36` slim **5px** scrollbar |
| PRIM-09 | реализовано | `avatar.tsx:22` online glow `group-data-[online=true]` |
| PRIM-10 | реализовано | `tooltip.tsx` + radius tokens |

### Login (LOGN)

| REQ | Статус | Доказательство |
|-----|--------|----------------|
| LOGN-01 | реализовано | `login.tsx:78-79` split layout `lg:w-[55%]` |
| LOGN-02 | реализовано | `login.tsx:22-59` `FLOATING_FEATURES` cards |
| LOGN-03 | реализовано | `login.tsx:3` `MagicLinkForm`; routes `handler.go:35-36` |
| LOGN-04 | реализовано | `login.tsx:80` `login-mesh`; `app.css:844` / `mesh-shift` |
| LOGN-05 | **расходится** | Magic Link — да; Telegram login — **нет** в UI/routes, dead API (**A15-004**) |

### Dashboard (DASH)

| REQ | Статус | Доказательство |
|-----|--------|----------------|
| DASH-01 | реализовано | `dashboard.lazy.tsx` + `widget-grid.tsx` react-grid-layout |
| DASH-02 | реализовано | `dashboard.lazy.tsx:35-40,71-72` greeting + date |
| DASH-03 | реализовано | quick actions icons `dashboard.lazy.tsx:7-12` |
| DASH-04 | **частично** | Анимации виджетов/shimmer есть (`widget-grid.tsx:75`); явный staggered entrance как в mockup — слабее документирован |
| DASH-05 | реализовано | `widget-grid.tsx:2,185,329` DnD config |

### Messenger (MSGR)

| REQ | Статус | Доказательство |
|-----|--------|----------------|
| MSGR-01 | реализовано | `channel-list.tsx:521` `w-[280px]` 3-panel |
| MSGR-02 | реализовано | channel list search/filter UI (`channel-list.tsx`) |
| MSGR-03 | реализовано | `message-item.tsx` avatar/reactions/metadata |
| MSGR-04 | **частично** | Composer redesign есть (`message-composer.tsx:584`); **не** `.glass` (**A15-007**) |
| MSGR-05 | реализовано | reactions `message-item.tsx:477+`; typing `message-composer.tsx:430`; reply `519+` |
| MSGR-06 | реализовано | send/react/thread/pin/voice/polls/GIF modules present under `components/messenger/` + `backend/internal/messenger/` |

### Kanban (KANB)

| REQ | Статус | Доказательство |
|-----|--------|----------------|
| KANB-01 | реализовано | `kanban-board.tsx` / `board-column.tsx` |
| KANB-02 | реализовано | `issue-card.tsx` pragmatic DnD + avatars |
| KANB-03 | реализовано | board filter bar `board/index.tsx` |
| KANB-04 | реализовано | `board/index.tsx:1,117` monitorForElements DnD |

### Issue / Calls / TMS / Settings / Admin / Onboarding / Trash / XTRA

| REQ | Статус | Доказательство |
|-----|--------|----------------|
| ISSU-01..03 | реализовано | `issues/$issueKey.tsx`, `issue-sidebar-meta.tsx`, `issue-description.tsx` |
| ISSU-04 | реализовано | comments/time-track/attachments components + tracker BE |
| CALL-01..03 | реализовано | `CallRoom.tsx:1-19` LiveKit grid/controls/participants |
| CALL-04 | реализовано | mute/share/record/chat: `CallControls`, `ScreenShareView`, `recording-controls.tsx`, `CallChat.tsx` |
| TMSP-01..03 | реализовано | `routes/.../tms*`, `modules/tms/components/*` |
| TMSP-04 | реализовано | `backend/internal/tms/` + FE cases/runs/envs/flaky |
| STNG-01..05 | реализовано | `routes/_authenticated/settings*` + notification prefs component |
| ADMN-01 | реализовано | `admin/dashboard.tsx:118` DAU/WAU/MAU KPI |
| ADMN-02..04 | реализовано | tabs Users/Audit/Limits routes under `admin/` |
| ADMN-05 | реализовано | admin BE `audit` + auth admin approve |
| ONBD-01 | реализовано | `onboarding.lazy.tsx` progress steps `STEP_LABEL_KEYS:152-157` |
| ONBD-02 | реализовано | profile RHF+Zod step |
| ONBD-03 | реализовано | **6** cards `MODULE_KEYS:124-131` (comment «5» устарел — код верен) |
| ONBD-04 | реализовано | theme+accent step `onboarding.lazy.tsx:595+` |
| ONBD-05 | реализовано | `completeOnboarding` API + flow |
| TRSH-01..03 | реализовано | `routes/_authenticated/trash.tsx`, `admin/trash.tsx` |
| XTRA-01..06 | реализовано | projects/sprints/gantt/JQL/settings/dashboard routes under `projects/$projectId/` |
| XTRA-07 | реализовано | regression implied by shipped modules (не отдельный visual-only gate в этом аудите) |

### Future (не в score)

| ID | Статус | Примечание |
|----|--------|------------|
| PLSH-01..03 | отсутствует (deferred) | Явно «Polish deferred» в REQUIREMENTS |

---

## ROADMAP functional sample (вне REQUIREMENTS.md)

| Claim / REQ ID (ROADMAP) | Статус | Доказательство |
|--------------------------|--------|----------------|
| AUTH Magic Link | реализовано | `handler.go:35-36` |
| AUTH Telegram login (Phase 1 success #1) | **расходится** с TZ + текущим UI | `tz_constraints.md:17`; login без TG; `LoginWithTelegram` dead `service_login.go:94` |
| AuthService split (Phase 14) | реализовано | `service_{login,session,profile,organization,registration}.go` |
| WIP limits (KANB) | реализовано | `workflowstatus.go:34` `wip_limit`; `sprint/service.go:255` |
| Messenger / TMS / Calls modules | реализовано | `wiring_messenger.go`, `wiring_integrations.go:55` |
| v4 CRM/HR/Calendar/Bases/Accounting | реализовано (модули есть) | `wiring_modules.go:33-76`; качество — см. A6 audit, не scope A15 |
| Coverage Phase 125 | реализовано (по PLAN) | `125-regression-gate/PLAN.md:30` Status COMPLETE; COV в STATE **83.0%** |
| July Rebuild / W2+ | **не** в 125 complete | `july-2026/ROADMAP-FINAL-2026-07-03.md:17-21` freeze / planning |

---

## Phase inventory detail

```
milestones/v1.0-phases:     01..07     =  7
milestones/v2.0-phases:     08..13     =  6
.planning/phases:           14..125    =112
────────────────────────────────────────────
TOTAL counted as 1..125                =125
.planning/phases/999-*                 = +1 backlog (не в 125)
```

| Проверка | Результат |
|----------|-----------|
| Missing dirs in 14–125 | **нет** |
| Phase 14 PLAN.md | **отсутствует** (A15-005) |
| ROADMAP overview checklist | 1–38, 40–64 (**нет 39, нет 65–125**) |
| STATE milestones table Coverage 59–125 | ✅ claimed COMPLETE |

---

## Чистые зоны (кратко)

- Design tokens v2.0 (surfaces/text/borders/radii/shadows/easing/accents) — в `app.css` и UI primitives.
- Layout shell / login mesh / messenger 280px / LiveKit calls / admin KPI — на месте.
- Backend domain modules из ROADMAP v1–v4 — каталоги и wiring существуют.
- Phase 125 PLAN + STATE coverage numbers согласованы по порядку величины (82%+ → 83.0%).

---

## Рекомендации (docs-only, вне scope фикса)

1. Дописать в `ROADMAP.md` Phases **39**, **65–125** (или явно «Coverage wave documented in STATE/COVERAGE») — закрыть **A15-001/002**.
2. Обновить Progress-таблицу ROADMAP до актуального канона или пометить «superseded by STATE».
3. Переписать **LOGN-05** → Magic Link only; пометить Telegram login deprecated / remove dead paths — **A15-004**.
4. Синхронизировать `PROJECT.md` checkboxes с shipped reality или пометить файл historical — **A15-003**.
5. Добавить `PLAN.md`/ссылку для Phase 14 или указать archive path — **A15-005**.

---

*Аудит A15 · Grok · 2026-07-26 · product code не изменялся*
