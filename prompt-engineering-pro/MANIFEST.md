---
manifest_version: 1
product: Touchstone
edition: pro
version: 0.3.0-beta.1
depends_on: ../pro-qa-ai/
status: foundation-ready
owner: automation-quality
---

# Манифест — Touchstone Pro

## Состав пакета

| Путь | Назначение | Владелец |
|------|------------|----------|
| `README.md`, `START-HERE.md`, `VERSION.md` | Вход и версия | automation-quality |
| `RELATIONSHIP.md` | Контракт Standard↔Pro | automation-quality |
| `SOURCES.md` | Карта источников Hub/GSD | automation-quality |
| `agents/` | 12 ролей разработки + corporate-шаблоны | automation-quality |
| `skills/` | phase-plan, wave-orchestrate, verify-goal | automation-quality |
| `playbooks/` | Процессы поставки продукта | automation-quality |
| `prompts/13-*` | Фрагменты корпоративной поставки | automation-quality |
| `adapters/cursor/` | Маппинг ролей в `.cursor/agents/` | владелец AI-платформы |
| `gsd-port/`, `PORTABLE-HUB-KIT.md` | Перенос kit | automation-quality |
| `layers/01`–`07` | Слои практик | automation-quality |
| `extracts/` | Переносимые тексты контрактов | automation-quality |

## Не входит в Pro package

- Сгенерированные `context/projects/*`
- Секреты, `.env`, approval records с живыми токенами
- Полные копии upstream GSD skill bodies
- Дубли того, что уже есть в Standard (context, rules, recipes, evals, scripts)
- Продуктовый код backend/frontend

## Acceptance curation pack

- [x] Слои 0–7 описаны
- [x] Роли разработки перенесены из Standard, ростер задокументирован
- [x] Граница изданий проведена по объекту работы
- [x] Extracts покрывают contract / delegate / wave / mutation / backlog
- [x] SOURCES ссылается на реальные пути Hub
- [x] Standard остаётся обязательной базой
- [ ] Paired Pro eval на внешнем репо (`NOT_RUN`)
