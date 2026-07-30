---
name: touchstone-standard
product: Touchstone
edition: standard
version: 0.2.0-rc.1
status: foundation-ready
release_date: TODO
owner: automation-quality
validation_status: STRUCTURAL_PASS
last_validation: 2026-07-31
release_eligibility: foundation-ready-eval-blocked
provenance: MANIFEST.md
supersedes: 0.1.0-alpha.1
---

# Версия — Touchstone Standard

Текущая semantic version: **`0.2.0-rc.1`**.

Это **foundation-ready** издание линейки Touchstone для тестовых фреймворков.
Корень поставки: эта папка `prompt-engineering/`. Продуктовая карта:
[TOUCHSTONE.md](TOUCHSTONE.md). Максимальное издание в Hub monorepo:
sibling folder `prompt-engineering-pro/` (не входит в portable package Standard).

## Что означает статус

| Термин | Значение |
|--------|----------|
| `foundation-ready` | Идентичность, вход, package, validator, governance готовы к переносу |
| `STRUCTURAL_PASS` | Дерево и валидатор прошли структурные проверки |
| `eval-blocked` | Paired effectiveness / ROI ещё `NOT_RUN` — GA `1.0.0` нельзя |

Это **не** заявление о completion effectiveness, conformance или ROI.

## Evidence (structural)

Замер 2026-07-31 на текущем дереве. Команды воспроизводимы из корня
`prompt-engineering/`.

| Проверка | Команда | Результат |
|----------|---------|-----------|
| Structural validation | `python scripts/validate_toolkit.py` | PASS · 0 errors / 0 warnings · exit 0 |
| Script tests | `python -m pytest scripts/tests -q` | 54 passed / 1 skipped (55 total); SKIP — symlink недоступен |
| Portable package | `python scripts/package_portable.py --root . --output <dst> --dry-run` | 174 files · exit 0 |

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
