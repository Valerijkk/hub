# TASK-120 — 🔌 Git*/1C/imex: secrets RAM + trash filter

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: На проверке
> Источник: W7 integrations · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008
> Закрыто: 2026-07-29 · be2 · ветка `fix/integrations/task-120-secrets-imex`

## [OUTCOME]

Нет ListAllWebhookSecrets RAM blast; GitLab body redact; 1C token encrypt; imex trash filter.

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/{github,gitlab,imex,onec}/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - plaintext tokens в settings
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: W7-A3-004/006/007/008 (4)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 4 ID + tests
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
