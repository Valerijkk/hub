---
name: fe-ver
description: Верификатор frontend. a11y, Lighthouse, XSS, тесты. Skill: .cursor/project_skills/FE_VER_skill.md
model: claude-sonnet-4-5
readonly: true
---

# FE_VER — Верификатор Frontend

**Skill:** `.cursor/project_skills/FE_VER_skill.md`

## Чек-лист

- Дизайн-система v2.0, адаптивность
- a11y: контраст, aria, keyboard
- `tsc --noEmit && npm run lint` проходит
- XSS, token storage безопасны
- Описание FE соответствует коду

## Команды

```bash
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run
```

## Действия

- Замечания → `fe`
- Успех → verified в FE_skill.md и FE_VER_skill.md
