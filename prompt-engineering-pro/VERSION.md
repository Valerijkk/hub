---
name: touchstone-pro
product: Touchstone
edition: pro
version: 0.3.0-beta.1
status: foundation-ready
depends_on: touchstone-standard>=0.4.0-beta.1
owner: automation-quality
validation_status: CURATED_PACK
release_eligibility: foundation-ready-eval-blocked
supersedes: 0.2.0-rc.1
---

# Версия — Touchstone Pro

**`0.3.0-beta.1`** — издание перестало быть набором документов и получило
работающий состав: 12 ролей разработки, три скила жизненного цикла фазы,
семь плейбуков поставки, фрагменты категории 13, Cursor-адаптер и
portable hub kit.

## Что изменилось против `0.2.0-rc.1`

Граница между изданиями проведена заново — по **объекту работы**, а не по
уровню «базовый / продвинутый».

| | `0.2.0-rc.1` | `0.3.0-beta.1` |
|---|---|---|
| Состав Pro | 20 файлов документации | 57 файлов, из них 26 переехали из Standard |
| Роли разработки | лежали в Standard | здесь |
| Объект Standard | тестовый фреймворк **и** разработка | только тестовый фреймворк |

Это **ломающее** изменение для тех, кто ставил Standard `0.2.0-rc.1` ради
корпоративного ростера: после обновления роли разработки в Standard не
найдутся, нужно ставить Pro рядом.

Почему не `rc`: release candidate означает «релизим, если не всплывёт
багов», а по собственным воротам издание к релизу не допущено, пока не
прогнан paired eval. `beta` честнее.

## Ворота

| Gate | Статус |
|------|--------|
| Standard structural | PASS (зависимость, проверяется валидатором Standard) |
| Pro curation completeness | CURATED_PACK — слои 0–7, extracts, ростер, скилы, плейбуки |
| Independent Pro package validator | N/A — Pro не дублирует `scripts/`, проверяется вместе со Standard |
| Paired Pro eval | `NOT_RUN` |

GA `1.0.0` для Pro = GA Standard **и** отдельный Pro smoke
(волна + мутационный extract на пилотном репозитории) + owner sign-off.

## Совместимость со Standard

Pro объявляет минимальную версию Standard, против которой проверен, а не
жёсткое равенство major/minor: издания покрывают разные объекты работы и
будут расходиться в темпе. Правила загрузки — [RELATIONSHIP.md](RELATIONSHIP.md).
