# W8 — GSD-RU (`.cursor/skills/gsd-ru/`)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Агент:** docs-writer (Grok 4.5) · **Дата:** 2026-07-23 · **Режим:** readonly repo-scan
> **Scope:** глубина pack `gsd-ru` (6 skills + README), сверка с P0 из `03-GSD-PORT-INVENTORY.md` / `GSD-INTEGRATION.md`, PE-зеркало, маршрутизация `gsd-index`.
> **Код не менялся.** Язык отчёта: **RU**.

## Вердикт (кратко)

- **Pack на месте:** 6 skills + README; ядро цикла `phase-plan → wave-run → goal-verify → handoff` + Hub-native `portable-kit` / `microservices-strangler`.
- **Глубина skills:** **PASS** — все `SKILL.md` **≥80 строк** (факт: 144–197), есть frontmatter, шаблон, worked example Hub, DoD, anti-patterns, «Связь».
- **P0 upstream → gsd-ru:** **6/21** закрыты локальным skill (**4** lifecycle + 2 Hub-native вне P0-счётчика). **15/21** намеренно **не** как RU skill — agents/practice/route; статус mid-Aug = **PARTIAL** (polish → 15.08).
- **PE-зеркало:** skills `phase-plan` / `wave-orchestrate` / `verify-goal` — **stubs (~13–16 LOC)** vs Hub `gsd-ru` (~144–159) — асимметрия глубины для portable consumers.
- **verified:** `false` · **historical_self_check:** PASS (pack depth facts scope); gaps P0/agents и PE stubs — open, не блокеры Must G4 (`07-ACCEPTANCE.md`).

---

## 1. Inventory pack (факт filesystem)

| Артефакт | Строк (файл) | Frontmatter | Когда | Шаблон/чек-лист | Worked example | DoD | Anti-patterns | Связь |
|----------|-------------:|:-----------:|:-----:|:---------------:|:--------------:|:---:|:-------------:|:-----:|
| `phase-plan/SKILL.md` | 159 | ✅ | ✅ | ✅ PLAN | ✅ bases | ✅ | ✅ | ✅ |
| `wave-run/SKILL.md` | 150 | ✅ | ✅ | ✅ §5.2 | ✅ be1 bases | ✅ | ✅ | ✅ |
| `goal-verify/SKILL.md` | 144 | ✅ | ✅ | ✅ отчёт | ✅ + негативный | ✅ | ✅ | ✅ |
| `handoff/SKILL.md` | 144 | ✅ | ✅ | ✅ STATE + §5.2 | ✅ bases→ver | ✅ | ✅ | ✅ |
| `portable-kit/SKILL.md` | 148 | ✅ | ✅ | ✅ copy/sanitize | ✅ Hub→репо | ✅ | ✅ | ✅ |
| `microservices-strangler/SKILL.md` | 197 | ✅ | ✅ | ✅ PLAN выноса | ✅ notification | ✅ | ✅ | ✅ |
| `README.md` | ~110 | — | — | DoD пакета | ✅ сквозной bases | ⚠️ чекбоксы `[ ]` | ✅ | ✅ |

**DoD пакета (README):** цель ≥80 строк — **выполнена** по всем 6 skills. Чекбоксы DoD в README **все unchecked** (`[ ]` ×6, `[x]` ×0) — гигиена статуса, не отсутствие содержания.

**Ссылки README (выборочно):** `56-regression-verify.mdc`, `team-lead.md`, `05-PORTABLE-KIT.md`, `ADR-ARCH-MICROSERVICES`, `gsd-index`, `test-triage` — **существуют**.

---

## 2. Покрытие P0 (21) ↔ gsd-ru

Источник: `.planning/archive/july-2026/ai-expansion/03-GSD-PORT-INVENTORY.md` §3 + `.cursor/GSD-INTEGRATION.md` § «P0/P1 port status mid-Aug».

### 2.1 Закрыто локальным gsd-ru skill (6 upstream P0 → 4 skills)

| Upstream P0 | gsd-ru | Оценка глубины |
|-------------|--------|----------------|
| `gsd-plan-phase`, `gsd-analyze-dependencies` | `phase-plan` | PASS — goal/Won't/depends/волны/rule 56/example bases |
| `gsd-execute-phase` | `wave-run` | PASS — полный §5.2, лимиты параллели, GZ-QA-13 |
| `gsd-verify-work` | `goal-verify` | PASS — goal-backward + негативный пример + матрица evidence |
| `gsd-progress`, `gsd-resume-work` (+ pause) | `handoff` | PASS — STATE + rule 55 + §5.2 handoff |

Hub-native (не из списка 21 P0, но в pack): `portable-kit`, `microservices-strangler` — **глубина OK**, planning-only strangler согласован с ADR.

### 2.2 Missing P0 как gsd-ru skill (15) — intentional port target = agents

| Upstream P0 | in Hub? (inventory) | Port target до 15.08 | Риск без RU skill |
|-------------|---------------------|----------------------|-------------------|
| `gsd-discuss-phase`, `gsd-spec-phase` | practice · ba1/ba2 | agents ba* + PE | Lifecycle «типовой день» в README **стартует с phase-plan** — discuss/spec вне lite-цикла |
| `gsd-plant-seed` | practice · `.planning/reference/seeds/` | ba2/product | Низкий — практика seeds есть |
| `gsd-workstreams` | practice · team-workflow | PE wave-execution + team-lead | Частично в `wave-run` (streams be1∥be2); отдельного skill нет |
| `gsd-quick`, `gsd-fast` | partial · route | be*/fe*/codex-micro | Без ~/.claude — только route в gsd-index |
| `gsd-add-tests`, `gsd-validate-phase` | partial · route | qa-auto* | Nyquist/add-tests не в RU |
| `gsd-audit-uat` | partial · route | qa-manual | Нет RU UAT-аудита |
| `gsd-code-review`, `gsd-code-review-fix` | practice · code-reviewer | code-reviewer | OK как agent practice |
| `gsd-secure-phase` | practice · security-auditor | security-auditor (+ strangler при ADR) | Strangler ≠ полный secure-phase |
| `gsd-debug` | practice · debugger + test-triage | debugger | Hub `/test-triage` закрывает ворота; полный gsd-debug — upstream |
| `gsd-docs-update` | practice · docs-writer + rule 55 | docs-writer | OK |
| `gsd-ship` | practice · devops | devops-engineer | Нет RU ship checklist |

**Итого:** для **локального lite без `~/.claude`** закрыт **ежедневный оркестрационный** контур; **quality/BA/ship P0** зависят от глубины секций в agents и наличия upstream — это заявленный gap M6 в `07-ACCEPTANCE.md` (PARTIAL).

---

## 3. Findings

| Sev | Место | Наблюдение | Рекомендация | Статус |
|-----|-------|------------|--------------|--------|
| **MEDIUM** | `prompt-engineering/skills/{phase-plan,wave-orchestrate,verify-goal}/SKILL.md` | PE-зеркало ядра цикла = **13–16 строк** (чеклист без шаблона/example/DoD). `gsd-ru` = **144–159**. Portable kit consumers получают «каркас», не operable skill. | До 15.08: расширить PE skills до ≥80 (project-agnostic шаблон + DoD + anti-patterns) **или** явно в `gsd-port/README` пометить «stub → копируй паттерн из playbooks». | open |
| **MEDIUM** | `gsd-ru/README.md` § «Типовой день» + P0 discuss/spec | Lite-цикл **пропускает** discuss→spec (P0). Оркестратор без ba* и без upstream не имеет RU skill на «что строим». | В README: явный шаг «до phase-plan → ba discuss/spec **или** upstream `gsd-discuss-phase`/`gsd-spec-phase`»; не раздувать pack без нужды. | open |
| **MEDIUM** | `.cursor/skills/gsd-index/SKILL.md` § роли | Таблица ролей всё ещё зовёт **upstream** `gsd-plan-phase` / `gsd-execute-phase` / `gsd-verify-work` первыми; `gsd-ru/*` только в отдельной таблице сверху. Риск: агент на машине **без** `~/.claude` идёт в «пустой» skill name. | В role-таблице для daily cycle поставить **primary = `gsd-ru/…`**, upstream — secondary. | open |
| **LOW** | `gsd-ru/README.md` DoD пакета | 6 чекбоксов DoD остаются `[ ]` при фактически выполненном объёме/структуре (версия README: 2026-07-23 max expansion). | Проставить `[x]` после точечной сверки (или «статус: PASS» одной строкой). | open |
| **LOW** | `GSD-INTEGRATION.md:107` | Строка «`gsd-handoff` via» — **нет** такого upstream skill в 79; канон = `progress`/`pause`/`resume` → `gsd-ru/handoff` + `/handoff-sync`. | Заменить на `gsd-ru/handoff` / `gsd-progress`+`pause`+`resume`. | open |
| **INFO** | Pack scope vs 21 P0 | **6/21** в gsd-ru — by design (inventory §7). Не считать «недопорт 15 skills в gsd-ru» дефектом pack. | Добивать **agents** (M6), не плодить 15 новых RU skills. | accepted |
| **INFO** | PE без `handoff` / `portable-kit` / `strangler` skills | Playbook `handoff-state` есть; Hub-specific skills намеренно в `.cursor/skills/gsd-ru/`. | Оставить; в portable-kit skill уже описан copy-path. | accepted |
| **INFO** | `07-ACCEPTANCE` G4 / M5 | G4 PASS (6 skills); M5 PASS (ядро); M6 PARTIAL. | Согласовано с этим сканом. | accepted |

### CRITICAL / HIGH

**0** — дефектов целостности pack (битые обязательные секции, skills <80 строк, отсутствующий skeleton) **не найдено**.

---

## 4. Сверка глубины vs DoD README

| Критерий DoD пакета | Факт |
|---------------------|------|
| Каждый skill ≥80 строк, RU | ✅ 144–197 |
| Полный шаблон + DoD + anti-patterns | ✅ |
| ≥1 worked example с путями Hub | ✅ (bases / notification / portable) |
| Ссылки §5.2 / rule 56 / ADR | ✅ (проверены пути) |
| README отражает 6 skills + сквозной пример | ✅ |
| Изменения в STATE | вне scope этого readonly-скана |

---

## 5. Связанные артефакты (не правились)

| Документ | Роль |
|----------|------|
| `.planning/archive/july-2026/ai-expansion/03-GSD-PORT-INVENTORY.md` | Канон 79 / P0=21 |
| `.cursor/GSD-INTEGRATION.md` | Маршрутизация + mid-Aug P0/P1 |
| `.planning/archive/july-2026/ai-expansion/07-ACCEPTANCE.md` | G4/M5/M6 |
| `prompt-engineering/gsd-port/README.md` | PE-зеркало |
| `.cursor/skills/gsd-index/SKILL.md` | Быстрый индекс |

---

## 6. Вердикт числами

| Метрика | Значение |
|---------|----------|
| gsd-ru skills | **6** + README |
| Upstream P0 закрыты gsd-ru | **6/21** (28.6%) |
| P0 → agents/route (не RU skill) | **15/21** |
| Skills <80 LOC | **0** |
| CRITICAL | **0** |
| HIGH | **0** |
| MEDIUM | **3** (PE stubs; discuss/spec вне lite; gsd-index primary) |
| LOW | **2** (README DoD unchecked; `gsd-handoff` typo) |
| INFO/accepted | **3** |

**Блокер Must AI-expansion (G4):** нет.
**До 15.08:** polish M6 (agents P0) + углубление PE stubs + правка маршрутизации gsd-index / typo INTEGRATION.

**verified: false** · **historical_self_check:** PASS (filesystem/inventory/ACCEPTANCE scope) — 2026-07-23 · readonly.
)
