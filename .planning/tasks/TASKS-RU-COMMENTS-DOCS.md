# TASKS — Русификация комментариев и документации

> Создано: 2026-06-21 · Цель: **100% комментариев и docs на русском**, подробно, актуально по коду  
> Верификация: 10 агентов после всех 30 исполнителей

## Правила (из `docs/code_style_guide.md`)

- Комментарии объясняют **ЗАЧЕМ**, не **ЧТО**
- Без placeholder-комментариев (`// Handle error`)
- Без эмодзи в коде
- Godoc / JSDoc — на русском
- Логику кода **не менять**, только комментарии и docs

---

## Wave RU-1 — Исполнители (30 агентов)

| ID | Область | Пути | Агент | Статус |
|----|---------|------|-------|--------|
| RU-D01 | Бизнес-контекст | `docs/vision_and_scope.md`, `docs/BUSINESS.md` | docs-writer | ✅ |
| RU-D02 | Системная спецификация | `docs/system_specification.md` | sa + docs-writer | ✅ |
| RU-D03 | Архитектура | `docs/architecture.md`, `docs/wiki/ARCHITECTURE.md` | arch + docs-writer | ✅ |
| RU-D04 | Тест-план | `docs/test_plan.md`, `docs/TESTING.md` | qa + docs-writer | ✅ |
| RU-D05 | Стек и инструменты | `docs/stack_and_tools.md`, `docs/TECHNICAL.md` | docs-writer | ✅ |
| RU-D06 | Стиль и безопасность | `docs/code_style_guide.md`, `docs/security_guidelines.md`, `docs/tz_constraints.md` | docs-writer | ✅ |
| RU-D07 | Wiki API/Deploy | `docs/wiki/API.md`, `docs/wiki/DEPLOY.md`, `docs/wiki/QUICKSTART.md` | docs-writer | ✅ |
| RU-D08 | Wiki прочее | `docs/wiki/HOME.md`, `docs/wiki/SHORTCUTS.md`, `docs/README.md` | docs-writer | ✅ |
| RU-D09 | Аналитика | `docs/competitive_audit.md`, `docs/competitive_analysis_2026.md`, `docs/user_needs_analysis_2026.md` | ba + docs-writer | ✅ |
| RU-D10 | Планирование | `.planning/STATE.md`, `.planning/COVERAGE.md`, `AGENTS.md`, `.cursor/REGULATION.md` | docs-writer | ✅ |
| RU-B01 | auth | `backend/internal/auth/**/*.go` (не `*_test.go` если только testframework) | be | ✅ |
| RU-B02 | ws + rbac | `backend/internal/ws`, `backend/internal/rbac` | be | ✅ |
| RU-B03 | audit + notification | `backend/internal/audit`, `backend/internal/notification` | be | ✅ |
| RU-B04 | tracker | `backend/internal/tracker` | be | ✅ |
| RU-B05 | tms | `backend/internal/tms` | be | ✅ |
| RU-B06 | project + sprint | `backend/internal/project`, `backend/internal/sprint` | be | ✅ |
| RU-B07 | messenger + calls | `backend/internal/messenger`, `backend/internal/calls` | be | ✅ |
| RU-B08 | search + storage + dashboard | `backend/internal/search`, `storage`, `dashboard` | be | ✅ |
| RU-B09 | platform | `backend/internal/platform` | be | ✅ |
| RU-B10 | apigateway + imex + backup | `backend/internal/apigateway`, `imex`, `backup` | be | ✅ |
| RU-B11 | github + gitlab | `backend/internal/github`, `gitlab` | be | ✅ |
| RU-B12 | ai + intranet + прочие | `backend/internal/ai`, `intranet`, `crm`, `mail`, … | be | ✅ |
| RU-B13 | ent schemas | `backend/ent/schema/*.go` | be | ✅ |
| RU-B14 | cmd + main | `backend/cmd/**` | be | ✅ |
| RU-B15 | testframework go + e2e | `backend/testframework/**`, `testframework/e2e/**` | be + qa | ✅ |
| RU-F01 | frontend api + hooks + stores | `frontend/src/api`, `hooks`, `stores` | fe | ✅ |
| RU-F02 | frontend modules | `frontend/src/modules` | fe | ✅ |
| RU-F03 | frontend components + routes | `frontend/src/components`, `routes` | fe | ✅ |
| RU-F04 | frontend testframework | `testframework/e2e`, `frontend/src/**/*.test.ts` | fe + qa | ✅ (e2e в B15) |
| RU-F05 | deployments + cli | `deployments/**`, `cli/**` | devops-engineer | ✅ |
| RU-M01 | README корень | `README.md`, `frontend/README.md` | docs-writer | ✅ (frontend README нет) |

---

## Wave RU-V — Верификаторы (10 агентов)

| ID | Проверка | Метод | Статус |
|----|----------|-------|--------|
| RU-V01 | docs/ полнота русского | grep `[A-Za-z]{4,}` в заголовках/абзацах, отчёт | ba-ver | ⏳ |
| RU-V02 | docs/ актуальность vs код | gsd-doc-verifier | sa-ver | ⏳ |
| RU-V03 | backend auth/ws/rbac/audit | grep `^// [A-Za-z]` + выборочное чтение | be-ver | ⏳ |
| RU-V04 | backend tracker/tms/project | то же | be-ver | ⏳ |
| RU-V05 | backend messenger/calls/platform | то же | be-ver | ⏳ |
| RU-V06 | backend остальные модули | то же | be-ver | ⏳ |
| RU-V07 | ent + cmd | то же | be-ver | ⏳ |
| RU-V08 | frontend src | grep `// [A-Za-z]` + JSX `{/* */}` | fe-ver | ⏳ |
| RU-V09 | testframework + e2e | комментарии в тестах/steps | qa-ver | ⏳ |
| RU-V10 | Итоговый gate | сводный отчёт 100% / gaps → TASKS | verifier | ⏳ |

**Критерий PASS RU-V10:** 0 английских комментариев (кроме `@swagger`, URL, имен полей API, `@example` JSON).

---

## Оркестрация

```
Wave RU-1: 30 пакетов → 9 агентов ✅ COMPLETE (2026-06-21)
    → Wave RU-V: 10 верификаторов (2 волны × 5) ⏳ NEXT
    → FIX gaps → повтор RU-V10
```

### Итог Wave RU-1 (9 агентов)

| Агент | Пакеты | Масштаб |
|-------|--------|---------|
| RU-D01–D03 | docs core | 5 md, gap-анализ doc↔code |
| RU-D04–D08 | test/stack/wiki | 14 md, Magic Link fix |
| RU-D09–D10 | planning | 9 md |
| RU-B01–B03 | auth/ws/rbac/audit/notification | 45 prod .go |
| RU-B04–B07 | tracker/tms/project/sprint/messenger/calls | ~70 prod .go |
| RU-B08–B14 | platform/ent/cmd/остальные | ~140 prod .go |
| RU-F01–F03 | frontend src | ~113 файлов |
| RU-B15 | testframework | 289 файлов (255 Go + 34 e2e) |
| RU-F05 | deployments/cli | 23 файла |

**Известные gaps для RU-V:** битая кодировка в assert-литералах старых messenger-тестов; `frontend/README.md` отсутствует; Swagger paths under-documented.

*Opus 4.8 недоступен — исполнители на доступных моделях.*
