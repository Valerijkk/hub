---
manifest_version: 1
product: Touchstone
edition: pro
version: 0.2.0-rc.1
depends_on: ../prompt-engineering/
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
| `layers/01`–`07` | Слои практик | automation-quality |
| `extracts/` | Переносимые тексты контрактов | automation-quality |

## Не входит в Pro package

- Сгенерированные `context/projects/*`
- Секреты, `.env`, approval records с живыми токенами
- Полные копии `.cursor/agents/*` и upstream GSD skill bodies
- Продуктовый код backend/frontend

## Acceptance curation pack

- [x] Слои 0–7 описаны
- [x] Extracts покрывают contract / delegate / wave / mutation / backlog
- [x] SOURCES ссылается на реальные пути Hub
- [x] Standard остаётся обязательной базой
- [ ] Paired Pro eval на внешнем репо (`NOT_RUN`)
