# TASK-101 — 📐 AI Ollama localhost consent vs SafeDial

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта · MUTATION-2026-07-29-W5
> Источник: A2W5-A2-005 · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Approval record: owner 2026-07-28 — отдельный ADR approve не требуется; код + ADR-AI-OLLAMA-LOOPBACK-2026

## [OUTCOME]

Ollama loopback разрешён только по явному consent/ADR без ослабления SSRF SafeDial для остальных URL.

## [ROLE]

- Role: product→be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/ai/**
  - platform/http safedial
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - глобальный bypass ValidateOutboundURL
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: owner 2026-07-28 — approve не требуется (trusted offline dialer)

## [REFERENCES]

- Finding IDs: A2W5-A2-005 P1-BE-U01 (1)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR P1-BE-U01
  - allowlist loopback only when enabled
  - SSRF tests intact
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
