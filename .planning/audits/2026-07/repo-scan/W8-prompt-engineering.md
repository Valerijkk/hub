# W8 — Prompt Engineering / Portable kit (readonly)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> Агент: `docs-writer` (Grok 4.5) · Дата: **2026-07-23** · Scope: `prompt-engineering/` — corporate agents, `prompts/13-corporate-delivery.md`, полнота PORTABLE kit.
> Код **не менялся**. Только findings. Язык: RU.

## Чек-лист (результат)

| Контроль | Статус | Evidence |
|----------|--------|----------|
| Corporate agents на диске (стиль A) | PASS | 11 файлов: orchestrator, product-owner, architect, backend-dev, frontend-dev, qa-manual, qa-auto, code-reviewer, debugger, security-reviewer, devops + 4 test-* |
| Ростер ↔ `agents/README.md` | PASS | таблица README совпадает с файлами |
| Обезличенность agents (нет Hub product-путей) | PASS | grep `backend/internal`, `frontend/src`, `be1`/`fe1`, `Fiber`, `LiveKit` в `agents/*.md` → 0 |
| `13-corporate-delivery.md` нумерация 1–112 | PASS | PowerShell: count=112, min=1, max=112, missing=∅ |
| Слоты / мастер-шаблон в №13 | PASS | слоты `{{goal}}`…; профиль разворота на `_MASTER-PROMPT-TEMPLATE.md` |
| GSD PE skills | PASS | `skills/{phase-plan,wave-orchestrate,verify-goal}/SKILL.md` + README |
| GSD PE playbooks | PASS | `gsd-lifecycle`, `wave-execution`, `handoff-state`, `portable-kit-dry-run` |
| Templates волны | PASS | `phase-kickoff`, `wave-delegate`, `corporate-wave` |
| User guide portable | PASS | `PORTABLE-HUB-KIT.md` (10 шагов, L1–L5, expected tree) |
| Context = template (не filled Hub) | PASS | `context/PROJECT-CONTEXT.md` → `status: template` |
| Ссылка `playbooks/corporate-delivery.md` | **FAIL** | файл **отсутствует**; ссылки в `orchestrator.md`, `product-owner.md` |
| README/INSTALL синхрон с corporate+GSD | PARTIAL | drift структуры/ростера/категорий (см. findings) |
| EXPANDED-EXAMPLES для категории 13 | FAIL | эталон только для 01–12 |
| Самодостаточный copy L2 без `ai-expansion/` | PARTIAL | dry-run/канон слоёв живут в `.planning/july-2026/` |

## Findings

| Sev | file:line / путь | Наблюдение | Рекомендация | Статус |
|-----|------------------|------------|--------------|--------|
| **HIGH** | `agents/orchestrator.md` §3; `agents/product-owner.md` §3 / §12 | Агенты грузят / ссылаются на `playbooks/corporate-delivery.md`, файла **нет** (`playbooks/` содержит 10 md, без `corporate-delivery`). Битая обязательная зависимость corporate-стиля. | Создать `playbooks/corporate-delivery.md` (конвейер PO→arch→dev→qa→security→handoff) **или** убрать ссылки, оставив `gsd-lifecycle` + `13-corporate-delivery` + `corporate-wave`. | open |
| **MEDIUM** | `README.md` (дерево §Структура, шапка) | Дерево всё ещё «тестовый тулкит»: 4 test-агента, skills без GSD, нет `gsd-port/`, `PORTABLE-HUB-KIT.md`; текст «~1000 … по **12** категориям» при наличии категории **13**. Ростер-таблица неполная (нет orchestrator/PO/arch/be/fe/security/devops). | Обновить дерево + ростер + счётчик категорий (12 test + 13 corporate). | open |
| **MEDIUM** | `INSTALL.md` целиком | INSTALL = 5 шагов только для test-smoke; **0** упоминаний `PORTABLE-HUB-KIT.md`, corporate agents, GSD skills/playbooks. Новый потребитель по INSTALL не видит corporate delivery path. | Добавить развилку: «только тесты» vs «Corporate Hub kit → PORTABLE-HUB-KIT.md»; smoke corporate: phase-kickoff docs-only. | open |
| **MEDIUM** | `prompts/EXPANDED-EXAMPLES.md` + `prompts/README.md` | Заявлены «12 эталонов» / «по одному на категорию»; категория 13 (112 промптов) **без** развёрнутого `#1`. Риск: corporate prompts остаются однострочниками без эталона профиля. | Добавить эталон `13 #1` (SMART-goal) + обновить таблицу профилей. | open |
| **MEDIUM** | `PORTABLE-HUB-KIT.md:13,41`; `playbooks/portable-kit-dry-run.md:15`; `playbooks/README.md:17` | Канон слоёв/dry-run — hard-path в `.planning/july-2026/ai-expansion/{05,08}`. После `cp -r prompt-engineering/` **без** L3 reference ссылки битые. Риск описан в PORTABLE §Риски, но L2 сам по себе не самодостаточен. | Вложить в PE краткий `gsd-port/LAYERS.md` (или mirror чек-листа DoD) **или** явно в INSTALL: «Must copy `ai-expansion/04–06,08`». | open (accepted risk, если L3 всегда копируют) |
| **MEDIUM** | `.planning/.../05-PORTABLE-KIT.md` §XOR («7+4 ролей») vs факт PE | Спека пишет «7+4»; на диске **11** corporate + **4** test. Drift манифеста → путаница при приёмке roster. | В `05` и acceptance: «11 corporate + 4 test» (как M2 в `07-ACCEPTANCE`). | open (planning drift) |
| **MEDIUM** | PE roster vs corporate prompts #74 / docs-sync | Нет роли `docs-writer` / tech-writer в PE; handoff/docs-sync промпты (#74) и dry-run Accept ссылаются на docs / `team-lead`. В стиле A docs-sync ложится на orchestrator/dev без явного владельца. | Добавить лёгкий `docs-writer.md` **или** явно закрепить docs-sync за `orchestrator` / `product-owner` в playbook. | open |
| **LOW** | `playbooks/portable-kit-dry-run.md:37` | Гейт Accept: `team-lead / docs` — Hub-имя роли в project-agnostic PE. | Заменить на `orchestrator / docs-writer` (или `{{orchestrator_role}}`). | open |
| **LOW** | Неравномерная глубина агентов | `orchestrator` / `backend-dev` / `architect` — полные §1–15 (DoD, few-shot); `qa-manual` / `qa-auto` / `code-reviewer` / `debugger` — короче, без явного DoD-чеклиста как у be/fe. | Выровнять минимальный скелет: DoD + «когда НЕ вызывать» + секция GSD (goal-backward / wave) по образцу backend-dev. | open (polish → 15.08) |
| **INFO** | `context/*.md` | `status: template` — корректно для kit; не копировать filled Hub context (совпадает с DoD `05`). | Оставить; при dry-run — intake на целевом репо. | accepted |
| **INFO** | Acceptance `07` G3/M1/M2/M3 | Must PE corporate + 112 prompts + dry-run docs — согласованы с фактом на диске (кроме битой ссылки corporate-delivery). | Закрыть HIGH выше → re-verify G3. | — |

### Закрытые / подтверждённые инварианты (не findings)

1. **Corporate roster полный для стиля A:** 11 агентов + разведение `qa-*` / `code-reviewer` / `debugger` vs `test-*` в `agents/README.md`.
2. **`13-corporate-delivery.md`:** секции Planning (1–15) → Waves (16–28) → Review (29–42) → Strangler (43–56) → Reduced GitHub (57–68) → Handoff (69–80) → Verification (81–92) → Portable-kit (93–102) → AI-team ops (103–112); композиции в хвосте валидны.
3. **Обезличенность тела агентов:** продукт-пути Hub в `agents/*.md` не найдены; факты только через `context/` + `{{слоты}}`.
4. **GSD-порт в PE:** `gsd-port/README.md` ↔ playbooks/skills/templates согласованы.
5. **Шаблоны делегирования:** `_MASTER-PROMPT-TEMPLATE`, `wave-delegate`, `phase-kickoff`, `corporate-wave` на месте.
6. **Portable user path:** `PORTABLE-HUB-KIT.md` покрывает copy → intake → XOR agents → smoke wave → kit_version в STATE.
7. **Skills README** включает GSD-триаду (phase-plan / wave-orchestrate / verify-goal).
8. **Согласие с `07-ACCEPTANCE` M1:** 112 ≥ 100 corporate prompts — PASS по факту.

## Полнота PORTABLE kit (сводка L1–L5 с точки зрения L2)

| Слой | Ожидание (`05`) | В `prompt-engineering/` | Вердикт |
|------|-----------------|-------------------------|---------|
| L2 PE целиком | Must | папка полная (~78 файлов) | PASS |
| L1 agents XOR | PE corporate | 11+4 agents | PASS (стиль A) |
| L1 prompt-contract | `.cursor/playbooks/` | **вне** PE (правильно для XOR Hub) | N/A в L2; Must при полном переносе |
| L1 gsd-ru | `.cursor/skills/gsd-ru/` | зеркало skills в PE (3) + gsd-port | PASS partial (полный RU — вне PE) |
| L3 planning templates | STATE/ROADMAP + ai-expansion 04/05/06/08 | только **ссылки** из PE | PARTIAL без copy L3 |
| L4 ADR stubs | Should | не в PE | out of L2 |
| L5 blueprints | `06` + strangler | промпты 43–56 + skill вне PE (`gsd-ru/microservices-strangler`) | PASS process / PARTIAL files |
| DoD: нет Hub-paths в PE agents | Must | PASS | PASS |
| DoD: context-intake + Verification | Must | skills + rules/40 | PASS (процесс) |
| DoD: dry-run | Should | playbook → `08` | PASS docs / зависит от L3 |

## Grep / проверки (кратко)

```
agents/*.md × Hub product paths     → 0
13-corporate-delivery numbered 1..112 → 112 contiguous
playbooks/corporate-delivery.md     → MISSING
context PROJECT-CONTEXT status      → template
EXPANDED-EXAMPLES category 13       → absent
INSTALL × PORTABLE|corporate|13     → 0 hits
```

## Вердикт

| Sev | Count |
|-----|-------|
| CRITICAL | 0 |
| HIGH | 1 (битая ссылка `corporate-delivery.md`) |
| MEDIUM | 6 |
| LOW | 2 |
| INFO/accepted | 2 |

**Итог:** corporate agents и библиотека `13` (112) — **содержательно PASS**; portable **user guide + GSD PE-зеркало** на месте. Блокер качества документации kit — **HIGH: отсутствующий playbook `corporate-delivery.md`**. Остальное — drift README/INSTALL/эталонов и зависимость L2→L3 для полного dry-run.

**verified: false** · **historical_self_check:** PASS (readonly audit scope) — 2026-07-23 · код не менялся.

### Next (docs / PE, не код продукта)

1. Создать или удалить ссылки на `playbooks/corporate-delivery.md` (HIGH).
2. Синхронизировать `README.md` + `INSTALL.md` с corporate/GSD/portable.
3. Эталон `EXPANDED-EXAMPLES` для `13 #1`.
4. Поправить «7+4» → «11+4» в `05-PORTABLE-KIT.md`.
5. (Polish) DoD-скелет для qa/debugger/code-reviewer; роль docs в стиле A.
