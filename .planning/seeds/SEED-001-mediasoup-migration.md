# SEED-001 — Миграция LiveKit → mediasoup (SFU)

> Planted: 2026-07-10 · **Status: REJECTED 2026-07-10** (решение пользователя) · Brief: `.planning/JULY-2026-REBUILD-BRIEF.md` §2

## Decision

**Не мигрируем. Остаёмся на LiveKit.**

Нет продуктового/ops драйвера; стоимость переписи calls (BE/FE/infra/recording/tests) не окупается. M4 `hub-calls` продолжает LiveKit-домен.

## Idea (archived)

Заменить LiveKit на mediasoup через adapter — отклонено.

## Trigger (if ever revisited)

Только при явном драйвере: лицензия/стоимость LiveKit, SIP/telephony, кастомный RTP pipeline, доказанный perf ceiling. Иначе не поднимать.

## Breadcrumbs

- BE/FE/infra LiveKit as-is — канон
- Roadmap 11.a: hub-calls = LiveKit-домен (без смены SFU)
