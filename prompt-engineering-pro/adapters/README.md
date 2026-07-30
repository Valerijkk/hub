# Adapters — Touchstone Pro

| Адаптер | Что мапит |
|---------|-----------|
| [cursor/](cursor/) | 12 ролей разработки в `.cursor/agents/<name>.md` |

Standard держит свой адаптер с четырьмя `test-*` ролями. Два маппинга
не конфликтуют: имена агентов не пересекаются, поэтому оба издания
устанавливаются в один `.cursor/agents/` рядом.

`readonly: true` выставлен для `architect`, `code-reviewer` и
`security-reviewer` — роли ревью не должны писать в репозиторий, иначе
независимость проверки теряется.

Правила выбора модели и wrapper-шаблон — общие со Standard:
[`playbooks/model-routing.md`](../../prompt-engineering/playbooks/model-routing.md).
