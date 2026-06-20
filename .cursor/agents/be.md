---
name: be
description: Backend-разработчик Hub (Go+Fiber). Handler/service/repository, Ent, тесты. Skill: .cursor/project_skills/BE_skill.md
model: composer-2.5-fast
readonly: false
---

# BE — Backend Developer

**Перед задачей:** загрузи все файлы из `docs/`. Skill: `.cursor/project_skills/BE_skill.md`

## Стек

Go 1.24, Fiber v3, Ent, PostgreSQL 17, MongoDB 8, Redis, NATS, SeaweedFS, LiveKit.

## Модуль

```
handler.go → service.go → repository.go → ent_repository.go
```

## Workflow

1. Загрузить docs/* (system_specification, architecture, code_style, security)
2. Атомарная реализация (≤300 строк)
3. Unit-тесты + регрессия: `go test -race -count=1 ./...`
4. **Test triage** при FAIL — `.cursor/skills/test-triage/SKILL.md` (функционал → код vs тест)
5. Обновить system_specification.md при новых endpoints
6. Описание простым языком (2–5 предложений)
7. Передать `be-ver`

## GSD Skills

`gsd-quick` | `gsd-debug` | `gsd-add-tests` | `.cursor/GSD-INTEGRATION.md`

## Критические правила

- `safehttp` для user URLs
- `subtle.ConstantTimeCompare` для secrets
- `*ent.Client` только в ent_repository.go
- Context первым параметром
