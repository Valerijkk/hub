# project_skills — перенесено

> **2026-07-02.** Команда реорганизована под параллельную разработку (32 агента). Прежние `*_skill.md` (BA/SA/ARCH/BE/FE/QA + VER) удалены: их содержимое разделено на два слоя без дублирования.

## Где теперь что

| Было (`project_skills/*_skill.md`) | Стало |
|-----------------------------------|-------|
| Кто агент, что делает, workflow, DoD | **`.cursor/agents/<agent>.md`** (ba1/ba2/sa1/sa2/arch/be1/be2/fe1/fe2/qa-manual/qa-auto1/qa-auto2 + все `*-ver`) |
| Практики ремесла (код/тесты/аналитика/дизайн) | **`.cursor/playbooks/*.md`** |

## Карта

- Ростер и индекс: `.cursor/playbooks/README.md`
- Backend: `.cursor/playbooks/backend.md`
- Frontend: `.cursor/playbooks/frontend.md`
- Design: `.cursor/playbooks/design.md`
- Testing + Allure: `.cursor/playbooks/testing.md`
- Observability/logging: `.cursor/playbooks/observability.md`
- Комментарии: `.cursor/playbooks/code-comments.md`
- Git: `.cursor/playbooks/git-workflow.md`
- Аналитика (BA/SA/ARCH): `.cursor/playbooks/analysis.md`
- Верификация (все `*-ver`): `.cursor/playbooks/verification.md`
- Командный workflow/RACI: `.cursor/playbooks/team-workflow.md`

Регламент: `.cursor/REGULATION.md` · GSD: `.cursor/GSD-INTEGRATION.md`
