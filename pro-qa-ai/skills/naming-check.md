---
name: naming-check
description: Аудит имён тестовых файлов и кейсов против конвенций активного project profile и pattern cards. Использовать на diff перед мержем или для scoped naming-аудита — никогда не навязывать общий вкус поверх канона проекта.
disable-model-invocation: true
---

# Аудит нейминга (`naming-check`)

Сравнивает имена в `{{scope}}` с **конвенциями проекта**, а не с личным стилем. Вердикты: **ERROR** (блокирует рекомендацию мержа diff) | **WARN** (только в отчёте) | **OK**.

## Когда запускать
- Перед мержем, когда изменились тестовые файлы или имена кейсов
- Scoped-аудит: один домен, каталог или полное дерево тестов
- После заполнения `bootstrap-test-context` — проверить, что новые тесты совпадают с наблюдаемым каноном

## Стоп-условия

- Правила нейминга в profile недоступны (`TEST-ARCHITECTURE.md` всё ещё `template` / всё `UNKNOWN`) → остановиться; сначала `bootstrap-test-context` и human review
- Нет scope → запросить пути diff или список каталогов
- Запрошен массовый rename legacy → отдельная chore-задача; этот skill только аудитит

## Вход

Область: дифф, каталог или срез дерева тестов. Плюс профиль — канон имён
берётся из него, а не из общих правил нейминга.

## Контекст
| Нужно | Primary |
|------|---------|
| Канон имён файлов/кейсов | `context/projects/<profile>/TEST-ARCHITECTURE.md` § Organization |
| Layout / levels | То же + `TEST-ARCHITECTURE.md` § Test topology |
| Approved examples | `context/projects/<profile>/PATTERN-INDEX.md` → связанные cards в `patterns/approved/` |
| Legacy exceptions | `PATTERN-INDEX.md` § Project-specific exceptions |
| Формат ID/traceability | `context/projects/<profile>/TRACEABILITY.md` § Identifier conventions |

Если profile и legacy расходятся, **побеждает profile** при `status: filled`; legacy — только fallback.

## Шаги
### 1. Собрать объекты

Вход: `{{scope}}` = diff | directory | срез дерева тестов.

1. Перечислить тестовые файлы в scope (паттерны из `TEST-ARCHITECTURE.md`, не угаданные).
2. Извлечь имена кейсов: `func Test…`, `it(…)`, `test(…)`, `def test_…` через поиск или чтение.

### 2. Имена файлов

Для каждого файла сверить с таблицей profile (не с общими правилами):

| Проверка | Вердикт при нарушении |
|-------|---------------------|
| Suffix/prefix соответствует level (unit/integration/e2e) по проекту | ERROR |
| Разделители и регистр совпадают с проектом | ERROR |
| Правило пары test↔code (если проект задаёт строгий mapping) | ERROR или WARN по profile |
| Каталог соответствует layout domain/level | ERROR |
| Путь в задокументированной зоне legacy exception | только WARN |

### 3. Имена кейсов / функций

По `TEST-ARCHITECTURE.md` и naming pattern cards:

1. Присутствует требуемая структура (напр. unit_scenario_expected — **как в проекте**, не изобретённая).
2. Язык совпадает с проектом (EN/RU/mixed, если задокументировано).
3. Нет в списке запрещённых проекта (`Test1`, `it('works')` и т.п. — **из profile**, не generic lint).
4. Имя читаемо в CI без открытия файла.
5. TC/REQ ID в имени, если этого требует `TRACEABILITY.md`.

Нарушение структуры или запрещённое имя → **ERROR**. Субъективная ясность → **WARN** только если profile допускает style notes.

### 4. Согласованность имя ↔ тело (выборка)

Режим diff: все изменённые кейсы. Полный аудит: выборка по доменам (размер из profile, напр. 10).

1. Тело реализует сценарий и ожидание, подразумеваемые именем.
2. Имя обещает error path, а тест только проверяет отсутствие panic → **ERROR** (lying name по [rules/30](../standards/rules/30-severity-policy.md)).
3. Лишние assertions сверх имени → **WARN**.

### 5. Helpers и fixtures (режим diff)

Если profile или pattern cards задают имена полей fixture/sut/want/got/table → **WARN** при нарушении (не блокирует, пока profile не говорит ERROR).

## Формат отчёта

```markdown
## Naming audit — {{scope}} — {{date}}

Profile: {{profile}} | Canon: TEST-ARCHITECTURE § Organization + pattern cards

Files: N checked
  ERROR: `path` — <rule ref> → <suggested name per project pattern>
  WARN:  `path` — <reason>

Cases: M checked
  ERROR: `path:line` `TestName` — <violation> → `<project-shaped name>`
  WARN:  …

Name ↔ body: K checked, mismatches: X

Verdict: PASS (0 ERROR) | FAIL (N ERROR)

Verification:
- Claims checked: file/case names vs profile canon and cards
- Evidence: paths read; canon sections cited
- Not run / N/A: …
- Risks / assumptions: legacy zones, incomplete profile
- Mode: self-check
- Verdict: PASS | FAIL | PARTIAL
```

## Критерий готовности

- [ ] Канон взят из профиля, а не из общих представлений
- [ ] Имена файлов и кейсов проверены отдельно
- [ ] Согласованность имени с телом проверена на выборке
- [ ] У каждого ERROR есть предложенное имя в форме проекта
- [ ] Задокументированные исключения профиля дали WARN, а не ERROR

## Вердикт

| Вердикт | Условие |
|---------|---------|
| PASS | Ноль ERROR |
| PARTIAL | ERROR нет, но есть WARN, требующие решения человека |
| FAIL | Есть ERROR, включая имя-врун |
| NOT_RUN | Канон нейминга в профиле недоступен — сверять не с чем |

## Границы
- **ERROR** блокирует рекомендацию мержа diff; **WARN** — нет.
- Каждый **ERROR** включает **project-shaped** имя-замену (из pattern card или таблицы нейминга), а не общий совет «переименовать».
- Массовые legacy-rename → отдельная задача; не расширять scope здесь.
- Не повышать WARN до ERROR ради личного предпочтения.

## Что дальше

- Naming baseline: [rules/10](../standards/rules/10-naming-conventions.md) (defaults тулкита; profile перекрывает)
- Severity (lying names): [rules/30](../standards/rules/30-severity-policy.md)
- Verification: [rules/40](../standards/rules/40-verification-protocol.md)
- Profile bootstrap: `bootstrap-test-context.md`
