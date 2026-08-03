---
name: touchstone-standard
product: Touchstone
edition: standard
version: 0.3.0-beta.1
status: foundation-ready
release_date: TODO
owner: automation-quality
validation_status: STRUCTURAL_PASS
last_validation: 2026-08-01
release_eligibility: foundation-ready-eval-blocked
provenance: MANIFEST.md
supersedes: 0.2.0-rc.1
---

# Версия — Touchstone Standard

Текущая semantic version: **`0.3.0-beta.1`**.

Это **foundation-ready** издание линейки Touchstone — AI-слой **для тестового
фреймворка**. Корень поставки: эта папка. Продуктовая карта:
[TOUCHSTONE.md](TOUCHSTONE.md).

Издание для разработки продукта — Pro, sibling `prompt-engineering-pro/`
(в portable package Standard не входит).

## Ломающее изменение в `0.3.0-beta.1`

Из Standard вынесены 26 файлов слоя разработки: 12 ролей продуктовой
команды, три скила жизненного цикла фазы, пять плейбуков поставки,
фрагменты категории 13, gsd-port и portable hub kit. Все они переехали
в Pro.

Причина: Standard раздавал корпоративный ростер автоматизаторам, которым
он не нужен, и одновременно делал вид, что покрывает разработку. Теперь
граница проходит по объекту работы — Standard про тестовый фреймворк,
Pro про продукт.

Кто ставил `0.2.0-rc.1` ради ростера — после обновления не найдёт его в
Standard; нужно поставить Pro рядом. Package: 174 → 148 файлов.

Почему `beta`, а не `rc`: release candidate обещает «релизим, если не
всплывёт багов», а по собственным воротам издание к релизу не допущено,
пока `paired eval` не прогнан.

## Что означает статус

| Термин | Значение |
|--------|----------|
| `foundation-ready` | Идентичность, вход, package, validator, governance готовы к переносу |
| `STRUCTURAL_PASS` | Дерево и валидатор прошли структурные проверки |
| `eval-blocked` | Paired effectiveness / ROI ещё `NOT_RUN` — GA `1.0.0` нельзя |

Это **не** заявление о completion effectiveness, conformance или ROI.

## Evidence (structural)

Замер 2026-08-01, перезамерено после замены мутации на проверку чувствительности. Команды воспроизводимы из корня
`pro-qa-ai/`.

| Проверка | Команда | Результат |
|----------|---------|-----------|
| Structural validation | `python scripts/validate_toolkit.py` | PASS · 0 errors / 0 warnings · exit 0 |
| Script tests | `python -m pytest scripts/tests -q` | 54 passed / 1 skipped (55 total); SKIP — symlink недоступен |
| Portable package | `python scripts/package_portable.py --root . --output <dst> --dry-run` | 148 files · exit 0 |

Package исключает generated Hub profile, smoke и raw feedback. Целостность
состава защищена `required_paths` в `scripts/package_manifest.json`: удаление
любого обязательного файла роняет пакер с exit 2 (проверено мутацией на
`TOUCHSTONE.md` 2026-07-31).

## Продвижение к GA `1.0.0`

1. Зелёные validator unit tests на intended distribution
2. Структурная валидация exact package
3. Approved source + governance review
4. Representative paired eval evidence
5. Sign-off accountable owner

Пока п.4–5 открыты: `validation_status = STRUCTURAL_PASS`,
`release_eligibility = foundation-ready-eval-blocked`.

История: [CHANGELOG.md](CHANGELOG.md) · владение: [MANIFEST.md](MANIFEST.md).
