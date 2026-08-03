# skills/ — процедурные скилы

> Скил — пошаговая процедура с проверяемым результатом. Не роль (это агенты), а чек-процесс: прогнал шаги → получил отчёт. Скилы запускает человек, агент или CI.

| Скил | Назначение | Когда запускать |
|------|-----------|-----------------|
| [bootstrap-test-context](bootstrap-test-context/SKILL.md) | Read-only скан заданных project/test roots → generated profile | Первый запуск слоя в проекте |
| [refresh-test-context](refresh-test-context/SKILL.md) | Инкрементальный refresh profile с provenance/freshness | После изменения framework/CI/config |
| [curate-review-feedback](curate-review-feedback/SKILL.md) | Проверка вручную добавленных before/comment/after examples | После ручного заполнения `review-feedback/inbox/` |
| [tf-init](tf-init/SKILL.md) | Первый запуск: изучить репозиторий, наполнить контекст | Контура в репозитории ещё нет |
| [tf-status](tf-status/SKILL.md) | Где я: свежесть контекста, покрытие, что дальше | Начало сессии, в любой непонятной ситуации |
| [tf-cases](tf-cases/SKILL.md) | Спроектировать тест-кейсы из требований техниками тест-дизайна | Есть требование, нужен набор кейсов |
| [tf-cover](tf-cover/SKILL.md) | Покрыть тестами ручки, модуль, требования | Пришло новое либо есть непокрытое |
| [tf-triage](tf-triage/SKILL.md) | Упал тест: code, test, infra, flaky, stop | Красный тест или сообщение о баге |
| [tf-review](tf-review/SKILL.md) | Ревью тестового кода до мержа | Автор считает работу законченной |
| [curate-pattern-cards](curate-pattern-cards/SKILL.md) | Candidate → approved/deprecated с human gate | После context refresh или feedback curation |
| [anti-ai-check](anti-ai-check/SKILL.md) | Скан качества generated-кода; heuristics → semantic confirmation | После AI-изменений; перед ревью |
| [naming-check](naming-check/SKILL.md) | Аудит нейминга тестовых функций и файлов против активного profile | Перед мержем тестов; периодический аудит |
| [test-verification](test-verification/SKILL.md) | Верификация сгенерированных тест-кейсов против документации | После `test-case-generator`; перед автоматизацией TC |

Всего: **13 skills**, каждый в собственной папке. Нормализация feedback и
approval паттернов — разные trust gates: первый skill не создаёт cards, второй
читает только human-approved feedback.

## Формат скила

```
skills/<name>/SKILL.md
---
name: <slug>
description: <одна строка — что делает>
---
# /<name>
Шаги с командами → формат отчёта → критерии PASS/FAIL
```

Каждый skill обязан: (1) читать `context/projects/<profile>/` вместо угадывания
путей; (2) завершаться evidence block и честным PASS/FAIL/PARTIAL/NOT_RUN;
(3) не менять source code без явного шага/разрешения; (4) применять
[rules/40](../rules/40-verification-protocol.md) по риску и
[rules/50](../rules/50-data-handling.md) ко всему scanned/manual content.
