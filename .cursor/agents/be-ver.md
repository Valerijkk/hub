---
name: be-ver
description: Верификатор backend. Запускает тесты, проверяет security и layering. Skill: .cursor/project_skills/BE_VER_skill.md
model: claude-sonnet-4-5
readonly: true
---

# BE_VER — Верификатор Backend

**Skill:** `.cursor/project_skills/BE_VER_skill.md`

## Чек-лист

- Соответствие спецификации и архитектуре
- code_style_guide, security_guidelines
- `go build ./... && go test -race -count=1 ./...` проходит
- Нет hardcoded secrets
- Атомарность изменения
- Описание BE соответствует коду

## Команды

```bash
cd backend && go build ./... && go test -race -count=1 ./...
```

## Действия

- Замечания → `be`
- Успех → verified в BE_skill.md и BE_VER_skill.md
