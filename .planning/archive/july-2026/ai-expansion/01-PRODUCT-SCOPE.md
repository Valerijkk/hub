# Product Scope — AI Expansion → 15.08.2026

## Vision

Собрать **переносимый AI quality layer для автоматизаторов**: generated context
существующего тестового фреймворка, approved patterns, вручную добавляемый опыт
коллег, agents/recipes, GSD-lite и измеримый eval. Corporate/reduced-GitHub и
microservices остаются расширением, а не ядром test-automation слоя.

## Goals (SMART, deadline 15.08.2026)

| ID | Goal | Metric |
|----|------|--------|
| G1 | Инвентарь GSD + план порта в планах Hub | `03-GSD-PORT-INVENTORY.md` полный, P0 закрыты в kit |
| G2 | Расширить все 32 agent-промпта секциями GSD-паттернов | 32/32 файлов содержат `GSD-паттерны (Hub-усиленные)` |
| G3 | PE: корпоративные агенты + GSD playbooks/skills | ≥7 corporate agents, ≥3 GSD playbooks, ≥3 skills |
| G4 | GSD-RU lite pack в `.cursor/skills/gsd-ru/` | 6 skills + README, индекс в `gsd-index` |
| G5 | Portable kit spec + reduced-GitHub blueprint | `05` + `06` + `PORTABLE-HUB-KIT.md` |
| G6 | Dynamic model routing | Sol Pro orchestrator + 31 `inherit`; catalog-backed decision/fallback log |

## MoSCoW

### Must (до 15.08)

- Зафиксировать в `.planning/archive/july-2026/ai-expansion/` полный план порта GSD → Hub/PE.
- Расширить промпты `.cursor/agents/*` практиками GSD (goal-backward, atomic, wave, handoff, anti-patterns).
- Расширить `prompt-engineering/`: GSD lifecycle/waves/handoff + corporate agents (не только тесты).
- Добавить path-based context bootstrap; закрытые MR не читать, примеры коллег
  человек добавляет вручную после bootstrap.
- Добавить structural validator и paired baseline-vs-assisted eval contract.
- GSD-RU skills pack (русский, с воротами Hub).
- Спека portable kit + blueprint «урезанный GitHub» + карта микросервисов (по ADR).
- Обновить STATE / INDEX / GSD-INTEGRATION.

### Should

- Библиотека промптов `prompts/13-corporate-delivery.md` (≥40).
- Шаблоны `phase-kickoff`, `wave-delegate`, `corporate-wave`.
- Чек-лист strangler выноса в skill `microservices-strangler`.

### Could

- Автогенерация agent-матрицы из prompt-contract §8.
- Примеры «перенос kit в пустой monorepo» (dry-run script, later).

### Won't (этот горизонт)

- Продовый код backend/frontend (July freeze).
- Реализация reduced-GitHub как runtime-модуля (только blueprint).
- Реальный вынос микросервисов (только docs/checklist).
- SSO/LDAP реализация (ждёт ADR-AUTH-SSO accept+implement).
- Копирование лицензионно-сомнительного upstream GSD кода целиком.

## Success metrics

| Metric | Status | Evidence |
|--------|--------|----------|
| `ai-expansion/` имеет scope→acceptance→rework chain | PASS | README + 01–09 |
| 32/32 Hub agents имеют компактный GSD routing | PASS | `.cursor/agents/` |
| PE GSD/corporate/portable layer | PASS structural | `gsd-port`, agents, recipes, packager |
| GSD-RU работает без обязательного upstream | PASS structural | skill names/routes validated |
| Portable package не содержит generated project data | PASS | 165-file package; Hub profile/smoke absent |
| Установка в физически пустой repo | NOT_RUN | release/pilot gate |
| Измеримое улучшение review/quality | NOT_RUN | paired eval required |

## Out of scope

Код продукта, деплой, FA-06 secret rotation, новые Ent schemas, UI redesign.

## Dependencies

| ADR / артефакт | Зачем |
|----------------|-------|
| ADR-ARCH-MICROSERVICES-2026 | Карта сервисов в portable arch |
| ADR-MODELS-GPT56-2026 | Cutover моделей в августе |
| prompt-contract.md | Единый контракт промпта |
| prompt-engineering/ (уже есть) | База PE, расширяем |
| GSD-INTEGRATION.md | Маршрутизация skills |

## Risks

| Risk | Mitigation |
|------|------------|
| Расползание в продовый код | Жёсткий Won't + file boundaries |
| Дублирование PE vs `.cursor/` | PE = project-agnostic; `.cursor` = Hub-specific |
| Upstream GSD drift | Портим паттерны, не vendor lock на `~/.claude` |
| Перегруз агентов текстом | Секции GSD ≤40 строк на роль |

## Verification

`product-ver` → scope/KPI; `arch-ver` → portable arch; `docs-writer` → links; финальная приёмка `team-lead`.
