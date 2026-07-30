---
name: touchstone-pro
product: Touchstone
edition: pro
version: 0.2.0-rc.1
status: foundation-ready
depends_on: touchstone-standard@0.2.0-rc.1
owner: automation-quality
validation_status: CURATED_PACK
release_eligibility: foundation-ready-eval-blocked
---

# Версия — Touchstone Pro

**`0.2.0-rc.1`** — первая публичная раскладка Prompt Engineering Pro для линейки
Touchstone. Совпадает по major/minor с Standard: Pro — надстройка той же волны.

| Gate | Статус |
|------|--------|
| Standard structural | PASS (зависимость) |
| Pro curation completeness | CURATED_PACK (слои 0–7 + extracts) |
| Independent Pro package validator | N/A (не дублирует scripts/; валидируется Standard) |
| Paired Pro eval | `NOT_RUN` |

GA `1.0.0` для Pro = GA Standard **и** отдельный Pro smoke
(wave + mutation extract на пилотном репо) + owner sign-off.
