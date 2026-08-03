---
agent: test-code-reviewer (или скил naming-check напрямую)
scenario: аудит нейминга тестов директории/домена без полного ревью
---

# Шаблон: аудит нейминга

```
Прогони аудит нейминга по skills/naming-check.md для указанного scope.
Полное ревью содержимого НЕ нужно — только Step 1–4 скила (файлы, функции,
«имя не врёт» выборочно).

## Scope
- Директория/домен: {{путь}}
- Глубина: {{все файлы | только изменённые с {{ref}} | выборка N}}
- Режим «имя ↔ содержимое»: {{все | 10 случайных | пропустить}}

## Канон
`context/projects/<profile>/TEST-ARCHITECTURE.md` § Organization,
`TRACEABILITY.md` и approved naming cards из `PATTERN-INDEX.md`.
Если profile отсутствует — bootstrap; если stale — refresh. Project-specific
exceptions из `PATTERN-INDEX.md` репорти как WARN, не ERROR.

## Ожидаемый выход
Отчёт по формату naming-check: ERROR/WARN со счётчиками, каждому ERROR —
готовое правильное имя. Вердикт PASS/FAIL.
```
